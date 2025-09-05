package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"navitone-cli/internal/models"
)

// MainView handles the main application view
type MainView struct {
	state  *models.AppState
	width  int
	height int
	styles Styles
}

// Styles contains styling for the UI
type Styles struct {
	TabActive        lipgloss.Style
	TabInactive      lipgloss.Style
	Header           lipgloss.Style
	Content          lipgloss.Style
	Footer           lipgloss.Style
	Help             lipgloss.Style
	SectionTitle     lipgloss.Style
	ActiveField      lipgloss.Style
	ActiveEditField  lipgloss.Style
	InactiveField    lipgloss.Style
	ErrorMessage     lipgloss.Style
	SuccessMessage   lipgloss.Style
	HelpText         lipgloss.Style
}

// NewMainView creates a new main view
func NewMainView(state *models.AppState) *MainView {
	styles := Styles{
		TabActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("63")).
			Padding(0, 1),
		TabInactive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Background(lipgloss.Color("235")).
			Padding(0, 1),
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("57")).
			Padding(0, 1).
			Width(100),
		Content: lipgloss.NewStyle().
			Padding(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")),
		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Background(lipgloss.Color("235")).
			Padding(0, 1),
		Help: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1).
			Background(lipgloss.Color("235")),
		SectionTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("33")),
		ActiveField: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("63")),
		ActiveEditField: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("196")),
		InactiveField: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		ErrorMessage: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true),
		SuccessMessage: lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true),
		HelpText: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true),
	}

	return &MainView{
		state:  state,
		styles: styles,
	}
}

// SetSize updates the view dimensions
func (v *MainView) SetSize(width, height int) {
	v.width = width
	v.height = height
}

// Render returns the complete view string
func (v *MainView) Render() string {
	if v.width == 0 || v.height == 0 {
		return "Loading..."
	}

	var sections []string

	// Header with tabs
	sections = append(sections, v.renderHeader())

	// Main content area
	sections = append(sections, v.renderContent())

	// Footer with playback controls
	sections = append(sections, v.renderFooter())
	
	// Log area at the bottom
	sections = append(sections, v.renderLogArea())

	// Help overlay if active
	if v.state.ShowHelp {
		return v.renderHelpOverlay(strings.Join(sections, "\n"))
	}

	// Modal overlays if active
	content := strings.Join(sections, "\n")
	if v.state.ShowAlbumModal {
		return v.renderAlbumModalOverlay(content)
	}
	if v.state.ShowArtistModal {
		return v.renderArtistModalOverlay(content)
	}

	return content
}

// renderHeader creates the header with tab navigation
func (v *MainView) renderHeader() string {
	var tabs []string

	for i := models.HomeTab; i <= models.ConfigTab; i++ {
		style := v.styles.TabInactive
		if i == v.state.CurrentTab {
			style = v.styles.TabActive
		}
		tabs = append(tabs, style.Render(i.String()))
	}

	tabBar := strings.Join(tabs, "")
	return v.styles.Header.Width(v.width).Render(tabBar)
}

// renderContent creates the main content area based on current tab
func (v *MainView) renderContent() string {
	contentHeight := v.height - 6 // Account for header, footer, and log area
	content := v.styles.Content.
		Width(v.width - 2).
		Height(contentHeight)

	switch v.state.CurrentTab {
	case models.HomeTab:
		return content.Render(v.renderHomeTab())
	case models.AlbumsTab:
		return content.Render(v.renderAlbumsTab())
	case models.ArtistsTab:
		return content.Render(v.renderArtistsTab())
	case models.TracksTab:
		return content.Render(v.renderTracksTab())
	case models.PlaylistsTab:
		return content.Render(v.renderPlaylistsTab())
	case models.QueueTab:
		return content.Render(v.renderQueueTab())
	case models.ConfigTab:
		return content.Render(v.renderConfigTab())
	default:
		return content.Render("Unknown tab")
	}
}

