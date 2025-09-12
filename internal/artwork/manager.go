package artwork

import (
	"crypto/md5"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"navitone-cli/internal/config"
	"navitone-cli/internal/models"
)

// Manager coordinates artwork conversion, caching, and API integration
type Manager struct {
	converter        *Converter
	cache            *Cache
	config           *config.Config
	mbClient         *MusicBrainzClient
	navidromeBaseURL string // Store base URL for constructing cover art URLs
	mu               sync.RWMutex
}

// NewManager creates a new artwork manager
func NewManager(cfg *config.Config) (*Manager, error) {
	converter := NewConverter(cfg)
	
	cache, err := NewCache()
	if err != nil {
		return nil, fmt.Errorf("failed to create cache: %w", err)
	}

	return &Manager{
		converter:        converter,
		cache:            cache,
		config:           cfg,
		mbClient:         NewMusicBrainzClient(),
		navidromeBaseURL: cfg.Navidrome.ServerURL,
	}, nil
}

// GetAlbumArtwork retrieves ASCII artwork for an album
func (m *Manager) GetAlbumArtwork(album models.Album) (string, error) {
	// Check if artwork is enabled
	if !m.converter.IsEnabled() {
		return "", fmt.Errorf("artwork display is disabled")
	}

	// Use Navidrome cover art if available
	if album.CoverArt != "" {
		coverURL := m.buildNavidromeCoverArtURL(album.CoverArt)
		return m.getArtworkFromURL(coverURL, album.ID)
	}

	// Fallback to MusicBrainz
	coverURL, err := m.mbClient.GetAlbumCoverArt(album)
	if err != nil {
		return "", fmt.Errorf("no cover art available from any source: %w", err)
	}

	return m.getArtworkFromURL(coverURL, album.ID)
}


// getArtworkFromURL converts artwork from URL with caching
func (m *Manager) getArtworkFromURL(url, id string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Generate cache key
	width, height := m.converter.GetArtworkSize()
	cacheKey := GetCacheKey(url, width, height)

	// Try to get from cache first
	if cachedArt, found := m.cache.Get(cacheKey); found {
		return cachedArt, nil
	}

	// Convert from URL
	ascii, err := m.converter.ConvertFromURL(url)
	if err != nil {
		return "", fmt.Errorf("failed to convert artwork: %w", err)
	}

	// Cache the result (don't fail if caching fails)
	if err := m.cache.Set(cacheKey, ascii); err != nil {
		log.Printf("Warning: failed to cache artwork: %v", err)
	}

	return ascii, nil
}

// IsEnabled returns whether artwork display is currently enabled
func (m *Manager) IsEnabled() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.converter.IsEnabled()
}

// CleanCache removes expired cache files
func (m *Manager) CleanCache() error {
	return m.cache.Clean()
}

// UpdateConfig updates the manager's configuration
func (m *Manager) UpdateConfig(cfg *config.Config) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = cfg
	m.converter.config = cfg
	m.navidromeBaseURL = cfg.Navidrome.ServerURL
}

// buildNavidromeCoverArtURL constructs a full authenticated Navidrome getCoverArt URL
func (m *Manager) buildNavidromeCoverArtURL(coverArtID string) string {
	if coverArtID == "" || m.navidromeBaseURL == "" {
		return ""
	}

	// Check if it's already a full URL
	if strings.HasPrefix(coverArtID, "http://") || strings.HasPrefix(coverArtID, "https://") {
		return coverArtID
	}

	// Generate authentication parameters (same as Navidrome client)
	salt := fmt.Sprintf("%d", time.Now().UnixNano())
	hash := md5.Sum([]byte(m.config.Navidrome.Password + salt))
	token := fmt.Sprintf("%x", hash)

	// Build authenticated URL
	params := url.Values{}
	params.Add("u", m.config.Navidrome.Username)
	params.Add("t", token)
	params.Add("s", salt)
	params.Add("c", "navitone-cli")
	params.Add("v", "1.16.1")
	params.Add("f", "json")
	params.Add("id", coverArtID)

	baseURL := strings.TrimSuffix(m.navidromeBaseURL, "/")
	return fmt.Sprintf("%s/rest/getCoverArt?%s", baseURL, params.Encode())
}