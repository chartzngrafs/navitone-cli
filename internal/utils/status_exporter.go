package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// PlaybackStatusExport holds track & playback information exported for desktop plugins
type PlaybackStatusExport struct {
	IsPlaying       bool   `json:"is_playing"`
	Title           string `json:"title"`
	Artist          string `json:"artist"`
	Album           string `json:"album"`
	Bitrate         int    `json:"bitrate"`
	PositionSeconds int    `json:"position_seconds"`
	DurationSeconds int    `json:"duration_seconds"`
	UpdatedAt       int64  `json:"updated_at"`
}

var (
	lastExportedStatus PlaybackStatusExport
	lastExportTime     time.Time
	exportMu           sync.Mutex
)

// ExportPlaybackStatus exports current playback status to ~/.cache/navitone-cli/state.json
func ExportPlaybackStatus(status PlaybackStatusExport) error {
	exportMu.Lock()
	defer exportMu.Unlock()

	statusChanged := status.IsPlaying != lastExportedStatus.IsPlaying ||
		status.Title != lastExportedStatus.Title ||
		status.Artist != lastExportedStatus.Artist ||
		status.Album != lastExportedStatus.Album ||
		status.Bitrate != lastExportedStatus.Bitrate ||
		status.DurationSeconds != lastExportedStatus.DurationSeconds

	// Rate limit position-only updates to once per second
	if !statusChanged && time.Since(lastExportTime) < time.Second {
		return nil
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = filepath.Join(os.Getenv("HOME"), ".cache")
	}

	dir := filepath.Join(cacheDir, "navitone-cli")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	targetFile := filepath.Join(dir, "state.json")
	tmpFile := targetFile + ".tmp"

	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}

	if err := os.Rename(tmpFile, targetFile); err != nil {
		return err
	}

	lastExportedStatus = status
	lastExportTime = time.Now()
	return nil
}

// ClearPlaybackStatus resets the status file when navitone-cli exits or stops playback
func ClearPlaybackStatus() {
	ExportPlaybackStatus(PlaybackStatusExport{
		IsPlaying: false,
		UpdatedAt: time.Now().Unix(),
	})
}
