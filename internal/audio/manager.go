package audio

import (
	"navitone-cli/internal/audio/mpv"
	"navitone-cli/internal/models"
	"navitone-cli/pkg/navidrome"
	"navitone-cli/pkg/scrobbling"
	"time"
)

// Manager is a wrapper around the MPV manager to maintain API compatibility
type Manager struct {
	mpvManager *mpv.Manager
}

// RepeatMode represents different repeat modes
type RepeatMode = mpv.RepeatMode

// Repeat mode constants
const (
	RepeatNone = mpv.RepeatNone
	RepeatOne  = mpv.RepeatOne
	RepeatAll  = mpv.RepeatAll
)

// NewManager creates a new MPV-based audio manager
func NewManager(navidromeClient *navidrome.Client, scrobbler *scrobbling.Manager) (*Manager, error) {
	mpvManager, err := mpv.NewManager(navidromeClient, scrobbler)
	if err != nil {
		return nil, err
	}

	manager := &Manager{
		mpvManager: mpvManager,
	}

	// Start the MPV backend
	if err := mpvManager.Start(); err != nil {
		return nil, err
	}

	return manager, nil
}

// SetStateCallback sets the callback function for state updates
func (m *Manager) SetStateCallback(callback func(*models.AppState)) {
	m.mpvManager.SetStateCallback(callback)
}

// SetLogCallback sets the callback function for log messages
func (m *Manager) SetLogCallback(callback func(string)) {
	m.mpvManager.SetLogCallback(callback)
}

// AddToQueue adds a track to the playback queue
func (m *Manager) AddToQueue(track models.Track) {
	m.mpvManager.AddToQueue(track)
}

// AddTracksToQueue adds multiple tracks to the playback queue
func (m *Manager) AddTracksToQueue(tracks []models.Track) {
	m.mpvManager.AddTracksToQueue(tracks)
}

// RemoveFromQueue removes a track from the queue at the specified index
func (m *Manager) RemoveFromQueue(index int) {
	m.mpvManager.RemoveFromQueue(index)
}

// ClearQueue removes all tracks from the queue
func (m *Manager) ClearQueue() {
	m.mpvManager.ClearQueue()
}

// PlayTrackAtIndex starts playing the track at the specified queue index
func (m *Manager) PlayTrackAtIndex(index int) error {
	return m.mpvManager.PlayTrackAtIndex(index)
}

// PlayCurrent plays the current track (or first track if none selected)
func (m *Manager) PlayCurrent() error {
	return m.mpvManager.PlayCurrent()
}

// Pause pauses the current playback
func (m *Manager) Pause() {
	m.mpvManager.Pause()
}

// Resume resumes the paused playback
func (m *Manager) Resume() {
	m.mpvManager.Resume()
}

// Stop stops the current playback
func (m *Manager) Stop() {
	m.mpvManager.Stop()
}

// TogglePlayPause toggles between play and pause
func (m *Manager) TogglePlayPause() error {
	return m.mpvManager.TogglePlayPause()
}

// NextTrack plays the next track in the queue
func (m *Manager) NextTrack() error {
	return m.mpvManager.NextTrack()
}

// PreviousTrack plays the previous track in the queue
func (m *Manager) PreviousTrack() error {
	return m.mpvManager.PreviousTrack()
}

// SeekForward seeks forward in the current track
func (m *Manager) SeekForward(seconds int) error {
	return m.mpvManager.SeekForward(seconds)
}

// SeekBackward seeks backward in the current track
func (m *Manager) SeekBackward(seconds int) error {
	return m.mpvManager.SeekBackward(seconds)
}

// SetVolume sets the playback volume (0.0 to 1.0)
func (m *Manager) SetVolume(volume float64) {
	m.mpvManager.SetVolume(volume)
}

// GetVolume returns the current playback volume (0.0 to 1.0)
func (m *Manager) GetVolume() float64 {
	return m.mpvManager.GetVolume()
}

// GetQueue returns a copy of the current queue
func (m *Manager) GetQueue() []models.Track {
	return m.mpvManager.GetQueue()
}

// GetCurrentTrack returns the currently playing track
func (m *Manager) GetCurrentTrack() *models.Track {
	return m.mpvManager.GetCurrentTrack()
}

// GetCurrentIndex returns the current track index
func (m *Manager) GetCurrentIndex() int {
	return m.mpvManager.GetCurrentIndex()
}

// IsPlaying returns whether audio is currently playing
func (m *Manager) IsPlaying() bool {
	return m.mpvManager.IsPlaying()
}

// GetPosition returns the current playback position
func (m *Manager) GetPosition() time.Duration {
	return m.mpvManager.GetPosition()
}

// GetDuration returns the duration of the current track
func (m *Manager) GetDuration() time.Duration {
	return m.mpvManager.GetDuration()
}

// Close closes the audio manager and releases resources
func (m *Manager) Close() error {
	return m.mpvManager.Shutdown()
}

// Additional methods that may have been used by the old system

// CheckStreamingPermissions verifies that the user has proper streaming access
func (m *Manager) CheckStreamingPermissions() error {
	// MPV handles streaming permissions through the navidrome client
	// This is a no-op for compatibility
	return nil
}

// ToggleShuffle toggles shuffle mode on/off (if implemented in MPV manager)
func (m *Manager) ToggleShuffle() {
	// TODO: Implement shuffle mode in MPV manager
}

// IsShuffleEnabled returns whether shuffle mode is enabled (if implemented in MPV manager)
func (m *Manager) IsShuffleEnabled() bool {
	// TODO: Implement shuffle mode in MPV manager
	return false
}