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
	if len(theme.Palette) >= 2 {
		p := make([]lipgloss.Color, len(theme.Palette))
		copy(p, theme.Palette)
		// Ensure seamless loop back to initial color
		if p[0] != p[len(p)-1] {
			p = append(p, p[0])
		}
		return p
	}

	// Standard multi-color palette cycle across key Omarchy theme colors
	palette := []lipgloss.Color{
		theme.Primary,
		theme.Secondary,
		theme.Accent,
		theme.Success,
		theme.Warning,
	}

	var clean []lipgloss.Color
	for _, c := range palette {
		if string(c) != "" && (len(clean) == 0 || clean[len(clean)-1] != c) {
			clean = append(clean, c)
		}
	}
	if len(clean) < 2 {
		return []lipgloss.Color{
			lipgloss.Color("#8BE9FD"),
			lipgloss.Color("#FF79C6"),
			lipgloss.Color("#BD93F9"),
			lipgloss.Color("#50FA7B"),
			lipgloss.Color("#F1FA8C"),
			lipgloss.Color("#8BE9FD"),
		}
	}
	if clean[0] != clean[len(clean)-1] {
		clean = append(clean, clean[0])
	}
	return clean
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

// GetGradientColorFloat calculates continuous TrueColor RGB interpolation with floating-point precision
func GetGradientColorFloat(step, totalSteps, tick float64, theme Theme) lipgloss.Color {
	if totalSteps <= 0 {
		totalSteps = 1.0
	}

	palette := OmarchyPalette(theme)
	if len(palette) < 2 {
		return theme.Primary
	}

	numSegments := float64(len(palette) - 1)

	// Smooth sub-frame animation offset (~25 seconds per full rotation)
	animOffset := tick * 0.03

	// Continuous normalized position [0.0, 1.0)
	normPos := math.Mod((step+animOffset)/totalSteps, 1.0)
	if normPos < 0 {
		normPos += 1.0
	}

	scaled := normPos * numSegments
	segmentIdx := int(scaled)
	if segmentIdx >= len(palette)-1 {
		segmentIdx = len(palette) - 2
	}
	if segmentIdx < 0 {
		segmentIdx = 0
	}
	factor := scaled - float64(segmentIdx)

	// Smoothstep S-curve for ultra-smooth continuous color blending
	smoothFactor := factor * factor * (3.0 - 2.0*factor)

	nextIdx := segmentIdx + 1
	if nextIdx >= len(palette) {
		nextIdx = len(palette) - 1
	}

	return interpolateRGB(palette[segmentIdx], palette[nextIdx], smoothFactor)
}

// GetGradientColor wraps GetGradientColorFloat for integer inputs
func GetGradientColor(step, totalSteps, tick int, theme Theme) lipgloss.Color {
	return GetGradientColorFloat(float64(step), float64(totalSteps), float64(tick), theme)
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

	// Split content into lines and ensure exactly innerHeight lines with exact innerWidth width
	rawLines := strings.Split(content, "\n")
	contentLines := make([]string, innerHeight)
	for i := 0; i < innerHeight; i++ {
		if i < len(rawLines) {
			line := rawLines[i]
			w := lipgloss.Width(line)
			if w > innerWidth {
				line = lipgloss.NewStyle().MaxWidth(innerWidth).Render(line)
				w = lipgloss.Width(line)
			}
			if w < innerWidth {
				line += strings.Repeat(" ", innerWidth-w)
			}
			contentLines[i] = line
		} else {
			contentLines[i] = strings.Repeat(" ", innerWidth)
		}
	}

	totalPerimeter := 2*width + 2*innerHeight
	step := 0

	var builder strings.Builder

	// Top Border Line
	cTopLeft := GetGradientColor(step, totalPerimeter, tick, theme)
	builder.WriteString(lipgloss.NewStyle().Foreground(cTopLeft).Render("╭"))
	step++

	for x := 0; x < innerWidth; x++ {
		c := GetGradientColor(step, totalPerimeter, tick, theme)
		builder.WriteString(lipgloss.NewStyle().Foreground(c).Render("─"))
		step++
	}

	cTopRight := GetGradientColor(step, totalPerimeter, tick, theme)
	builder.WriteString(lipgloss.NewStyle().Foreground(cTopRight).Render("╮"))
	step++
	builder.WriteString("\n")

	// Middle Content Lines with Left/Right Borders
	leftStepStart := totalPerimeter - 1

	for i := 0; i < innerHeight; i++ {
		cLeft := GetGradientColor(leftStepStart-i, totalPerimeter, tick, theme)
		builder.WriteString(lipgloss.NewStyle().Foreground(cLeft).Render("│"))

		builder.WriteString(contentLines[i])

		cRight := GetGradientColor(step, totalPerimeter, tick, theme)
		builder.WriteString(lipgloss.NewStyle().Foreground(cRight).Render("│"))
		step++

		builder.WriteString("\n")
	}

	// Bottom Border Line
	cBottomLeft := GetGradientColor(leftStepStart-innerHeight, totalPerimeter, tick, theme)
	builder.WriteString(lipgloss.NewStyle().Foreground(cBottomLeft).Render("╰"))

	bottomStepStart := leftStepStart - innerHeight - 1
	for x := 0; x < innerWidth; x++ {
		c := GetGradientColor(bottomStepStart-x, totalPerimeter, tick, theme)
		builder.WriteString(lipgloss.NewStyle().Foreground(c).Render("─"))
	}

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

	gradColor := GetGradientColorFloat(0, 60, float64(tick)*0.5, theme)
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
	totalSteps := 60.0
	tickF := float64(tick) * 0.5

	var builder strings.Builder
	for i, r := range runes {
		c := GetGradientColorFloat(float64(i), totalSteps, tickF, theme)
		builder.WriteString(lipgloss.NewStyle().Foreground(c).Bold(true).Render(string(r)))
	}
	return builder.String()
}