// renderFooter creates the footer with playback controls
func (v *MainView) renderFooter() string {
	var parts []string

	if v.state.CurrentTrack != nil {
		parts = append(parts, fmt.Sprintf("♪ %s - %s", 
			v.state.CurrentTrack.Artist, v.state.CurrentTrack.Title))
	} else {
		parts = append(parts, "♪ No track playing")
	}

	if v.state.IsPlaying {
		parts = append(parts, "▶")
	} else {
		parts = append(parts, "⏸")
	}

	parts = append(parts, fmt.Sprintf("Vol: %d%%", v.state.Volume))
	parts = append(parts, fmt.Sprintf("Queue: %d", len(v.state.Queue)))

	footer := strings.Join(parts, " | ")
	return v.styles.Footer.Width(v.width).Render(footer)
}

// Tab-specific render functions
func (v *MainView) renderHomeTab() string {
	var content strings.Builder
	content.WriteString("🏠 Home\n\n")
	
	// Show queue status
	if len(v.state.Queue) > 0 {
		content.WriteString(fmt.Sprintf("🔄 Queue: %d tracks", len(v.state.Queue)))
		if v.state.CurrentTrack != nil {
			playStatus := "⏸"
			if v.state.IsPlaying {
				playStatus = "▶"
			}
			content.WriteString(fmt.Sprintf(" | %s %s - %s", 
				playStatus, v.state.CurrentTrack.Artist, v.state.CurrentTrack.Title))
		}
		content.WriteString("\n\n")
	}
	
	// Library stats
	content.WriteString("📊 Library Overview\n")
	content.WriteString(fmt.Sprintf("Albums: %d | Artists: %d | Tracks: %d\n\n", 
		len(v.state.Albums), len(v.state.Artists), len(v.state.Tracks)))
	
	// Recently added albums (show first few)
	content.WriteString("💿 Recently Added Albums\n")
	if len(v.state.Albums) == 0 {
		content.WriteString("Load albums data by visiting the Albums tab\n\n")
	} else {
		maxShow := 5
		if len(v.state.Albums) < maxShow {
			maxShow = len(v.state.Albums)
		}
		for i := 0; i < maxShow; i++ {
			album := v.state.Albums[i]
			yearStr := ""
			if album.Year > 0 {
				yearStr = fmt.Sprintf(" (%d)", album.Year)
			}
			content.WriteString(fmt.Sprintf("  • %s - %s%s\n", album.Artist, album.Name, yearStr))
		}
		if len(v.state.Albums) > maxShow {
			content.WriteString(fmt.Sprintf("  ... and %d more (visit Albums tab)\n", len(v.state.Albums)-maxShow))
		}
		content.WriteString("\n")
	}
	
	// Top artists by album count
	content.WriteString("🎤 Top Artists\n")
	if len(v.state.Artists) == 0 {
		content.WriteString("Load artists data by visiting the Artists tab\n\n")
	} else {
		// Sort artists by album count (simple bubble sort for top 5)
		topArtists := make([]models.Artist, len(v.state.Artists))
		copy(topArtists, v.state.Artists)
		
		// Simple sort by album count (descending)
		for i := 0; i < len(topArtists)-1; i++ {
			for j := 0; j < len(topArtists)-i-1; j++ {
				if topArtists[j].AlbumCount < topArtists[j+1].AlbumCount {
					topArtists[j], topArtists[j+1] = topArtists[j+1], topArtists[j]
				}
			}
		}
		
		maxShow := 5
		if len(topArtists) < maxShow {
			maxShow = len(topArtists)
		}
		for i := 0; i < maxShow; i++ {
			artist := topArtists[i]
			star := ""
			if artist.StarredAt != nil {
				star = "★ "
			}
			albumText := "album"
			if artist.AlbumCount != 1 {
				albumText = "albums"
			}
			content.WriteString(fmt.Sprintf("  • %s%s (%d %s)\n", 
				star, artist.Name, artist.AlbumCount, albumText))
		}
		if len(v.state.Artists) > maxShow {
			content.WriteString(fmt.Sprintf("  ... and %d more (visit Artists tab)\n", len(v.state.Artists)-maxShow))
		}
		content.WriteString("\n")
	}
	
	// Navigation hint
	content.WriteString("💡 Navigate with Tab/Shift+Tab • F1/? for help")
	
	return content.String()
}

