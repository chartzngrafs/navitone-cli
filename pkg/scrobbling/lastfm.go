package scrobbling

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const LastFMAPIURL = "https://ws.audioscrobbler.com/2.0/"

// LastFMClient handles submissions to Last.fm
type LastFMClient struct {
	apiKey     string
	secret     string
	username   string
	password   string
	sessionKey string
	httpClient *http.Client
}

// NewLastFMClient creates a new Last.fm client
func NewLastFMClient(apiKey, secret, username, password string) *LastFMClient {
	return &LastFMClient{
		apiKey:   apiKey,
		secret:   secret,
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetTimeout sets the HTTP client timeout
func (c *LastFMClient) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// Authenticate performs authentication with Last.fm to get a session key
func (c *LastFMClient) Authenticate(ctx context.Context) error {
	// Get auth token first
	token, err := c.getAuthToken(ctx)
	if err != nil {
		return fmt.Errorf("getting auth token: %w", err)
	}

	// Get session key using username/password authentication
	sessionKey, err := c.getMobileSession(ctx, token)
	if err != nil {
		return fmt.Errorf("getting session key: %w", err)
	}

	c.sessionKey = sessionKey
	return nil
}

// getAuthToken gets an authentication token from Last.fm
func (c *LastFMClient) getAuthToken(ctx context.Context) (string, error) {
	params := map[string]string{
		"method":  "auth.getToken",
		"api_key": c.apiKey,
	}

	resp, err := c.makeRequest(ctx, params, false)
	if err != nil {
		return "", err
	}

	var tokenResp struct {
		Token string `json:"token"`
		Error int    `json:"error,omitempty"`
		Message string `json:"message,omitempty"`
	}

	if err := json.Unmarshal(resp, &tokenResp); err != nil {
		return "", fmt.Errorf("parsing token response: %w", err)
	}

	if tokenResp.Error != 0 {
		return "", fmt.Errorf("Last.fm error %d: %s", tokenResp.Error, tokenResp.Message)
	}

	return tokenResp.Token, nil
}

// getMobileSession gets a session key using mobile authentication
func (c *LastFMClient) getMobileSession(ctx context.Context, token string) (string, error) {
	// Generate auth token hash
	authToken := fmt.Sprintf("%x", md5.Sum([]byte(c.username+fmt.Sprintf("%x", md5.Sum([]byte(c.password))))))

	params := map[string]string{
		"method":     "auth.getMobileSession",
		"api_key":    c.apiKey,
		"username":   c.username,
		"authToken":  authToken,
	}

	resp, err := c.makeRequest(ctx, params, true)
	if err != nil {
		return "", err
	}

	var sessionResp struct {
		Session struct {
			Name string `json:"name"`
			Key  string `json:"key"`
		} `json:"session"`
		Error   int    `json:"error,omitempty"`
		Message string `json:"message,omitempty"`
	}

	if err := json.Unmarshal(resp, &sessionResp); err != nil {
		return "", fmt.Errorf("parsing session response: %w", err)
	}

	if sessionResp.Error != 0 {
		return "", fmt.Errorf("Last.fm error %d: %s", sessionResp.Error, sessionResp.Message)
	}

	return sessionResp.Session.Key, nil
}

// Scrobble submits a completed track play to Last.fm
func (c *LastFMClient) Scrobble(ctx context.Context, track ScrobbleTrack) error {
	if c.sessionKey == "" {
		return fmt.Errorf("not authenticated - call Authenticate() first")
	}

	params := map[string]string{
		"method":    "track.scrobble",
		"api_key":   c.apiKey,
		"sk":        c.sessionKey,
		"artist":    track.Artist,
		"track":     track.Title,
		"timestamp": strconv.FormatInt(track.Timestamp, 10),
	}

	if track.Album != "" {
		params["album"] = track.Album
	}
	if track.Duration > 0 {
		params["duration"] = strconv.Itoa(track.Duration)
	}
	if track.TrackNumber > 0 {
		params["trackNumber"] = strconv.Itoa(track.TrackNumber)
	}
	if track.MBID != "" {
		params["mbid"] = track.MBID
	}

	_, err := c.makeRequest(ctx, params, true)
	return err
}

// UpdateNowPlaying updates the "now playing" status on Last.fm
func (c *LastFMClient) UpdateNowPlaying(ctx context.Context, track ScrobbleTrack) error {
	if c.sessionKey == "" {
		return fmt.Errorf("not authenticated - call Authenticate() first")
	}

	params := map[string]string{
		"method":  "track.updateNowPlaying",
		"api_key": c.apiKey,
		"sk":      c.sessionKey,
		"artist":  track.Artist,
		"track":   track.Title,
	}

	if track.Album != "" {
		params["album"] = track.Album
	}
	if track.Duration > 0 {
		params["duration"] = strconv.Itoa(track.Duration)
	}
	if track.TrackNumber > 0 {
		params["trackNumber"] = strconv.Itoa(track.TrackNumber)
	}
	if track.MBID != "" {
		params["mbid"] = track.MBID
	}

	_, err := c.makeRequest(ctx, params, true)
	return err
}

// GetUserInfo gets information about the authenticated user (for testing)
func (c *LastFMClient) GetUserInfo(ctx context.Context) (*UserInfo, error) {
	if c.sessionKey == "" {
		return nil, fmt.Errorf("not authenticated - call Authenticate() first")
	}

	params := map[string]string{
		"method":  "user.getInfo",
		"api_key": c.apiKey,
		"sk":      c.sessionKey,
		"user":    c.username,
	}

	resp, err := c.makeRequest(ctx, params, true)
	if err != nil {
		return nil, err
	}

	var userResp struct {
		User UserInfo `json:"user"`
		Error   int    `json:"error,omitempty"`
		Message string `json:"message,omitempty"`
	}

	if err := json.Unmarshal(resp, &userResp); err != nil {
		return nil, fmt.Errorf("parsing user info response: %w", err)
	}

	if userResp.Error != 0 {
		return nil, fmt.Errorf("Last.fm error %d: %s", userResp.Error, userResp.Message)
	}

	return &userResp.User, nil
}

// makeRequest makes an authenticated request to the Last.fm API
func (c *LastFMClient) makeRequest(ctx context.Context, params map[string]string, signed bool) ([]byte, error) {
	// Add format
	params["format"] = "json"

	// Generate API signature if required
	if signed {
		params["api_sig"] = c.generateSignature(params)
	}

	// Convert params to URL values
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", LastFMAPIURL, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "navitone-cli/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %d, body: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// generateSignature generates the API signature for authenticated requests
func (c *LastFMClient) generateSignature(params map[string]string) string {
	// Sort parameters alphabetically
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "format" && k != "callback" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// Build parameter string
	var paramStr strings.Builder
	for _, k := range keys {
		paramStr.WriteString(k)
		paramStr.WriteString(params[k])
	}

	// Append secret and hash
	paramStr.WriteString(c.secret)
	return fmt.Sprintf("%x", md5.Sum([]byte(paramStr.String())))
}