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

	// Control channels
	stopCh   chan struct{}
	pauseCh  chan struct{}
	resumeCh chan struct{}

	// Event callback
	eventCallback func(PlaybackEvent)

	// Synchronization
	mu sync.RWMutex
	wg sync.WaitGroup
}

// NewPlayer creates a new audio player
func NewPlayer() (*Player, error) {
	// Initialize Oto context with reasonable defaults
	// Let's try different settings to match common audio formats
	op := &oto.NewContextOptions{
		SampleRate:   44100,
		ChannelCount: 2,
		Format:       oto.FormatSignedInt16LE,
	}

	fmt.Printf("[AUDIO DEBUG] Creating Oto context with: SampleRate=%d, Channels=%d, Format=%v\n",
		op.SampleRate, op.ChannelCount, op.Format)

	ctx, readyChan, err := oto.NewContext(op)
	if err != nil {
		return nil, fmt.Errorf("failed to create audio context: %w", err)
	}

	// Wait for the context to be ready
	<-readyChan
	fmt.Printf("[AUDIO DEBUG] Oto context ready\n")

	player := &Player{
		context:    ctx,
		httpClient: &http.Client{Timeout: 30 * time.Second},
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
	p.mu.Lock()
	defer p.mu.Unlock()

	fmt.Printf("[AUDIO DEBUG] PlayWithFormat called with URL: %s, trackID: %s, format: %s\n", streamURL, trackID, formatHint)

	// Stop current playback if any
	if p.state == StatePlaying || p.state == StatePaused {
		fmt.Printf("[AUDIO DEBUG] Stopping current playback\n")
		p.stopPlayback()
	}

	p.currentURL = streamURL
	p.currentID = trackID
	p.position = 0

	// Store format hint for playback loop
	p.formatHint = formatHint

	// Start new playback
	fmt.Printf("[AUDIO DEBUG] Starting playback loop\n")
	p.wg.Add(1)
	go p.playbackLoop()

	p.state = StatePlaying
	p.emitEvent("play_started", trackID, 0, 0)

	fmt.Printf("[AUDIO DEBUG] Play method completed, state: %d\n", p.state)
	return nil
}

// Pause pauses the current playback
func (p *Player) Pause() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state == StatePlaying {
		p.state = StatePaused
		select {
		case p.pauseCh <- struct{}{}:
		default:
		}
		p.emitEvent("paused", p.currentID, p.position, p.duration)
	}
}

// Resume resumes the paused playback
func (p *Player) Resume() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state == StatePaused {
		p.state = StatePlaying
		select {
		case p.resumeCh <- struct{}{}:
		default:
		}
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

// Close closes the audio player and releases resources
func (p *Player) Close() error {
	p.Stop()
	p.wg.Wait()

	if p.context != nil {
		p.context.Suspend()
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

	fmt.Printf("[AUDIO DEBUG] Playback loop started for URL: %s\n", p.currentURL)

	// Create HTTP request for the stream
	req, err := http.NewRequest("GET", p.currentURL, nil)
	if err != nil {
		fmt.Printf("[AUDIO DEBUG] Failed to create HTTP request: %v\n", err)
		p.emitEvent("error", p.currentID, 0, 0)
		return
	}

	// Add range header to support seeking (if needed in the future)
	req.Header.Set("User-Agent", "navitone-cli/1.0")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		fmt.Printf("[AUDIO DEBUG] HTTP request failed: %v\n", err)
		p.emitEvent("error", p.currentID, 0, 0)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("[AUDIO DEBUG] HTTP response status: %d\n", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("[AUDIO DEBUG] Bad HTTP status: %d\n", resp.StatusCode)
		p.emitEvent("error", p.currentID, 0, 0)
		return
	}

	// Detect audio format from URL, content-type, or format hint
	format := p.detectAudioFormat(p.currentURL, resp.Header.Get("Content-Type"))
	fmt.Printf("[AUDIO DEBUG] Detected format: %s, Content-Type: %s, Format hint: %s\n", format, resp.Header.Get("Content-Type"), p.formatHint)

	// Use appropriate decoder based on format
	var audioReader io.Reader
	if format != "" {
		decoder, err := NewDecoder(format)
		if err != nil {
			fmt.Printf("[AUDIO DEBUG] Failed to create decoder for %s: %v\n", format, err)
			p.emitEvent("error", p.currentID, 0, 0)
			return
		}

		decodedReader, err := decoder.Decode(resp.Body)
		if err != nil {
			fmt.Printf("[AUDIO DEBUG] Failed to decode %s: %v\n", format, err)
			p.emitEvent("error", p.currentID, 0, 0)
			return
		}

		fmt.Printf("[AUDIO DEBUG] Successfully created %s decoder\n", format)
		fmt.Printf("[AUDIO DEBUG] Decoder sample rate: %d\n", decoder.SampleRate())
		fmt.Printf("[AUDIO DEBUG] Decoder channels: %d\n", decoder.Channels())

		audioReader = decodedReader
	} else {
		fmt.Printf("[AUDIO DEBUG] Unknown format, cannot decode\n")
		p.emitEvent("error", p.currentID, 0, 0)
		return
	}

	// Create a new Oto player for this stream
	p.mu.Lock()
	fmt.Printf("[AUDIO DEBUG] Creating Oto player from audioReader\n")
	p.player = p.context.NewPlayer(audioReader)
	fmt.Printf("[AUDIO DEBUG] Oto player created successfully\n")
	p.mu.Unlock()

	fmt.Printf("[AUDIO DEBUG] Starting playback\n")
	// Start playback
	p.player.Play()
	fmt.Printf("[AUDIO DEBUG] Called player.Play() - audio should be playing now\n")

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
			p.player.Pause()
			pauseStart := time.Now()

			// Wait for resume or stop
			select {
			case <-p.resumeCh:
				p.player.Play()
				pausedDuration += time.Since(pauseStart)
			case <-p.stopCh:
				return
			}

		case <-ticker.C:
			if p.GetState() == StatePlaying {
				p.mu.Lock()
				p.position = time.Since(startTime) - pausedDuration
				p.mu.Unlock()

				p.emitEvent("position_update", p.currentID, p.position, p.duration)
			}

		default:
			// Check if playbook finished
			if p.player != nil && !p.player.IsPlaying() {
				p.emitEvent("finished", p.currentID, p.position, p.duration)
				return
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
