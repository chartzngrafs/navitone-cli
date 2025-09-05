package navidrome

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client represents a Navidrome API client
type Client struct {
	baseURL    string
	username   string
	password   string
	token      string
	salt       string
	httpClient *http.Client
}

// NewClient creates a new Navidrome API client
func NewClient(serverURL, username, password string) *Client {
	// Ensure server URL has no trailing slash
	baseURL := strings.TrimSuffix(serverURL, "/")

	return &Client{
		baseURL:  baseURL,
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetTimeout sets the HTTP client timeout
func (c *Client) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// Ping tests the connection and authenticates with the server
func (c *Client) Ping(ctx context.Context) error {
	params := url.Values{}
	params.Add("u", c.username)
	params.Add("p", c.password)
	params.Add("c", "navitone-cli")
	params.Add("v", "1.16.1") // Subsonic API version
	params.Add("f", "json")

	reqURL := fmt.Sprintf("%s/rest/ping?%s", c.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return fmt.Errorf("creating ping request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ping request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ping failed with status: %d", resp.StatusCode)
	}

	var pingResp struct {
		SubsonicResponse struct {
			Status string `json:"status"`
			Error  *struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			} `json:"error,omitempty"`
		} `json:"subsonic-response"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading ping response: %w", err)
	}

	if err := json.Unmarshal(body, &pingResp); err != nil {
		return fmt.Errorf("parsing ping response: %w", err)
	}

	if pingResp.SubsonicResponse.Status != "ok" {
		if pingResp.SubsonicResponse.Error != nil {
			return fmt.Errorf("ping error: %s", pingResp.SubsonicResponse.Error.Message)
		}
		return fmt.Errorf("ping failed with status: %s", pingResp.SubsonicResponse.Status)
	}

	return nil
}

// authenticate generates authentication parameters for API requests
func (c *Client) authenticate() (url.Values, error) {
	// Generate salt
	c.salt = fmt.Sprintf("%d", time.Now().UnixNano())

	// Generate token (MD5 hash of password + salt)
	hash := md5.Sum([]byte(c.password + c.salt))
	c.token = fmt.Sprintf("%x", hash)

	params := url.Values{}
	params.Add("u", c.username)
	params.Add("t", c.token)
	params.Add("s", c.salt)
	params.Add("c", "navitone-cli")
	params.Add("v", "1.16.1")
	params.Add("f", "json")

	return params, nil
}

// makeRequest performs an authenticated API request
func (c *Client) makeRequest(ctx context.Context, endpoint string, params url.Values) (*http.Response, error) {
	authParams, err := c.authenticate()
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Merge auth params with request params
	for key, values := range params {
		for _, value := range values {
			authParams.Add(key, value)
		}
	}

	reqURL := fmt.Sprintf("%s/rest/%s?%s", c.baseURL, endpoint, authParams.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// GetAlbums retrieves albums from the server
func (c *Client) GetAlbums(ctx context.Context, limit, offset int) (*AlbumsResponse, error) {
	params := url.Values{}
	params.Add("type", "newest") // Required parameter for getAlbumList2
	if limit > 0 {
		params.Add("size", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		params.Add("offset", fmt.Sprintf("%d", offset))
	}

	resp, err := c.makeRequest(ctx, "getAlbumList2", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading albums response: %w", err)
	}

	var albumsResp AlbumsResponse
	if err := json.Unmarshal(body, &albumsResp); err != nil {
		return nil, fmt.Errorf("parsing albums response: %w", err)
	}

	if albumsResp.SubsonicResponse.Status != "ok" {
		if albumsResp.SubsonicResponse.Error != nil {
			return nil, fmt.Errorf("albums error: %s", albumsResp.SubsonicResponse.Error.Message)
		}
		return nil, fmt.Errorf("albums failed with status: %s", albumsResp.SubsonicResponse.Status)
	}

	return &albumsResp, nil
}

// GetArtists retrieves artists from the server
func (c *Client) GetArtists(ctx context.Context) (*ArtistsResponse, error) {
	params := url.Values{}

	resp, err := c.makeRequest(ctx, "getArtists", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading artists response: %w", err)
	}

	var artistsResp ArtistsResponse
	if err := json.Unmarshal(body, &artistsResp); err != nil {
		return nil, fmt.Errorf("parsing artists response: %w", err)
	}

	if artistsResp.SubsonicResponse.Status != "ok" {
		if artistsResp.SubsonicResponse.Error != nil {
			return nil, fmt.Errorf("artists error: %s", artistsResp.SubsonicResponse.Error.Message)
		}
		return nil, fmt.Errorf("artists failed with status: %s", artistsResp.SubsonicResponse.Status)
	}

	return &artistsResp, nil
}

// GetSongs retrieves songs/tracks from the server
func (c *Client) GetSongs(ctx context.Context, limit, offset int) (*SongsResponse, error) {
	params := url.Values{}
	if limit > 0 {
		params.Add("size", fmt.Sprintf("%d", limit))
	}

	resp, err := c.makeRequest(ctx, "getRandomSongs", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading songs response: %w", err)
	}

	var songsResp RandomSongsResponse
	if err := json.Unmarshal(body, &songsResp); err != nil {
		return nil, fmt.Errorf("parsing songs response: %w", err)
	}

	if songsResp.SubsonicResponse.Status != "ok" {
		if songsResp.SubsonicResponse.Error != nil {
			return nil, fmt.Errorf("songs error: %s", songsResp.SubsonicResponse.Error.Message)
		}
		return nil, fmt.Errorf("songs failed with status: %s", songsResp.SubsonicResponse.Status)
	}

	// Convert to expected format
	convertedResp := &SongsResponse{
		SubsonicResponse: struct {
			BaseResponse
			SongsByGenre SongsList `json:"songsByGenre"`
		}{
			BaseResponse: songsResp.SubsonicResponse.BaseResponse,
			SongsByGenre: songsResp.SubsonicResponse.RandomSongs,
		},
	}

	return convertedResp, nil
}

// GetAlbumTracks retrieves tracks from a specific album
func (c *Client) GetAlbumTracks(ctx context.Context, albumID string) (*SongsResponse, error) {
	params := url.Values{}
	params.Add("id", albumID)

	resp, err := c.makeRequest(ctx, "getMusicDirectory", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading album tracks response: %w", err)
	}

	var directoryResp struct {
		SubsonicResponse struct {
			BaseResponse
			Directory struct {
				ID    string `json:"id"`
				Name  string `json:"name"`
				Child []Song `json:"child"`
			} `json:"directory"`
		} `json:"subsonic-response"`
	}

	if err := json.Unmarshal(body, &directoryResp); err != nil {
		return nil, fmt.Errorf("parsing album tracks response: %w", err)
	}

	if directoryResp.SubsonicResponse.Status != "ok" {
		if directoryResp.SubsonicResponse.Error != nil {
			return nil, fmt.Errorf("album tracks error: %s", directoryResp.SubsonicResponse.Error.Message)
		}
		return nil, fmt.Errorf("album tracks failed with status: %s", directoryResp.SubsonicResponse.Status)
	}

	// Convert to expected format
	convertedResp := &SongsResponse{
		SubsonicResponse: struct {
			BaseResponse
			SongsByGenre SongsList `json:"songsByGenre"`
		}{
			BaseResponse: directoryResp.SubsonicResponse.BaseResponse,
			SongsByGenre: SongsList{Song: directoryResp.SubsonicResponse.Directory.Child},
		},
	}

	return convertedResp, nil
}

// GetArtistTracks retrieves all tracks from an artist by getting all their albums and tracks
func (c *Client) GetArtistTracks(ctx context.Context, artistID string) (*SongsResponse, error) {
	// First get all albums by the artist
	albumsResp, err := c.GetArtistAlbums(ctx, artistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get artist albums: %w", err)
	}

	var allTracks []Song

	// For each album, get its tracks
	for _, album := range albumsResp.SubsonicResponse.AlbumList2.Album {
		tracksResp, err := c.GetAlbumTracks(ctx, album.ID)
		if err != nil {
			// Log error but continue with other albums
			fmt.Printf("Warning: failed to get tracks for album %s: %v\n", album.Name, err)
			continue
		}

		allTracks = append(allTracks, tracksResp.SubsonicResponse.SongsByGenre.Song...)
	}

	// Return all tracks in the expected format
	convertedResp := &SongsResponse{
		SubsonicResponse: struct {
			BaseResponse
			SongsByGenre SongsList `json:"songsByGenre"`
		}{
			BaseResponse: BaseResponse{Status: "ok"},
			SongsByGenre: SongsList{Song: allTracks},
		},
	}

	return convertedResp, nil
}

// GetArtistAlbums retrieves albums by a specific artist
func (c *Client) GetArtistAlbums(ctx context.Context, artistID string) (*AlbumsResponse, error) {
	params := url.Values{}
	params.Add("id", artistID)

	resp, err := c.makeRequest(ctx, "getArtist", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading artist albums response: %w", err)
	}

	var artistResp struct {
		SubsonicResponse struct {
			BaseResponse
			Artist ArtistWithAlbums `json:"artist"`
		} `json:"subsonic-response"`
	}

	if err := json.Unmarshal(body, &artistResp); err != nil {
		return nil, fmt.Errorf("parsing artist albums response: %w", err)
	}

	if artistResp.SubsonicResponse.Status != "ok" {
		if artistResp.SubsonicResponse.Error != nil {
			return nil, fmt.Errorf("artist albums error: %s", artistResp.SubsonicResponse.Error.Message)
		}
		return nil, fmt.Errorf("artist albums failed with status: %s", artistResp.SubsonicResponse.Status)
	}

	// Convert to expected format
	convertedResp := &AlbumsResponse{
		SubsonicResponse: struct {
			BaseResponse
			AlbumList2 AlbumList `json:"albumList2"`
		}{
			BaseResponse: artistResp.SubsonicResponse.BaseResponse,
			AlbumList2:   AlbumList{Album: artistResp.SubsonicResponse.Artist.Album},
		},
	}

	return convertedResp, nil
}

// GetStreamURL returns the streaming URL for a song
func (c *Client) GetStreamURL(songID string) string {
	params, _ := c.authenticate()
	params.Add("id", songID)
	return fmt.Sprintf("%s/rest/stream?%s", c.baseURL, params.Encode())
}

// Scrobble submits a scrobble to the server
func (c *Client) Scrobble(ctx context.Context, songID string, submission bool) error {
	params := url.Values{}
	params.Add("id", songID)
	if submission {
		params.Add("submission", "true")
	}

	resp, err := c.makeRequest(ctx, "scrobble", params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("scrobble failed with status: %d", resp.StatusCode)
	}

	return nil
}
