package audio

import (
	"fmt"
	"log"
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
	queue           []models.Track
	currentIndex    int
	isPlaying       bool
	isShuffled      bool
	repeatMode      RepeatMode
	
	// Callbacks
	stateCallback   func(*models.AppState)
	
	// Synchronization
	mu              sync.RWMutex
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

// SetStateCallback sets the callback function for state updates
func (m *Manager) SetStateCallback(callback func(*models.AppState)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stateCallback = callback
}

// AddToQueue adds a track to the playback queue
func (m *Manager) AddToQueue(track models.Track) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.queue = append(m.queue, track)
	log.Printf("Added track to queue: %s - %s", track.Artist, track.Title)
	m.notifyStateChange()
}

// AddTracksToQueue adds multiple tracks to the playback queue
func (m *Manager) AddTracksToQueue(tracks []models.Track) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.queue = append(m.queue, tracks...)
	log.Printf("Added %d tracks to queue", len(tracks))
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
	log.Printf("Removed track from queue at index %d", index)
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
	log.Println("Cleared playback queue")
	m.notifyStateChange()
}

// PlayTrackAtIndex starts playing the track at the specified queue index
func (m *Manager) PlayTrackAtIndex(index int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	fmt.Printf("[AUDIO DEBUG] PlayTrackAtIndex called with index: %d, queue length: %d\n", index, len(m.queue))
	
	if index < 0 || index >= len(m.queue) {
		fmt.Printf("[AUDIO DEBUG] Invalid queue index: %d\n", index)
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
		log.Println("Paused playback")
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
		log.Println("Resumed playback")
		m.notifyStateChange()
	}
}

// Stop stops the current playback
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.player.Stop()
	m.isPlaying = false
	log.Println("Stopped playback")
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
	log.Println("Reached end of queue")
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
	log.Printf("Set volume to %.0f%%", volume*100)
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
	streamURL := m.navidromeClient.GetStreamURL(track.ID)
	
	fmt.Printf("[AUDIO DEBUG] Playing track: %s - %s (ID: %s)\n", track.Artist, track.Title, track.ID)
	fmt.Printf("[AUDIO DEBUG] Track suffix/format: %s\n", track.Suffix)
	fmt.Printf("[AUDIO DEBUG] Stream URL: %s\n", streamURL)
	
	// Pass the track format hint to the player
	err := m.player.PlayWithFormat(streamURL, track.ID, track.Suffix)
	if err != nil {
		fmt.Printf("[AUDIO DEBUG] Failed to play track: %v\n", err)
		return fmt.Errorf("failed to play track: %w", err)
	}
	
	m.currentIndex = index
	m.isPlaying = true
	
	fmt.Printf("[AUDIO DEBUG] Successfully started playback, currentIndex: %d, isPlaying: %v\n", m.currentIndex, m.isPlaying)
	log.Printf("Playing track: %s - %s", track.Artist, track.Title)
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

// handlePlayerEvent handles events from the audio player
func (m *Manager) handlePlayerEvent(event PlaybackEvent) {
	switch event.Type {
	case "finished":
		// Track finished, play next track
		go func() {
			time.Sleep(100 * time.Millisecond) // Small delay to avoid race conditions
			
			// Scrobble the finished track
			if m.scrobbler != nil {
				track := m.GetCurrentTrack()
				if track != nil {
					scrobbleTrack := scrobbling.ScrobbleTrack{
						Title:    track.Title,
						Artist:   track.Artist,
						Album:    track.Album,
						Duration: track.Duration,
						Timestamp: time.Now().Unix(),
					}
					m.scrobbler.Scrobble(scrobbleTrack)
				}
			}
			
			// Play next track
			m.NextTrack()
		}()
		
	case "error":
		log.Printf("Playback error for track: %s", event.TrackID)
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