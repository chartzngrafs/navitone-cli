package audio

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ebitengine/oto/v3"
)

// StreamTracker wraps an io.Reader to track when the HTTP stream ends
type StreamTracker struct {
	reader io.Reader
	id     string
}

func (st *StreamTracker) Read(p []byte) (n int, err error) {
	n, err = st.reader.Read(p)
	return n, err
}

// PlaybackState represents the current state of playback
type PlaybackState int

const (
	StateStopped PlaybackState = iota
	StatePlaying
	StatePaused
)

// PlaybackEvent represents events emitted by the player
type PlaybackEvent struct {
	Type     string
	TrackID  string
	Position time.Duration
	Duration time.Duration
}

// Player represents the audio player
type Player struct {
	context    *oto.Context
	player     *oto.Player
	httpClient *http.Client

	// State
	state      PlaybackState
	currentURL string
	currentID  string
	formatHint string
	volume     float64
	position   time.Duration
	duration   time.Duration
	byteOffset int64  // HTTP Range byte offset for seeking

	// Control channels
	stopCh   chan struct{}
	pauseCh  chan struct{}
	resumeCh chan struct{}

	// Event callback
	eventCallback func(PlaybackEvent)


	// Position offset for seeking simulation
	positionOffset time.Duration

	// Synchronization
	mu sync.RWMutex
	wg sync.WaitGroup
}

// NewPlayer creates a new audio player
func NewPlayer() (*Player, error) {
	// Initialize Oto context with reasonable defaults
	// Use conservative settings to minimize audio issues
	op := &oto.NewContextOptions{
		SampleRate:   44100,
		ChannelCount: 2,
		Format:       oto.FormatSignedInt16LE,
		// Add buffer size configuration to prevent underruns
		BufferSize: time.Millisecond * 100, // 100ms buffer
	}


	ctx, readyChan, err := oto.NewContext(op)
	if err != nil {
		return nil, fmt.Errorf("failed to create audio context: %w", err)
	}

	// Wait for the context to be ready
	<-readyChan

	player := &Player{
		context:    ctx,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		state:      StateStopped,
		volume:     0.7, // Default volume 70%
		stopCh:     make(chan struct{}),
		pauseCh:    make(chan struct{}),
		resumeCh:   make(chan struct{}),
	}

	return player, nil
}

// Play starts playing a track from the given stream URL
func (p *Player) Play(streamURL, trackID string) error {
	return p.PlayWithFormat(streamURL, trackID, "")
}

// PlayWithFormat starts playing a track with a format hint
func (p *Player) PlayWithFormat(streamURL, trackID, formatHint string) error {
	return p.PlayWithFormatAndDuration(streamURL, trackID, formatHint, 0)
}

// PlayWithFormatAndDuration starts playing a track with format hint and duration
func (p *Player) PlayWithFormatAndDuration(streamURL, trackID, formatHint string, duration time.Duration) error {
	return p.PlayWithRange(streamURL, trackID, formatHint, duration, 0)
}

// PlayWithRange starts playing a track from a specific byte offset using HTTP Range headers
func (p *Player) PlayWithRange(streamURL, trackID, formatHint string, duration time.Duration, byteOffset int64) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Stop current playback if any
	if p.state == StatePlaying || p.state == StatePaused {
		p.stopPlayback()
	}

	p.currentURL = streamURL
	p.currentID = trackID
	p.position = 0
	p.duration = duration
	p.byteOffset = byteOffset  // Store byte offset for range requests
	p.positionOffset = 0       // Reset position offset for new track

	// Store format hint for playback loop
	p.formatHint = formatHint

	// Start new playback
	p.wg.Add(1)
	go p.playbackLoop()

	p.state = StatePlaying
	p.emitEvent("play_started", trackID, 0, 0)

	return nil
}

// Pause pauses the current playback
func (p *Player) Pause() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state == StatePlaying && p.player != nil {
		p.state = StatePaused
		// Directly pause the oto player
		p.player.Pause()
		p.emitEvent("paused", p.currentID, p.position, p.duration)
	}
}

// Resume resumes the paused playback
func (p *Player) Resume() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state == StatePaused && p.player != nil {
		p.state = StatePlaying
		// Directly resume the oto player
		p.player.Play()
		p.emitEvent("resumed", p.currentID, p.position, p.duration)
	}
}

// Stop stops the current playback
func (p *Player) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state == StatePlaying || p.state == StatePaused {
		p.stopPlayback()
		p.state = StateStopped
		p.emitEvent("stopped", p.currentID, p.position, p.duration)
	}
}

// SetVolume sets the playback volume (0.0 to 1.0)
func (p *Player) SetVolume(volume float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if volume < 0 {
		volume = 0
	}
	if volume > 1 {
		volume = 1
	}

	p.volume = volume
	// TODO: Apply volume to current player if playing
}

// GetState returns the current playback state
func (p *Player) GetState() PlaybackState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state
}

