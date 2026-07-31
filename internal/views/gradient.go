package views

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// OmarchyPalette returns the color stops for the active theme's gradient sequence
func OmarchyPalette(theme Theme) []lipgloss.Color {
	// Standard palette cycle across key Omarchy theme colors
	return []lipgloss.Color{
		theme.Primary,
		theme.Secondary,
		theme.Accent,
		theme.Success,
		theme.Warning,
		theme.Primary, // Loop back to primary
	}
}

// parseHexColor converts a lipgloss.Color hex string to float RGB values (0-255)
func parseHexColor(c lipgloss.Color) (float64, float64, float64) {
	hex := strings.TrimPrefix(string(c), "#")
	if len(hex) == 3 {
		// Expand shorthand hex e.g. "f00" -> "ff0000"
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	if len(hex) != 6 {
		return 139, 233, 253 // Fallback cyan
	}
	r, err1 := strconv.ParseUint(hex[0:2], 16, 8)
	g, err2 := strconv.ParseUint(hex[2:4], 16, 8)
	b, err3 := strconv.ParseUint(hex[4:6], 16, 8)
	if err1 != nil || err2 != nil || err3 != nil {
		return 139, 233, 253
	}
	return float64(r), float64(g), float64(b)
}

// interpolateRGB performs linear RGB interpolation between two lipgloss colors
func interpolateRGB(c1, c2 lipgloss.Color, factor float64) lipgloss.Color {
	factor = math.Max(0, math.Min(1, factor))
	r1, g1, b1 := parseHexColor(c1)
	r2, g2, b2 := parseHexColor(c2)

	r := uint8(math.Round(r1 + (r2-r1)*factor))
	g := uint8(math.Round(g1 + (g2-g1)*factor))
	b := uint8(math.Round(b1 + (b2-b1)*factor))

	return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", r, g, b))
}

// GetGradientColor calculates an interpolated color along the Omarchy palette sequence
func GetGradientColor(step, totalSteps, tick int, theme Theme) lipgloss.Color {
	if totalSteps <= 0 {
		totalSteps = 1
	}

	palette := OmarchyPalette(theme)
	numSegments := len(palette) - 1

	// Scale tick down for 60 FPS master framerate to maintain serene color morphing
	slowTick := tick / 5

	// Calculate shifted normalized position [0.0, 1.0)
	normPos := math.Mod(float64(step+slowTick)/float64(totalSteps), 1.0)
	if normPos < 0 {
		normPos += 1.0
	}

	// Determine segment and internal interpolation factor
	scaled := normPos * float64(numSegments)
	segmentIdx := int(scaled)
	if segmentIdx >= numSegments {
		segmentIdx = numSegments - 1
	}
	factor := scaled - float64(segmentIdx)

	return interpolateRGB(palette[segmentIdx], palette[segmentIdx+1], factor)
}

// RenderGradientBox wraps content in a box with a moving Omarchy gradient border
func RenderGradientBox(content string, width, height, tick int, theme Theme) string {
	if width < 4 {
		width = 4
	}
	if height < 3 {
		height = 3
	}

	innerWidth := width - 2
	innerHeight := height - 2

	// Split content into lines and ensure exactly innerHeight lines
	rawLines := strings.Split(content, "\n")
	contentLines := make([]string, innerHeight)
	for i := 0; i < innerHeight; i++ {
		if i < len(rawLines) {
			line := rawLines[i]
			w := lipgloss.Width(line)
			if w < innerWidth {
				line += strings.Repeat(" ", innerWidth-w)
			}
			contentLines[i] = line
		} else {
			contentLines[i] = strings.Repeat(" ", innerWidth)
		}
	}

	// Calculate perimeter steps
	// Top: width, Right: innerHeight, Bottom: width, Left: innerHeight
	totalPerimeter := 2*width + 2*innerHeight
	step := 0

	var builder strings.Builder

	// --- Top Border Line ---
	// Corner ╭
	cTopLeft := GetGradientColor(step, totalPerimeter, tick, theme)
	builder.WriteString(lipgloss.NewStyle().Foreground(cTopLeft).Render("╭"))
	step++

	// Horizontal top line ─
	for x := 0; x < innerWidth; x++ {
		c := GetGradientColor(step, totalPerimeter, tick, theme)
		builder.WriteString(lipgloss.NewStyle().Foreground(c).Render("─"))
		step++
	}

	// Corner ╮
	cTopRight := GetGradientColor(step, totalPerimeter, tick, theme)
	builder.WriteString(lipgloss.NewStyle().Foreground(cTopRight).Render("╮"))
	step++
	builder.WriteString("\n")

	// --- Middle Content Lines with Left/Right Borders ---
	leftStepStart := totalPerimeter - 1

	for i := 0; i < innerHeight; i++ {
		// Left border │
		cLeft := GetGradientColor(leftStepStart-i, totalPerimeter, tick, theme)
		builder.WriteString(lipgloss.NewStyle().Foreground(cLeft).Render("│"))

		// Content line
		builder.WriteString(contentLines[i])

		// Right border │
		cRight := GetGradientColor(step, totalPerimeter, tick, theme)
		builder.WriteString(lipgloss.NewStyle().Foreground(cRight).Render("│"))
		step++

		builder.WriteString("\n")
	}

	// --- Bottom Border Line ---
	// Corner ╰
	cBottomLeft := GetGradientColor(leftStepStart-innerHeight, totalPerimeter, tick, theme)
	builder.WriteString(lipgloss.NewStyle().Foreground(cBottomLeft).Render("╰"))

	// Bottom line ─
	bottomStepStart := leftStepStart - innerHeight - 1
	for x := 0; x < innerWidth; x++ {
		c := GetGradientColor(bottomStepStart-x, totalPerimeter, tick, theme)
		builder.WriteString(lipgloss.NewStyle().Foreground(c).Render("─"))
	}

	// Corner ╯
	cBottomRight := GetGradientColor(step, totalPerimeter, tick, theme)
	builder.WriteString(lipgloss.NewStyle().Foreground(cBottomRight).Render("╯"))

	return builder.String()
}

// RenderGradientLine creates a horizontal separator line with a moving gradient
func RenderGradientLine(width, tick int, theme Theme) string {
	if width <= 0 {
		return ""
	}
	var builder strings.Builder
	for x := 0; x < width; x++ {
		c := GetGradientColor(x, width, tick, theme)
		builder.WriteString(lipgloss.NewStyle().Foreground(c).Render("─"))
	}
	return builder.String()
}

// ApplySelectionShimmer applies a dynamic pulsing gradient highlight to selected items
func ApplySelectionShimmer(text string, isSelected bool, tick int, theme Theme) string {
	if !isSelected {
		return text
	}

	// Slow, smooth color transition across 120 steps (~10 seconds per cycle)
	gradColor := GetGradientColor(tick/6, 120, 0, theme)
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Background).
		Background(gradColor)

	return style.Render(text)
}

