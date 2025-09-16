package mpv

import (
    "fmt"
    "math/rand"
    "navitone-cli/internal/models"
    "navitone-cli/pkg/navidrome"
    "navitone-cli/pkg/scrobbling"
    "sync"
    "time"
)

// Manager handles MPV-based audio playback and queue management
type Manager struct {
	process           *MPVProcess
	ipc              *IPCClient
	commands         *CommandWrapper
	eventProcessor   *EventProcessor
	navidromeClient  *navidrome.Client
	scrobbler        *scrobbling.Manager

	// State management
	queue            []models.Track
	originalQueue    []models.Track // Store original order for when shuffle is disabled
	currentIndex     int
	isPlaying        bool
	isPaused         bool
	repeatMode       RepeatMode
    shuffleMode      bool
	position         time.Duration
	duration         time.Duration
	volume           float64

	// Callbacks
	stateCallback    func(*models.AppState)
	logCallback      func(string)

	// Synchronization
	mu               sync.RWMutex
	eventWg          sync.WaitGroup
	stopEventLoop    chan struct{}
}

// RepeatMode represents different repeat modes
type RepeatMode int

const (
	RepeatNone RepeatMode = iota
	RepeatOne
	RepeatAll
)

// NewManager creates a new MPV-based audio manager
func NewManager(navidromeClient *navidrome.Client, scrobbler *scrobbling.Manager) (*Manager, error) {
	// Create MPV process
	process := NewMPVProcess("")
	
	manager := &Manager{
		process:         process,
		navidromeClient: navidromeClient,
		scrobbler:       scrobbler,
		queue:           make([]models.Track, 0),
		currentIndex:    -1,
		repeatMode:      RepeatNone,
		volume:          1.0, // Default 100% volume
		stopEventLoop:   make(chan struct{}),
	}

	return manager, nil
}

// Start initializes and starts the MPV backend
func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Start MPV process
	if err := m.process.Start(nil); err != nil {
		return fmt.Errorf("failed to start MPV process: %w", err)
	}

	// Get IPC client
	m.ipc = m.process.GetIPC()
	if m.ipc == nil {
		m.process.Stop()
		return fmt.Errorf("failed to get IPC client")
	}

	// Create command wrapper
	m.commands = NewCommandWrapper(m.ipc)

	// Create event processor
	m.eventProcessor = NewEventProcessor()
	m.eventProcessor.SetEventCallback(m.handlePlaybackEvent)

	// Set initial volume
	if err := m.commands.SetVolume(m.volume * 100); err != nil {
		m.logMessage(fmt.Sprintf("Failed to set initial volume: %v", err))
	}

	// Set up property observations for real-time updates
	if err := m.commands.ObserveProperty(1, "playback-time"); err != nil {
		m.logMessage(fmt.Sprintf("Failed to observe playback-time: %v", err))
	}
	if err := m.commands.ObserveProperty(2, "duration"); err != nil {
		m.logMessage(fmt.Sprintf("Failed to observe duration: %v", err))
	}
	if err := m.commands.ObserveProperty(3, "pause"); err != nil {
		m.logMessage(fmt.Sprintf("Failed to observe pause: %v", err))
	}

	// Start event processing loop
	m.eventWg.Add(1)
	go m.eventLoop()

	m.logMessage("MPV backend started successfully")
	return nil
}

// Shutdown shuts down the MPV backend
func (m *Manager) Shutdown() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop event loop (if not already stopped)
	select {
	case <-m.stopEventLoop:
		// Already closed
	default:
		close(m.stopEventLoop)
	}
	m.eventWg.Wait()

	// Stop MPV process
	if m.process != nil {
		return m.process.Stop()
	}

	m.logMessage("MPV backend stopped")
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
        // Add to current queue and shuffle new tracks
        newTracksStart := len(m.queue)
        m.queue = append(m.queue, tracks...)
        m.shuffleSlice(m.queue[newTracksStart:])
    } else {
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
		if m.commands != nil {
			m.commands.Stop()
		}
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

	if m.commands != nil {
		m.commands.Stop()
	}
	m.queue = make([]models.Track, 0)
	m.originalQueue = make([]models.Track, 0)
	m.currentIndex = -1
	m.isPlaying = false
	m.isPaused = false
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

	if m.isPlaying && !m.isPaused && m.commands != nil {
		if err := m.commands.Pause(); err != nil {
			m.logMessage(fmt.Sprintf("Failed to pause: %v", err))
		} else {
			m.isPaused = true
			m.logMessage("Paused playback")
			m.notifyStateChange()
		}
	}
}

// Resume resumes the paused playback
func (m *Manager) Resume() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isPlaying && m.isPaused && m.commands != nil {
		if err := m.commands.Play(); err != nil {
			m.logMessage(fmt.Sprintf("Failed to resume: %v", err))
		} else {
			m.isPaused = false
			m.logMessage("Resumed playback")
			m.notifyStateChange()
		}
	}
}

