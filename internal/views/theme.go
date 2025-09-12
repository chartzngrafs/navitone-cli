package views

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

// Theme contains the color palette for the application
type Theme struct {
    // Keep placeholders for future customization; we avoid hardcoded hex colors
    // and rely on terminal palette or default styles.
    Primary        lipgloss.Color
    Accent         lipgloss.Color
    Error          lipgloss.Color
    Success        lipgloss.Color
    Info           lipgloss.Color
    AccentIndex    int
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

// NewDarkTheme creates the default dark theme with our palette
func NewDarkTheme() Theme {
    // Map to basic ANSI palette indices so themes can override them.
    return Theme{
        Primary: lipgloss.Color("4"),   // blue
        Accent:  lipgloss.Color("6"),   // cyan
        Error:   lipgloss.Color("1"),   // red
        Success: lipgloss.Color("2"),   // green
        Info:    lipgloss.Color("4"),   // blue
        AccentIndex: -1,
    }
}

// NewLightTheme creates a light variant using the same palette with adjusted contrast
func NewLightTheme() Theme {
    return NewDarkTheme()
}

// ThemedStyles contains all styled components using the theme
type ThemedStyles struct {
	// Tab Navigation
	TabActive        lipgloss.Style
	TabInactive      lipgloss.Style
	
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
	
	// Content Sections
	SectionTitle        lipgloss.Style
	ActiveSectionTitle  lipgloss.Style
	HelpText            lipgloss.Style
	
	// Status Messages
	ErrorMessage     lipgloss.Style
	SuccessMessage   lipgloss.Style
	InfoMessage      lipgloss.Style
	
	// Modal Components
	ModalBorder      lipgloss.Style
	ModalContent     lipgloss.Style
	
	// Special States
	CurrentTrack     lipgloss.Style
	QueueNumber      lipgloss.Style
}

// NewThemedStyles creates a complete set of themed styles
func NewThemedStyles(theme Theme) ThemedStyles {
    // Helper: active highlight style uses either reverse-video or a palette background
    active := lipgloss.NewStyle().Bold(true)
    if theme.AccentIndex >= 0 {
        idx := fmt.Sprintf("%d", theme.AccentIndex)
        active = active.Foreground(lipgloss.Color("0")).Background(lipgloss.Color(idx))
    } else {
        active = active.Reverse(true)
    }
    return ThemedStyles{
        // Tab Navigation (use reverse/bold so it adapts to terminal colors)
        TabActive: active.Copy().
            Padding(0, 1),
        TabInactive: lipgloss.NewStyle().
            Foreground(lipgloss.Color("8")).
            Padding(0, 1),

        // Layout Components (avoid hard-coded backgrounds)
        Header: lipgloss.NewStyle().
            Bold(true).
            Reverse(true).
            Padding(0, 1).
            Width(100),
        Content: lipgloss.NewStyle().
            Padding(1).
            Border(lipgloss.RoundedBorder()),
        Footer: lipgloss.NewStyle().
            Bold(false).
            Reverse(true). // use terminal selection-style background
            Border(lipgloss.RoundedBorder()).
            BorderTop(false).
            BorderBottom(false).
            Padding(0, 1),
        Player: lipgloss.NewStyle().
            Padding(0, 1).
            Bold(true),
        LogArea: lipgloss.NewStyle().
            Padding(0, 1),

        // Interactive Elements
        ActiveField: active.Copy(),
        ActiveEditField: active.Copy().Underline(true),
        InactiveField: lipgloss.NewStyle(),

        // Content Sections
        SectionTitle: lipgloss.NewStyle().
            Bold(true),
        ActiveSectionTitle: lipgloss.NewStyle().
            Bold(true).
            Underline(true),
        HelpText: lipgloss.NewStyle().
            Faint(true).
            Italic(true),

        // Status Messages (use ANSI palette indices)
        ErrorMessage: lipgloss.NewStyle().
            Foreground(theme.Error).
            Bold(true),
        SuccessMessage: lipgloss.NewStyle().
            Foreground(theme.Success).
            Bold(true),
        InfoMessage: lipgloss.NewStyle().
            Foreground(theme.Info).
            Bold(true),

        // Modal Components
        ModalBorder: lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            Padding(1),
        ModalContent: lipgloss.NewStyle(),

        // Special States
        CurrentTrack: lipgloss.NewStyle().
            Bold(true),
        QueueNumber: lipgloss.NewStyle().
            Faint(true),
    }
}

// GetThemeInfo returns a formatted string showing the current theme colors
func (t Theme) GetThemeInfo() string {
    info := fmt.Sprintf("Theme Colors (ANSI-based):\n")
    info += fmt.Sprintf("  Primary: %s\n", string(t.Primary))
    info += fmt.Sprintf("  Accent: %s\n", string(t.Accent))
    info += fmt.Sprintf("  Error: %s\n", string(t.Error))
    info += fmt.Sprintf("  Success: %s\n", string(t.Success))
    info += fmt.Sprintf("  Info: %s\n", string(t.Info))
    return info
}