// GetPosition returns the current playback position
func (p *Player) GetPosition() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.position
}

// GetDuration returns the duration of the current track
func (p *Player) GetDuration() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.duration
}

// GetCurrentTrack returns the ID of the currently playing track
func (p *Player) GetCurrentTrack() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.currentID
}

// SetEventCallback sets the callback function for playback events
func (p *Player) SetEventCallback(callback func(PlaybackEvent)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.eventCallback = callback
}

// AdjustPositionOffset adjusts the position offset for simulated seeking
func (p *Player) AdjustPositionOffset(offset time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.positionOffset += offset
}

// Close closes the audio player and releases resources
func (p *Player) Close() error {
	p.Stop()
	p.wg.Wait()

	if p.context != nil {
		p.context.Suspend()
	}
	return nil
}

// SeekForward seeks forward by the specified duration
func (p *Player) SeekForward(duration time.Duration) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.state != StatePlaying && p.state != StatePaused {
		return fmt.Errorf("no track currently playing")
	}
	
	// For now, seeking with streaming audio requires restarting playback
	// This is a limitation of the current implementation 
	newPosition := p.position + duration
	if newPosition >= p.duration && p.duration > 0 {
		newPosition = p.duration - time.Second // Stop 1 second before end
	}
	
	return p.seekToPosition(newPosition)
}

// SeekBackward seeks backward by the specified duration  
func (p *Player) SeekBackward(duration time.Duration) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.state != StatePlaying && p.state != StatePaused {
		return fmt.Errorf("no track currently playing")
	}
	
	newPosition := p.position - duration
	if newPosition < 0 {
		newPosition = 0
	}
	
	return p.seekToPosition(newPosition)
}

// seekToPosition seeks to a specific position (must be called with lock held)
func (p *Player) seekToPosition(position time.Duration) error {
	if position < 0 {
		position = 0
	}
	if p.duration > 0 && position > p.duration {
		position = p.duration
	}
	
	// For streaming audio, we restart playback from the desired position
	// This works by requesting a new stream with a time offset
	offsetSeconds := int(position.Seconds())
	
	// Stop current playback
	wasPlaying := p.state == StatePlaying
	p.stopPlayback()
	
	// Create new stream URL with time offset  
	// We need access to navidrome client for this, so this is a limitation
	// For now, just update position and emit event
	p.position = position
	p.emitEvent("seek", p.currentID, p.position, p.duration)
	
	// If we were playing, we should restart playback from the new position
	// This requires coordination with the audio manager
	if wasPlaying {
		// Signal that seeking occurred and we need to restart
		return fmt.Errorf("seek to %d seconds - requires playback restart", offsetSeconds)
	}
	
	return nil
}

// stopPlayback stops the current playback (must be called with lock held)
func (p *Player) stopPlayback() {
	if p.player != nil {
		p.player.Close()
		p.player = nil
	}

	select {
	case p.stopCh <- struct{}{}:
	default:
	}
}

// emitEvent emits a playback event (must be called with lock held)
func (p *Player) emitEvent(eventType, trackID string, position, duration time.Duration) {
	if p.eventCallback != nil {
		event := PlaybackEvent{
			Type:     eventType,
			TrackID:  trackID,
			Position: position,
			Duration: duration,
		}
		// Emit event in goroutine to avoid blocking
		go p.eventCallback(event)
	}
}