func (v *MainView) renderAlbumsTab() string {
	if v.state.LoadingAlbums {
		return "💿 Albums\n\nLoading albums..."
	}
	
	if v.state.LoadingError != "" {
		return fmt.Sprintf("💿 Albums\n\n❌ Error: %s\n\nPress 'r' to retry", v.state.LoadingError)
	}
	
	if len(v.state.Albums) == 0 {
		return "💿 Albums\n\nNo albums found.\n\nMake sure your Navidrome server is configured in the Config tab."
	}
	
	var content strings.Builder
	content.WriteString("💿 Albums\n\n")
	
	// Show instructions
	content.WriteString("↑↓ Navigate • Enter to view tracks • Alt+Enter/A to queue album • R to refresh\n\n")
	
	// Render album list
	startIdx := 0
	endIdx := len(v.state.Albums)
	
	// Limit visible items (simple pagination)
	maxVisible := 20
	if endIdx > maxVisible {
		if v.state.SelectedAlbumIndex >= maxVisible {
			startIdx = v.state.SelectedAlbumIndex - maxVisible + 1
		}
		endIdx = startIdx + maxVisible
		if endIdx > len(v.state.Albums) {
			endIdx = len(v.state.Albums)
		}
	}
	
	for i := startIdx; i < endIdx; i++ {
		album := v.state.Albums[i]
		line := v.formatAlbumLine(album, i == v.state.SelectedAlbumIndex)
		content.WriteString(line)
		content.WriteString("\n")
	}
	
	// Show pagination info if needed
	if len(v.state.Albums) > maxVisible {
		content.WriteString(fmt.Sprintf("\nShowing %d-%d of %d albums", 
			startIdx+1, endIdx, len(v.state.Albums)))
	}
	
	return content.String()
}

func (v *MainView) formatAlbumLine(album models.Album, selected bool) string {
	// Format: [YEAR] Artist - Album Name (Tracks)
	yearStr := ""
	if album.Year > 0 {
		yearStr = fmt.Sprintf("[%d] ", album.Year)
	}
	
	line := fmt.Sprintf("%s%s - %s (%d tracks)", 
		yearStr, album.Artist, album.Name, album.TrackCount)
	
	if selected {
		return v.styles.ActiveField.Render("> " + line)
	}
	
	return "  " + line
}

func (v *MainView) renderArtistsTab() string {
	if v.state.LoadingArtists {
		return "🎤 Artists\n\nLoading artists..."
	}
	
	if v.state.LoadingError != "" {
		return fmt.Sprintf("🎤 Artists\n\n❌ Error: %s\n\nPress 'r' to retry", v.state.LoadingError)
	}
	
	if len(v.state.Artists) == 0 {
		return "🎤 Artists\n\nNo artists found.\n\nMake sure your Navidrome server is configured in the Config tab."
	}
	
	var content strings.Builder
	content.WriteString("🎤 Artists\n\n")
	
	// Show instructions
	content.WriteString("↑↓ Navigate • Enter to view albums • R to refresh\n\n")
	
	// Render artist list
	startIdx := 0
	endIdx := len(v.state.Artists)
	
	// Limit visible items (simple pagination)
	maxVisible := 20
	if endIdx > maxVisible {
		if v.state.SelectedArtistIndex >= maxVisible {
			startIdx = v.state.SelectedArtistIndex - maxVisible + 1
		}
		endIdx = startIdx + maxVisible
		if endIdx > len(v.state.Artists) {
			endIdx = len(v.state.Artists)
		}
	}
	
	for i := startIdx; i < endIdx; i++ {
		artist := v.state.Artists[i]
		line := v.formatArtistLine(artist, i == v.state.SelectedArtistIndex)
		content.WriteString(line)
		content.WriteString("\n")
	}
	
	// Show pagination info if needed
	if len(v.state.Artists) > maxVisible {
		content.WriteString(fmt.Sprintf("\nShowing %d-%d of %d artists", 
			startIdx+1, endIdx, len(v.state.Artists)))
	}
	
	return content.String()
}

