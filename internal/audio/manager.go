package audio

import (
	"context"
	"fmt"
	"navitone-cli/internal/models"
	"navitone-cli/pkg/navidrome"
	"navitone-cli/pkg/scrobbling"
	"sync"
	"time"
)

// Manager handles audio playback and queue management
type Manager struct {
	player          *Player
	navidromeClient *navidrome.Client
	scrobbler       *scrobbling.Manager

	// State
	queue        []models.Track
	currentIndex int
	isPlaying    bool
	repeatMode   RepeatMode

	// Callbacks
	stateCallback func(*models.AppState)
	logCallback   func(string)

	// Synchronization
	mu sync.RWMutex
}

// RepeatMode represents different repeat modes
type RepeatMode int

const (
	RepeatNone RepeatMode = iota
	RepeatOne
	RepeatAll
)

// NewManager creates a new audio manager
func NewManager(navidromeClient *navidrome.Client, scrobbler *scrobbling.Manager) (*Manager, error) {
	player, err := NewPlayer()
	if err != nil {
		return nil, fmt.Errorf("failed to create audio player: %w", err)
	}

	manager := &Manager{
		player:          player,
		navidromeClient: navidromeClient,
		scrobbler:       scrobbler,
		queue:           make([]models.Track, 0),
		currentIndex:    -1,
		repeatMode:      RepeatNone,
	}

	// Set up player event callback
	player.SetEventCallback(manager.handlePlayerEvent)

	return manager, nil
}

// CheckStreamingPermissions verifies that the user has proper streaming access
func (m *Manager) CheckStreamingPermissions() error {
	err := m.navidromeClient.CheckUserPermissions(context.TODO())
	if err != nil {
		return err
	}
	return nil
}

// SetStateCallback sets the callback function for state updates
func (m *Manager) SetStateCallback(callback func(*models.AppState)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stateCallback = callback
}

// SetLogCallback sets the callback function for log messages
func (m *Manager) SetLogCallback(callback func(string)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logCallback = callback
}

// logMessage sends a message to the log callback if available
func (m *Manager) logMessage(message string) {
	if m.logCallback != nil {
		go m.logCallback(message) // Call in goroutine to avoid blocking
	}
}

// AddToQueue adds a track to the playback queue
func (m *Manager) AddToQueue(track models.Track) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.queue = append(m.queue, track)
	m.logMessage(fmt.Sprintf("Added track to queue: %s - %s", track.Artist, track.Title))
	m.notifyStateChange()
}

// AddTracksToQueue adds multiple tracks to the playback queue
func (m *Manager) AddTracksToQueue(tracks []models.Track) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.queue = append(m.queue, tracks...)
	m.logMessage(fmt.Sprintf("Added %d tracks to queue", len(tracks)))
	m.notifyStateChange()
}

// RemoveFromQueue removes a track from the queue at the specified index
func (m *Manager) RemoveFromQueue(index int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if index < 0 || index >= len(m.queue) {
		return
	}

	// Adjust current index if necessary
	if index < m.currentIndex {
		m.currentIndex--
	} else if index == m.currentIndex && m.isPlaying {
		// If removing currently playing track, stop playback
		m.player.Stop()
		m.isPlaying = false
	}

	m.queue = append(m.queue[:index], m.queue[index+1:]...)
	m.logMessage(fmt.Sprintf("Removed track from queue at index %d", index))
	m.notifyStateChange()
}

// ClearQueue removes all tracks from the queue
func (m *Manager) ClearQueue() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.player.Stop()
	m.queue = make([]models.Track, 0)
	m.currentIndex = -1
	m.isPlaying = false
	m.logMessage("Cleared playback queue")
	m.notifyStateChange()
}

// PlayTrackAtIndex starts playing the track at the specified queue index
func (m *Manager) PlayTrackAtIndex(index int) error {
	m.mu.Lock()
	defer m.mu.Unlock()


	if index < 0 || index >= len(m.queue) {
		return fmt.Errorf("invalid queue index: %d", index)
	}

	return m.playTrackAtIndexLocked(index)
}

// PlayCurrent plays the current track (or first track if none selected)
func (m *Manager) PlayCurrent() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.queue) == 0 {
		return fmt.Errorf("queue is empty")
	}

	if m.currentIndex < 0 {
		m.currentIndex = 0
	}

	return m.playTrackAtIndexLocked(m.currentIndex)
}

// Pause pauses the current playback
func (m *Manager) Pause() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isPlaying {
		m.player.Pause()
		m.isPlaying = false
		m.logMessage("Paused playback")
		m.notifyStateChange()
	}
}

// Resume resumes the paused playback
func (m *Manager) Resume() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isPlaying && m.player.GetState() == StatePaused {
		m.player.Resume()
		m.isPlaying = true
		m.logMessage("Resumed playback")
		m.notifyStateChange()
	}
}

// Stop stops the current playback
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.player.Stop()
	m.isPlaying = false
	m.logMessage("Stopped playback")
	m.notifyStateChange()
}

// NextTrack plays the next track in the queue
func (m *Manager) NextTrack() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.queue) == 0 {
		return fmt.Errorf("queue is empty")
	}

	nextIndex := m.getNextTrackIndex()
	if nextIndex >= 0 {
		return m.playTrackAtIndexLocked(nextIndex)
	}

	// End of queue
	m.player.Stop()
	m.isPlaying = false
	m.logMessage("Reached end of queue")
	m.notifyStateChange()
	return nil
}

