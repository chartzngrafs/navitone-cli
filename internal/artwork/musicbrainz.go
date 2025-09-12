package artwork

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"navitone-cli/internal/models"
)

// MusicBrainzClient handles MusicBrainz API requests
type MusicBrainzClient struct {
	client   *http.Client
	baseURL  string
	coverURL string
}

// NewMusicBrainzClient creates a new MusicBrainz API client
func NewMusicBrainzClient() *MusicBrainzClient {
	return &MusicBrainzClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL:  "https://musicbrainz.org/ws/2",
		coverURL: "https://coverartarchive.org",
	}
}

// MusicBrainzRelease represents a release from MusicBrainz API
type MusicBrainzRelease struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Score int    `json:"score,omitempty"`
}

// MusicBrainzSearchResponse represents search results from MusicBrainz
type MusicBrainzSearchResponse struct {
	Releases []MusicBrainzRelease `json:"releases"`
	Count    int                  `json:"count"`
}

// FindAlbumMBID searches for an album's MusicBrainz ID
func (c *MusicBrainzClient) FindAlbumMBID(album models.Album) (string, error) {
	// Build search query: artist + album name
	query := fmt.Sprintf(`artist:"%s" AND release:"%s"`, 
		strings.ReplaceAll(album.Artist, `"`, `\"`), 
		strings.ReplaceAll(album.Name, `"`, `\"`))
	
	// URL encode the query
	searchURL := fmt.Sprintf("%s/release/?query=%s&fmt=json&limit=1", 
		c.baseURL, url.QueryEscape(query))
	
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set User-Agent as required by MusicBrainz
	req.Header.Set("User-Agent", "Navitone-CLI/1.0 (https://github.com/yourusername/navitone-cli)")
	
	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to search MusicBrainz: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("MusicBrainz API error: %d", resp.StatusCode)
	}
	
	var searchResp MusicBrainzSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	
	// Return the first match if any
	if len(searchResp.Releases) > 0 {
		return searchResp.Releases[0].ID, nil
	}
	
	return "", fmt.Errorf("no releases found for %s - %s", album.Artist, album.Name)
}

// GetCoverArtURL retrieves cover art URL from Cover Art Archive
func (c *MusicBrainzClient) GetCoverArtURL(mbid string) (string, error) {
	coverURL := fmt.Sprintf("%s/release/%s/front-250", c.coverURL, mbid)
	
	// Make a HEAD request to check if cover art exists
	req, err := http.NewRequest("HEAD", coverURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to check cover art: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("cover art not available (HTTP %d)", resp.StatusCode)
	}
	
	return coverURL, nil
}

// GetAlbumCoverArt finds and returns cover art URL for an album
func (c *MusicBrainzClient) GetAlbumCoverArt(album models.Album) (string, error) {
	// First, find the MusicBrainz ID
	mbid, err := c.FindAlbumMBID(album)
	if err != nil {
		return "", fmt.Errorf("failed to find MBID: %w", err)
	}
	
	// Then get the cover art URL
	coverURL, err := c.GetCoverArtURL(mbid)
	if err != nil {
		return "", fmt.Errorf("failed to get cover art: %w", err)
	}
	
	return coverURL, nil
}