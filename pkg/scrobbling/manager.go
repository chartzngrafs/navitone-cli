package scrobbling

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"navitone-cli/internal/config"
)

// Manager handles scrobbling to multiple services
type Manager struct {
	config         *config.Config
	lastfm         *LastFMClient
	listenbrainz   *ListenBrainzClient
	queuedScrobbles []QueuedScrobble
	mutex          sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
}

// NewManager creates a new scrobbling manager
func NewManager(cfg *config.Config) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	
	manager := &Manager{
		config:          cfg,
		queuedScrobbles: make([]QueuedScrobble, 0),
		ctx:             ctx,
		cancel:          cancel,
	}

	// Initialize clients if enabled
	if cfg.Scrobbling.LastFM.Enabled {
		manager.lastfm = NewLastFMClient(
			cfg.Scrobbling.LastFM.APIKey,
			cfg.Scrobbling.LastFM.Secret,
			cfg.Scrobbling.LastFM.Username,
			cfg.Scrobbling.LastFM.Password,
		)
	}

	if cfg.Scrobbling.ListenBrainz.Enabled {
		manager.listenbrainz = NewListenBrainzClient(cfg.Scrobbling.ListenBrainz.Token)
	}

	// Start retry worker
	go manager.retryWorker()

	return manager
}

// Close shuts down the scrobbling manager
func (m *Manager) Close() {
	m.cancel()
}

// Scrobble submits a scrobble to all enabled services
func (m *Manager) Scrobble(track ScrobbleTrack) []ScrobbleResult {
	var results []ScrobbleResult
	var wg sync.WaitGroup

	resultsChan := make(chan ScrobbleResult, 2)

	// Scrobble to Last.fm
	if m.lastfm != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := m.scrobbleLastFM(track)
			resultsChan <- result
		}()
	}

	// Scrobble to ListenBrainz
	if m.listenbrainz != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := m.scrobbleListenBrainz(track)
			resultsChan <- result
		}()
	}

	// Wait for all scrobbles and collect results
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	for result := range resultsChan {
		results = append(results, result)
		
		// Queue failed scrobbles for retry
		if !result.Success {
			m.queueForRetry(result.Track, result.Service)
		}
	}

	return results
}

// UpdateNowPlaying updates now playing status on all enabled services
func (m *Manager) UpdateNowPlaying(track ScrobbleTrack) []ScrobbleResult {
	var results []ScrobbleResult
	var wg sync.WaitGroup

	resultsChan := make(chan ScrobbleResult, 2)

	// Update Last.fm
	if m.lastfm != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := m.updateNowPlayingLastFM(track)
			resultsChan <- result
		}()
	}

	// Update ListenBrainz
	if m.listenbrainz != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := m.updateNowPlayingListenBrainz(track)
			resultsChan <- result
		}()
	}

	// Wait for all updates and collect results
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	for result := range resultsChan {
		results = append(results, result)
	}

	return results
}

// scrobbleLastFM handles Last.fm scrobbling
func (m *Manager) scrobbleLastFM(track ScrobbleTrack) ScrobbleResult {
	result := ScrobbleResult{
		Service:   "Last.fm",
		Track:     track,
		Timestamp: time.Now().Unix(),
	}

	// Authenticate if needed
	if m.lastfm.sessionKey == "" {
		if err := m.lastfm.Authenticate(m.ctx); err != nil {
			result.Error = fmt.Errorf("authentication failed: %w", err)
			return result
		}
	}

	// Submit scrobble
	if err := m.lastfm.Scrobble(m.ctx, track); err != nil {
		result.Error = err
		return result
	}

	result.Success = true
	return result
}

// scrobbleListenBrainz handles ListenBrainz scrobbling
func (m *Manager) scrobbleListenBrainz(track ScrobbleTrack) ScrobbleResult {
	result := ScrobbleResult{
		Service:   "ListenBrainz",
		Track:     track,
		Timestamp: time.Now().Unix(),
	}

	// Convert to ListenBrainz format
	listen := Listen{
		ListenedAt: int(track.Timestamp),
		TrackMetadata: TrackMetadata{
			ArtistName:  track.Artist,
			TrackName:   track.Title,
			ReleaseName: track.Album,
			AdditionalInfo: map[string]interface{}{
				"duration_ms": track.Duration * 1000,
			},
		},
	}

	if track.TrackNumber > 0 {
		listen.TrackMetadata.AdditionalInfo["tracknumber"] = track.TrackNumber
	}

	// Submit listen
	if err := m.listenbrainz.SubmitListen(m.ctx, listen); err != nil {
		result.Error = err
		return result
	}

	result.Success = true
	return result
}