func (v *MainView) formatArtistLine(artist models.Artist, selected bool) string {
	// Format: Artist Name (X albums)
	albumText := "album"
	if artist.AlbumCount != 1 {
		albumText = "albums"
	}
	
	line := fmt.Sprintf("%s (%d %s)", artist.Name, artist.AlbumCount, albumText)
	
	// Add star indicator if starred
	if artist.StarredAt != nil {
		line = "★ " + line
	}
	
	if selected {
		return v.styles.ActiveField.Render("> " + line)
	}
	
	return "  " + line
}

func (v *MainView) renderTracksTab() string {
	if v.state.LoadingTracks {
		return "🎵 Tracks\n\nLoading tracks..."
	}
	
	if v.state.LoadingError != "" {
		return fmt.Sprintf("🎵 Tracks\n\n❌ Error: %s\n\nPress 'r' to retry", v.state.LoadingError)
	}
	
	if len(v.state.Tracks) == 0 {
		return "🎵 Tracks\n\nNo tracks found.\n\nMake sure your Navidrome server is configured in the Config tab."
	}
	
	var content strings.Builder
	content.WriteString("🎵 Tracks\n\n")
	
	// Show instructions
	content.WriteString("↑↓ Navigate • Enter to add to queue • R to refresh\n\n")
	
	// Render track list
	startIdx := 0
	endIdx := len(v.state.Tracks)
	
	// Limit visible items (simple pagination)
	maxVisible := 15 // Fewer tracks per page since they take more space
	if endIdx > maxVisible {
		if v.state.SelectedTrackIndex >= maxVisible {
			startIdx = v.state.SelectedTrackIndex - maxVisible + 1
		}
		endIdx = startIdx + maxVisible
		if endIdx > len(v.state.Tracks) {
			endIdx = len(v.state.Tracks)
		}
	}
	
	for i := startIdx; i < endIdx; i++ {
		track := v.state.Tracks[i]
		line := v.formatTrackLine(track, i == v.state.SelectedTrackIndex)
		content.WriteString(line)
		content.WriteString("\n")
	}
	
	// Show pagination info if needed
	if len(v.state.Tracks) > maxVisible {
		content.WriteString(fmt.Sprintf("\nShowing %d-%d of %d tracks", 
			startIdx+1, endIdx, len(v.state.Tracks)))
	}
	
	return content.String()
}

func (v *MainView) formatTrackLine(track models.Track, selected bool) string {
	// Format: Track# Artist - Title (Album) [Duration]
	trackNum := ""
	if track.Track > 0 {
		trackNum = fmt.Sprintf("%02d. ", track.Track)
	}
	
	// Format duration (seconds to mm:ss)
	duration := ""
	if track.Duration > 0 {
		minutes := track.Duration / 60
		seconds := track.Duration % 60
		duration = fmt.Sprintf(" [%d:%02d]", minutes, seconds)
	}
	
	line := fmt.Sprintf("%s%s - %s (%s)%s", 
		trackNum, track.Artist, track.Title, track.Album, duration)
	
	if selected {
		return v.styles.ActiveField.Render("> " + line)
	}
	
	return "  " + line
}

func (v *MainView) renderPlaylistsTab() string {
	return "📋 Playlists Tab\n\n(Coming soon)"
}

