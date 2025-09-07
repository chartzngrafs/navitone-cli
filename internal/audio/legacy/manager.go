package audio

import (
	"context"
	"fmt"
	"math/rand"
	"navitone-cli/internal/models"
	"navitone-cli/pkg/navidrome"
	"navitone-cli/pkg/scrobbling"
	"net/http"
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
	originalQueue []models.Track // Store original order for when shuffle is disabled
	currentIndex int
	isPlaying    bool
	repeatMode   RepeatMode
	shuffleMode  bool
	isSeeking    bool  // Flag to prevent auto-advance during seeking

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

	// Random number generator is automatically seeded in Go 1.20+

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

	if m.shuffleMode && len(m.originalQueue) > 0 {
		// If shuffle is on, add to both queues
		m.originalQueue = append(m.originalQueue, tracks...)
		
		// Add to current queue and re-shuffle the new tracks only
		newTracksStart := len(m.queue)
		m.queue = append(m.queue, tracks...)
		
		// Shuffle only the newly added tracks to maintain current position
		newTracks := m.queue[newTracksStart:]
		for i := len(newTracks) - 1; i > 0; i-- {
			j := rand.Intn(i + 1)
			newTracks[i], newTracks[j] = newTracks[j], newTracks[i]
		}
	} else {
		// Normal mode - just add tracks
		m.queue = append(m.queue, tracks...)
	}
	
	m.logMessage(fmt.Sprintf("Added %d tracks to queue (shuffle: %v)", len(tracks), m.shuffleMode))
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

	m.logMessage(fmt.Sprintf("Pause() called - isPlaying: %v, playerState: %v", m.isPlaying, m.player.GetState()))
	
	if m.isPlaying {
		m.logMessage("Calling m.player.Pause()")
		m.player.Pause()
		m.isPlaying = false
		m.logMessage("Paused playback - set isPlaying to false")
		m.notifyStateChange()
		
		// Verify player state after pause
		newState := m.player.GetState()
		m.logMessage(fmt.Sprintf("Player state after pause: %v", newState))
	} else {
		m.logMessage("Already paused - no action taken")
	}
}

// Resume resumes the paused playback
func (m *Manager) Resume() {
	m.mu.Lock()
	defer m.mu.Unlock()

	playerState := m.player.GetState()
	m.logMessage(fmt.Sprintf("Resume() called - isPlaying: %v, playerState: %v", m.isPlaying, playerState))

	if !m.isPlaying && playerState == StatePaused {
		m.logMessage("Calling m.player.Resume()")
		m.player.Resume()
		m.isPlaying = true
		m.logMessage("Resumed playback - set isPlaying to true")
		m.notifyStateChange()
		
		// Verify player state after resume
		newState := m.player.GetState()
		m.logMessage(fmt.Sprintf("Player state after resume: %v", newState))
	} else {
		m.logMessage(fmt.Sprintf("Cannot resume - isPlaying: %v, playerState: %v", m.isPlaying, playerState))
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
	playerState := m.player.GetState()
	currentIndex := m.currentIndex
	queueLen := len(m.queue)
	m.mu.RUnlock()

	m.logMessage(fmt.Sprintf("TogglePlayPause - playing: %v, playerState: %v, index: %d, queue: %d", 
		playing, playerState, currentIndex, queueLen))

	if playing {
		m.logMessage("Calling Pause()")
		m.Pause()
	} else {
		if playerState == StatePaused {
			m.logMessage("Player paused - calling Resume()")
			m.Resume()
		} else {
			m.logMessage("Player not paused - calling PlayCurrent()")
			return m.PlayCurrent()
		}
	}

	// Check state after toggle
	m.mu.RLock()
	newPlaying := m.isPlaying
	newPlayerState := m.player.GetState()
	m.mu.RUnlock()
	
	m.logMessage(fmt.Sprintf("After toggle - playing: %v, playerState: %v", newPlaying, newPlayerState))
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

// ToggleShuffle toggles shuffle mode on/off
func (m *Manager) ToggleShuffle() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.shuffleMode = !m.shuffleMode
	
	if m.shuffleMode {
		// Enabling shuffle: save original order and shuffle the queue
		m.originalQueue = make([]models.Track, len(m.queue))
		copy(m.originalQueue, m.queue)
		
		// Find currently playing track before shuffle
		var currentTrack *models.Track
		if m.currentIndex >= 0 && m.currentIndex < len(m.queue) {
			currentTrack = &m.queue[m.currentIndex]
		}
		
		// Shuffle the queue using Fisher-Yates algorithm
		for i := len(m.queue) - 1; i > 0; i-- {
			j := rand.Intn(i + 1)
			m.queue[i], m.queue[j] = m.queue[j], m.queue[i]
		}
		
		// Find the new index of the currently playing track
		if currentTrack != nil {
			for i, track := range m.queue {
				if track.ID == currentTrack.ID {
					m.currentIndex = i
					break
				}
			}
		}
		
		m.logMessage(fmt.Sprintf("Shuffle enabled - queue randomized (%d tracks)", len(m.queue)))
	} else {
		// Disabling shuffle: restore original order
		if len(m.originalQueue) > 0 {
			// Find currently playing track before restoring order
			var currentTrack *models.Track
			if m.currentIndex >= 0 && m.currentIndex < len(m.queue) {
				currentTrack = &m.queue[m.currentIndex]
			}
			
			// Restore original order
			copy(m.queue, m.originalQueue)
			
			// Find the new index of the currently playing track in original order
			if currentTrack != nil {
				for i, track := range m.queue {
					if track.ID == currentTrack.ID {
						m.currentIndex = i
						break
					}
				}
			}
			
			m.originalQueue = nil // Clear the backup
		}
		m.logMessage("Shuffle disabled - original order restored")
	}
	
	m.notifyStateChange()
}

// IsShuffleEnabled returns whether shuffle mode is enabled
func (m *Manager) IsShuffleEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.shuffleMode
}