// Stop stops the current playback
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.commands != nil {
		if err := m.commands.Stop(); err != nil {
			m.logMessage(fmt.Sprintf("Failed to stop: %v", err))
		}
	}
	m.isPlaying = false
	m.isPaused = false
	m.logMessage("Stopped playback")
	m.notifyStateChange()
}

// TogglePlayPause toggles between play and pause
func (m *Manager) TogglePlayPause() error {
	m.mu.RLock()
	playing := m.isPlaying
	paused := m.isPaused
	currentIndex := m.currentIndex
	queueLen := len(m.queue)
	m.mu.RUnlock()

	m.logMessage(fmt.Sprintf("TogglePlayPause - playing: %v, paused: %v, index: %d, queue: %d",
		playing, paused, currentIndex, queueLen))

	if playing {
		if paused {
			m.Resume()
		} else {
			m.Pause()
		}
	} else {
		return m.PlayCurrent()
	}

	return nil
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
	if m.commands != nil {
		m.commands.Stop()
	}
	m.isPlaying = false
	m.isPaused = false
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

// SeekForward seeks forward in the current track
func (m *Manager) SeekForward(seconds int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currentIndex >= 0 && m.currentIndex < len(m.queue) && m.commands != nil {
		m.logMessage(fmt.Sprintf("Seeking forward %d seconds", seconds))
		return m.commands.SeekRelative(float64(seconds))
	}

	return fmt.Errorf("no track currently playing")
}

// SeekBackward seeks backward in the current track
func (m *Manager) SeekBackward(seconds int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currentIndex >= 0 && m.currentIndex < len(m.queue) && m.commands != nil {
		m.logMessage(fmt.Sprintf("Seeking backward %d seconds", seconds))
		return m.commands.SeekRelative(float64(-seconds))
	}

	return fmt.Errorf("no track currently playing")
}

// SetVolume sets the playback volume
func (m *Manager) SetVolume(volume float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if volume < 0 {
		volume = 0
	}
	if volume > 1 {
		volume = 1
	}

	m.volume = volume

	if m.commands != nil {
		if err := m.commands.SetVolume(volume * 100); err != nil {
			m.logMessage(fmt.Sprintf("Failed to set volume: %v", err))
		} else {
			m.logMessage(fmt.Sprintf("Set volume to %.0f%%", volume*100))
		}
	}
}

// GetVolume returns the current playback volume
func (m *Manager) GetVolume() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.volume
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
	return m.isPlaying && !m.isPaused
}

// GetPosition returns the current playback position
func (m *Manager) GetPosition() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.position
}

// GetDuration returns the duration of the current track
func (m *Manager) GetDuration() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.duration
}

// Close closes the audio manager and releases resources
func (m *Manager) Close() error {
	m.Stop()
	return nil
}

// Private methods

// logMessage sends a message to the log callback if available
func (m *Manager) logMessage(message string) {
	if m.logCallback != nil {
		m.logCallback(message)
	}
}