func (v *MainView) renderQueueTab() string {
	var content strings.Builder
	content.WriteString("🔄 Queue\n\n")
	
	if len(v.state.Queue) == 0 {
		content.WriteString("Queue is empty.\n\n")
		content.WriteString("Add tracks by navigating to Albums, Artists, or Tracks tabs and pressing Enter.")
		return content.String()
	}
	
	// Show instructions
	content.WriteString(fmt.Sprintf("↑↓ Navigate • Enter/Space to play • X/Del to remove • C to clear all (%d tracks)\n\n", len(v.state.Queue)))
	
	// Show current playing track if any
	if v.state.CurrentTrack != nil {
		playStatus := "⏸"
		if v.state.IsPlaying {
			playStatus = "▶"
		}
		content.WriteString(fmt.Sprintf("Now Playing: %s %s - %s\n\n", 
			playStatus, v.state.CurrentTrack.Artist, v.state.CurrentTrack.Title))
	}
	
	// Render queue list
	for i, track := range v.state.Queue {
		line := v.formatQueueLine(track, i, i == v.state.SelectedQueueIndex)
		content.WriteString(line)
		content.WriteString("\n")
	}
	
	return content.String()
}

func (v *MainView) formatQueueLine(track models.Track, index int, selected bool) string {
	// Format: #. Artist - Title (Album) [Duration]
	queueNum := fmt.Sprintf("%2d. ", index+1)
	
	// Format duration (seconds to mm:ss)
	duration := ""
	if track.Duration > 0 {
		minutes := track.Duration / 60
		seconds := track.Duration % 60
		duration = fmt.Sprintf(" [%d:%02d]", minutes, seconds)
	}
	
	line := fmt.Sprintf("%s%s - %s (%s)%s", 
		queueNum, track.Artist, track.Title, track.Album, duration)
	
	// Indicate if this is the currently playing track
	if v.state.CurrentTrack != nil && track.ID == v.state.CurrentTrack.ID {
		if v.state.IsPlaying {
			line = "▶ " + line[2:] // Replace queue number with play indicator
		} else {
			line = "⏸ " + line[2:] // Replace queue number with pause indicator
		}
	}
	
	if selected {
		return v.styles.ActiveField.Render("> " + line)
	}
	
	return "  " + line
}

func (v *MainView) renderConfigTab() string {
	cf := v.state.ConfigForm
	if cf == nil {
		return "Configuration not loaded"
	}

	var sections []string

	// Header
	sections = append(sections, "⚙️  Configuration")
	sections = append(sections, "")

	// Navidrome section
	sections = append(sections, v.renderConfigSection("Navidrome Server Settings", []models.ConfigFormField{
		models.ServerURLField,
		models.UsernameField,
		models.PasswordField,
	}, cf))

	sections = append(sections, "")

	// Scrobbling section
	sections = append(sections, v.renderConfigSection("Scrobbling Settings", []models.ConfigFormField{
		models.LastFMEnabledField,
		models.LastFMUsernameField,
		models.LastFMPasswordField,
		models.ListenBrainzEnabledField,
		models.ListenBrainzTokenField,
	}, cf))

	sections = append(sections, "")

	// Audio section
	sections = append(sections, v.renderConfigSection("Audio Settings", []models.ConfigFormField{
		models.VolumeField,
		models.AudioDeviceField,
		models.BufferSizeField,
	}, cf))

	sections = append(sections, "")

	// Status messages
	if cf.ValidationError != "" {
		sections = append(sections, v.styles.ErrorMessage.Render("❌ "+cf.ValidationError))
		sections = append(sections, "")
	}

	if cf.ConnectionStatus != "" {
		style := v.styles.SuccessMessage
		if strings.Contains(cf.ConnectionStatus, "❌") {
			style = v.styles.ErrorMessage
		}
		sections = append(sections, style.Render(cf.ConnectionStatus))
		sections = append(sections, "")
	}

	// Help text
	if cf.EditMode {
		sections = append(sections, v.styles.HelpText.Render("Enter to save • Esc to cancel"))
	} else {
		sections = append(sections, v.styles.HelpText.Render("↑↓ Navigate • Enter to edit • F2 Save • F3 Test connection"))
	}

	return strings.Join(sections, "\n")
}

