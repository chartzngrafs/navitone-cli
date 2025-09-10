package views

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

// Theme contains the color palette for the application
type Theme struct {
	// Base Colors
	Jet            lipgloss.Color // #333333 - Dark base/background
	Aquamarine     lipgloss.Color // #a9fbd7 - Bright accent/active states
	AfricanViolet  lipgloss.Color // #9f87af - Secondary elements
	BlueMunsell    lipgloss.Color // #048ba8 - Headers/primary actions
	Plum           lipgloss.Color // #9c528b - Errors/special states
	
	// Computed Colors for better contrast
	White          lipgloss.Color // For text on dark backgrounds
	LightGray      lipgloss.Color // For secondary text
	DarkGray       lipgloss.Color // For borders/subtle elements
}

// NewTheme creates a new theme with the specified color palette and variant
func NewTheme(variant string) Theme {
	switch variant {
	case "light":
		return NewLightTheme()
	case "dark", "":
		fallthrough
	default:
		return NewDarkTheme()
	}
}

// NewDarkTheme creates the default dark theme with our palette
func NewDarkTheme() Theme {
	return Theme{
		// Primary palette
		Jet:           lipgloss.Color("#333333"),
		Aquamarine:    lipgloss.Color("#a9fbd7"),
		AfricanViolet: lipgloss.Color("#9f87af"),
		BlueMunsell:   lipgloss.Color("#048ba8"),
		Plum:          lipgloss.Color("#9c528b"),
		
		// Additional colors for better UI contrast
		White:         lipgloss.Color("#ffffff"),
		LightGray:     lipgloss.Color("#cccccc"),
		DarkGray:      lipgloss.Color("#666666"),
	}
}

// NewLightTheme creates a light variant using the same palette with adjusted contrast
func NewLightTheme() Theme {
	return Theme{
		// Primary palette with roles adjusted for light theme
		Jet:           lipgloss.Color("#f5f5f5"), // Light background
		Aquamarine:    lipgloss.Color("#048ba8"), // Darker for contrast on light
		AfricanViolet: lipgloss.Color("#9f87af"),
		BlueMunsell:   lipgloss.Color("#333333"), // Dark for text on light
		Plum:          lipgloss.Color("#9c528b"),
		
		// Additional colors for light theme
		White:         lipgloss.Color("#333333"), // Dark text
		LightGray:     lipgloss.Color("#666666"), // Darker secondary text
		DarkGray:      lipgloss.Color("#cccccc"), // Light borders
	}
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
	return ThemedStyles{
		// Tab Navigation - Using Blue Munsell for headers/primary navigation
		TabActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.White).
			Background(theme.BlueMunsell).
			Padding(0, 1),
		TabInactive: lipgloss.NewStyle().
			Foreground(theme.LightGray).
			Background(theme.Jet).
			Padding(0, 1),

		// Layout Components
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.White).
			Background(theme.Plum). // Dark purple header
			Padding(0, 1).
			Width(100),
		Content: lipgloss.NewStyle().
			Padding(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.DarkGray).
			Background(theme.Jet),
		Footer: lipgloss.NewStyle().
			Foreground(theme.White).
			Background(theme.AfricanViolet). // Lighter purple footer
			Padding(0, 1),
		Player: lipgloss.NewStyle().
			Foreground(theme.White).
			Background(theme.AfricanViolet).
			Padding(0, 1).
			Bold(true),
		LogArea: lipgloss.NewStyle().
			Foreground(theme.LightGray).
			Background(theme.Jet).
			Padding(0, 1),

		// Interactive Elements - Aquamarine for active states
		ActiveField: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Jet).
			Background(theme.Aquamarine),
		ActiveEditField: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.White).
			Background(theme.Plum), // Plum for edit mode
		InactiveField: lipgloss.NewStyle().
			Foreground(theme.LightGray),

		// Content Sections
		SectionTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.BlueMunsell),
		ActiveSectionTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.White).
			Background(theme.BlueMunsell). // Different from selection highlight
			Padding(0, 1),
		HelpText: lipgloss.NewStyle().
			Foreground(theme.AfricanViolet).
			Italic(true),

		// Status Messages
		ErrorMessage: lipgloss.NewStyle().
			Foreground(theme.Plum).
			Bold(true),
		SuccessMessage: lipgloss.NewStyle().
			Foreground(theme.Aquamarine).
			Bold(true),
		InfoMessage: lipgloss.NewStyle().
			Foreground(theme.AfricanViolet).
			Bold(true),

		// Modal Components
		ModalBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.BlueMunsell).
			Background(theme.Jet).
			Padding(1),
		ModalContent: lipgloss.NewStyle().
			Foreground(theme.White).
			Background(theme.Jet),

		// Special States
		CurrentTrack: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Aquamarine). // Bright highlight for currently playing
			Background(theme.Jet),
		QueueNumber: lipgloss.NewStyle().
			Foreground(theme.AfricanViolet),
	}
}

// GetThemeInfo returns a formatted string showing the current theme colors
func (t Theme) GetThemeInfo() string {
	info := fmt.Sprintf("Theme Colors:\n")
	info += fmt.Sprintf("  Jet: %s (base/background)\n", string(t.Jet))
	info += fmt.Sprintf("  Aquamarine: %s (active/highlights)\n", string(t.Aquamarine))
	info += fmt.Sprintf("  African Violet: %s (secondary)\n", string(t.AfricanViolet))
	info += fmt.Sprintf("  Blue Munsell: %s (headers/primary)\n", string(t.BlueMunsell))
	info += fmt.Sprintf("  Plum: %s (errors/special)\n", string(t.Plum))
	return info
}