// playTrackAtIndexLocked plays the track at the specified index (must be called with lock held)
func (m *Manager) playTrackAtIndexLocked(index int) error {
	if index < 0 || index >= len(m.queue) {
		return fmt.Errorf("invalid queue index: %d", index)
	}

	track := m.queue[index]

	// Get stream URL from Navidrome
	streamURL := m.navidromeClient.GetStreamURL(track.ID)

	// Update event processor with current track
	if m.eventProcessor != nil {
		m.eventProcessor.SetCurrentTrackID(track.ID)
	}

	// Load file in MPV
	if m.commands != nil {
		if err := m.commands.LoadFile(streamURL, "replace"); err != nil {
			// Fallback to download URL
			downloadURL := m.navidromeClient.GetDownloadURL(track.ID)
			if err := m.commands.LoadFile(downloadURL, "replace"); err != nil {
				return fmt.Errorf("failed to load track: %w", err)
			}
		}
	}

	m.currentIndex = index
	m.isPlaying = true
	m.isPaused = false
	m.duration = time.Duration(track.Duration) * time.Second

	m.logMessage(fmt.Sprintf("Playing track: %s - %s", track.Artist, track.Title))
	m.notifyStateChange()

    // Submit "Now Playing" (routes to server/client based on method)
    if m.scrobbler != nil {
        scrobbleTrack := scrobbling.ScrobbleTrack{
            Title:    track.Title,
            Artist:   track.Artist,
            Album:    track.Album,
            Duration: track.Duration,
        }
        go m.scrobbler.NowPlaying(track.ID, scrobbleTrack)
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

// shuffleSlice shuffles a slice of tracks in place
func (m *Manager) shuffleSlice(tracks []models.Track) {
    // Fisher-Yates shuffle with math/rand
    for i := len(tracks) - 1; i > 0; i-- {
        j := rand.Intn(i + 1)
        tracks[i], tracks[j] = tracks[j], tracks[i]
    }
}

// ToggleShuffle toggles shuffle mode on/off
func (m *Manager) ToggleShuffle() {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.shuffleMode = !m.shuffleMode

    if m.shuffleMode {
        // Seed RNG for shuffle randomness
        rand.Seed(time.Now().UnixNano())
        // Save original order
        m.originalQueue = make([]models.Track, len(m.queue))
        copy(m.originalQueue, m.queue)

        // Track currently playing item
        var currentTrack *models.Track
        if m.currentIndex >= 0 && m.currentIndex < len(m.queue) {
            ct := m.queue[m.currentIndex]
            currentTrack = &ct
        }

        // Shuffle entire queue
        m.shuffleSlice(m.queue)

        // Re-locate current track index after shuffle
        if currentTrack != nil {
            for i, t := range m.queue {
                if t.ID == currentTrack.ID {
                    m.currentIndex = i
                    break
                }
            }
        }

        m.logMessage(fmt.Sprintf("Shuffle enabled - queue randomized (%d tracks)", len(m.queue)))
    } else {
        // Restore original order if available
        if len(m.originalQueue) > 0 {
            var currentTrack *models.Track
            if m.currentIndex >= 0 && m.currentIndex < len(m.queue) {
                ct := m.queue[m.currentIndex]
                currentTrack = &ct
            }

            copy(m.queue, m.originalQueue)
            // Trim in case sizes differ (shouldn't, but safe)
            if len(m.queue) > len(m.originalQueue) {
                m.queue = m.queue[:len(m.originalQueue)]
            }

            if currentTrack != nil {
                for i, t := range m.queue {
                    if t.ID == currentTrack.ID {
                        m.currentIndex = i
                        break
                    }
                }
            }
            m.originalQueue = nil
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

// notifyStateChange notifies the UI about state changes (must be called with lock held)
func (m *Manager) notifyStateChange() {
	if m.stateCallback != nil {
		go func() {
			m.stateCallback(nil)
		}()
	}
}

// eventLoop processes MPV events
func (m *Manager) eventLoop() {
	defer m.eventWg.Done()

	if m.ipc == nil {
		return
	}

	events := m.ipc.GetEvents()

	for {
		select {
		case <-m.stopEventLoop:
			return

		case event, ok := <-events:
			if !ok {
				return // Events channel closed
			}

			if m.eventProcessor != nil {
				m.eventProcessor.ProcessEvent(event)
			}
		}
	}
}

// handlePlaybackEvent handles processed playback events
func (m *Manager) handlePlaybackEvent(event PlaybackEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch event.Type {
	case EventTrackStarted:
		m.logMessage("Track started")

	case EventTrackFinished:
		m.logMessage("Track finished")
		
        // Submit scrobble for completed track (routes to server/client)
        if m.scrobbler != nil && m.currentIndex >= 0 && m.currentIndex < len(m.queue) {
            track := m.queue[m.currentIndex]
            scrobbleTrack := scrobbling.ScrobbleTrack{
                Title:       track.Title,
                Artist:      track.Artist,
                Album:       track.Album,
                Duration:    track.Duration,
                TrackNumber: track.Track,
                Timestamp:   time.Now().Unix(),
            }
            m.logMessage(fmt.Sprintf("Scrobbling completed track: %s - %s", track.Artist, track.Title))
            go m.scrobbler.SubmitScrobble(track.ID, scrobbleTrack)
        }
		
		// Auto-advance to next track
		go func() {
			time.Sleep(100 * time.Millisecond) // Brief delay
			m.NextTrack()
		}()

	case EventTrackError:
		m.logMessage(fmt.Sprintf("Track error: %v", event.Data))
		// Try next track
		go func() {
			time.Sleep(100 * time.Millisecond)
			m.NextTrack()
		}()

	case EventPositionUpdate:
		if event.Position > 0 {
			m.position = event.Position
		}
		if event.Duration > 0 {
			m.duration = event.Duration
		}

	case EventStateChange:
		// Handle state changes from MPV
		if dataMap, ok := event.Data.(map[string]interface{}); ok {
			if paused, exists := dataMap["paused"]; exists {
				if pausedBool, ok := paused.(bool); ok {
					m.isPaused = pausedBool
				}
			}
			if idle, exists := dataMap["idle"]; exists {
				if idleBool, ok := idle.(bool); ok && idleBool {
					m.isPlaying = false
					m.isPaused = false
				}
			}
		}
	}

	m.notifyStateChange()
}
