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
	theme  Theme
	styles ThemedStyles
}


// NewMainView creates a new main view
func NewMainView(state *models.AppState, themeVariant string) *MainView {
	theme := NewTheme(themeVariant)
	styles := NewThemedStyles(theme)

	return &MainView{
		state:  state,
		theme:  theme,
		styles: styles,
		width:  80,  // Default width
		height: 24,  // Default height
	}
}

// SetSize updates the view dimensions
func (v *MainView) SetSize(width, height int) {
	// Debug logging to track size changes
	if width == 0 || height == 0 {
		// Ignore invalid dimensions completely
		return
	}
	v.width = width
	v.height = height
}

// Render returns the complete view string
func (v *MainView) Render() string {
	// Ensure we always have valid dimensions
	if v.width <= 0 {
		v.width = 80
	}
	if v.height <= 0 {
		v.height = 24
	}
	

	var sections []string

	// Header with tabs - always render header first
	header := v.renderHeader()
	
	
	sections = append(sections, header)

	// Main content area
	sections = append(sections, v.renderContent())

	// Footer with playback controls
	sections = append(sections, v.renderFooter())
	
	// Log area at the bottom
	sections = append(sections, v.renderLogArea())


	// Modal overlays if active
	content := strings.Join(sections, "\n")
	
	
	if v.state.ShowAlbumModal {
		return v.renderAlbumModalOverlay(content)
	}
	if v.state.ShowArtistModal {
		return v.renderArtistModalOverlay(content)
	}
	if v.state.ShowSearchModal {
		return v.renderSearchModalOverlay(content)
	}
	if v.state.ShowSortModal {
		return v.renderSortModalOverlay(content)
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
	
	// Ensure header has a valid width
	headerWidth := v.width
	if headerWidth <= 0 {
		headerWidth = 80 // Fallback width
	}
	
	return v.styles.Header.Width(headerWidth).Render(tabBar)
}

// renderContent creates the main content area based on current tab
func (v *MainView) renderContent() string {
	// Ensure we have valid dimensions
	width := v.width
	height := v.height
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}
	
	contentHeight := height - 6 // Account for header, footer, and log area
	contentWidth := width - 2
	if contentWidth < 10 {
		contentWidth = 10 // Minimum content width
	}
	if contentHeight < 5 {
		contentHeight = 5 // Minimum content height
	}
	
	content := v.styles.Content.
		Width(contentWidth).
		Height(contentHeight)

	switch v.state.CurrentTab {
	case models.HomeTab:
		return content.Render(v.renderHomeTab())
	case models.AlbumsTab:
		return content.Render(v.renderAlbumsTab())
	case models.ArtistsTab:
		return content.Render(v.renderArtistsTab())
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

// renderFooter creates a simple footer with basic info (player module handles playback details)
func (v *MainView) renderFooter() string {
	// Simple footer with just navigation hints since player module handles playback
	footer := "Tab/Shift+Tab: Switch tabs | Shift+F: Global Search | Shift+S: Sort | Ctrl+C/q: Quit"
	
	// Ensure footer has a valid width
	footerWidth := v.width
	if footerWidth <= 0 {
		footerWidth = 80 // Fallback width
	}
	
	return v.styles.Footer.Width(footerWidth).Render(footer)
}

// Tab-specific render functions
func (v *MainView) renderHomeTab() string {
	if v.state.LoadingHomeData {
		return "🏠 Home\n\nLoading home data..."
	}
	
	if v.state.LoadingError != "" {
		return fmt.Sprintf("🏠 Home\n\n❌ Error: %s\n\nPress 'r' to retry", v.state.LoadingError)
	}
	
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
	
	// Instructions
	content.WriteString("↑↓ Navigate • PgUp/PgDn Jump sections • Enter/Shift+Enter to select • R to refresh\n\n")
	
	// Render all four sections vertically
	content.WriteString(v.renderHomeSections())
	
	return content.String()
}

// renderHomeSections renders all four home sections vertically with interactive navigation
func (v *MainView) renderHomeSections() string {
	var sections strings.Builder
	
	// Use full width for vertical layout
	sectionWidth := v.width - 4 // Leave space for padding
	if sectionWidth < 40 {
		sectionWidth = 40
	}
	
	// Render all sections vertically
	sections.WriteString(v.renderRecentlyAddedSection(sectionWidth))
	sections.WriteString("\n")
	sections.WriteString(v.renderTopArtistsSection(sectionWidth))
	sections.WriteString("\n")
	sections.WriteString(v.renderMostPlayedAlbumsSection(sectionWidth))
	sections.WriteString("\n")
	sections.WriteString(v.renderTopTracksSection(sectionWidth))
	
	return sections.String()
}

// renderRecentlyAddedSection renders the Recently Added Albums section
func (v *MainView) renderRecentlyAddedSection(width int) string {
	var content strings.Builder
	isActiveSection := v.state.HomeSelectedSection == 0
	
	// Section title with indicator if active
	title := "💿 Recently Added Albums"
	if isActiveSection {
		title = v.styles.ActiveSectionTitle.Render(title)
	} else {
		title = v.styles.SectionTitle.Render(title)
	}
	content.WriteString(title + "\n")
	
	if len(v.state.RecentlyAddedAlbums) == 0 {
		content.WriteString("  No albums loaded\n")
		return content.String()
	}
	
	// Show albums with selection highlighting
	maxShow := 6 // More items since we have vertical space
	if len(v.state.RecentlyAddedAlbums) < maxShow {
		maxShow = len(v.state.RecentlyAddedAlbums)
	}
	
	for i := 0; i < maxShow; i++ {
		album := v.state.RecentlyAddedAlbums[i]
		yearStr := ""
		if album.Year > 0 {
			yearStr = fmt.Sprintf(" (%d)", album.Year)
		}
		
		line := fmt.Sprintf("%s - %s%s", album.Artist, album.Name, yearStr)
		if isActiveSection && v.state.HomeSelectedIndex == i {
			line = v.styles.ActiveField.Render("> " + line)
		} else {
			line = "  " + line
		}
		content.WriteString(line + "\n")
	}
	
	if len(v.state.RecentlyAddedAlbums) > maxShow {
		content.WriteString(fmt.Sprintf("  ... %d more\n", len(v.state.RecentlyAddedAlbums)-maxShow))
	}
	
	return content.String()
}

// renderTopArtistsSection renders the Top Artists section
func (v *MainView) renderTopArtistsSection(width int) string {
	var content strings.Builder
	isActiveSection := v.state.HomeSelectedSection == 1
	
	// Section title with indicator if active
	title := "🎤 Top Artists"
	if isActiveSection {
		title = v.styles.ActiveSectionTitle.Render(title)
	} else {
		title = v.styles.SectionTitle.Render(title)
	}
	content.WriteString(title + "\n")
	
	if len(v.state.TopArtistsByPlays) == 0 {
		content.WriteString("  No artists loaded\n")
		return content.String()
	}
	
	// Show artists with selection highlighting
	maxShow := 5 // Show all 5 top artists
	if len(v.state.TopArtistsByPlays) < maxShow {
		maxShow = len(v.state.TopArtistsByPlays)
	}
	
	for i := 0; i < maxShow; i++ {
		artist := v.state.TopArtistsByPlays[i]
		star := ""
		if artist.StarredAt != nil {
			star = "★ "
		}
		
		albumText := "album"
		if artist.AlbumCount != 1 {
			albumText = "albums"
		}
		
		line := fmt.Sprintf("%s%s (%d %s)", star, artist.Name, artist.AlbumCount, albumText)
		if isActiveSection && v.state.HomeSelectedIndex == i {
			line = v.styles.ActiveField.Render("> " + line)
		} else {
			line = "  " + line
		}
		content.WriteString(line + "\n")
	}
	
	if len(v.state.TopArtistsByPlays) > maxShow {
		content.WriteString(fmt.Sprintf("  ... %d more\n", len(v.state.TopArtistsByPlays)-maxShow))
	}
	
	return content.String()
}

// renderMostPlayedAlbumsSection renders the Most Played Albums section
func (v *MainView) renderMostPlayedAlbumsSection(width int) string {
	var content strings.Builder
	isActiveSection := v.state.HomeSelectedSection == 2
	
	// Section title with indicator if active
	title := "🔥 Most Played Albums"
	if isActiveSection {
		title = v.styles.ActiveSectionTitle.Render(title)
	} else {
		title = v.styles.SectionTitle.Render(title)
	}
	content.WriteString(title + "\n")
	
	if len(v.state.MostPlayedAlbums) == 0 {
		content.WriteString("  No albums loaded\n")
		return content.String()
	}
	
	// Show albums with selection highlighting
	maxShow := 6 // More items since we have vertical space
	if len(v.state.MostPlayedAlbums) < maxShow {
		maxShow = len(v.state.MostPlayedAlbums)
	}
	
	for i := 0; i < maxShow; i++ {
		album := v.state.MostPlayedAlbums[i]
		yearStr := ""
		if album.Year > 0 {
			yearStr = fmt.Sprintf(" (%d)", album.Year)
		}
		
		line := fmt.Sprintf("%s - %s%s", album.Artist, album.Name, yearStr)
		if isActiveSection && v.state.HomeSelectedIndex == i {
			line = v.styles.ActiveField.Render("> " + line)
		} else {
			line = "  " + line
		}
		content.WriteString(line + "\n")
	}
	
	if len(v.state.MostPlayedAlbums) > maxShow {
		content.WriteString(fmt.Sprintf("  ... %d more\n", len(v.state.MostPlayedAlbums)-maxShow))
	}
	
	return content.String()
}

// renderTopTracksSection renders the Top Tracks section
func (v *MainView) renderTopTracksSection(width int) string {
	var content strings.Builder
	isActiveSection := v.state.HomeSelectedSection == 3
	
	// Section title with indicator if active
	title := "🎵 Top Tracks"
	if isActiveSection {
		title = v.styles.ActiveSectionTitle.Render(title)
	} else {
		title = v.styles.SectionTitle.Render(title)
	}
	content.WriteString(title + "\n")
	
	if len(v.state.TopTracks) == 0 {
		content.WriteString("  No tracks loaded\n")
		return content.String()
	}
	
	// Show tracks with selection highlighting  
	maxShow := 8 // More tracks since we have vertical space
	if len(v.state.TopTracks) < maxShow {
		maxShow = len(v.state.TopTracks)
	}
	
	for i := 0; i < maxShow; i++ {
		track := v.state.TopTracks[i]
		
		// Format duration (seconds to mm:ss)
		duration := ""
		if track.Duration > 0 {
			minutes := track.Duration / 60
			seconds := track.Duration % 60
			duration = fmt.Sprintf(" [%d:%02d]", minutes, seconds)
		}
		
		line := fmt.Sprintf("%s - %s%s", track.Artist, track.Title, duration)
		if isActiveSection && v.state.HomeSelectedIndex == i {
			line = v.styles.ActiveField.Render("> " + line)
		} else {
			line = "  " + line
		}
		content.WriteString(line + "\n")
	}
	
	if len(v.state.TopTracks) > maxShow {
		content.WriteString(fmt.Sprintf("  ... %d more\n", len(v.state.TopTracks)-maxShow))
	}
	
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
	instructions := "↑↓ Navigate • PgUp/PgDn Jump 25 • Enter to view tracks • Alt+Enter/A to queue album • R to refresh • Shift+S to sort"
	content.WriteString(instructions + "\n\n")
	
	// Render all albums with smart viewport for large lists
	startIdx := 0
	endIdx := len(v.state.Albums)
	
	// For very large lists, show a window around the selected item
	maxVisible := 25 // Show more items since we removed pagination
	if len(v.state.Albums) > maxVisible {
		// Center the viewport around the selected item
		viewportStart := v.state.SelectedAlbumIndex - maxVisible/2
		if viewportStart < 0 {
			viewportStart = 0
		}
		if viewportStart+maxVisible > len(v.state.Albums) {
			viewportStart = len(v.state.Albums) - maxVisible
		}
		startIdx = viewportStart
		endIdx = viewportStart + maxVisible
	}
	
	for i := startIdx; i < endIdx; i++ {
		album := v.state.Albums[i]
		line := v.formatAlbumLine(album, i == v.state.SelectedAlbumIndex)
		content.WriteString(line)
		content.WriteString("\n")
	}
	
	// Show total count
	if len(v.state.Albums) > 0 {
		if len(v.state.Albums) > maxVisible {
			content.WriteString(fmt.Sprintf("\nShowing %d-%d of %d albums", 
				startIdx+1, endIdx, len(v.state.Albums)))
		} else {
			content.WriteString(fmt.Sprintf("\n%d albums total", len(v.state.Albums)))
		}
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
	content.WriteString("↑↓ Navigate • PgUp/PgDn Jump 25 • Enter to view albums • R to refresh • Shift+S to sort\n\n")
	
	// Render all artists with smart viewport for large lists
	startIdx := 0
	endIdx := len(v.state.Artists)
	
	// For very large lists, show a window around the selected item
	maxVisible := 25 // Show more items since we removed pagination
	if len(v.state.Artists) > maxVisible {
		// Center the viewport around the selected item
		viewportStart := v.state.SelectedArtistIndex - maxVisible/2
		if viewportStart < 0 {
			viewportStart = 0
		}
		if viewportStart+maxVisible > len(v.state.Artists) {
			viewportStart = len(v.state.Artists) - maxVisible
		}
		startIdx = viewportStart
		endIdx = viewportStart + maxVisible
	}
	
	for i := startIdx; i < endIdx; i++ {
		artist := v.state.Artists[i]
		line := v.formatArtistLine(artist, i == v.state.SelectedArtistIndex)
		content.WriteString(line)
		content.WriteString("\n")
	}
	
	// Show total count
	if len(v.state.Artists) > 0 {
		if len(v.state.Artists) > maxVisible {
			content.WriteString(fmt.Sprintf("\nShowing %d-%d of %d artists", 
				startIdx+1, endIdx, len(v.state.Artists)))
		} else {
			content.WriteString(fmt.Sprintf("\n%d artists total", len(v.state.Artists)))
		}
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
	content.WriteString(fmt.Sprintf("↑↓ Navigate • PgUp/PgDn Jump 25 • Enter/Space to play • X/Del to remove • C to clear all (%d tracks)\n\n", len(v.state.Queue)))
	
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
	// Format duration (seconds to mm:ss)
	duration := ""
	if track.Duration > 0 {
		minutes := track.Duration / 60
		seconds := track.Duration % 60
		duration = fmt.Sprintf(" [%d:%02d]", minutes, seconds)
	}
	
	// Build the track info without styling first
	trackInfo := fmt.Sprintf("%s - %s (%s)%s", track.Artist, track.Title, track.Album, duration)
	
	var line string
	
	// Indicate if this is the currently playing track
	if v.state.CurrentTrack != nil && track.ID == v.state.CurrentTrack.ID {
		if v.state.IsPlaying {
			line = "▶ " + trackInfo
		} else {
			line = "⏸ " + trackInfo
		}
		// Use special styling for current track
		if selected {
			return v.styles.ActiveField.Render("> " + line)
		}
		return v.styles.CurrentTrack.Render("  " + line)
	} else {
		// Regular queue entry with styled number
		queueNum := v.styles.QueueNumber.Render(fmt.Sprintf("%2d. ", index+1))
		line = queueNum + trackInfo
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
		} else if strings.Contains(cf.ConnectionStatus, "ℹ") {
			style = v.styles.InfoMessage
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


// renderPlayer creates the persistent player module
func (v *MainView) renderPlayer() string {
	// Ensure player has a valid width
	playerWidth := v.width
	if playerWidth <= 0 {
		playerWidth = 80 // Fallback width
	}
	
	playerStyle := v.styles.Player.Copy().Width(playerWidth)

	if v.state.CurrentTrack == nil {
		// Show empty player with current state
		var status []string
		
		if v.state.IsPlaying {
			status = append(status, "▶ Playing")
		} else {
			status = append(status, "⏸ Stopped")
		}
		
		status = append(status, fmt.Sprintf("Vol: %d%%", v.state.Volume))
		status = append(status, fmt.Sprintf("Queue: %d", len(v.state.Queue)))
		
		if v.state.IsShuffleMode {
			status = append(status, "🔀 SHUFFLE ON")
		} else {
			status = append(status, "🔀 Shuffle off")
		}
		
		statusStr := strings.Join(status, " | ")
		playerContent := fmt.Sprintf("♪ No track loaded | %s\nSPACE: Play/Pause | Alt+←/→: Skip | Alt+S: Shuffle | Shift+↑/↓: Volume", statusStr)
		return playerStyle.Render(playerContent)
	}

	var parts []string

	// Current track info
	trackInfo := fmt.Sprintf("♪ %s - %s", v.state.CurrentTrack.Artist, v.state.CurrentTrack.Title)
	if v.state.CurrentTrack.Album != "" {
		trackInfo += fmt.Sprintf(" (%s)", v.state.CurrentTrack.Album)
	}
	parts = append(parts, trackInfo)

	// Playback status and controls
	var controls []string
	if v.state.IsPlaying {
		controls = append(controls, "▶ Playing")
	} else {
		controls = append(controls, "⏸ Paused")
	}

	// Volume
	controls = append(controls, fmt.Sprintf("Vol: %d%%", v.state.Volume))

	// Queue info
	controls = append(controls, fmt.Sprintf("Queue: %d", len(v.state.Queue)))

	// Shuffle indicator
	if v.state.IsShuffleMode {
		controls = append(controls, "🔀 Shuffle")
	}

	// Progress bar placeholder (for now, we'll show a simple indicator)
	if v.state.CurrentTrack.Duration > 0 {
		// TODO: Add actual position tracking
		controls = append(controls, "[▓▓▓▓░░░░░░]")
	}

	controlStr := strings.Join(controls, " | ")
	parts = append(parts, controlStr)

	// Keybindings hint
	parts = append(parts, "SPACE: Play/Pause | Alt+←/→: Skip | Alt+S: Shuffle | ←/→: Scrub | Shift+↑/↓: Volume")

	playerContent := strings.Join(parts, "\n")
	return playerStyle.Render(playerContent)
}

// renderLogArea creates the log area at the bottom showing recent events
func (v *MainView) renderLogArea() string {
	// Ensure log area has a valid width
	logWidth := v.width
	if logWidth <= 0 {
		logWidth = 80 // Fallback width
	}
	
	logStyle := v.styles.LogArea.Copy().Width(logWidth)

	if len(v.state.LogMessages) == 0 {
		return logStyle.Render("Ready • Press SPACE to play/pause, Alt+S for shuffle, or navigate with Tab")
	}

	// Show up to 2 most recent log messages
	var logLines []string
	messageCount := len(v.state.LogMessages)
	
	if messageCount > 0 {
		// Show the most recent messages (up to 2)
		startIndex := 0
		if messageCount > 2 {
			startIndex = messageCount - 2
		}
		
		for i := startIndex; i < messageCount; i++ {
			msg := v.state.LogMessages[i]
			// Truncate very long messages to fit nicely
			if len(msg) > v.width-4 {
				msg = msg[:v.width-7] + "..."
			}
			logLines = append(logLines, msg)
		}
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
	// Ensure we have valid dimensions
	width := v.width
	height := v.height
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}
	
	// Use lipgloss to properly position the modal
	modalStyle := v.styles.ModalBorder.Copy().
		Width(modalWidth-4). // Account for border and padding
		Height(modalHeight-4).
		Align(lipgloss.Center, lipgloss.Center)
	
	// Position the modal in the center of the available space
	positionedModal := lipgloss.Place(
		width, height,
		lipgloss.Center, lipgloss.Center,
		modalStyle.Render(modal),
	)
	
	return positionedModal
}

// renderSearchModalOverlay renders the search modal overlay
func (v *MainView) renderSearchModalOverlay(background string) string {
	var content strings.Builder
	
	// Modal header
	content.WriteString("🔍 Global Search\n\n")
	
	// Search input box
	content.WriteString(fmt.Sprintf("Search: %s█\n\n", v.state.SearchQuery))
	
	if v.state.LoadingSearchResults {
		content.WriteString("Searching...")
	} else if len(v.state.SearchQuery) == 0 {
		content.WriteString("Type to search across artists, albums, and tracks\n")
		content.WriteString("↑↓ Navigate • Enter to select • Esc to close")
	} else {
		results := v.state.SearchResults
		
		if len(results.Artists) == 0 && len(results.Albums) == 0 && len(results.Tracks) == 0 {
			content.WriteString("No results found")
		} else {
			content.WriteString("↑↓ Navigate • Enter: Play & queue remaining • Shift+Enter: Queue only • Esc to close\n\n")
			
			currentIndex := 0
			
			// Artists section
			if len(results.Artists) > 0 {
				content.WriteString("🎤 Artists:\n")
				for _, artist := range results.Artists {
					selected := currentIndex == v.state.SelectedSearchIndex
					line := v.formatSearchArtistLine(artist, selected)
					content.WriteString(line)
					content.WriteString("\n")
					currentIndex++
				}
				// Add MORE option if we have exactly 5 artists (indicating more might be available)
				if len(results.Artists) == 5 {
					selected := currentIndex == v.state.SelectedSearchIndex
					line := "  " + "→ MORE artists..."
					if selected {
						line = v.styles.ActiveField.Render("> " + "→ MORE artists...")
					}
					content.WriteString(line)
					content.WriteString("\n")
					currentIndex++
				}
				content.WriteString("\n")
			}
			
			// Albums section
			if len(results.Albums) > 0 {
				content.WriteString("💿 Albums:\n")
				for _, album := range results.Albums {
					selected := currentIndex == v.state.SelectedSearchIndex
					line := v.formatSearchAlbumLine(album, selected)
					content.WriteString(line)
					content.WriteString("\n")
					currentIndex++
				}
				// Add MORE option if we have exactly 5 albums
				if len(results.Albums) == 5 {
					selected := currentIndex == v.state.SelectedSearchIndex
					line := "  " + "→ MORE albums..."
					if selected {
						line = v.styles.ActiveField.Render("> " + "→ MORE albums...")
					}
					content.WriteString(line)
					content.WriteString("\n")
					currentIndex++
				}
				content.WriteString("\n")
			}
			
			// Tracks section
			if len(results.Tracks) > 0 {
				content.WriteString("🎵 Tracks:\n")
				for _, track := range results.Tracks {
					selected := currentIndex == v.state.SelectedSearchIndex
					line := v.formatSearchTrackLine(track, selected)
					content.WriteString(line)
					content.WriteString("\n")
					currentIndex++
				}
				// Add MORE option if we have exactly 5 tracks
				if len(results.Tracks) == 5 {
					selected := currentIndex == v.state.SelectedSearchIndex
					line := "  " + "→ MORE tracks..."
					if selected {
						line = v.styles.ActiveField.Render("> " + "→ MORE tracks...")
					}
					content.WriteString(line)
					content.WriteString("\n")
					currentIndex++
				}
			}
		}
	}

	// Center the modal overlay
	return v.overlayModal(background, content.String(), 80, 25)
}

// formatSearchArtistLine formats an artist line for search results
func (v *MainView) formatSearchArtistLine(artist models.Artist, selected bool) string {
	starred := ""
	if artist.StarredAt != nil {
		starred = "★ "
	}
	
	line := fmt.Sprintf("%s%s (%d albums)", starred, artist.Name, artist.AlbumCount)
	
	if selected {
		return v.styles.ActiveField.Render("> " + line)
	}
	
	return "  " + line
}

// formatSearchAlbumLine formats an album line for search results
func (v *MainView) formatSearchAlbumLine(album models.Album, selected bool) string {
	year := ""
	if album.Year > 0 {
		year = fmt.Sprintf("[%d] ", album.Year)
	}
	
	line := fmt.Sprintf("%s%s - %s (%d tracks)", year, album.Artist, album.Name, album.TrackCount)
	
	if selected {
		return v.styles.ActiveField.Render("> " + line)
	}
	
	return "  " + line
}

// formatSearchTrackLine formats a track line for search results
func (v *MainView) formatSearchTrackLine(track models.Track, selected bool) string {
	// Format duration (seconds to mm:ss)
	duration := ""
	if track.Duration > 0 {
		minutes := track.Duration / 60
		seconds := track.Duration % 60
		duration = fmt.Sprintf(" [%d:%02d]", minutes, seconds)
	}
	
	line := fmt.Sprintf("%s - %s (%s)%s", track.Artist, track.Title, track.Album, duration)
	
	if selected {
		return v.styles.ActiveField.Render("> " + line)
	}
	
	return "  " + line
}

// renderSortModalOverlay renders the sort modal overlay
func (v *MainView) renderSortModalOverlay(background string) string {
	var content strings.Builder
	
	// Modal header
	contextName := ""
	switch v.state.CurrentSortContext {
	case "albums":
		contextName = "Albums"
	case "artists":
		contextName = "Artists"
	case "playlists":
		contextName = "Playlists"
	}
	content.WriteString(fmt.Sprintf("🔧 Sort %s\n\n", contextName))
	
	// Instructions
	content.WriteString("↑↓ Navigate • Enter to apply sort • Esc to cancel\n\n")
	
	// Get available sort options for current context
	availableOptions := v.getAvailableSortOptions()
	
	if len(availableOptions) == 0 {
		content.WriteString("No sort options available for this context")
	} else {
		// Render sort options
		for i, option := range availableOptions {
			selected := i == v.state.SelectedSortIndex
			
			line := option.DisplayName
			if selected {
				line = v.styles.ActiveField.Render("> " + line)
			} else {
				line = "  " + line
			}
			
			content.WriteString(line)
			content.WriteString("\n")
		}
	}
	
	// Center the modal overlay
	return v.overlayModal(background, content.String(), 50, 15)
}

// getAvailableSortOptions returns sort options available for the current context (view helper)
func (v *MainView) getAvailableSortOptions() []models.SortOption {
	var available []models.SortOption
	for _, option := range models.SortOptions {
		for _, applicable := range option.Applicable {
			if applicable == v.state.CurrentSortContext {
				available = append(available, option)
				break
			}
		}
	}
	return available
}