// renderConfigSection renders a section of configuration fields
func (v *MainView) renderConfigSection(title string, fields []models.ConfigFormField, cf *models.ConfigFormState) string {
	var lines []string
	
	// Section title
	lines = append(lines, v.styles.SectionTitle.Render(title))
	
	// Top border
	lines = append(lines, "┌" + strings.Repeat("─", 45) + "┐")
	
	// Fields
	for _, field := range fields {
		lines = append(lines, v.renderConfigField(field, cf))
	}
	
	// Bottom border
	lines = append(lines, "└" + strings.Repeat("─", 45) + "┘")
	
	return strings.Join(lines, "\n")
}

// renderConfigField renders a single configuration field
func (v *MainView) renderConfigField(field models.ConfigFormField, cf *models.ConfigFormState) string {
	isActive := cf.ActiveField == field
	label := cf.GetFieldLabel(field)
	
	var line string
	
	if cf.IsCheckboxField(field) {
		// Checkbox field
		checked := cf.GetCheckboxValue(field)
		checkbox := "□"
		if checked {
			checkbox = "☑"
		}
		
		line = fmt.Sprintf("│ %s %s", checkbox, label)
	} else {
		// Text input field
		value := cf.GetFieldValue(field)
		if cf.EditMode && isActive {
			value = cf.CurrentInput
		}
		
		// Pad the value to fit in the field
		fieldWidth := 45 - len(label) - 6
		if len(value) > fieldWidth {
			value = value[:fieldWidth-3] + "..."
		}
		
		line = fmt.Sprintf("│ %s: [%-*s] │", label, fieldWidth, value)
	}
	
	// Highlight active field
	if isActive {
		if cf.EditMode {
			return v.styles.ActiveEditField.Render(line)
		} else {
			return v.styles.ActiveField.Render(line)
		}
	}
	
	// Complete the checkbox line format
	if cf.IsCheckboxField(field) {
		padding := 45 - len(line) + 1
		line += strings.Repeat(" ", padding) + "│"
	}
	
	return v.styles.InactiveField.Render(line)
}

// renderHelpOverlay creates the context-aware help overlay
func (v *MainView) renderHelpOverlay(background string) string {
	helpContent := v.getContextualHelp()
	
	help := v.styles.Help.
		Width(60).
		Height(20).
		Render(helpContent)
	
	// Center the help overlay
	x := (v.width - 60) / 2
	y := (v.height - 20) / 2
	
	// Simple overlay positioning (could be enhanced)
	lines := strings.Split(background, "\n")
	helpLines := strings.Split(help, "\n")
	
	for i, helpLine := range helpLines {
		if y+i < len(lines) && y+i >= 0 {
			line := lines[y+i]
			if x >= 0 && x+len(helpLine) <= len(line) {
				lines[y+i] = line[:x] + helpLine + line[x+len(helpLine):]
			}
		}
	}
	
	return strings.Join(lines, "\n")
}

