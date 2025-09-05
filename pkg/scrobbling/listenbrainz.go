package scrobbling

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const ListenBrainzAPIURL = "https://api.listenbrainz.org"

// ListenBrainzClient handles submissions to ListenBrainz
type ListenBrainzClient struct {
	token      string
	httpClient *http.Client
}

// NewListenBrainzClient creates a new ListenBrainz client
func NewListenBrainzClient(token string) *ListenBrainzClient {
	return &ListenBrainzClient{
		token: token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetTimeout sets the HTTP client timeout
func (c *ListenBrainzClient) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// Listen represents a single listening event
type Listen struct {
	ListenedAt    int                    `json:"listened_at"`
	TrackMetadata TrackMetadata          `json:"track_metadata"`
	RecordingMSID string                 `json:"recording_msid,omitempty"`
	UserName      string                 `json:"user_name,omitempty"`
}

// TrackMetadata contains metadata about the track
type TrackMetadata struct {
	ArtistName           string                 `json:"artist_name"`
	TrackName            string                 `json:"track_name"`
	ReleaseName          string                 `json:"release_name,omitempty"`
	AdditionalInfo       map[string]interface{} `json:"additional_info,omitempty"`
}

// ListenPayload represents the payload for listen submissions
type ListenPayload struct {
	ListenType string   `json:"listen_type"`
	Listens    []Listen `json:"listens"`
}

// SubmitListen submits a single listen to ListenBrainz
func (c *ListenBrainzClient) SubmitListen(ctx context.Context, listen Listen) error {
	payload := ListenPayload{
		ListenType: "single",
		Listens:    []Listen{listen},
	}

	return c.submitPayload(ctx, "/1/submit-listens", payload)
}

// SubmitPlayingNow submits a "playing now" notification
func (c *ListenBrainzClient) SubmitPlayingNow(ctx context.Context, metadata TrackMetadata) error {
	listen := Listen{
		TrackMetadata: metadata,
	}
	
	payload := ListenPayload{
		ListenType: "playing_now",
		Listens:    []Listen{listen},
	}

	return c.submitPayload(ctx, "/1/submit-listens", payload)
}

// SubmitListens submits multiple listens at once
func (c *ListenBrainzClient) SubmitListens(ctx context.Context, listens []Listen) error {
	payload := ListenPayload{
		ListenType: "import",
		Listens:    listens,
	}

	return c.submitPayload(ctx, "/1/submit-listens", payload)
}

// submitPayload handles the actual HTTP submission
func (c *ListenBrainzClient) submitPayload(ctx context.Context, endpoint string, payload ListenPayload) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", ListenBrainzAPIURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Token "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "navitone-cli/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("submission request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("submission failed with status: %d", resp.StatusCode)
	}

	return nil
}

// ValidateToken validates the ListenBrainz token
func (c *ListenBrainzClient) ValidateToken(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", ListenBrainzAPIURL+"/1/validate-token", nil)
	if err != nil {
		return fmt.Errorf("creating validation request: %w", err)
	}

	req.Header.Set("Authorization", "Token "+c.token)
	req.Header.Set("User-Agent", "navitone-cli/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var result struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Valid   bool   `json:"valid"`
		}
		
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("parsing validation response: %w", err)
		}
		
		if !result.Valid {
			return fmt.Errorf("token is invalid: %s", result.Message)
		}
		
		return nil
	}

	return fmt.Errorf("token validation failed with status: %d", resp.StatusCode)
}

// GetUserListens retrieves recent listens for the user (for testing/verification)
func (c *ListenBrainzClient) GetUserListens(ctx context.Context, username string, count int) ([]Listen, error) {
	url := fmt.Sprintf("%s/1/user/%s/listens?count=%d", ListenBrainzAPIURL, username, count)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "navitone-cli/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Payload struct {
			Count   int      `json:"count"`
			Listens []Listen `json:"listens"`
		} `json:"payload"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return result.Payload.Listens, nil
}