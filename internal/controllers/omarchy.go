package controllers

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"navitone-cli/internal/config"
	"navitone-cli/internal/views"
)

// Hex color extraction regexes matching #RRGGBB, #RGB, or 0xRRGGBB
var (
	hex6Regex = regexp.MustCompile(`(?i)(?:#|0x)([a-f0-9]{6})`)
	hex3Regex = regexp.MustCompile(`(?i)#([a-f0-9]{3})\b`)
)

// CleanHexColor normalizes any color string representation into standard #RRGGBB format
func CleanHexColor(val string) string {
	val = strings.TrimSpace(val)
	// Match 6-digit hex
	if match := hex6Regex.FindStringSubmatch(val); len(match) == 2 {
		return "#" + strings.ToUpper(match[1])
	}
	// Match 3-digit hex e.g. #0F0 -> #00FF00
	if match := hex3Regex.FindStringSubmatch(val); len(match) == 2 {
		h := match[1]
		return "#" + strings.ToUpper(string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]}))
	}
	return ""
}

// FindOmarchyColorsPath resolves the active colors.toml file path and theme name across Omarchy Quattro and legacy structures.
func FindOmarchyColorsPath(home string) (colorsPath string, themeName string) {
	// Active theme name resolution - Omarchy Quattro stores state in ~/.local/state/omarchy/
	namePaths := []string{
		filepath.Join(home, ".local", "state", "omarchy", "current", "theme.name"),
		filepath.Join(home, ".local", "state", "omarchy", "theme.name"),
		filepath.Join(home, ".local", "state", "omarchy", "active_theme"),
		filepath.Join(home, ".local", "state", "omarchy", "current", "theme", "theme.name"),
		filepath.Join(home, ".config", "omarchy", "current", "theme.name"),
		filepath.Join(home, ".config", "omarchy", "theme.name"),
		filepath.Join(home, ".config", "omarchy", "active_theme"),
		filepath.Join(home, ".config", "omarchy", "current", "theme", "theme.name"),
	}

	for _, np := range namePaths {
		if data, err := os.ReadFile(np); err == nil {
			name := strings.TrimSpace(string(data))
			if name != "" {
				themeName = name
				break
			}
		}
	}

	// Candidate direct paths to colors.toml
	candidatePaths := []string{
		filepath.Join(home, ".local", "state", "omarchy", "current", "theme", "colors.toml"),
		filepath.Join(home, ".local", "state", "omarchy", "current", "colors.toml"),
		filepath.Join(home, ".local", "state", "omarchy", "theme", "colors.toml"),
		filepath.Join(home, ".config", "omarchy", "current", "theme", "colors.toml"),
		filepath.Join(home, ".config", "omarchy", "current", "colors.toml"),
		filepath.Join(home, ".config", "omarchy", "theme", "colors.toml"),
		filepath.Join(home, ".config", "omarchy", "colors.toml"),
	}

	for _, p := range candidatePaths {
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			colorsPath = p
			break
		}
	}

	currentThemeLink := filepath.Join(home, ".local", "state", "omarchy", "current", "theme")
	if themeName == "" {
		if target, err := os.Readlink(currentThemeLink); err == nil {
			themeName = filepath.Base(target)
		} else if fi, err := os.Stat(currentThemeLink); err == nil && fi.IsDir() {
			themeName = "current"
		}
	}

	if colorsPath == "" && themeName != "" {
		namedPaths := []string{
			filepath.Join(home, ".config", "omarchy", "themes", themeName, "colors.toml"),
			filepath.Join(home, ".local", "share", "omarchy", "themes", themeName, "colors.toml"),
		}
		for _, p := range namedPaths {
			if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
				colorsPath = p
				break
			}
		}
	}

	if themeName == "" {
		themeName = "omarchy-theme"
	}

	return colorsPath, themeName
}

// FindOmarchyCavaPath resolves active Cava theme or Cava system config paths.
func FindOmarchyCavaPath(home, themeName string) string {
	candidates := []string{
		filepath.Join(home, ".local", "state", "omarchy", "current", "theme", "cava_theme"),
		filepath.Join(home, ".local", "state", "omarchy", "current", "cava_theme"),
		filepath.Join(home, ".config", "omarchy", "current", "theme", "cava_theme"),
		filepath.Join(home, ".config", "omarchy", "themes", themeName, "cava_theme"),
		filepath.Join(home, ".config", "omarchy", "current", "cava_theme"),
		filepath.Join(home, ".config", "omarchy", "cava_theme"),
	}

	for _, p := range candidates {
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			return p
		}
	}
	return ""
}

