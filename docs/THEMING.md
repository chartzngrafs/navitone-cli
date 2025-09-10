# Navitone CLI - Theming System

## Overview

Navitone CLI now features a sophisticated theming system built around your custom color palette:

### 🎨 Color Palette

- **Jet (#333333)** - Dark base color for backgrounds and inactive elements
- **Aquamarine (#a9fbd7)** - Bright accent color for active selections and highlights
- **African Violet (#9f87af)** - Secondary color for metadata and help text  
- **Blue Munsell (#048ba8)** - Primary color for headers and navigation
- **Plum (#9c528b)** - Special state color for errors and edit modes

## Theme Variants

The theming system supports multiple variants that can be configured in your `config.toml`:

### Dark Theme (Default)
```toml
[ui]
theme = "dark"
```

- **Background**: Dark jet (#333333)
- **Active selections**: Bright aquamarine highlights
- **Headers**: Blue munsell for primary navigation
- **Current track**: Special aquamarine highlighting
- **Errors**: Plum coloring for warnings and edit states

### Light Theme
```toml
[ui]
theme = "light"  
```

- **Background**: Light gray (#f5f5f5)  
- **Text**: Dark colors for optimal readability
- **Active selections**: Darker blue munsell for contrast
- **Maintains the same color relationships with adjusted contrast

## UI Component Mapping

| Component | Dark Theme | Light Theme | Purpose |
|-----------|------------|-------------|---------|
| **Header** | **Plum bg** | **Plum bg** | **Top header with tabs** |
| **Footer** | **African Violet bg** | **African Violet bg** | **Bottom footer** |
| **Tabs (Active)** | Blue Munsell bg | Blue Munsell bg | Primary navigation |
| **Tabs (Inactive)** | Jet bg | Light gray bg | Secondary navigation |
| **Active Selections** | Aquamarine bg | Blue Munsell bg | Current item highlight |
| **Active Section Headers** | **Blue Munsell bg** | **Blue Munsell bg** | **Home tab section headers** |
| **Current Track** | Aquamarine text | Blue Munsell text | Playing track indicator |
| **Player Bar** | African Violet bg | African Violet bg | Persistent player |
| **Error Messages** | Plum text | Plum text | Warnings and errors |
| **Section Titles** | Blue Munsell text | Jet text | Content headers |
| **Help Text** | African Violet text | African Violet text | Secondary information |

## Visual Examples

### Dark Theme Features:
- **Top header** uses dark purple (Plum) background for strong visual presence
- **Bottom footer** uses lighter purple (African Violet) to create visual hierarchy
- **Active album/track selections** pop with bright aquamarine highlights  
- **Home section headers** use blue munsell background when active (different from selections)
- **Currently playing track** stands out with aquamarine text
- **Player bar** uses subtle african violet background
- **Error states** use distinctive plum coloring

### Light Theme Features:
- **Inverted contrast** maintains readability on light backgrounds
- **Same color relationships** with adjusted brightness levels
- **Consistent visual hierarchy** across both variants

## Configuration

The theme is automatically loaded from your configuration file:

```toml
[ui]
theme = "dark"          # or "light"
show_album_art = false
home_album_count = 8
```

Changes take effect on restart. The system gracefully falls back to dark theme if an invalid theme is specified.

## Technical Implementation

- **Color definitions**: Located in `internal/views/theme.go`
- **Style application**: Handled by `ThemedStyles` struct
- **Dynamic loading**: Theme variant loaded from config during startup
- **Fallback behavior**: Invalid themes default to dark variant

The theming system provides a cohesive visual experience that makes the terminal interface as intuitive as modern GUI applications while maintaining excellent readability and visual hierarchy.