// SeekForward seeks forward in the current track
func (m *Manager) SeekForward(seconds int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.currentIndex >= 0 && m.currentIndex < len(m.queue) && m.player != nil {
		m.logMessage(fmt.Sprintf("Seeking forward %d seconds", seconds))
		
		// Get current position and calculate new position
		currentPosition := m.player.GetPosition()
		newPosition := currentPosition + time.Duration(seconds)*time.Second
		
		return m.seekToPosition(newPosition)
	}
	
	m.logMessage("No track playing to seek in")
	return fmt.Errorf("no track currently playing")
}

// SeekBackward seeks backward in the current track  
func (m *Manager) SeekBackward(seconds int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.currentIndex >= 0 && m.currentIndex < len(m.queue) && m.player != nil {
		m.logMessage(fmt.Sprintf("Seeking backward %d seconds", seconds))
		
		// Get current position and calculate new position
		currentPosition := m.player.GetPosition()
		newPosition := currentPosition - time.Duration(seconds)*time.Second
		if newPosition < 0 {
			newPosition = 0
		}
		
		return m.seekToPosition(newPosition)
	}
	
	m.logMessage("No track playing to seek in")
	return fmt.Errorf("no track currently playing")
}

// seekToPosition seeks to a specific position using HTTP Range requests
func (m *Manager) seekToPosition(position time.Duration) error {
	if m.currentIndex < 0 || m.currentIndex >= len(m.queue) {
		return fmt.Errorf("no track currently playing")
	}
	
	track := m.queue[m.currentIndex]
	trackDuration := time.Duration(track.Duration) * time.Second
	
	// Clamp position to valid range
	if position < 0 {
		position = 0
	}
	if position > trackDuration {
		position = trackDuration
	}
	
	m.logMessage(fmt.Sprintf("Seeking to position %v (%d seconds) using HTTP Range", position, int(position.Seconds())))
	
	// Set seeking flag to prevent auto-advance on errors
	// Keep it set for a few seconds to handle async errors
	m.isSeeking = true
	go func() {
		time.Sleep(3 * time.Second) // Wait 3 seconds before clearing flag
		m.mu.Lock()
		m.isSeeking = false
		m.mu.Unlock()
		m.logMessage("Clearing seeking flag after timeout")
	}()
	
	// Get current stream URL
	streamURL := m.navidromeClient.GetStreamURL(track.ID)
	wasPlaying := m.isPlaying
	
	// For compressed formats like FLAC, MP3, OGG - seeking with HTTP Range doesn't work well
	// because decoders need to start from frame boundaries
	// Instead, we'll adjust the position offset for the UI display
	if track.Suffix == "flac" || track.Suffix == "mp3" || track.Suffix == "ogg" {
		m.logMessage(fmt.Sprintf("Compressed format (%s) - adjusting position offset for seeking", track.Suffix))
		
		// Calculate the offset we want to apply
		currentRealPosition := m.player.GetPosition()
		seekOffset := position - currentRealPosition
		
		// Tell the player to adjust its position calculations by this offset
		m.player.AdjustPositionOffset(seekOffset)
		
		m.logMessage(fmt.Sprintf("Applied position offset of %v (target: %v, current: %v)", seekOffset, position, currentRealPosition))
		m.notifyStateChange()
		return nil
	}
	
	// For uncompressed formats (WAV, AIFF), try actual HTTP Range seeking
	// Stop current playback only for formats where range seeking works
	m.player.Stop()
	m.isPlaying = false
	
	// Calculate estimated byte position
	bytePosition, err := m.estimateBytePosition(streamURL, position, trackDuration)
	if err != nil {
		m.logMessage(fmt.Sprintf("Range seeking failed, restarting from beginning: %v", err))
		// Fallback: restart from beginning but keep playing
		err = m.player.PlayWithFormatAndDuration(streamURL, track.ID, track.Suffix, trackDuration)
		if err != nil {
			return fmt.Errorf("failed to start playback: %w", err)
		}
		position = 0
	} else {
		m.logMessage(fmt.Sprintf("Estimated byte position: %d of content for %s format", bytePosition, track.Suffix))
		
		// Try range playback for uncompressed formats
		err = m.player.PlayWithRange(streamURL, track.ID, track.Suffix, trackDuration, bytePosition)
		if err != nil {
			m.logMessage(fmt.Sprintf("Range playback failed, restarting from beginning: %v", err))
			// Ultimate fallback: restart from beginning but keep playing
			err = m.player.PlayWithFormatAndDuration(streamURL, track.ID, track.Suffix, trackDuration)
			if err != nil {
				return fmt.Errorf("failed to start playback: %w", err)
			}
			position = 0
		}
	}
	
	// Update position tracking to account for the seek offset
	m.player.mu.Lock()
	m.player.position = position
	m.player.mu.Unlock()
	
	// Restore playing state - always resume playback
	if wasPlaying {
		m.isPlaying = true
		m.logMessage(fmt.Sprintf("Successfully seeked to %v and resumed playback", position))
	} else {
		// Even if paused, start playing then immediately pause to ensure audio is ready
		m.isPlaying = true
		time.Sleep(100 * time.Millisecond) // Brief delay to ensure playback starts
		m.player.Pause()
		m.isPlaying = false
		m.logMessage(fmt.Sprintf("Successfully seeked to %v while paused", position))
	}
	
	m.notifyStateChange()
	return nil
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
	// Shuffle mode doesn't change navigation logic - queue is already shuffled
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
	// Shuffle mode doesn't change navigation logic - queue is already shuffled
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
		// Only advance to next track on error if we're not seeking
		if !m.isSeeking {
			m.logMessage("Advancing to next track due to playback error")
			go m.NextTrack()
		} else {
			m.logMessage("Ignoring error during seeking operation")
		}
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

// estimateBytePosition estimates the byte position for a given time offset
func (m *Manager) estimateBytePosition(streamURL string, targetTime, totalDuration time.Duration) (int64, error) {
	// Get content length by making a HEAD request
	req, err := http.NewRequest("HEAD", streamURL, nil)
	if err != nil {
		return 0, fmt.Errorf("creating HEAD request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("HEAD request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check if server supports range requests
	if resp.Header.Get("Accept-Ranges") != "bytes" {
		return 0, fmt.Errorf("server does not support byte ranges")
	}

	contentLength := resp.ContentLength
	if contentLength <= 0 {
		return 0, fmt.Errorf("unknown content length")
	}

	// Simple linear estimation: byte_pos = (targetTime / totalDuration) * contentLength
	if totalDuration <= 0 {
		return 0, fmt.Errorf("unknown track duration")
	}

	ratio := float64(targetTime) / float64(totalDuration)
	estimatedByte := int64(ratio * float64(contentLength))

	m.logMessage(fmt.Sprintf("Estimated byte position: %d of %d (%.1f%%)", 
		estimatedByte, contentLength, ratio*100))

	return estimatedByte, nil
}

// createRangeStreamURL creates a stream URL that uses HTTP Range headers
func (m *Manager) createRangeStreamURL(baseURL string, byteOffset int64) string {
	// For this implementation, we'll modify the HTTP request in the player
	// rather than changing the URL. Return the base URL for now.
	return baseURL
}