// getContextualHelp returns help content based on current context
func (v *MainView) getContextualHelp() string {
	help := "NAVITONE-CLI HELP\n\n"
	
	// Check if we're in a modal first
	if v.state.ShowAlbumModal {
		help += "Album Tracks Modal:\n"
		help += "↑↓ / j/k       - Navigate tracks\n"
		help += "Enter          - Play track & queue remainder\n"
		help += "A              - Add all tracks to queue\n"
		help += "Esc / q        - Close modal\n\n"
	} else if v.state.ShowArtistModal {
		help += "Artist Albums Modal:\n"
		help += "↑↓ / j/k       - Navigate albums\n"
		help += "Enter          - View album tracks (modal)\n"
		help += "A / Alt+Enter  - Add all albums to queue\n"
		help += "Esc / q        - Close modal\n\n"
	}
	
	// Global shortcuts (always shown)
	help += "Global Navigation:\n"
	help += "Tab/Shift+Tab  - Switch tabs\n"
	help += "F1 / ?         - Toggle this help\n"
	help += "Ctrl+C / q     - Quit application\n\n"
	
	// Global playback controls (always shown)
	help += "Playback Controls:\n"
	help += "Ctrl+P         - Play/Pause\n"
	help += "Ctrl+N         - Next track\n"
	help += "Ctrl+B         - Previous track\n"
	help += "Ctrl+S         - Stop\n\n"
	
	// Tab-specific shortcuts
	switch v.state.CurrentTab {
	case models.HomeTab:
		help += "Home Tab:\n"
		help += "View library overview and recently added content\n"
	case models.AlbumsTab:
		help += "Albums Tab:\n"
		help += "↑↓ / j/k       - Navigate albums\n"
		help += "Enter          - View album tracks (modal)\n"
		help += "Alt+Enter/A    - Queue entire album\n"
		help += "R              - Refresh albums list\n"
	case models.ArtistsTab:
		help += "Artists Tab:\n"
		help += "↑↓ / j/k       - Navigate artists\n"
		help += "Enter          - View artist albums (modal)\n"
		help += "R              - Refresh artists list\n"
	case models.TracksTab:
		help += "Tracks Tab:\n"
		help += "↑↓ / j/k       - Navigate tracks\n"
		help += "Enter          - Add track to queue\n"
		help += "R              - Refresh tracks list\n"
	case models.QueueTab:
		help += "Queue Tab:\n"
		help += "↑↓ / j/k       - Navigate queue\n"
		help += "Enter/Space    - Play selected track\n"
		help += "Del / X        - Remove selected track\n"
		help += "C              - Clear entire queue\n"
	case models.ConfigTab:
		help += "Config Tab:\n"
		help += "↑↓ / j/k       - Navigate fields\n"
		help += "Enter          - Edit field / toggle checkbox\n"
		help += "F2             - Save configuration\n"
		help += "F3             - Test Navidrome connection\n"
	case models.PlaylistsTab:
		help += "Playlists Tab:\n"
		help += "(Coming soon - playlist management)\n"
	}
	
	help += "\nPress F1 or ? to close help"
	return help
}

// renderLogArea creates the log area at the bottom showing recent events
func (v *MainView) renderLogArea() string {
	logStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Background(lipgloss.Color("235")).
		Padding(0, 1).
		Width(v.width)

	if len(v.state.LogMessages) == 0 {
		// Show empty log area to maintain consistent layout
		return logStyle.Render("")
	}

	// Show up to 2 most recent log messages
	var logLines []string
	for _, msg := range v.state.LogMessages {
		logLines = append(logLines, msg)
	}
	
	// Pad to always show 2 lines for consistent layout
	for len(logLines) < 2 {
		logLines = append(logLines, "")
	}

	logContent := strings.Join(logLines, "\n")
	return logStyle.Render(logContent)
}

// renderAlbumModalOverlay renders the album tracks modal overlay
func (v *MainView) renderAlbumModalOverlay(background string) string {
	if v.state.SelectedAlbum == nil {
		return background
	}

	var content strings.Builder
	
	// Modal header
	content.WriteString(fmt.Sprintf("🎵 %s - %s (%d)\n\n", 
		v.state.SelectedAlbum.Artist, v.state.SelectedAlbum.Name, v.state.SelectedAlbum.Year))

	if v.state.LoadingModalContent {
		content.WriteString("Loading tracks...")
	} else if len(v.state.AlbumTracks) == 0 {
		content.WriteString("No tracks found.")
	} else {
		// Instructions
		content.WriteString("↑↓ Navigate • Enter to play & queue remainder • A to add all • Esc to close\n\n")
		
		// Track list
		for i, track := range v.state.AlbumTracks {
			line := v.formatModalTrackLine(track, i, i == v.state.SelectedModalIndex)
			content.WriteString(line)
			content.WriteString("\n")
		}
	}

	// Center the modal overlay (styling is applied in overlayModal)
	return v.overlayModal(background, content.String(), 80, 25)
}

