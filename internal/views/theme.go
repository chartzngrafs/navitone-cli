package views

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

// Theme contains the color palette for the application
type Theme struct {
    // Enhanced color palette with rich theming support
    Primary        lipgloss.Color // Primary UI color (tabs, borders)
    Accent         lipgloss.Color // Primary accent color (headers, highlights)
    Secondary      lipgloss.Color // Secondary accent (selections, focus)
    Success        lipgloss.Color // Success states (play, connected)
    Warning        lipgloss.Color // Warning states (loading, partial)
    Error          lipgloss.Color // Error states (failed, disconnected)
    Background     lipgloss.Color // Background color
    Foreground     lipgloss.Color // Foreground/text color

    // Theme metadata
    Name           string         // Theme name (e.g., "omarchy-dracula")
    Source         string         // "omarchy", "manual", or "builtin"
    AccentIndex    int            // Legacy ANSI index support
}

// NewTheme creates a new theme with the specified color palette and variant
func NewTheme(variant string, accentIndex int) Theme {
    switch variant {
    case "light":
        t := NewLightTheme()
        t.AccentIndex = accentIndex
        return t
    case "dark", "":
        fallthrough
    default:
        t := NewDarkTheme()
        t.AccentIndex = accentIndex
        return t
    }
}

// NewThemeFromConfig creates a theme from configuration data
func NewThemeFromConfig(themeConfig interface{}) Theme {
    // Import the config package type here for proper type assertion
    // We'll handle this through reflection since we can't import config package
    // due to circular dependency
    return themeFromReflection(themeConfig)
}

// NewThemeFromConfigStruct creates a theme from a concrete ThemeConfig struct
func NewThemeFromConfigStruct(name, source, background, foreground string, colors map[string]string) Theme {
    theme := NewDarkTheme() // Start with defaults

    // Set metadata
    theme.Name = name
    theme.Source = source

    // Set colors from map
    if accent, ok := colors["accent"]; ok {
        theme.Accent = lipgloss.Color(accent)
    }
    if primary, ok := colors["primary"]; ok {
        theme.Primary = lipgloss.Color(primary)
    }
    if secondary, ok := colors["secondary"]; ok {
        theme.Secondary = lipgloss.Color(secondary)
    }
    if success, ok := colors["success"]; ok {
        theme.Success = lipgloss.Color(success)
    }
    if warning, ok := colors["warning"]; ok {
        theme.Warning = lipgloss.Color(warning)
    }
    if error, ok := colors["error"]; ok {
        theme.Error = lipgloss.Color(error)
    }

    // Set background and foreground
    if background != "" {
        theme.Background = lipgloss.Color(background)
    }
    if foreground != "" {
        theme.Foreground = lipgloss.Color(foreground)
    }

    return theme
}

// themeFromReflection uses reflection to extract theme data from any struct
func themeFromReflection(themeConfig interface{}) Theme {
    // For now, just return default and we'll fix this properly
    return NewDarkTheme()
}

// themeFromMap creates a theme from a map structure
func themeFromMap(configMap map[string]interface{}) Theme {
    theme := NewDarkTheme() // Start with defaults

    // Extract theme metadata
    if name, ok := configMap["name"].(string); ok {
        theme.Name = name
    }
    if source, ok := configMap["source"].(string); ok {
        theme.Source = source
    }

    // Extract colors if they exist
    if colorsMap, ok := configMap["colors"].(map[string]interface{}); ok {
        if accent, ok := colorsMap["accent"].(string); ok {
            theme.Accent = lipgloss.Color(accent)
        }
        if primary, ok := colorsMap["primary"].(string); ok {
            theme.Primary = lipgloss.Color(primary)
        }
        if secondary, ok := colorsMap["secondary"].(string); ok {
            theme.Secondary = lipgloss.Color(secondary)
        }
        if success, ok := colorsMap["success"].(string); ok {
            theme.Success = lipgloss.Color(success)
        }
        if warning, ok := colorsMap["warning"].(string); ok {
            theme.Warning = lipgloss.Color(warning)
        }
        if err, ok := colorsMap["error"].(string); ok {
            theme.Error = lipgloss.Color(err)
        }
    }

    // Extract background and foreground
    if background, ok := configMap["background"].(string); ok {
        theme.Background = lipgloss.Color(background)
    }
    if foreground, ok := configMap["foreground"].(string); ok {
        theme.Foreground = lipgloss.Color(foreground)
    }

    return theme
}