// playbackLoop handles the actual audio playback
func (p *Player) playbackLoop() {
	defer p.wg.Done()


	// Create HTTP request for the stream with no timeout for streaming
	req, err := http.NewRequest("GET", p.currentURL, nil)
	if err != nil {
		p.emitEvent("error", p.currentID, 0, 0)
		return
	}

	// Add range header for seeking if byte offset is specified
	if p.byteOffset > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", p.byteOffset))
	}
	req.Header.Set("User-Agent", "navitone-cli/1.0")

	// Use a client with no timeout for streaming (different from the default httpClient)
	streamingClient := &http.Client{
		Timeout: 0, // No timeout for streaming audio
	}
	
	resp, err := streamingClient.Do(req)
	if err != nil {
		p.emitEvent("error", p.currentID, 0, 0)
		return
	}
	defer resp.Body.Close()

	
	
	if resp.StatusCode != http.StatusOK {
		p.emitEvent("error", p.currentID, 0, 0)
		return
	}

	// Detect audio format from URL, content-type, or format hint
	format := p.detectAudioFormat(p.currentURL, resp.Header.Get("Content-Type"))

	// Use appropriate decoder based on format
	var audioReader io.Reader
	if format != "" {
		decoder, err := NewDecoder(format)
		if err != nil {
			// Fallback to MP3 decoder
			decoder, err = NewDecoder("mp3")
			if err != nil {
				p.emitEvent("error", p.currentID, 0, 0)
				return
			}
		}

		
		decodedReader, err := decoder.Decode(resp.Body)
		if err != nil {
			// Reset response body - we need to make a new request
			resp.Body.Close()
			
			// Make new request for fallback
			req2, err := http.NewRequest("GET", p.currentURL, nil)
			if err != nil {
				p.emitEvent("error", p.currentID, 0, 0)
				return
			}
			req2.Header.Set("User-Agent", "navitone-cli/1.0")
			
			resp2, err := streamingClient.Do(req2)
			if err != nil {
				p.emitEvent("error", p.currentID, 0, 0)
				return
			}
			defer resp2.Body.Close()
			
			if resp2.StatusCode != http.StatusOK {
				p.emitEvent("error", p.currentID, 0, 0)
				return
			}
			
			// Try MP3 decoder
			mp3Decoder, err := NewDecoder("mp3")
			if err != nil {
				p.emitEvent("error", p.currentID, 0, 0)
				return
			}
			
			decodedReader, err = mp3Decoder.Decode(resp2.Body)
			if err != nil {
				p.emitEvent("error", p.currentID, 0, 0)
				return
			}
			
		} else {
			}


		audioReader = decodedReader
	} else {
		p.emitEvent("error", p.currentID, 0, 0)
		return
	}

	// Create a new Oto player for this stream
	p.mu.Lock()
	p.player = p.context.NewPlayer(audioReader)
	p.mu.Unlock()

	// Start playback
	p.player.Play()

	// Position tracking loop
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	startTime := time.Now()
	var pausedDuration time.Duration

	for {
		select {
		case <-p.stopCh:
			return

		case <-p.pauseCh:
			// Pause handling is now done directly in Pause() method
			continue
		
		case <-p.resumeCh:
			// Resume handling is now done directly in Resume() method
			continue

		case <-ticker.C:
			if p.GetState() == StatePlaying {
				p.mu.Lock()
				p.position = time.Since(startTime) - pausedDuration + p.positionOffset
				p.mu.Unlock()

				p.emitEvent("position_update", p.currentID, p.position, p.duration)
			}

		default:
			// Check if track finished - use multiple criteria to avoid false positives
			if p.player != nil {
				isPlaying := p.player.IsPlaying()
				
				// Get current state for debugging
				p.mu.RLock()
				currentPosition := p.position
				trackDuration := p.duration
				trackID := p.currentID
				p.mu.RUnlock()
				
				
				if !isPlaying {
					// Only consider track "finished" if:
					// 1. Player reports not playing AND
					// 2. We've played for at least 30 seconds (to avoid false positives) AND
					// 3. Either we've played 90% of the track OR we've exceeded the track duration
					minPlayTime := 30 * time.Second
					finishThreshold := time.Duration(float64(trackDuration) * 0.9) // 90% of track
					
					hasPlayedMinimum := currentPosition >= minPlayTime
					hasPlayedMostOfTrack := currentPosition >= finishThreshold
					hasExceededDuration := trackDuration > 0 && currentPosition >= trackDuration
					
					if hasPlayedMinimum && (hasPlayedMostOfTrack || hasExceededDuration) {
						p.emitEvent("finished", trackID, currentPosition, trackDuration)
						return
					}
					// Don't emit finished event for premature stops
				}
			}
		}
	}
}

// detectAudioFormat detects the audio format from URL, content-type, or format hint
func (p *Player) detectAudioFormat(url, contentType string) string {
	// First priority: Use format hint from track metadata
	if p.formatHint != "" {
		hint := strings.ToLower(p.formatHint)
		// Normalize common extensions
		switch hint {
		case "mp3", "mpeg":
			return "mp3"
		case "flac":
			return "flac"
		case "ogg", "oga", "vorbis":
			return "ogg"
		case "wav", "wave":
			return "wav"
		case "m4a", "aac":
			return "mp3" // Fallback to MP3 decoder for now
		}
	}

	// Second priority: Try to detect from URL extension
	url = strings.ToLower(url)
	if strings.Contains(url, ".mp3") || strings.Contains(url, "format=mp3") {
		return "mp3"
	}
	if strings.Contains(url, ".flac") || strings.Contains(url, "format=flac") {
		return "flac"
	}
	if strings.Contains(url, ".ogg") || strings.Contains(url, "format=ogg") {
		return "ogg"
	}
	if strings.Contains(url, ".wav") || strings.Contains(url, "format=wav") {
		return "wav"
	}

	// Third priority: Try to detect from content-type
	contentType = strings.ToLower(contentType)
	if strings.Contains(contentType, "audio/mpeg") || strings.Contains(contentType, "audio/mp3") {
		return "mp3"
	}
	if strings.Contains(contentType, "audio/flac") {
		return "flac"
	}
	if strings.Contains(contentType, "audio/ogg") || strings.Contains(contentType, "audio/vorbis") {
		return "ogg"
	}
	if strings.Contains(contentType, "audio/wav") || strings.Contains(contentType, "audio/wave") {
		return "wav"
	}

	// Last resort: Default to MP3
	return "mp3"
}