// renderArtistModalOverlay renders the artist albums modal overlay
func (v *MainView) renderArtistModalOverlay(background string) string {
	if v.state.SelectedArtist == nil {
		return background
	}

	var content strings.Builder
	
	// Modal header
	albumText := "album"
	if v.state.SelectedArtist.AlbumCount != 1 {
		albumText = "albums"
	}
	content.WriteString(fmt.Sprintf("🎤 %s (%d %s)\n\n", 
		v.state.SelectedArtist.Name, v.state.SelectedArtist.AlbumCount, albumText))

	if v.state.LoadingModalContent {
		content.WriteString("Loading albums...")
	} else if len(v.state.ArtistAlbums) == 0 {
		content.WriteString("No albums found.")
	} else {
		// Instructions
		content.WriteString("↑↓ Navigate • Enter to view tracks • A/Alt+Enter to queue all • Esc to close\n\n")
		
		// Album list
		for i, album := range v.state.ArtistAlbums {
			line := v.formatModalAlbumLine(album, i == v.state.SelectedModalIndex)
			content.WriteString(line)
			content.WriteString("\n")
		}
	}

	// Center the modal overlay (styling is applied in overlayModal)
	return v.overlayModal(background, content.String(), 80, 25)
}

// formatModalTrackLine formats a track line for modal display
func (v *MainView) formatModalTrackLine(track models.Track, index int, selected bool) string {
	// Format: Track# Title [Duration]
	trackNum := ""
	if track.Track > 0 {
		trackNum = fmt.Sprintf("%02d. ", track.Track)
	} else {
		trackNum = fmt.Sprintf("%2d. ", index+1)
	}
	
	// Format duration (seconds to mm:ss)
	duration := ""
	if track.Duration > 0 {
		minutes := track.Duration / 60
		seconds := track.Duration % 60
		duration = fmt.Sprintf(" [%d:%02d]", minutes, seconds)
	}
	
	line := fmt.Sprintf("%s%s%s", trackNum, track.Title, duration)
	
	if selected {
		return v.styles.ActiveField.Render("> " + line)
	}
	
	return "  " + line
}

// formatModalAlbumLine formats an album line for modal display  
func (v *MainView) formatModalAlbumLine(album models.Album, selected bool) string {
	// Format: [Year] Album Name (Tracks)
	yearStr := ""
	if album.Year > 0 {
		yearStr = fmt.Sprintf("[%d] ", album.Year)
	}
	
	line := fmt.Sprintf("%s%s (%d tracks)", yearStr, album.Name, album.TrackCount)
	
	if selected {
		return v.styles.ActiveField.Render("> " + line)
	}
	
	return "  " + line
}

// overlayModal overlays a modal on the background content using lipgloss positioning
func (v *MainView) overlayModal(_ /* background */, modal string, modalWidth, modalHeight int) string {
	// Use lipgloss to properly position the modal
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Background(lipgloss.Color("235")).
		Padding(1).
		Width(modalWidth-4). // Account for border and padding
		Height(modalHeight-4).
		Align(lipgloss.Center, lipgloss.Center)
	
	// Position the modal in the center of the available space
	positionedModal := lipgloss.Place(
		v.width, v.height,
		lipgloss.Center, lipgloss.Center,
		modalStyle.Render(modal),
	)
	
	return positionedModal
}