// RenderGradientText colors each character of a string with a slow, smooth moving gradient wave
func RenderGradientText(text string, tick int, theme Theme) string {
	if text == "" {
		return ""
	}
	runes := []rune(text)
	totalSteps := 60
	slowTick := tick / 5 // Gentle 400ms step progression

	var builder strings.Builder
	for i, r := range runes {
		c := GetGradientColor(i, totalSteps, slowTick, theme)
		builder.WriteString(lipgloss.NewStyle().Foreground(c).Bold(true).Render(string(r)))
	}
	return builder.String()
}

// RenderInAppCava renders an integrated multi-bar Cava audio spectrum visualizer matching system Cava config
func RenderInAppCava(width, height int, bars []int, tick int, isPlaying bool, theme Theme) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	// Exact Unicode sub-cell fractional block characters from 1/8th to 8/8ths
	subBlocks := []string{
		" ",      // 0: Empty space
		"\u2581", // 1: Lower 1/8th block
		"\u2582", // 2: Lower 2/8ths block
		"\u2583", // 3: Lower 3/8ths block
		"\u2584", // 4: Lower 4/8ths block
		"\u2585", // 5: Lower 5/8ths block
		"\u2586", // 6: Lower 6/8ths block
		"\u2587", // 7: Lower 7/8ths block
		"\u2588", // 8: Full block
	}

	barWidth := 2
	barSpacing := 1
	step := barWidth + barSpacing
	numBars := width / step
	if numBars <= 0 {
		numBars = 1
	}

	// Determine max range of raw Cava output (default 1000.0)
	maxRange := 1000.0
	if len(bars) > 0 {
		maxVal := 0
		for _, b := range bars {
			if b > maxVal {
				maxVal = b
			}
		}
		if maxVal <= 100 && maxVal > 0 {
			maxRange = 100.0
		}
	}

	// Map raw bars to normalized heights with 1.2x balanced sensitivity gain
	sensGain := 1.2
	maxH := float64(height)
	rawHeights := make([]float64, numBars)
	if len(bars) >= numBars {
		for i := 0; i < numBars; i++ {
			val := (float64(bars[i]) / maxRange) * sensGain * maxH
			if val > maxH {
				val = maxH
			}
			rawHeights[i] = val
		}
	} else if len(bars) > 0 {
		for i := 0; i < numBars; i++ {
			idx := int((float64(i) / float64(numBars)) * float64(len(bars)))
			if idx >= len(bars) {
				idx = len(bars) - 1
			}
			val := (float64(bars[idx]) / maxRange) * sensGain * maxH
			if val > maxH {
				val = maxH
			}
			rawHeights[i] = val
		}
	} else {
		// Dynamic fallback matching noise reduction
		t := float64(tick) * 0.12
		for i := 0; i < numBars; i++ {
			fi := float64(i)
			if isPlaying {
				val := (maxH * 0.65) + (maxH * 0.35 * math.Sin(0.2*fi+t)) + (maxH * 0.25 * math.Cos(0.4*fi-0.7*t))
				if val < 0.2 {
					val = 0.2
				}
				if val > maxH {
					val = maxH
				}
				rawHeights[i] = val
			} else {
				val := 0.4 + 0.3*math.Sin(0.1*fi+0.4*t)
				rawHeights[i] = val
			}
		}
	}

	// 3-tap Gaussian spatial smoothing filter across bars
	barHeights := make([]float64, numBars)
	for i := 0; i < numBars; i++ {
		left := rawHeights[max(0, i-1)]
		mid := rawHeights[i]
		right := rawHeights[min(numBars-1, i+1)]
		barHeights[i] = 0.2*left + 0.6*mid + 0.2*right
	}

	lines := make([]strings.Builder, height)

	// Static vertical gradient colors from bottom row (0) to top row (height-1)
	rowColors := make([]lipgloss.Color, height)
	for r := 0; r < height; r++ {
		rowColors[r] = GetGradientColor(r, height, 0, theme)
	}

	for row := 0; row < height; row++ {
		r := height - 1 - row // 0 is bottom row, height-1 is top row
		rowStyle := lipgloss.NewStyle().Foreground(rowColors[r]).Bold(true)
		cellBottomSub := float64(r * 8)
		cellTopSub := float64((r + 1) * 8)

		var lineBuilder strings.Builder
		for i := 0; i < numBars; i++ {
			subPixels := barHeights[i] * 8.0
			var barSegment string

			if subPixels >= cellTopSub {
				// Completely filled cell: 2 full blocks
				barSegment = "\u2588\u2588"
			} else if subPixels > cellBottomSub {
				// Top peak sub-cell: 2 fractional blocks
				subIdx := int(math.Round(subPixels - cellBottomSub))
				if subIdx < 1 {
					subIdx = 1
				}
				if subIdx > 8 {
					subIdx = 8
				}
				bChar := subBlocks[subIdx]
				barSegment = bChar + bChar
			} else {
				// Empty cell
				barSegment = "  "
			}

			lineBuilder.WriteString(rowStyle.Render(barSegment))
			if i < numBars-1 {
				lineBuilder.WriteString(" ") // 1 space gap between bars
			}
		}

		renderedW := lipgloss.Width(lineBuilder.String())
		if renderedW < width {
			lineBuilder.WriteString(strings.Repeat(" ", width-renderedW))
		}
		lines[row].WriteString(lineBuilder.String())
	}

	var result strings.Builder
	for i := 0; i < height; i++ {
		result.WriteString(lines[i].String())
		if i < height-1 {
			result.WriteString("\n")
		}
	}
	return result.String()
}
