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

// NewMainViewWithTheme creates a new main view with enhanced theme config
func NewMainViewWithTheme(state *models.AppState, themeConfig interface{}) *MainView {
    theme := NewThemeFromConfig(themeConfig)
    styles := NewThemedStyles(theme)

    return &MainView{
        state:  state,
        theme:  theme,
        styles: styles,
        width:  80, // Default width
        height: 24, // Default height
    }
}

// NewMainViewWithDirectTheme creates a new main view with pre-built theme and styles
func NewMainViewWithDirectTheme(state *models.AppState, theme Theme, styles ThemedStyles) *MainView {
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

	// Render Zen Mode (full-screen artwork only)
	if v.state.ZenMode {
		return v.renderZenMode()
	}

	var sections []string

	// Header with tabs - always render header first
	header := v.renderHeader()

	sections = append(sections, header)

	// Main content area
	sections = append(sections, v.renderContent())

	// Streamlined Control Bar (Player, Timeline, Activity Log & Hotkeys)
	sections = append(sections, v.renderControlBar())

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

    // Compute content height dynamically accounting for header (1), content borders (2), and controlBar (5)
    overhead := 8
    contentHeight := height - overhead
    contentWidth := width - 2
	if contentWidth < 10 {
		contentWidth = 10 // Minimum content width
	}
	if contentHeight < 3 {
		contentHeight = 3 // Minimum content height for small terminals
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
	case models.RadioTab:
		return content.Render(v.renderRadioTab())
	case models.QueueTab:
		return content.Render(v.renderQueueTab())
	case models.ConfigTab:
		return content.Render(v.renderConfigTab())
	default:
		return content.Render("Unknown tab")
	}
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

// renderProgressBar creates a dynamic progress bar based on current position
func (v *MainView) renderProgressBar() string {
	if v.state.CurrentTrack == nil || v.state.CurrentTrack.Duration <= 0 {
		return "[░░░░░░░░░░]"
	}

	// Calculate progress ratio
	totalDuration := float64(v.state.CurrentTrack.Duration)
	currentPosition := v.state.Position.Seconds()
	
	// Ensure position doesn't exceed duration
	if currentPosition > totalDuration {
		currentPosition = totalDuration
	}
	
	progressRatio := currentPosition / totalDuration
	
	// Create progress bar with 10 segments
	barWidth := 10
	filledSegments := int(progressRatio * float64(barWidth))
	
	if filledSegments > barWidth {
		filledSegments = barWidth
	}
	
	// Build progress bar
	var bar strings.Builder
	bar.WriteString("[")
	
	for i := 0; i < barWidth; i++ {
		if i < filledSegments {
			bar.WriteString("▓")
		} else {
			bar.WriteString("░")
		}
	}
	
	bar.WriteString("]")
	return bar.String()
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

    // Footer displays navigation instructions

	// Render all four sections vertically with height constraints
	homeSections := v.renderHomeSections()
	content.WriteString(homeSections)

	// Ensure content fits within available height
	fullContent := content.String()

	// Get the content height that was calculated in renderContent()
	contentHeight := v.height - 12
	if contentHeight < 3 {
		contentHeight = 3
	}

	// Split content into lines and truncate if necessary
	lines := strings.Split(fullContent, "\n")

	// Account for border and padding overhead (approximately 4 lines)
	maxContentLines := contentHeight - 4
	if maxContentLines < 1 {
		maxContentLines = 1
	}

	if len(lines) > maxContentLines {
		// Truncate and add scroll indicator
		lines = lines[:maxContentLines-1]
		lines = append(lines, "... (use ↑↓ to navigate sections)")
	}

	return strings.Join(lines, "\n")
}

// renderHomeSections renders all four home sections vertically with interactive navigation
func (v *MainView) renderHomeSections() string {
	var sections strings.Builder

	// Use full width for vertical layout
	sectionWidth := v.width - 4 // Leave space for padding
	if sectionWidth < 40 {
		sectionWidth = 40
	}

	// Calculate available height for sections (conservative approach)
	availableHeight := v.height - 20 // Very conservative to ensure fit
	if availableHeight < 10 {
		availableHeight = 10
	}

	// Render all sections vertically with exactly 4 items each
	maxItemsPerSection := 4 // Always show exactly 4 items per section

	sections.WriteString(v.renderRecentlyAddedSectionConstrained(sectionWidth, maxItemsPerSection))
	sections.WriteString("\n")
	sections.WriteString(v.renderTopArtistsSectionConstrained(sectionWidth, maxItemsPerSection))
	sections.WriteString("\n")
	sections.WriteString(v.renderMostPlayedAlbumsSectionConstrained(sectionWidth, maxItemsPerSection))
	sections.WriteString("\n")
	sections.WriteString(v.renderTopTracksSectionConstrained(sectionWidth, maxItemsPerSection))

	return sections.String()
}

// renderRecentlyAddedSectionConstrained renders the Recently Added Albums section with item limit
func (v *MainView) renderRecentlyAddedSectionConstrained(width int, maxItems int) string {
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

	// Show albums with selection highlighting, constrained by maxItems
	maxShow := maxItems
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

	// No "more" indicator - always show exactly 4 items

	return content.String()
}

// renderTopArtistsSectionConstrained renders the Top Artists section with item limit
func (v *MainView) renderTopArtistsSectionConstrained(width int, maxItems int) string {
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

	// Show artists with selection highlighting, constrained by maxItems
	maxShow := maxItems
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

	// No "more" indicator - always show exactly 4 items

	return content.String()
}

// renderMostPlayedAlbumsSectionConstrained renders the Most Played Albums section with item limit
func (v *MainView) renderMostPlayedAlbumsSectionConstrained(width int, maxItems int) string {
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

	// Show albums with selection highlighting, constrained by maxItems
	maxShow := maxItems
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

	// No "more" indicator - always show exactly 4 items

	return content.String()
}

// renderTopTracksSectionConstrained renders the Top Tracks section with item limit
func (v *MainView) renderTopTracksSectionConstrained(width int, maxItems int) string {
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

	// Show tracks with selection highlighting, constrained by maxItems
	maxShow := maxItems
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

	// No "more" indicator - always show exactly 4 items

	return content.String()
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
	// Adjust maxVisible based on artwork display
	maxVisible := 25 // Default items visible
	if v.state.ShowArtwork && v.state.CurrentArtwork != "" {
		maxVisible = 15 // Reduce visible items when showing artwork
	}
	
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

	// Show artwork if enabled and available
	if v.state.ShowArtwork && len(v.state.Albums) > 0 {
		content.WriteString(v.renderAlbumArtwork())
	}

	return content.String()
}

func (v *MainView) formatAlbumLine(album models.Album, selected bool) string {
    yearStr := ""
    if album.Year > 0 { yearStr = fmt.Sprintf("[%d] ", album.Year) }
    left := fmt.Sprintf("%s%s - %s", yearStr, album.Artist, album.Name)
    
    // Format with play count: "X tracks (Y plays)"
    unit := "track"; if album.TrackCount != 1 { unit = "tracks" }
    right := fmt.Sprintf("%2d %s (%4d plays)", album.TrackCount, unit, album.PlayCount)
    
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
	// Artists tab always shows 25 items (no artwork)
	maxVisible := 25
	
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

	// Note: Artist artwork removed - only Albums tab shows artwork

	return content.String()
}

func (v *MainView) formatArtistLine(artist models.Artist, selected bool) string {
    unit := "album"; if artist.AlbumCount != 1 { unit = "albums" }
    star := ""
    if artist.StarredAt != nil { star = "★ " }
    left := star + artist.Name
    
    // Format with play count: "X albums (Y plays)"
    right := fmt.Sprintf("%2d %s (%4d plays)", artist.AlbumCount, unit, artist.PlayCount)
    
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

func (v *MainView) renderRadioTab() string {
	if v.state.LoadingRadio {
		return "📻 Internet Radio Stations\n\n⏳ Loading radio stations..."
	}

	if v.state.LoadingError != "" {
		return fmt.Sprintf("📻 Internet Radio Stations\n\n❌ Error: %s\n\nPress 'Shift+R' to retry", v.state.LoadingError)
	}

	if len(v.state.RadioStations) == 0 {
		return "📻 Internet Radio Stations\n\nNo internet radio stations found.\n\nAdd radio stations in Navidrome Web UI under Settings -> Radio Stations."
	}

	var content strings.Builder
	content.WriteString("📻 Internet Radio Stations\n\n")

	startIdx := 0
	endIdx := len(v.state.RadioStations)

	maxVisible := 25
	if len(v.state.RadioStations) > maxVisible {
		viewportStart := v.state.SelectedRadioIndex - maxVisible/2
		if viewportStart < 0 {
			viewportStart = 0
		}
		if viewportStart+maxVisible > len(v.state.RadioStations) {
			viewportStart = len(v.state.RadioStations) - maxVisible
		}
		startIdx = viewportStart
		endIdx = viewportStart + maxVisible
	}

	for i := startIdx; i < endIdx; i++ {
		station := v.state.RadioStations[i]
		line := v.formatRadioLine(station, i == v.state.SelectedRadioIndex)
		content.WriteString(line)
		content.WriteString("\n")
	}

	if len(v.state.RadioStations) > 0 {
		if len(v.state.RadioStations) > maxVisible {
			content.WriteString(fmt.Sprintf("\nShowing %d-%d of %d radio stations",
				startIdx+1, endIdx, len(v.state.RadioStations)))
		} else {
			content.WriteString(fmt.Sprintf("\n%d radio stations total", len(v.state.RadioStations)))
		}
	}

	return content.String()
}

func (v *MainView) formatRadioLine(station models.RadioStation, selected bool) string {
	var prefix string
	if selected {
		prefix = "→ "
	} else {
		prefix = "  "
	}

	left := fmt.Sprintf("%s📻 %s", prefix, station.Name)
	right := station.HomePageURL

	return v.formatRow(left, right, selected, prefix)
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

	// UI section
	sections = append(sections, v.renderConfigSection("UI Settings", []models.ConfigFormField{
		models.VimModeField,
		models.ShowArtworkField,
		models.ArtworkQualityField,
		models.ArtworkColorField,
		models.ArtworkSizeField,
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

    // Join all sections and ensure content fits within available height
    fullContent := strings.Join(sections, "\n")

    // Get the content height that was calculated in renderContent()
    contentHeight := v.height - 12
    if contentHeight < 3 {
        contentHeight = 3
    }

    // Split content into lines and truncate if necessary
    lines := strings.Split(fullContent, "\n")

    // Account for border and padding overhead (approximately 4 lines)
    maxContentLines := contentHeight - 4
    if maxContentLines < 1 {
        maxContentLines = 1
    }

    if len(lines) > maxContentLines {
        // Truncate and add scroll indicator
        lines = lines[:maxContentLines-1]
        lines = append(lines, "... (content truncated - scroll to see more)")
    }

    return strings.Join(lines, "\n")
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

// renderControlBar creates a unified, 3-line control panel combining player controls, timeline progress, and status/key hints
func (v *MainView) renderControlBar() string {
	width := v.width
	if width <= 0 {
		width = 80
	}
	boxWidth := width - 2
	if boxWidth < 20 {
		boxWidth = 20
	}
	innerWidth := boxWidth - 4 // account for borders & padding
	if innerWidth <= 15 {
		innerWidth = 15
	}

	boldStyle := lipgloss.NewStyle().Bold(true)
	var controlLines []string

	// --- Line 1: Status Badge + Track Metadata + Player Stats ---
	var statusBadge string
	if v.state.IsPlaying {
		statusBadge = v.styles.PlayingIndicator.Render("▶ Playing")
	} else if v.state.CurrentTrack != nil {
		statusBadge = v.styles.PausedIndicator.Render("⏸ Paused")
	} else {
		statusBadge = v.styles.HelpText.Render("⏹ Stopped")
	}

	var trackText string
	if v.state.CurrentTrack != nil {
		trackText = fmt.Sprintf("🎵 %s — %s", v.state.CurrentTrack.Title, v.state.CurrentTrack.Artist)
		if v.state.CurrentTrack.Album != "" {
			trackText += fmt.Sprintf(" (%s)", v.state.CurrentTrack.Album)
		}
	} else {
		trackText = "♪ No track selected"
	}

	// Player Stats
	var statsParts []string
	bitrate := 0
	if v.state.CurrentTrack != nil && v.state.CurrentTrack.BitRate > 0 {
		bitrate = v.state.CurrentTrack.BitRate
	} else if v.state.Bitrate > 0 {
		bitrate = v.state.Bitrate
	}
	if bitrate > 0 {
		statsParts = append(statsParts, fmt.Sprintf("%d kbps", bitrate))
	}
	statsParts = append(statsParts, fmt.Sprintf("Vol: %d%%", v.state.Volume))
	if v.state.IsShuffleMode {
		statsParts = append(statsParts, "🔀 ON")
	}
	statsParts = append(statsParts, fmt.Sprintf("Q: %d", len(v.state.Queue)))

	statsStr := v.styles.HelpText.Render(strings.Join(statsParts, " • "))

	// Combine Line 1 with truncation/spacing
	maxTrackWidth := innerWidth - lipgloss.Width(statusBadge) - lipgloss.Width(statsStr) - 4
	if maxTrackWidth > 10 && utf8.RuneCountInString(trackText) > maxTrackWidth {
		runes := []rune(trackText)
		trackText = string(runes[:maxTrackWidth-3]) + "..."
	}
	line1Left := statusBadge + "  " + boldStyle.Render(trackText)
	line1Spacer := innerWidth - lipgloss.Width(line1Left) - lipgloss.Width(statsStr)
	if line1Spacer < 1 {
		line1Spacer = 1
	}
	line1 := line1Left + strings.Repeat(" ", line1Spacer) + statsStr
	controlLines = append(controlLines, line1)

	// --- Line 2: Timeline Progress Bar ---
	var currentSec, totalSec int
	if v.state.CurrentTrack != nil && v.state.CurrentTrack.Duration > 0 {
		currentSec = int(v.state.Position.Seconds())
		totalSec = v.state.CurrentTrack.Duration
	}
	timeStart := fmt.Sprintf("%02d:%02d", currentSec/60, currentSec%60)
	timeEnd := fmt.Sprintf("%02d:%02d", totalSec/60, totalSec%60)

	// Dynamic bar width
	barLen := innerWidth - len(timeStart) - len(timeEnd) - 6 // space around brackets
	if barLen < 5 {
		barLen = 5
	}
	progressBar := v.renderCustomProgressBar(barLen)
	line2 := fmt.Sprintf("%s  %s  %s", v.styles.HelpText.Render(timeStart), progressBar, v.styles.HelpText.Render(timeEnd))
	controlLines = append(controlLines, line2)

	// --- Line 3: Latest Event Log (Left) + Contextual Shortcuts (Right) ---
	latestLog := "⚡ Ready"
	if len(v.state.LogMessages) > 0 {
		latestLog = "⚡ " + v.state.LogMessages[len(v.state.LogMessages)-1]
	}

	hotkeys := v.controlBarHotkeys()
	hotkeysW := lipgloss.Width(hotkeys)
	
	var line3Left string
	availForLog := innerWidth - hotkeysW - 2
	if availForLog >= 12 {
		logText := latestLog
		if utf8.RuneCountInString(logText) > availForLog {
			runes := []rune(logText)
			logText = string(runes[:availForLog-3]) + "..."
		}
		line3Left = v.styles.HelpText.Render(logText)
	}

	line3Right := v.styles.HelpText.Render(hotkeys)
	line3Spacer := innerWidth - lipgloss.Width(line3Left) - lipgloss.Width(line3Right)
	if line3Spacer < 1 {
		line3Spacer = 1
	}
	line3 := line3Left + strings.Repeat(" ", line3Spacer) + line3Right
	controlLines = append(controlLines, line3)

	panelContent := strings.Join(controlLines, "\n")

	// Render within rounded themed container
	boxStyle := lipgloss.NewStyle().
		Width(boxWidth).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(v.theme.Primary).
		Padding(0, 1)

	return boxStyle.Render(panelContent)
}

// controlBarHotkeys returns concise context-aware shortcuts for line 3 of the control bar
func (v *MainView) controlBarHotkeys() string {
	if v.state.ShowAlbumModal || v.state.ShowArtistModal || v.state.ShowPlaylistModal || v.state.ShowSearchModal || v.state.ShowSortModal {
		return "Esc Close • Enter Select"
	}
	switch v.state.CurrentTab {
	case models.HomeTab:
		return "Enter Select • Shift+Enter Queue • Shift+R Refresh • / Search • Shift+Z Zen"
	case models.AlbumsTab:
		return "Enter Tracks • A Queue • Shift+S Sort • Shift+R Refresh • / Search • Shift+Z Zen"
	case models.ArtistsTab:
		return "Enter Albums • A Queue • Shift+S Sort • Shift+R Refresh • / Search • Shift+Z Zen"
	case models.PlaylistsTab:
		return "Enter Tracks • A Queue • Shift+S Sort • Shift+R Refresh • / Search • Shift+Z Zen"
	case models.RadioTab:
		return "Enter Play • A Queue • Shift+R Refresh • / Search • Shift+Z Zen"
	case models.QueueTab:
		return "Space Play • Alt+←/→ Skip • Shift+↑/↓ Vol • X Del • C Clear • Shift+Z Zen"
	case models.ConfigTab:
		return "Enter Edit • F2 Save • F3 Test • Shift+Z Zen"
	default:
		return "Space Play • Alt+←/→ Skip • / Search • Shift+Z Zen • q Quit"
	}
}

// renderCustomProgressBar renders a progress bar of exact segment length
func (v *MainView) renderCustomProgressBar(width int) string {
	if width <= 0 {
		return "[]"
	}
	if v.state.CurrentTrack == nil || v.state.CurrentTrack.Duration <= 0 {
		return "[" + strings.Repeat("░", width) + "]"
	}

	totalDuration := float64(v.state.CurrentTrack.Duration)
	currentPosition := v.state.Position.Seconds()
	if currentPosition > totalDuration {
		currentPosition = totalDuration
	}

	progressRatio := currentPosition / totalDuration
	filledSegments := int(progressRatio * float64(width))
	if filledSegments > width {
		filledSegments = width
	}

	barFilled := strings.Repeat("▓", filledSegments)
	barEmpty := strings.Repeat("░", width-filledSegments)

	styledFill := v.styles.ProgressFill.Render(barFilled)
	styledEmpty := v.styles.ProgressBar.Render(barEmpty)

	return "[" + styledFill + styledEmpty + "]"
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

// renderAlbumArtwork renders ASCII artwork for the currently selected album
func (v *MainView) renderAlbumArtwork() string {
	if v.state.LoadingArtwork {
		return "\n\n⏳ Loading artwork..."
	}

	if v.state.CurrentArtwork == "" {
		return "" // No artwork to display
	}

	var content strings.Builder
	content.WriteString("\n\n")

	// Add selected album info if we have it
	if len(v.state.Albums) > v.state.SelectedAlbumIndex {
		album := v.state.Albums[v.state.SelectedAlbumIndex]
		content.WriteString(fmt.Sprintf("🎨 %s - %s", album.Artist, album.Name))
		if album.Year > 0 {
			content.WriteString(fmt.Sprintf(" (%d)", album.Year))
		}
		content.WriteString("\n\n")
	}

	// Add the ASCII artwork
	content.WriteString(v.state.CurrentArtwork)

	return content.String()
}

// renderZenMode renders a clean Zen Mode interface showing full-screen colored album art, playing track info, and playback progress
func (v *MainView) renderZenMode() string {
	var zenParts []string

	// 1. Album Artwork (Color)
	var artContent string
	if v.state.LoadingArtwork {
		artContent = "⏳ Loading artwork..."
	} else if v.state.CurrentArtwork != "" {
		artContent = v.state.CurrentArtwork
	} else {
		artContent = "[ No Album Art Available ]"
	}
	zenParts = append(zenParts, artContent)

	boldStyle := lipgloss.NewStyle().Bold(true)

	// 2. Track & Playing Info
	var infoLines []string
	if v.state.CurrentTrack != nil {
		trackTitle := boldStyle.Render("🎵 " + v.state.CurrentTrack.Title)
		artistAlbum := fmt.Sprintf("👤 %s", v.state.CurrentTrack.Artist)
		if v.state.CurrentTrack.Album != "" {
			artistAlbum += fmt.Sprintf(" • 💿 %s", v.state.CurrentTrack.Album)
		}
		if v.state.CurrentTrack.Year > 0 {
			artistAlbum += fmt.Sprintf(" (%d)", v.state.CurrentTrack.Year)
		}

		infoLines = append(infoLines, trackTitle)
		infoLines = append(infoLines, v.styles.HelpText.Render(artistAlbum))
	} else if len(v.state.Albums) > 0 && v.state.SelectedAlbumIndex < len(v.state.Albums) {
		alb := v.state.Albums[v.state.SelectedAlbumIndex]
		albInfo := fmt.Sprintf("💿 %s • %s", alb.Artist, alb.Name)
		if alb.Year > 0 {
			albInfo += fmt.Sprintf(" (%d)", alb.Year)
		}
		infoLines = append(infoLines, boldStyle.Render(albInfo))
	}
	if len(infoLines) > 0 {
		zenParts = append(zenParts, strings.Join(infoLines, "\n"))
	}

	// 3. Playing State & Progress Bar
	var progressLines []string
	statusStr := "⏹ Stopped"
	if v.state.IsPlaying {
		statusStr = "▶ Playing"
	} else if v.state.CurrentTrack != nil {
		statusStr = "⏸ Paused"
	}

	var metaParts []string
	metaParts = append(metaParts, statusStr)
	bitrate := 0
	if v.state.CurrentTrack != nil && v.state.CurrentTrack.BitRate > 0 {
		bitrate = v.state.CurrentTrack.BitRate
	} else if v.state.Bitrate > 0 {
		bitrate = v.state.Bitrate
	}
	if bitrate > 0 {
		metaParts = append(metaParts, fmt.Sprintf("%d kbps", bitrate))
	}
	metaParts = append(metaParts, fmt.Sprintf("Vol: %d%%", v.state.Volume))
	if v.state.IsShuffleMode {
		metaParts = append(metaParts, "🔀 Shuffle")
	}
	progressLines = append(progressLines, v.styles.HelpText.Render(strings.Join(metaParts, "  •  ")))

	if v.state.CurrentTrack != nil && v.state.CurrentTrack.Duration > 0 {
		currentSec := int(v.state.Position.Seconds())
		totalSec := v.state.CurrentTrack.Duration
		timeStr := fmt.Sprintf("%02d:%02d / %02d:%02d", currentSec/60, currentSec%60, totalSec/60, totalSec%60)
		progressBar := v.renderProgressBar()
		progressLines = append(progressLines, fmt.Sprintf("%s %s", timeStr, progressBar))
	}
	zenParts = append(zenParts, strings.Join(progressLines, "\n"))

	// 4. Subtle Zen Mode key hint
	hint := v.styles.HelpText.Render("Shift+Z Exit Zen Mode • Space Play/Pause • Alt+←/→ Skip • Shift+↑/↓ Volume")
	zenParts = append(zenParts, hint)

	// Combine into a clean centered layout block
	fullContent := strings.Join(zenParts, "\n\n")

	return lipgloss.Place(
		v.width,
		v.height,
		lipgloss.Center,
		lipgloss.Center,
		fullContent,
	)
}

