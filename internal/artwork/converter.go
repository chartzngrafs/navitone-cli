package artwork

import (
	"fmt"
	"image"
	"strings"

	"github.com/TheZoraiz/ascii-image-converter/aic_package"
	"navitone-cli/internal/config"
)

// Converter handles ASCII art conversion from images
type Converter struct {
	config *config.Config
}

// QualitySettings defines ASCII art conversion quality parameters
type QualitySettings struct {
	// Character set options
	UseComplex    bool   // Use 69 characters instead of 10 for better detail
	CustomCharMap string // Custom character mapping for optimal contrast
	
	// Dimension and resolution
	Dimensions []int // [width, height] for ASCII art
	
	// Color and visual quality
	UseColor     bool // Enable colored ASCII art
	UseGrayscale bool // Use grayscale instead of monochrome
	UseBraille   bool // Use braille characters for higher resolution
	
	// Post-processing
	MaxHeight int // Maximum height limit for terminal compatibility
}

// NewConverter creates a new artwork converter
func NewConverter(cfg *config.Config) *Converter {
	return &Converter{
		config: cfg,
	}
}

// ConvertFromURL downloads an image from URL and converts it to ASCII art
func (c *Converter) ConvertFromURL(url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("empty URL provided")
	}

	// Get quality settings
	quality := c.getQualitySettings()
	
	// Configure ASCII converter with optimized settings
	flags := aic_package.DefaultFlags()
	
	// QUALITY OPTIMIZATIONS:
	
	// 1. Use complex character set for better detail (69 vs 10 characters)
	flags.Complex = quality.UseComplex
	
	// 2. Optimize dimensions for terminal display
	flags.Dimensions = quality.Dimensions
	
	// 3. Enable colors for better visual quality (if supported)
	flags.Colored = quality.UseColor
	flags.Grayscale = quality.UseGrayscale
	
	// 4. Braille option for high resolution (if supported) 
	flags.Braille = quality.UseBraille
	
	// 5. Custom character mapping for optimal contrast
	if quality.CustomCharMap != "" {
		flags.CustomMap = quality.CustomCharMap
	}
	
	// Convert directly from URL
	asciiArt, err := aic_package.Convert(url, flags)
	if err != nil {
		return "", fmt.Errorf("failed to convert to ASCII: %w", err)
	}

	// Apply post-processing optimizations
	return c.optimizeASCII(asciiArt, quality), nil
}

// ConvertFromBytes converts image bytes to ASCII art
// Note: This implementation would require saving bytes to temp file first
func (c *Converter) ConvertFromBytes(data []byte) (string, error) {
	return "", fmt.Errorf("bytes conversion not implemented - use ConvertFromURL instead")
}

// ConvertFromImage converts an image.Image to ASCII art
func (c *Converter) ConvertFromImage(img image.Image) (string, error) {
	// This would require additional encoding step, implement if needed
	return "", fmt.Errorf("direct image conversion not implemented yet")
}

// getQualitySettings returns optimized quality settings based on config
func (c *Converter) getQualitySettings() QualitySettings {
	// Get size settings
	dimensions := c.getDimensionsForSize(c.config.UI.ArtworkSize)
	
	// Base settings based on quality level
	settings := QualitySettings{
		Dimensions:  dimensions,
		UseColor:    c.config.UI.ArtworkColor,
		MaxHeight:   20, // Keep reasonable for terminal
	}
	
	// Configure quality based on user setting
	switch c.config.UI.ArtworkQuality {
	case "low":
		settings.UseComplex = false    // 10 characters only
		settings.UseGrayscale = false  // Pure B&W
		settings.UseBraille = false
		settings.CustomCharMap = ""    // Use default
		
	case "medium":
		settings.UseComplex = true     // 69 characters for better detail
		settings.UseGrayscale = true   // Better than pure B&W  
		settings.UseBraille = false
		settings.CustomCharMap = ""    // Use default complex set
		
	case "high":
		settings.UseComplex = true     // 69 characters
		settings.UseGrayscale = true
		settings.UseBraille = false    // Keep compatible for now
		// Optimized character mapping for maximum contrast density
		settings.CustomCharMap = "$@B%8&WM#*oahkbdpqwmZO0QLCJUYXzcvunxrjft/\\|()1{}[]?-_+~<>i!lI;:,\"^`'. "
		
	case "ultra":
		settings.UseComplex = true     // 69 characters
		settings.UseGrayscale = true
		settings.UseBraille = true     // High resolution with braille (requires UTF-8)
		// High contrast character mapping
		settings.CustomCharMap = "@@##**++==--::.. "
		settings.MaxHeight = 25        // Allow more height for ultra quality
		
	default:
		// Default to high quality
		settings.UseComplex = true
		settings.UseGrayscale = true  
		settings.UseBraille = false
		settings.CustomCharMap = "$@B%8&WM#*oahkbdpqwmZO0QLCJUYXzcvunxrjft/\\|()1{}[]?-_+~<>i!lI;:,\"^`'. "
	}
	
	return settings
}

// getDimensionsForSize returns appropriate dimensions based on size setting
func (c *Converter) getDimensionsForSize(size string) []int {
	switch size {
	case "small":
		return []int{35, 18}   // Compact for small terminals
	case "medium": 
		return []int{50, 25}   // Balanced detail and space
	case "large":
		return []int{70, 35}   // High detail for large terminals
	default:
		return []int{50, 25}   // Default to medium
	}
}

// optimizeASCII applies post-processing optimizations to improve quality
func (c *Converter) optimizeASCII(ascii string, quality QualitySettings) string {
	lines := strings.Split(ascii, "\n")
	
	// Remove empty lines at start and end
	start := 0
	end := len(lines) - 1
	
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	
	for end >= 0 && strings.TrimSpace(lines[end]) == "" {
		end--
	}
	
	if start > end {
		return "" // All lines were empty
	}
	
	// Keep only the cleaned lines
	cleanedLines := lines[start : end+1]
	
	// Apply height limit from quality settings
	if len(cleanedLines) > quality.MaxHeight {
		cleanedLines = cleanedLines[:quality.MaxHeight]
	}
	
	// TODO: Apply additional post-processing optimizations here:
	// - Contrast enhancement
	// - Character density optimization
	// - Edge sharpening
	
	return strings.Join(cleanedLines, "\n")
}

// GetArtworkSize returns the configured artwork size
func (c *Converter) GetArtworkSize() (width, height int) {
	quality := c.getQualitySettings()
	if len(quality.Dimensions) >= 2 {
		return quality.Dimensions[0], quality.Dimensions[1]
	}
	return 50, 25 // Improved default size
}

// IsEnabled returns whether artwork display is enabled in config
func (c *Converter) IsEnabled() bool {
	return c.config.UI.ShowAlbumArt
}