// GetCavaGradientColorByPos calculates smooth vertical gradient colors by normalized position [0.0, 1.0]
func GetCavaGradientColorByPos(normPos float64, theme Theme) lipgloss.Color {
	normPos = math.Max(0, math.Min(1, normPos))

	palette := theme.CavaPalette
	if len(palette) < 2 {
		palette = OmarchyPalette(theme)
	}

	numColors := len(palette)
	if numColors == 0 {
		return theme.Primary
	}
	if numColors == 1 {
		return palette[0]
	}

	scaled := normPos * float64(numColors-1)
	idx := int(scaled)
	if idx >= numColors-1 {
		idx = numColors - 2
	}
	if idx < 0 {
		idx = 0
	}
	factor := scaled - float64(idx)

	// Smoothstep S-curve for silky smooth Cava color transitions between rows
	smoothFactor := factor * factor * (3.0 - 2.0*factor)

	nextIdx := idx + 1
	if nextIdx >= numColors {
		nextIdx = numColors - 1
	}

	return interpolateRGB(palette[idx], palette[nextIdx], smoothFactor)
}

// GetCavaGradientColor calculates vertical gradient colors for Cava spectrum bars
func GetCavaGradientColor(row, totalRows int, theme Theme) lipgloss.Color {
	if totalRows <= 0 {
		totalRows = 1
	}
	if totalRows == 1 {
		return GetCavaGradientColorByPos(0.5, theme)
	}
	normPos := float64(row) / float64(totalRows-1)
	return GetCavaGradientColorByPos(normPos, theme)
}

// RenderInAppCava renders an integrated multi-bar Cava audio spectrum visualizer matching system Cava config
func RenderInAppCava(width, height int, bars []int, tick int, isPlaying bool, theme Theme) string {
	if width <= 0 || height <= 0 {
		return ""
	}

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

	maxRange := 1000.0
	if len(bars) > 0 {
		maxVal := 0
		for _, b := range bars {
			if b > maxVal {
				maxVal = b
			}
		}
		if float64(maxVal) > maxRange {
			maxRange = float64(maxVal)
		}
	}

	sensGain := 1.2
	maxH := float64(height)
	rawHeights := make([]float64, numBars)
	if len(bars) > 0 {
		for i := 0; i < numBars; i++ {
			var val float64
			if len(bars) == 1 {
				val = (float64(bars[0]) / maxRange) * sensGain * maxH
			} else {
				pos := float64(i) * float64(len(bars)-1) / float64(max(1, numBars-1))
				idx := int(pos)
				frac := pos - float64(idx)
				if idx >= len(bars)-1 {
					val = (float64(bars[len(bars)-1]) / maxRange) * sensGain * maxH
				} else {
					interpolated := (1.0-frac)*float64(bars[idx]) + frac*float64(bars[idx+1])
					val = (interpolated / maxRange) * sensGain * maxH
				}
			}
			if val > maxH {
				val = maxH
			}
			rawHeights[i] = val
		}
	} else {
		t := float64(tick) * 0.08
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

	barHeights := make([]float64, numBars)
	for i := 0; i < numBars; i++ {
		left := rawHeights[max(0, i-1)]
		mid := rawHeights[i]
		right := rawHeights[min(numBars-1, i+1)]
		barHeights[i] = 0.15*left + 0.70*mid + 0.15*right
	}

	lines := make([]strings.Builder, height)

	// Continuous vertical gradient color interpolation per row
	rowColors := make([]lipgloss.Color, height)
	for r := 0; r < height; r++ {
		midPos := (float64(r) + 0.5) / float64(height)
		rowColors[r] = GetCavaGradientColorByPos(midPos, theme)
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
				barSegment = "\u2588\u2588"
			} else if subPixels > cellBottomSub {
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
				barSegment = "  "
			}

			lineBuilder.WriteString(rowStyle.Render(barSegment))
			if i < numBars-1 {
				lineBuilder.WriteString(" ")
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
