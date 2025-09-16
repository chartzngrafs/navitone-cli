package views

import (
	"fmt"
	"strings"
)

// ProgressBarConfig contains configuration for rendering progress bars
type ProgressBarConfig struct {
	Width      int     // Total width of the progress bar
	Progress   float64 // Progress as a percentage (0.0 to 1.0)
	ShowText   bool    // Whether to show percentage text
	Character  rune    // Character to use for the filled portion
}

// DefaultProgressBarConfig returns sensible defaults
func DefaultProgressBarConfig() ProgressBarConfig {
	return ProgressBarConfig{
		Width:     30,
		Progress:  0.0,
		ShowText:  true,
		Character: '█',
	}
}

// RenderProgressBar creates a themed progress bar
func (s *ThemedStyles) RenderProgressBar(config ProgressBarConfig) string {
	if config.Width <= 0 {
		return ""
	}

	// Calculate filled and empty portions
	filled := int(float64(config.Width) * config.Progress)
	if filled > config.Width {
		filled = config.Width
	}

	// Build the progress bar
	bar := strings.Repeat(string(config.Character), filled)
	empty := strings.Repeat("░", config.Width-filled)

	// Apply styling
	styledFill := s.ProgressFill.Render(bar)
	styledEmpty := s.ProgressBar.Render(empty)
	progressBar := styledFill + styledEmpty

	// Add percentage text if requested
	if config.ShowText {
		percentage := fmt.Sprintf(" %3.0f%%", config.Progress*100)
		progressBar += s.ProgressBar.Render(percentage)
	}

	return progressBar
}

// RenderVolumeIndicator creates a themed volume indicator
func (s *ThemedStyles) RenderVolumeIndicator(volume int) string {
	// Volume should be 0-100
	if volume < 0 {
		volume = 0
	}
	if volume > 100 {
		volume = 100
	}

	config := DefaultProgressBarConfig()
	config.Width = 20
	config.Progress = float64(volume) / 100.0
	config.Character = '▌'

	volumeBar := s.RenderProgressBar(config)

	// Add volume icon based on level
	var icon string
	switch {
	case volume == 0:
		icon = "🔇"
	case volume < 30:
		icon = "🔈"
	case volume < 70:
		icon = "🔉"
	default:
		icon = "🔊"
	}

	return s.VolumeIndicator.Render(fmt.Sprintf("%s %s", icon, volumeBar))
}

// StatusIndicator represents different status states
type StatusIndicator int

const (
	StatusConnected StatusIndicator = iota
	StatusDisconnected
	StatusLoading
	StatusError
	StatusPlaying
	StatusPaused
	StatusStopped
)

// RenderStatusIndicator creates a themed status indicator
func (s *ThemedStyles) RenderStatusIndicator(status StatusIndicator, text string) string {
	var icon string
	var style = s.InfoMessage // default

	switch status {
	case StatusConnected:
		icon = "●"
		style = s.SuccessMessage
	case StatusDisconnected:
		icon = "●"
		style = s.ErrorMessage
	case StatusLoading:
		icon = "⟳"
		style = s.WarningMessage
	case StatusError:
		icon = "✗"
		style = s.ErrorMessage
	case StatusPlaying:
		icon = "▶"
		style = s.PlayingIndicator
	case StatusPaused:
		icon = "⏸"
		style = s.PausedIndicator
	case StatusStopped:
		icon = "⏹"
		style = s.InfoMessage
	default:
		icon = "?"
	}

	return style.Render(fmt.Sprintf("%s %s", icon, text))
}

// RenderConnectionStatus creates a connection status indicator
func (s *ThemedStyles) RenderConnectionStatus(connected bool, serverURL string) string {
	if connected {
		return s.RenderStatusIndicator(StatusConnected, fmt.Sprintf("Connected to %s", serverURL))
	}
	return s.RenderStatusIndicator(StatusDisconnected, "Disconnected")
}

// PlaybackState represents the current playback state
type PlaybackState struct {
	IsPlaying bool
	IsPaused  bool
	Track     string
	Position  float64 // 0.0 to 1.0
	Volume    int     // 0 to 100
}

// RenderPlaybackStatus creates a comprehensive playback status display
func (s *ThemedStyles) RenderPlaybackStatus(state PlaybackState) string {
	var status StatusIndicator
	var statusText string

	switch {
	case state.IsPlaying:
		status = StatusPlaying
		statusText = "Playing"
	case state.IsPaused:
		status = StatusPaused
		statusText = "Paused"
	default:
		status = StatusStopped
		statusText = "Stopped"
	}

	// Build status line
	statusLine := s.RenderStatusIndicator(status, statusText)

	if state.Track != "" {
		statusLine += " " + s.CurrentTrack.Render(state.Track)
	}

	// Add progress bar if playing or paused
	if state.IsPlaying || state.IsPaused {
		config := DefaultProgressBarConfig()
		config.Width = 25
		config.Progress = state.Position
		config.ShowText = false
		config.Character = '━'

		progressBar := s.RenderProgressBar(config)
		statusLine += "\n" + progressBar
	}

	// Add volume indicator
	volumeIndicator := s.RenderVolumeIndicator(state.Volume)
	statusLine += " " + volumeIndicator

	return statusLine
}

// CategoryBadge creates a colored badge for content categories
func (s *ThemedStyles) CategoryBadge(category string, text string) string {
	var style = s.TrackStyle // default

	switch strings.ToLower(category) {
	case "album", "albums":
		style = s.AlbumStyle
	case "artist", "artists":
		style = s.ArtistStyle
	case "playlist", "playlists":
		style = s.PlaylistStyle
	case "track", "tracks":
		style = s.TrackStyle
	}

	return style.Render(fmt.Sprintf("[%s] %s", strings.ToUpper(category), text))
}