// PreviousTrack plays the previous track in the queue
func (m *Manager) PreviousTrack() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.queue) == 0 {
		return fmt.Errorf("queue is empty")
	}

	prevIndex := m.getPreviousTrackIndex()
	if prevIndex >= 0 {
		return m.playTrackAtIndexLocked(prevIndex)
	}

	return nil
}

// TogglePlayPause toggles between play and pause
func (m *Manager) TogglePlayPause() error {
	m.mu.RLock()
	playing := m.isPlaying
	m.mu.RUnlock()

	if playing {
		m.Pause()
	} else {
		if m.player.GetState() == StatePaused {
			m.Resume()
		} else {
			return m.PlayCurrent()
		}
	}

	return nil
}

// GetQueue returns a copy of the current queue
func (m *Manager) GetQueue() []models.Track {
	m.mu.RLock()
	defer m.mu.RUnlock()

	queue := make([]models.Track, len(m.queue))
	copy(queue, m.queue)
	return queue
}

// GetCurrentTrack returns the currently playing track
func (m *Manager) GetCurrentTrack() *models.Track {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.currentIndex >= 0 && m.currentIndex < len(m.queue) {
		track := m.queue[m.currentIndex]
		return &track
	}

	return nil
}

// GetCurrentIndex returns the current track index
func (m *Manager) GetCurrentIndex() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentIndex
}

// IsPlaying returns whether audio is currently playing
func (m *Manager) IsPlaying() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isPlaying
}

// SetVolume sets the playback volume
func (m *Manager) SetVolume(volume float64) {
	m.player.SetVolume(volume)
	m.logMessage(fmt.Sprintf("Set volume to %.0f%%", volume*100))
}

// Close closes the audio manager and releases resources
func (m *Manager) Close() error {
	m.Stop()
	return m.player.Close()
}

// playTrackAtIndexLocked plays the track at the specified index (must be called with lock held)
func (m *Manager) playTrackAtIndexLocked(index int) error {
	if index < 0 || index >= len(m.queue) {
		return fmt.Errorf("invalid queue index: %d", index)
	}

	track := m.queue[index]
	
	// Check streaming permissions for the first track only to avoid spam
	if index == 0 || m.currentIndex == -1 {
	}
	
	// Use stream URL with proper parameters for full track access
	streamURL := m.navidromeClient.GetStreamURL(track.ID)


	// Convert duration from seconds to time.Duration
	trackDuration := time.Duration(track.Duration) * time.Second

	// Pass the track format hint and duration to the player
	err := m.player.PlayWithFormatAndDuration(streamURL, track.ID, track.Suffix, trackDuration)
	if err != nil {
		// Fallback to download URL
		downloadURL := m.navidromeClient.GetDownloadURL(track.ID)
		err = m.player.PlayWithFormatAndDuration(downloadURL, track.ID, track.Suffix, trackDuration)
		if err != nil {
			return fmt.Errorf("failed to play track: %w", err)
		}
	}

	m.currentIndex = index
	m.isPlaying = true

	m.logMessage(fmt.Sprintf("Playing track: %s - %s", track.Artist, track.Title))
	m.notifyStateChange()

	// Submit "Now Playing" to scrobbling services
	if m.scrobbler != nil {
		scrobbleTrack := scrobbling.ScrobbleTrack{
			Title:    track.Title,
			Artist:   track.Artist,
			Album:    track.Album,
			Duration: track.Duration,
		}
		go m.scrobbler.UpdateNowPlaying(scrobbleTrack)
	}

	return nil
}

// getNextTrackIndex returns the index of the next track to play
func (m *Manager) getNextTrackIndex() int {
	switch m.repeatMode {
	case RepeatOne:
		return m.currentIndex
	case RepeatAll:
		if m.currentIndex+1 >= len(m.queue) {
			return 0 // Loop back to beginning
		}
		return m.currentIndex + 1
	default: // RepeatNone
		if m.currentIndex+1 < len(m.queue) {
			return m.currentIndex + 1
		}
		return -1 // End of queue
	}
}

// getPreviousTrackIndex returns the index of the previous track to play
func (m *Manager) getPreviousTrackIndex() int {
	switch m.repeatMode {
	case RepeatOne:
		return m.currentIndex
	case RepeatAll:
		if m.currentIndex-1 < 0 {
			return len(m.queue) - 1 // Loop to end
		}
		return m.currentIndex - 1
	default: // RepeatNone
		if m.currentIndex-1 >= 0 {
			return m.currentIndex - 1
		}
		return -1 // Beginning of queue
	}
}

// handlePlayerEvent processes audio player events
func (m *Manager) handlePlayerEvent(event PlaybackEvent) {
	
	switch event.Type {
	case "finished":
		// Start next track in background
		go func() {

			// Play next track
			m.NextTrack()
		}()

	case "error":
		m.logMessage(fmt.Sprintf("Playback error for track: %s", event.TrackID))
		// Try next track on error
		go m.NextTrack()
	}
}

// notifyStateChange notifies the UI about state changes (must be called with lock held)
func (m *Manager) notifyStateChange() {
	if m.stateCallback != nil {
		// Create updated state - this is a simplified version
		// The actual implementation will depend on the full AppState structure
		go func() {
			// Call the callback in a goroutine to avoid blocking
			// The callback will need to be implemented to update the AppState
			// m.stateCallback(updatedState)
		}()
	}
}