// ParseOmarchyCavaTOML parses gradient_color_1 .. gradient_color_N from cava_theme or cava config content.
func ParseOmarchyCavaTOML(data []byte) []string {
	lines := strings.Split(string(data), "\n")
	gradMap := make(map[int]string)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		rawVal := parts[1]

		if idx := strings.LastIndex(key, "."); idx != -1 {
			key = key[idx+1:]
		}
		key = strings.TrimSpace(key)

		// Check for gradient_color_1 .. gradient_color_16
		if strings.HasPrefix(key, "gradient_color_") {
			numStr := key[15:]
			if num, err := strconv.Atoi(numStr); err == nil && num > 0 && num <= 32 {
				hexColor := CleanHexColor(rawVal)
				if hexColor != "" {
					gradMap[num] = hexColor
				}
			}
		}
	}

	if len(gradMap) > 0 {
		var keys []int
		for k := range gradMap {
			keys = append(keys, k)
		}
		sort.Ints(keys)
		var palette []string
		for _, k := range keys {
			palette = append(palette, gradMap[k])
		}
		return palette
	}

	return nil
}

// ParseOmarchyColorsTOML parses Omarchy colors.toml content supporting 24-color Quattro schemas, sections, and inline comments.
func ParseOmarchyColorsTOML(data []byte) (colorsMap map[string]string, palette []string) {
	colorsMap = make(map[string]string)
	lines := strings.Split(string(data), "\n")

	colorIndices := make(map[int]string)
	baseIndices := make(map[int]string)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || (strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]")) {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		rawKey := strings.TrimSpace(parts[0])
		rawVal := parts[1]

		key := rawKey
		if idx := strings.LastIndex(rawKey, "."); idx != -1 {
			key = rawKey[idx+1:]
		}
		key = strings.TrimSpace(key)

		hexColor := CleanHexColor(rawVal)
		if hexColor == "" {
			continue
		}

		colorsMap[key] = hexColor

		if strings.HasPrefix(key, "color") {
			if num, err := strconv.Atoi(key[5:]); err == nil && num >= 0 && num <= 31 {
				colorIndices[num] = hexColor
			}
		}
		if strings.HasPrefix(key, "base") {
			hexStr := key[4:]
			if num, err := strconv.ParseInt(hexStr, 16, 64); err == nil && num >= 0 && num <= 31 {
				baseIndices[int(num)] = hexColor
			}
		}
	}

	if len(colorIndices) > 0 {
		var keys []int
		for k := range colorIndices {
			keys = append(keys, k)
		}
		sort.Ints(keys)
		for _, k := range keys {
			palette = append(palette, colorIndices[k])
		}
	} else if len(baseIndices) > 0 {
		var keys []int
		for k := range baseIndices {
			keys = append(keys, k)
		}
		sort.Ints(keys)
		for _, k := range keys {
			palette = append(palette, baseIndices[k])
		}
	}

	if len(palette) <= 1 {
		for _, k := range []string{"accent", "primary", "secondary", "success", "warning", "error"} {
			if c, ok := colorsMap[k]; ok {
				palette = append(palette, c)
			}
		}
	}

	return colorsMap, palette
}