// NewDarkTheme creates the default dark theme with our palette
func NewDarkTheme() Theme {
    // Default dark theme with rich colors
    return Theme{
        Primary:     lipgloss.Color("#8be9fd"), // Cyan
        Accent:      lipgloss.Color("#6272a4"), // Muted blue
        Secondary:   lipgloss.Color("#ff79c6"), // Pink
        Success:     lipgloss.Color("#50fa7b"), // Green
        Warning:     lipgloss.Color("#f1fa8c"), // Yellow
        Error:       lipgloss.Color("#ff5555"), // Red
        Background:  lipgloss.Color("#282a36"), // Dark background
        Foreground:  lipgloss.Color("#f8f8f2"), // Light foreground
        Name:        "builtin-dark",
        Source:      "builtin",
        AccentIndex: -1,
    }
}

// NewLightTheme creates a light variant with appropriate contrast
func NewLightTheme() Theme {
    return Theme{
        Primary:     lipgloss.Color("#0184bc"), // Darker cyan for light backgrounds
        Accent:      lipgloss.Color("#44475a"), // Dark blue-gray
        Secondary:   lipgloss.Color("#bd93f9"), // Purple
        Success:     lipgloss.Color("#50a14f"), // Darker green
        Warning:     lipgloss.Color("#986801"), // Darker yellow
        Error:       lipgloss.Color("#e45649"), // Darker red
        Background:  lipgloss.Color("#fafafa"), // Light background
        Foreground:  lipgloss.Color("#383a42"), // Dark foreground
        Name:        "builtin-light",
        Source:      "builtin",
        AccentIndex: -1,
    }
}

// ThemedStyles contains all styled components using the theme
type ThemedStyles struct {
	// Tab Navigation
	TabActive        lipgloss.Style
	TabInactive      lipgloss.Style
	TabHover         lipgloss.Style

	// Layout Components
	Header           lipgloss.Style
	Content          lipgloss.Style
	Footer           lipgloss.Style
	Player           lipgloss.Style
	LogArea          lipgloss.Style

	// Interactive Elements
	ActiveField      lipgloss.Style
	ActiveEditField  lipgloss.Style
	InactiveField    lipgloss.Style
	FocusedField     lipgloss.Style

	// Content Sections
	SectionTitle        lipgloss.Style
	ActiveSectionTitle  lipgloss.Style
	HelpText            lipgloss.Style

	// Status Messages
	ErrorMessage     lipgloss.Style
	SuccessMessage   lipgloss.Style
	InfoMessage      lipgloss.Style
	WarningMessage   lipgloss.Style

	// Modal Components
	ModalBorder      lipgloss.Style
	ModalContent     lipgloss.Style
	ModalHeader      lipgloss.Style

	// Special States
	CurrentTrack     lipgloss.Style
	QueueNumber      lipgloss.Style
	PlayingIndicator lipgloss.Style
	PausedIndicator  lipgloss.Style

	// Progress and Status Indicators
	ProgressBar      lipgloss.Style
	ProgressFill     lipgloss.Style
	VolumeIndicator  lipgloss.Style
	ConnectionStatus lipgloss.Style

	// Content Categories (for visual differentiation)
	AlbumStyle       lipgloss.Style
	ArtistStyle      lipgloss.Style
	PlaylistStyle    lipgloss.Style
	TrackStyle       lipgloss.Style

	// Enhanced Interactive States
	Highlighted      lipgloss.Style
	Selected         lipgloss.Style
	Disabled         lipgloss.Style
}

