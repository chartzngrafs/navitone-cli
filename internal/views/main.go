package views

import (
    "fmt"
    "strings"
    "unicode/utf8"

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
func NewMainView(state *models.AppState, themeVariant string, accentIndex int) *MainView {
    theme := NewTheme(themeVariant, accentIndex)
    styles := NewThemedStyles(theme)

    return &MainView{
        state:  state,
        theme:  theme,
        styles: styles,
        width:  80, // Default width
        height: 24, // Default height
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
	if v.state.ShowPlaylistModal {
		return v.renderPlaylistModalOverlay(content)
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
    // Single-line pill-style tabs within a highlighted header bar
    var tabs []string
    for i := models.HomeTab; i <= models.ConfigTab; i++ {
        style := v.styles.TabInactive
        if i == v.state.CurrentTab { style = v.styles.TabActive }
        tabs = append(tabs, style.Render(i.String()))
    }
    pills := strings.Join(tabs, "")
    headerWidth := v.width
    if headerWidth <= 0 { headerWidth = 80 }
    return v.styles.Header.Width(headerWidth).Render(pills)
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

    // Compute content height accounting for header (1), footer (1), log (1),
    // and content box overhead (border top/bottom + padding top/bottom = 4)
    contentHeight := height - 7
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
    // Context-aware footer with concise key hints
    footer := v.footerHint()

	// Ensure footer has a valid width
	footerWidth := v.width
	if footerWidth <= 0 {
		footerWidth = 80 // Fallback width
	}

	return v.styles.Footer.Width(footerWidth).Render(footer)
}

// footerHint composes global and context-specific key hints
func (v *MainView) footerHint() string {
    global := "Tab/Shift+Tab Switch • Shift+F Search • Shift+S Sort • Ctrl+C/q Quit"

    if v.state.ShowAlbumModal || v.state.ShowArtistModal || v.state.ShowPlaylistModal || v.state.ShowSearchModal || v.state.ShowSortModal {
        return global + " • Esc close • Enter select"
    }

    var ctx string
    switch v.state.CurrentTab {
    case models.HomeTab:
        ctx = "↑↓ Navigate • PgUp/PgDn Sections • Enter select • Shift+Enter queue • R Refresh"
    case models.AlbumsTab:
        ctx = "↑↓ Navigate • PgUp/PgDn Jump 25 • Enter view tracks • A/Alt+Enter queue album • R Refresh • Shift+S Sort"
    case models.ArtistsTab:
        ctx = "↑↓ Navigate • PgUp/PgDn Jump 25 • Enter view albums • R Refresh • Shift+S Sort"
    case models.PlaylistsTab:
        ctx = "↑↓ Navigate • PgUp/PgDn Jump 25 • Enter view tracks • A/Alt+Enter queue • R Refresh • Shift+S Sort"
    case models.QueueTab:
        ctx = "↑↓ Navigate • PgUp/PgDn Jump 25 • Enter/Space play • X/Del remove • C clear • Alt+←/→ prev/next • Shift+↑/↓ volume"
    case models.ConfigTab:
        ctx = "↑↓ Move • Enter edit/toggle • F2 save • F3 test • Esc cancel"
    }

    if ctx != "" {
        return global + " • " + ctx
    }
    return global
}

// formatRow renders a consistent list row with an optional right-aligned metadata column
func (v *MainView) formatRow(left string, right string, selected bool, leading string) string {
    // Approximate inner width
    width := v.width
    if width <= 0 { width = 80 }
    maxLine := width - 6 // account for frame and prefix
    if maxLine < 20 { maxLine = width - 2 }

    prefix := "  "
    if selected { prefix = "> " }

    if leading != "" { leading += " " }
    baseWidth := lipgloss.Width(prefix + leading)

    leftText := left
    rightText := right

    if rightText != "" {
        // ensure at least a space between left and right
        rem := maxLine - baseWidth - 1 - lipgloss.Width(rightText)
        if rem < 1 {
            leftText = v.truncateToWidth(leftText, max(1, rem))
        }
        rem = maxLine - baseWidth - 1 - lipgloss.Width(rightText) - lipgloss.Width(leftText)
        if rem < 0 { rem = 0 }
        line := prefix + leading + leftText + strings.Repeat(" ", rem+1) + rightText
        if selected {
            return v.styles.ActiveField.Render(line)
        }
        return line
    }

    content := prefix + leading + v.truncateToWidth(leftText, maxLine-baseWidth)
    if selected { return v.styles.ActiveField.Render(content) }
    return content
}

// truncateToWidth truncates a string to a visual width and appends an ellipsis when needed
func (v *MainView) truncateToWidth(s string, w int) string {
    if w <= 0 { return "" }
    if lipgloss.Width(s) <= w { return s }

    target := w - 1
    if target < 1 { target = 1 }

    var b strings.Builder
    width := 0
    for _, r := range s {
        rw := runeWidth(r)
        if width+rw > target { break }
        b.WriteRune(r)
        width += rw
    }
    return b.String() + "…"
}

func runeWidth(r rune) int {
    // Basic approximation; most ASCII and narrow runes are width 1
    _ = utf8.RuneLen(r)
    return 1
}

func max(a, b int) int { if a > b { return a }; return b }

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

    // Footer displays navigation instructions

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

    // Footer displays instructions; keep content focused

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
    yearStr := ""
    if album.Year > 0 { yearStr = fmt.Sprintf("[%d] ", album.Year) }
    left := fmt.Sprintf("%s%s - %s", yearStr, album.Artist, album.Name)
    right := fmt.Sprintf("%d tracks", album.TrackCount)
    return v.formatRow(left, right, selected, "")
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

    // Footer displays instructions

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
    unit := "album"; if artist.AlbumCount != 1 { unit = "albums" }
    star := ""
    if artist.StarredAt != nil { star = "★ " }
    left := star + artist.Name
    right := fmt.Sprintf("%d %s", artist.AlbumCount, unit)
    return v.formatRow(left, right, selected, "")
}

func (v *MainView) formatPlaylistLine(playlist models.Playlist, selected bool) string {
    // Format with right-aligned counts and owner
    unit := "song"; if playlist.SongCount != 1 { unit = "songs" }
    icon := "🔒"; if playlist.Public { icon = "🌐" }
    left := icon + " " + playlist.Name
    right := fmt.Sprintf("%d %s", playlist.SongCount, unit)
    if playlist.Owner != "" { right += fmt.Sprintf(" • by %s", playlist.Owner) }
    return v.formatRow(left, right, selected, "")
}

func (v *MainView) renderPlaylistsTab() string {
	if v.state.LoadingPlaylists {
		return "📋 Playlists\n\nLoading playlists..."
	}

	if v.state.LoadingError != "" {
		return fmt.Sprintf("📋 Playlists\n\n❌ Error: %s\n\nPress 'r' to retry", v.state.LoadingError)
	}

	if len(v.state.Playlists) == 0 {
		return "📋 Playlists\n\nNo playlists found.\n\nMake sure your Navidrome server is configured in the Config tab."
	}

	var content strings.Builder
	content.WriteString("📋 Playlists\n\n")

    // Footer displays instructions

	// Render all playlists with smart viewport for large lists
	startIdx := 0
	endIdx := len(v.state.Playlists)

	// For very large lists, show a window around the selected item
	maxVisible := 25
	if len(v.state.Playlists) > maxVisible {
		// Center the viewport around the selected item
		viewportStart := v.state.SelectedPlaylistIndex - maxVisible/2
		if viewportStart < 0 {
			viewportStart = 0
		}
		if viewportStart+maxVisible > len(v.state.Playlists) {
			viewportStart = len(v.state.Playlists) - maxVisible
		}
		startIdx = viewportStart
		endIdx = viewportStart + maxVisible
	}

	for i := startIdx; i < endIdx; i++ {
		playlist := v.state.Playlists[i]
		line := v.formatPlaylistLine(playlist, i == v.state.SelectedPlaylistIndex)
		content.WriteString(line)
		content.WriteString("\n")
	}

	// Show total count
	if len(v.state.Playlists) > 0 {
		if len(v.state.Playlists) > maxVisible {
			content.WriteString(fmt.Sprintf("\nShowing %d-%d of %d playlists",
				startIdx+1, endIdx, len(v.state.Playlists)))
		} else {
			content.WriteString(fmt.Sprintf("\n%d playlists total", len(v.state.Playlists)))
		}
	}

	return content.String()
}

func (v *MainView) renderQueueTab() string {
	var content strings.Builder
	content.WriteString("🔄 Queue\n\n")

	if len(v.state.Queue) == 0 {
		content.WriteString("Queue is empty.\n\n")
		content.WriteString("Add tracks by navigating to Albums, Artists, or Tracks tabs and pressing Enter.")
		return content.String()
	}

    // Footer displays instructions

	// Show current playing track if any
	if v.state.CurrentTrack != nil {
		playStatus := "⏸"
		if v.state.IsPlaying {
			playStatus = "▶"
		}
		content.WriteString(fmt.Sprintf("Now Playing: %s %s - %s\n\n",
			playStatus, v.state.CurrentTrack.Artist, v.state.CurrentTrack.Title))
	}

	// Render queue list with smart viewport for large lists
	startIdx := 0
	endIdx := len(v.state.Queue)

	// For very large lists, show a window around the selected item
	maxVisible := 25 // Show more items since we removed pagination
	if len(v.state.Queue) > maxVisible {
		// Center the viewport around the selected item
		viewportStart := v.state.SelectedQueueIndex - maxVisible/2
		if viewportStart < 0 {
			viewportStart = 0
		}
		if viewportStart+maxVisible > len(v.state.Queue) {
			viewportStart = len(v.state.Queue) - maxVisible
		}
		startIdx = viewportStart
		endIdx = viewportStart + maxVisible
	}

	for i := startIdx; i < endIdx; i++ {
		track := v.state.Queue[i]
		line := v.formatQueueLine(track, i, i == v.state.SelectedQueueIndex)
		content.WriteString(line)
		content.WriteString("\n")
	}

	// Show total count
	if len(v.state.Queue) > 0 {
		if len(v.state.Queue) > maxVisible {
			content.WriteString(fmt.Sprintf("\nShowing %d-%d of %d tracks",
				startIdx+1, endIdx, len(v.state.Queue)))
		} else {
			content.WriteString(fmt.Sprintf("\n%d tracks total", len(v.state.Queue)))
		}
	}

	return content.String()
}

func (v *MainView) formatQueueLine(track models.Track, index int, selected bool) string {
    // Duration right column
    right := ""
    if track.Duration > 0 {
        minutes := track.Duration / 60
        seconds := track.Duration % 60
        right = fmt.Sprintf("%d:%02d", minutes, seconds)
    }

    // Leading: index or play/pause glyph
    leading := fmt.Sprintf("%2d.", index+1)
    playing := v.state.CurrentTrack != nil && track.ID == v.state.CurrentTrack.ID
    if playing {
        if v.state.IsPlaying { leading = "▶" } else { leading = "⏸" }
    }

    left := fmt.Sprintf("%s - %s (%s)", track.Artist, track.Title, track.Album)
    line := v.formatRow(left, right, selected, leading)
    if playing && !selected {
        return v.styles.CurrentTrack.Render(line)
    }
    return line
}

func (v *MainView) renderConfigTab() string {
    cf := v.state.ConfigForm
    if cf == nil {
        return "Configuration not loaded"
    }

    var sections []string

    // Header (avoid emojis to keep borders aligned in all fonts)
    sections = append(sections, "Configuration")
    sections = append(sections, "")

    // Navidrome section
    sections = append(sections, v.renderConfigSection("Navidrome Server Settings", []models.ConfigFormField{
        models.ServerURLField,
        models.UsernameField,
        models.PasswordField,
    }, cf))

    sections = append(sections, "")

    // Server scrobbling status (display above scrobbling settings)
    if cf.ServerScrobblingDetected {
        if cf.ServerScrobblingEnabled {
            sections = append(sections, "[OK] Server scrobbling enabled (configured in Navidrome)")
        } else {
            sections = append(sections, "[X] Server scrobbling disabled (configure in Navidrome)")
        }
    } else {
        sections = append(sections, "[i] Server scrobbling status unavailable")
    }
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

    return strings.Join(sections, "\n")
}

// renderConfigSection renders a section of configuration fields
func (v *MainView) renderConfigSection(title string, fields []models.ConfigFormField, cf *models.ConfigFormState) string {
    var lines []string
    // Section title
    lines = append(lines, v.styles.SectionTitle.Render(title))

    // Fixed inner width for box content
    boxWidth := 45
    // Top border
    lines = append(lines, "┌"+strings.Repeat("─", boxWidth)+"┐")

    // Fields
    for _, field := range fields {
        lines = append(lines, v.renderConfigFieldLine(field, cf, boxWidth))
        // Insert a spacer line between Last.fm and ListenBrainz groups
        if title == "Scrobbling Settings" && field == models.LastFMPasswordField {
            lines = append(lines, "│"+strings.Repeat(" ", boxWidth)+"│")
        }
    }

    // Bottom border
    lines = append(lines, "└"+strings.Repeat("─", boxWidth)+"┘")

    return strings.Join(lines, "\n")
}

// renderConfigFieldLine renders a single configuration field within a fixed-width box
func (v *MainView) renderConfigFieldLine(field models.ConfigFormField, cf *models.ConfigFormState, boxWidth int) string {
    isActive := cf.ActiveField == field
    label := cf.GetFieldLabel(field)

    // Build inner content (without borders) with a leading space for padding
    inner := ""

    if cf.IsCheckboxField(field) {
        // Checkbox field (ASCII to avoid font issues)
        checked := cf.GetCheckboxValue(field)
        box := "[ ]"
        if checked { box = "[x]" }
        inner = fmt.Sprintf(" %s %s", box, label)
    } else {
        // Text input field
        value := cf.GetFieldValue(field)
        if cf.EditMode && isActive {
            value = cf.CurrentInput
        }
        // Compute value width budget inside brackets
        prefix := " " + label + ": ["
        suffix := "]"
        maxVal := boxWidth - lipgloss.Width(prefix) - lipgloss.Width(suffix)
        if maxVal < 0 { maxVal = 0 }
        if lipgloss.Width(value) > maxVal {
            value = v.truncateToWidth(value, maxVal)
        }
        valPadded := value + strings.Repeat(" ", maxVal-lipgloss.Width(value))
        inner = prefix + valPadded + suffix
    }

    // Pad inner to full box width and add borders
    pad := boxWidth - lipgloss.Width(inner)
    if pad < 0 { pad = 0 }
    line := "│" + inner + strings.Repeat(" ", pad) + "│"

    // Highlight active field
    if isActive {
        if cf.EditMode {
            return v.styles.ActiveEditField.Render(line)
        } else {
            return v.styles.ActiveField.Render(line)
        }
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
		content.WriteString("↑↓ Navigate • PgUp/PgDn Jump • Enter to play & queue remainder • A to add all • Esc to close\n\n")

		// Track list with viewport scrolling for large albums
		startIdx := 0
		endIdx := len(v.state.AlbumTracks)

		// For large track lists, show a window around the selected item
		maxVisible := 15 // Show fewer items in modal to fit better
		if len(v.state.AlbumTracks) > maxVisible {
			// Center the viewport around the selected item
			viewportStart := v.state.SelectedModalIndex - maxVisible/2
			if viewportStart < 0 {
				viewportStart = 0
			}
			if viewportStart+maxVisible > len(v.state.AlbumTracks) {
				viewportStart = len(v.state.AlbumTracks) - maxVisible
			}
			startIdx = viewportStart
			endIdx = viewportStart + maxVisible
		}

		for i := startIdx; i < endIdx; i++ {
			track := v.state.AlbumTracks[i]
			line := v.formatModalTrackLine(track, i, i == v.state.SelectedModalIndex)
			content.WriteString(line)
			content.WriteString("\n")
		}

		// Show scroll indicator if there are more tracks
		if len(v.state.AlbumTracks) > maxVisible {
			content.WriteString(fmt.Sprintf("\nShowing %d-%d of %d tracks",
				startIdx+1, endIdx, len(v.state.AlbumTracks)))
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

// renderPlaylistModalOverlay renders the playlist tracks modal overlay
func (v *MainView) renderPlaylistModalOverlay(background string) string {
	if v.state.SelectedPlaylist == nil {
		return background
	}

	var content strings.Builder

	// Modal header - simplified to match album modal pattern
	content.WriteString(fmt.Sprintf("📋 %s (%d tracks)\n\n",
		v.state.SelectedPlaylist.Name, v.state.SelectedPlaylist.SongCount))

	if v.state.LoadingModalContent {
		content.WriteString("Loading tracks...")
	} else if len(v.state.PlaylistTracks) == 0 {
		content.WriteString("No tracks found.")
	} else {
		// Instructions
		content.WriteString("↑↓ Navigate • PgUp/PgDn Jump • Enter to play & queue remainder • A to add all • Esc to close\n\n")

		// Track list with viewport scrolling for large playlists
		startIdx := 0
		endIdx := len(v.state.PlaylistTracks)

		// For large track lists, show a window around the selected item
		maxVisible := 15 // Show fewer items in modal to fit better
		if len(v.state.PlaylistTracks) > maxVisible {
			// Center the viewport around the selected item
			viewportStart := v.state.SelectedModalIndex - maxVisible/2
			if viewportStart < 0 {
				viewportStart = 0
			}
			if viewportStart+maxVisible > len(v.state.PlaylistTracks) {
				viewportStart = len(v.state.PlaylistTracks) - maxVisible
			}
			startIdx = viewportStart
			endIdx = viewportStart + maxVisible
		}

		for i := startIdx; i < endIdx; i++ {
			track := v.state.PlaylistTracks[i]
			line := v.formatModalTrackLine(track, i, i == v.state.SelectedModalIndex)
			content.WriteString(line)
			content.WriteString("\n")
		}

		// Show scroll indicator if there are more tracks
		if len(v.state.PlaylistTracks) > maxVisible {
			content.WriteString(fmt.Sprintf("\nShowing %d-%d of %d tracks",
				startIdx+1, endIdx, len(v.state.PlaylistTracks)))
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

	line := fmt.Sprintf("%s%s - %s%s", trackNum, track.Artist, track.Title, duration)

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
