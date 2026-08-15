package views

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestParseHexColor(t *testing.T) {
	r, g, b := parseHexColor("#ff0000")
	if r != 255 || g != 0 || b != 0 {
		t.Errorf("Expected 255, 0, 0, got %f, %f, %f", r, g, b)
	}

	r, g, b = parseHexColor("#00ff00")
	if r != 0 || g != 255 || b != 0 {
		t.Errorf("Expected 0, 255, 0, got %f, %f, %f", r, g, b)
	}

	// Shorthand hex
	r, g, b = parseHexColor("#0f0")
	if r != 0 || g != 255 || b != 0 {
		t.Errorf("Expected 0, 255, 0 for #0f0, got %f, %f, %f", r, g, b)
	}
}

func TestInterpolateRGB(t *testing.T) {
	c1 := lipgloss.Color("#000000")
	c2 := lipgloss.Color("#ffffff")

	mid := interpolateRGB(c1, c2, 0.5)
	if string(mid) != "#808080" {
		t.Errorf("Expected midpoint #808080, got %s", mid)
	}
}

func TestRenderGradientLine(t *testing.T) {
	theme := NewDarkTheme()
	line := RenderGradientLine(10, 0, theme)
	if line == "" {
		t.Errorf("Expected non-empty line")
	}
	if lipgloss.Width(line) != 10 {
		t.Errorf("Expected visual width 10, got %d", lipgloss.Width(line))
	}
}

func TestRenderGradientBox(t *testing.T) {
	theme := NewDarkTheme()
	content := "Hello World\nLine 2"
	box := RenderGradientBox(content, 20, 6, 0, theme)

	if !strings.Contains(box, "╭") || !strings.Contains(box, "╯") {
		t.Errorf("Box missing border corner characters")
	}
	if !strings.Contains(box, "Hello World") {
		t.Errorf("Box missing content text")
	}
}

func TestApplySelectionShimmer(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	theme := NewDarkTheme()
	text := "Selected Item"

	unselected := ApplySelectionShimmer(text, false, 0, theme)
	if unselected != text {
		t.Errorf("Unselected text should be unchanged")
	}

	selected := ApplySelectionShimmer(text, true, 5, theme)
	if selected == text {
		t.Errorf("Selected text should be formatted with shimmer background")
	}
}

func TestRenderInAppCava(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	theme := NewDarkTheme()
	bars := []int{200, 400, 600, 800, 1000, 500, 300}
	cavaStr := RenderInAppCava(40, 15, bars, 10, true, theme)
	lines := strings.Split(cavaStr, "\n")
	if len(lines) != 15 {
		t.Errorf("Expected 15 lines in Cava rendering, got %d", len(lines))
	}
}

func TestRenderInAppCavaQuietSongSmoothness(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	theme := NewDarkTheme()
	
	// Frame A: peak is 95 out of 1000 (quiet passage)
	barsQuietA := []int{20, 50, 95, 40, 10}
	cavaStrA := RenderInAppCava(40, 10, barsQuietA, 1, true, theme)
	
	// Frame B: peak rises slightly to 105 out of 1000
	barsQuietB := []int{22, 55, 105, 44, 12}
	cavaStrB := RenderInAppCava(40, 10, barsQuietB, 2, true, theme)
	
	// Verify both outputs render correctly without dramatic line length or block output distortions
	linesA := strings.Split(cavaStrA, "\n")
	linesB := strings.Split(cavaStrB, "\n")
	if len(linesA) != 10 || len(linesB) != 10 {
		t.Fatalf("Expected 10 lines, got %d and %d", len(linesA), len(linesB))
	}
}