// ExtractThemeColors maps raw parsed colors to Navitone's semantic UI color roles.
func ExtractThemeColors(colorsMap map[string]string) map[string]string {
	themeColors := make(map[string]string)

	if accent, ok := colorsMap["accent"]; ok {
		themeColors["accent"] = accent
	} else if c4, ok := colorsMap["color4"]; ok {
		themeColors["accent"] = c4
	} else if c12, ok := colorsMap["color12"]; ok {
		themeColors["accent"] = c12
	} else if c6, ok := colorsMap["color6"]; ok {
		themeColors["accent"] = c6
	} else if bD, ok := colorsMap["base0D"]; ok {
		themeColors["accent"] = bD
	}

	if primary, ok := colorsMap["primary"]; ok {
		themeColors["primary"] = primary
	} else if c4, ok := colorsMap["color4"]; ok {
		themeColors["primary"] = c4
	} else if c12, ok := colorsMap["color12"]; ok {
		themeColors["primary"] = c12
	} else if accent, ok := colorsMap["accent"]; ok {
		themeColors["primary"] = accent
	}

	if secondary, ok := colorsMap["secondary"]; ok {
		themeColors["secondary"] = secondary
	} else if c5, ok := colorsMap["color5"]; ok {
		themeColors["secondary"] = c5
	} else if c13, ok := colorsMap["color13"]; ok {
		themeColors["secondary"] = c13
	} else if c2, ok := colorsMap["color2"]; ok {
		themeColors["secondary"] = c2
	} else if bE, ok := colorsMap["base0E"]; ok {
		themeColors["secondary"] = bE
	}

	if success, ok := colorsMap["success"]; ok {
		themeColors["success"] = success
	} else if c2, ok := colorsMap["color2"]; ok {
		themeColors["success"] = c2
	} else if c10, ok := colorsMap["color10"]; ok {
		themeColors["success"] = c10
	} else if bB, ok := colorsMap["base0B"]; ok {
		themeColors["success"] = bB
	}

	if warning, ok := colorsMap["warning"]; ok {
		themeColors["warning"] = warning
	} else if c3, ok := colorsMap["color3"]; ok {
		themeColors["warning"] = c3
	} else if c11, ok := colorsMap["color11"]; ok {
		themeColors["warning"] = c11
	} else if bA, ok := colorsMap["base0A"]; ok {
		themeColors["warning"] = bA
	}

	if errColor, ok := colorsMap["error"]; ok {
		themeColors["error"] = errColor
	} else if c1, ok := colorsMap["color1"]; ok {
		themeColors["error"] = c1
	} else if c9, ok := colorsMap["color9"]; ok {
		themeColors["error"] = c9
	} else if b8, ok := colorsMap["base08"]; ok {
		themeColors["error"] = b8
	}

	return themeColors
}

// reloadOmarchyTheme checks and reloads Omarchy theme colors directly from system config or config.toml
func (a *App) reloadOmarchyTheme() {
	home, err := os.UserHomeDir()
	if err == nil {
		colorsPath, themeName := FindOmarchyColorsPath(home)
		if colorsPath != "" {
			if data, err := os.ReadFile(colorsPath); err == nil {
				colorsMap, palette := ParseOmarchyColorsTOML(data)
				if len(colorsMap) > 0 {
					themeColors := ExtractThemeColors(colorsMap)
					bg := colorsMap["background"]
					fg := colorsMap["foreground"]

					// Check for cava_theme or cava config gradient colors
					cavaPath := FindOmarchyCavaPath(home, themeName)
					var cavaPalette []string
					if cavaPath != "" {
						if cavaData, cErr := os.ReadFile(cavaPath); cErr == nil {
							cavaPalette = ParseOmarchyCavaTOML(cavaData)
						}
					}
					if len(cavaPalette) == 0 {
						for _, c := range palette {
							if bg != "" && strings.EqualFold(c, bg) {
								continue
							}
							cavaPalette = append(cavaPalette, c)
						}
						if len(cavaPalette) == 0 {
							cavaPalette = palette
						}
					}

					newTheme := views.NewThemeFromConfigStructWithCavaPalette(
						themeName,
						"omarchy",
						bg,
						fg,
						themeColors,
						palette,
						cavaPalette,
					)
					newStyles := views.NewThemedStyles(newTheme)
					a.view.SetTheme(newTheme, newStyles)
					a.logMessage(fmt.Sprintf("Refreshed Omarchy theme: %s (%d colors, %d Cava colors)", themeName, len(palette), len(cavaPalette)))
					return
				}
			}
		}
	}

	// Fallback to reloading config.toml
	cfg, err := config.Load()
	if err == nil && cfg != nil {
		var newTheme views.Theme
		if cfg.Theme.Name != "" && cfg.Theme.Colors.Accent != "" {
			colors := map[string]string{
				"accent":    cfg.Theme.Colors.Accent,
				"primary":   cfg.Theme.Colors.Primary,
				"secondary": cfg.Theme.Colors.Secondary,
				"success":   cfg.Theme.Colors.Success,
				"warning":   cfg.Theme.Colors.Warning,
				"error":     cfg.Theme.Colors.Error,
			}
			newTheme = views.NewThemeFromConfigStructWithCavaPalette(
				cfg.Theme.Name,
				cfg.Theme.Source,
				cfg.Theme.Background,
				cfg.Theme.Foreground,
				colors,
				cfg.Theme.Colors.Palette,
				cfg.Theme.Colors.CavaPalette,
			)
		} else {
			newTheme = views.NewTheme(cfg.UI.Theme, cfg.UI.AccentIndex)
		}
		newStyles := views.NewThemedStyles(newTheme)
		a.view.SetTheme(newTheme, newStyles)
		a.logMessage(fmt.Sprintf("Refreshed theme: %s", newTheme.Name))
	}
}