// NewThemedStyles creates a complete set of themed styles
func NewThemedStyles(theme Theme) ThemedStyles {
    return ThemedStyles{
        // Tab Navigation with rich theming
        TabActive: lipgloss.NewStyle().
            Bold(true).
            Foreground(theme.Background).
            Background(theme.Accent).
            Padding(0, 1),
        TabInactive: lipgloss.NewStyle().
            Foreground(theme.Foreground).
            Faint(true).
            Padding(0, 1),
        TabHover: lipgloss.NewStyle().
            Foreground(theme.Secondary).
            Bold(true).
            Padding(0, 1),

        // Layout Components with theme integration
        Header: lipgloss.NewStyle().
            Bold(true).
            Foreground(theme.Background).
            Background(theme.Primary).
            Padding(0, 1).
            Width(100),
        Content: lipgloss.NewStyle().
            Foreground(theme.Foreground).
            Padding(1).
            Border(lipgloss.RoundedBorder()).
            BorderForeground(theme.Primary),
        Footer: lipgloss.NewStyle().
            Foreground(theme.Background).
            Background(theme.Accent).
            Border(lipgloss.RoundedBorder()).
            BorderTop(false).
            BorderBottom(false).
            BorderForeground(theme.Primary).
            Padding(0, 1),
        Player: lipgloss.NewStyle().
            Foreground(theme.Primary).
            Padding(0, 1).
            Bold(true),
        LogArea: lipgloss.NewStyle().
            Foreground(theme.Foreground).
            Padding(0, 1),

        // Interactive Elements with enhanced feedback
        ActiveField: lipgloss.NewStyle().
            Bold(true).
            Foreground(theme.Background).
            Background(theme.Secondary),
        ActiveEditField: lipgloss.NewStyle().
            Bold(true).
            Foreground(theme.Background).
            Background(theme.Secondary).
            Underline(true),
        InactiveField: lipgloss.NewStyle().
            Foreground(theme.Foreground),
        FocusedField: lipgloss.NewStyle().
            Foreground(theme.Primary).
            Bold(true),

        // Content Sections with visual hierarchy
        SectionTitle: lipgloss.NewStyle().
            Bold(true).
            Foreground(theme.Accent),
        ActiveSectionTitle: lipgloss.NewStyle().
            Bold(true).
            Foreground(theme.Primary).
            Underline(true),
        HelpText: lipgloss.NewStyle().
            Foreground(theme.Foreground).
            Faint(true).
            Italic(true),

        // Status Messages with semantic colors
        ErrorMessage: lipgloss.NewStyle().
            Foreground(theme.Error).
            Bold(true),
        SuccessMessage: lipgloss.NewStyle().
            Foreground(theme.Success).
            Bold(true),
        InfoMessage: lipgloss.NewStyle().
            Foreground(theme.Primary).
            Bold(true),
        WarningMessage: lipgloss.NewStyle().
            Foreground(theme.Warning).
            Bold(true),

        // Modal Components with theme consistency
        ModalBorder: lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(theme.Primary).
            Padding(1),
        ModalContent: lipgloss.NewStyle().
            Foreground(theme.Foreground),
        ModalHeader: lipgloss.NewStyle().
            Bold(true).
            Foreground(theme.Background).
            Background(theme.Primary).
            Padding(0, 1),

        // Special States with rich indicators
        CurrentTrack: lipgloss.NewStyle().
            Bold(true).
            Foreground(theme.Primary),
        QueueNumber: lipgloss.NewStyle().
            Foreground(theme.Foreground).
            Faint(true),
        PlayingIndicator: lipgloss.NewStyle().
            Foreground(theme.Success).
            Bold(true),
        PausedIndicator: lipgloss.NewStyle().
            Foreground(theme.Warning).
            Bold(true),

        // Progress and Status Indicators
        ProgressBar: lipgloss.NewStyle().
            Foreground(theme.Foreground).
            Faint(true),
        ProgressFill: lipgloss.NewStyle().
            Foreground(theme.Primary).
            Bold(true),
        VolumeIndicator: lipgloss.NewStyle().
            Foreground(theme.Secondary),
        ConnectionStatus: lipgloss.NewStyle().
            Foreground(theme.Success).
            Bold(true),

        // Content Categories for visual differentiation
        AlbumStyle: lipgloss.NewStyle().
            Foreground(theme.Primary),
        ArtistStyle: lipgloss.NewStyle().
            Foreground(theme.Secondary),
        PlaylistStyle: lipgloss.NewStyle().
            Foreground(theme.Accent),
        TrackStyle: lipgloss.NewStyle().
            Foreground(theme.Foreground),

        // Enhanced Interactive States
        Highlighted: lipgloss.NewStyle().
            Foreground(theme.Background).
            Background(theme.Primary),
        Selected: lipgloss.NewStyle().
            Foreground(theme.Background).
            Background(theme.Secondary).
            Bold(true),
        Disabled: lipgloss.NewStyle().
            Foreground(theme.Foreground).
            Faint(true),
    }
}

// GetThemeInfo returns a formatted string showing the current theme colors
func (t Theme) GetThemeInfo() string {
    info := fmt.Sprintf("Theme: %s (%s)\n", t.Name, t.Source)
    info += fmt.Sprintf("Colors:\n")
    info += fmt.Sprintf("  Primary: %s\n", string(t.Primary))
    info += fmt.Sprintf("  Accent: %s\n", string(t.Accent))
    info += fmt.Sprintf("  Secondary: %s\n", string(t.Secondary))
    info += fmt.Sprintf("  Success: %s\n", string(t.Success))
    info += fmt.Sprintf("  Warning: %s\n", string(t.Warning))
    info += fmt.Sprintf("  Error: %s\n", string(t.Error))
    info += fmt.Sprintf("  Background: %s\n", string(t.Background))
    info += fmt.Sprintf("  Foreground: %s\n", string(t.Foreground))
    return info
}