// updateNowPlayingLastFM handles Last.fm now playing updates
func (m *Manager) updateNowPlayingLastFM(track ScrobbleTrack) ScrobbleResult {
	result := ScrobbleResult{
		Service:   "Last.fm (Now Playing)",
		Track:     track,
		Timestamp: time.Now().Unix(),
	}

	// Authenticate if needed
	if m.lastfm.sessionKey == "" {
		if err := m.lastfm.Authenticate(m.ctx); err != nil {
			result.Error = fmt.Errorf("authentication failed: %w", err)
			return result
		}
	}

	// Update now playing
	if err := m.lastfm.UpdateNowPlaying(m.ctx, track); err != nil {
		result.Error = err
		return result
	}

	result.Success = true
	return result
}

// updateNowPlayingListenBrainz handles ListenBrainz playing now updates
func (m *Manager) updateNowPlayingListenBrainz(track ScrobbleTrack) ScrobbleResult {
	result := ScrobbleResult{
		Service:   "ListenBrainz (Playing Now)",
		Track:     track,
		Timestamp: time.Now().Unix(),
	}

	// Convert to ListenBrainz format
	metadata := TrackMetadata{
		ArtistName:  track.Artist,
		TrackName:   track.Title,
		ReleaseName: track.Album,
		AdditionalInfo: map[string]interface{}{
			"duration_ms": track.Duration * 1000,
		},
	}

	if track.TrackNumber > 0 {
		metadata.AdditionalInfo["tracknumber"] = track.TrackNumber
	}

	// Submit playing now
	if err := m.listenbrainz.SubmitPlayingNow(m.ctx, metadata); err != nil {
		result.Error = err
		return result
	}

	result.Success = true
	return result
}

// queueForRetry adds a failed scrobble to the retry queue
func (m *Manager) queueForRetry(track ScrobbleTrack, service string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	queued := QueuedScrobble{
		Track:      track,
		Service:    service,
		Attempts:   1,
		LastTry:    time.Now().Unix(),
		MaxRetries: 3,
	}

	m.queuedScrobbles = append(m.queuedScrobbles, queued)
}

// retryWorker periodically retries failed scrobbles
func (m *Manager) retryWorker() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.retryQueuedScrobbles()
		}
	}
}

// retryQueuedScrobbles attempts to retry failed scrobbles
func (m *Manager) retryQueuedScrobbles() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var remaining []QueuedScrobble

	for _, queued := range m.queuedScrobbles {
		// Skip if max retries reached
		if queued.Attempts >= queued.MaxRetries {
			log.Printf("Dropping scrobble after %d attempts: %s - %s", 
				queued.Attempts, queued.Track.Artist, queued.Track.Title)
			continue
		}

		// Skip if too recent
		if time.Now().Unix()-queued.LastTry < 60 {
			remaining = append(remaining, queued)
			continue
		}

		// Attempt retry
		var result ScrobbleResult
		switch queued.Service {
		case "Last.fm":
			result = m.scrobbleLastFM(queued.Track)
		case "ListenBrainz":
			result = m.scrobbleListenBrainz(queued.Track)
		}

		if result.Success {
			log.Printf("Retry successful: %s - %s via %s", 
				queued.Track.Artist, queued.Track.Title, queued.Service)
		} else {
			// Update retry info and keep in queue
			queued.Attempts++
			queued.LastTry = time.Now().Unix()
			remaining = append(remaining, queued)
			log.Printf("Retry failed (%d/%d): %s - %s via %s: %v", 
				queued.Attempts, queued.MaxRetries, 
				queued.Track.Artist, queued.Track.Title, 
				queued.Service, result.Error)
		}
	}

	m.queuedScrobbles = remaining
}

// GetQueueStats returns statistics about the retry queue
func (m *Manager) GetQueueStats() (int, int) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	total := len(m.queuedScrobbles)
	failed := 0
	
	for _, queued := range m.queuedScrobbles {
		if queued.Attempts >= queued.MaxRetries {
			failed++
		}
	}
	
	return total, failed
}