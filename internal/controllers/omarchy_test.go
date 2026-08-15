package controllers

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCleanHexColor(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"#ff0000", "#FF0000"},
		{"#00ff00 # green color", "#00FF00"},
		{"\"#123456\"", "#123456"},
		{"'#abcdef'", "#ABCDEF"},
		{"0x1A1515", "#1A1515"},
		{"color = #f00", "#FF0000"},
		{"invalid", ""},
	}

	for _, tt := range tests {
		got := CleanHexColor(tt.input)
		if got != tt.expected {
			t.Errorf("CleanHexColor(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestParseOmarchyColorsTOML(t *testing.T) {
	tomlData := `
# Sample Omarchy Quattro 24-color Theme
name = "quattro-rally"

[colors]
accent = "#AE3F82" # Omarchy UI accent
cursor = "#ffe9c7"
foreground = "#DCC9BC"
background = "#1A1515"

[colors.normal]
color0 = "#1A1515"
color1 = "#F07342"
color2 = "#959A6B"
color3 = "#E39C45"
color4 = "#756D94"
color5 = "#7B3D79"
color6 = "#AE3F82"
color7 = "#DCC9BC"

[colors.bright]
color8  = "#453636"
color9  = "#CF223E"
color10 = "#959A6B"
color11 = "#E39C45"
color12 = "#756D94"
color13 = "#7B3D79"
color14 = "#AE3F82"
color15 = "#ffe9c7"

[colors.extended]
color16 = "#E2687C"
color17 = "#C24F40"
color18 = "#F5ECDF"
color19 = "#6B72A8"
color20 = "#E21D63"
color21 = "#F0D17C"
color22 = "#DC9B68"
color23 = "#C7D3E0"
`

	colorsMap, palette := ParseOmarchyColorsTOML([]byte(tomlData))

	if colorsMap["accent"] != "#AE3F82" {
		t.Errorf("Expected accent #AE3F82, got %q", colorsMap["accent"])
	}
	if colorsMap["color16"] != "#E2687C" {
		t.Errorf("Expected color16 #E2687C, got %q", colorsMap["color16"])
	}
	if colorsMap["color23"] != "#C7D3E0" {
		t.Errorf("Expected color23 #C7D3E0, got %q", colorsMap["color23"])
	}

	if len(palette) != 24 {
		t.Errorf("Expected 24 colors in palette for Quattro theme, got %d", len(palette))
	}
}

func TestParseOmarchyBase24TOML(t *testing.T) {
	tomlData := `
background = "#181818"
foreground = "#E8E8E8"
accent = "#7CA6B7"

base00 = "#181818"
base01 = "#282828"
base02 = "#383838"
base03 = "#585858"
base04 = "#B8B8B8"
base05 = "#E8E8E8"
base06 = "#F8F8F8"
base07 = "#F8F8F8"
base08 = "#AB4642"
base09 = "#DC9656"
base0A = "#F7CA88"
base0B = "#A1B56C"
base0C = "#86C1B9"
base0D = "#7CA6B7"
base0E = "#BA8BAF"
base0F = "#A16946"
base10 = "#101010"
base11 = "#151515"
base12 = "#202020"
base13 = "#303030"
base14 = "#505050"
base15 = "#808080"
base16 = "#B0B0B0"
base17 = "#E0E0E0"
`

	colorsMap, palette := ParseOmarchyColorsTOML([]byte(tomlData))

	if colorsMap["base00"] != "#181818" {
		t.Errorf("Expected base00 #181818, got %q", colorsMap["base00"])
	}

	if len(palette) != 24 {
		t.Errorf("Expected 24 colors in Base24 palette, got %d", len(palette))
	}
}

func TestExtractThemeColors(t *testing.T) {
	colorsMap := map[string]string{
		"accent":    "#AE3F82",
		"color4":    "#756D94",
		"color5":    "#7B3D79",
		"color2":    "#959A6B",
		"color3":    "#E39C45",
		"color1":    "#F07342",
		"background": "#1A1515",
		"foreground": "#DCC9BC",
	}

	themeColors := ExtractThemeColors(colorsMap)

	expected := map[string]string{
		"accent":    "#AE3F82",
		"primary":   "#756D94",
		"secondary": "#7B3D79",
		"success":   "#959A6B",
		"warning":   "#E39C45",
		"error":     "#F07342",
	}

	if !reflect.DeepEqual(themeColors, expected) {
		t.Errorf("ExtractThemeColors() = %v; want %v", themeColors, expected)
	}
}

func TestFindOmarchyColorsPath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "omarchy_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	themeDir := filepath.Join(tempDir, ".config", "omarchy", "themes", "quattro-test")
	err = os.MkdirAll(themeDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create theme dir: %v", err)
	}

	colorsFile := filepath.Join(themeDir, "colors.toml")
	err = os.WriteFile(colorsFile, []byte("accent = \"#123456\""), 0644)
	if err != nil {
		t.Fatalf("Failed to write colors.toml: %v", err)
	}

	activeThemeFile := filepath.Join(tempDir, ".config", "omarchy", "active_theme")
	err = os.WriteFile(activeThemeFile, []byte("quattro-test"), 0644)
	if err != nil {
		t.Fatalf("Failed to write active_theme: %v", err)
	}

	resolvedPath, themeName := FindOmarchyColorsPath(tempDir)

	if resolvedPath != colorsFile {
		t.Errorf("FindOmarchyColorsPath() path = %q; want %q", resolvedPath, colorsFile)
	}
	if themeName != "quattro-test" {
		t.Errorf("FindOmarchyColorsPath() themeName = %q; want %q", themeName, "quattro-test")
	}
}

func TestParseOmarchyCavaTOML(t *testing.T) {
	cavaData := `
[color]
gradient = 1
gradient_color_1 = '#5C637A'
gradient_color_2 = '#6B72A8'
gradient_color_3 = '#9099a5'
gradient_color_4 = '#e2687c'
`
	palette := ParseOmarchyCavaTOML([]byte(cavaData))
	if len(palette) != 4 {
		t.Fatalf("Expected 4 gradient colors, got %d", len(palette))
	}
	if palette[0] != "#5C637A" || palette[3] != "#E2687C" {
		t.Errorf("Unexpected Cava palette values: %v", palette)
	}
}
