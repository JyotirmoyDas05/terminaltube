package renderer

import (
	"fmt"
	"image"
	"image/color"
	"strings"
	"terminaltube/pkg/types"

	"github.com/nfnt/resize"
) // ASCIIRenderer implements ASCII-based rendering
type ASCIIRenderer struct {
	mode        types.RenderMode
	grayRamp    string
	initialized bool
}

// NewASCIIRenderer creates a new ASCII renderer
func NewASCIIRenderer(mode types.RenderMode) *ASCIIRenderer {
	return &ASCIIRenderer{
		mode:     mode,
		grayRamp: " .:-=+*#%@",
	}
}

// Name returns the renderer name
func (r *ASCIIRenderer) Name() string {
	return fmt.Sprintf("ASCII_%s", r.mode.String())
}

// SupportsMode checks if this renderer supports the given mode
func (r *ASCIIRenderer) SupportsMode(mode types.RenderMode) bool {
	return mode == types.ASCII_COLOR || mode == types.ASCII_GRAY
}

// Initialize sets up the renderer
func (r *ASCIIRenderer) Initialize() error {
	r.initialized = true
	return nil
}

// Cleanup performs cleanup
func (r *ASCIIRenderer) Cleanup() error {
	r.initialized = false
	return nil
}

// Render converts an image to ASCII representation
func (r *ASCIIRenderer) Render(img image.Image, options types.RenderOptions) (string, error) {
	if !r.initialized {
		return "", fmt.Errorf("renderer not initialized")
	}

	width, height := options.Width, options.Height

	if width == 0 || height == 0 {
		return "", fmt.Errorf("invalid dimensions: width=%d, height=%d", width, height)
	}

	// Resize the image directly to target dimensions
	// Aspect ratio is already handled by the caller
	resizedImg := resize.Resize(uint(width), uint(height), img, resize.Lanczos3)

	// Convert to ASCII
	switch r.mode {
	case types.ASCII_COLOR:
		return r.renderColorASCII(resizedImg, options)
	case types.ASCII_GRAY:
		return r.renderGrayASCII(resizedImg, options)
	default:
		return "", fmt.Errorf("unsupported mode: %s", r.mode.String())
	}
}

// renderColorASCII renders using colored ASCII blocks
func (r *ASCIIRenderer) renderColorASCII(img image.Image, options types.RenderOptions) (string, error) {
	bounds := img.Bounds()
	var result strings.Builder

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			red, green, blue, _ := c.RGBA()

			// Convert to 8-bit values (fix the bit shifting)
			r8, g8, b8 := uint8(red>>8), uint8(green>>8), uint8(blue>>8)

			// Apply brightness and contrast adjustments
			r8 = r.adjustPixel(r8, options.Brightness, options.Contrast)
			g8 = r.adjustPixel(g8, options.Brightness, options.Contrast)
			b8 = r.adjustPixel(b8, options.Brightness, options.Contrast)

			// Convert to ANSI 256-color
			ansiColor := r.rgbToAnsi256(r8, g8, b8)

			// Choose character based on brightness for better visibility
			brightness := (int(r8) + int(g8) + int(b8)) / 3
			char := "█" // Default block
			if brightness < 64 {
				char = "█" // Full block for dark colors
			} else if brightness < 128 {
				char = "▓" // Dark shade
			} else if brightness < 192 {
				char = "▒" // Medium shade
			} else {
				char = "░" // Light shade
			}

			// Use foreground color for better contrast
			result.WriteString(fmt.Sprintf("\033[38;5;%dm%s\033[0m", ansiColor, char))
		}
		result.WriteString("\n")
	}

	return result.String(), nil
}

// renderGrayASCII renders using grayscale ASCII characters
func (r *ASCIIRenderer) renderGrayASCII(img image.Image, options types.RenderOptions) (string, error) {
	bounds := img.Bounds()
	var result strings.Builder

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			gray := color.GrayModel.Convert(c).(color.Gray)

			// Apply brightness and contrast adjustments
			adjusted := r.adjustPixel(gray.Y, options.Brightness, options.Contrast)

			// Map to character ramp
			charIndex := int(float64(adjusted) / 255.0 * float64(len(r.grayRamp)-1))
			if charIndex >= len(r.grayRamp) {
				charIndex = len(r.grayRamp) - 1
			}

			result.WriteByte(r.grayRamp[charIndex])
		}
		result.WriteString("\n")
	}

	return result.String(), nil
}

// adjustPixel applies brightness and contrast adjustments
func (r *ASCIIRenderer) adjustPixel(value uint8, brightness, contrast float64) uint8 {
	// Apply brightness (additive)
	adjusted := float64(value) + brightness*255.0

	// Apply contrast (multiplicative around midpoint)
	adjusted = ((adjusted/255.0-0.5)*contrast + 0.5) * 255.0

	// Clamp to valid range
	if adjusted < 0 {
		adjusted = 0
	} else if adjusted > 255 {
		adjusted = 255
	}

	return uint8(adjusted)
}

// rgbToAnsi256 converts RGB values to the closest ANSI 256-color code
func (r *ASCIIRenderer) rgbToAnsi256(red, green, blue uint8) int {
	// Handle grayscale colors (232-255 are grayscale)
	if red == green && green == blue {
		// Map to grayscale range (232-255, 24 levels)
		if red < 8 {
			return 16 // Black
		} else if red > 248 {
			return 231 // White (standard colors)
		} else {
			// Map to grayscale palette (232-255)
			return 232 + int((red-8)*23/240)
		}
	}

	// Convert RGB to 6x6x6 color cube (colors 16-231)
	// Each component is mapped to 0-5 range with better rounding
	r6 := int((red*5 + 127) / 255)
	g6 := int((green*5 + 127) / 255)
	b6 := int((blue*5 + 127) / 255)

	// Clamp values to valid range
	if r6 > 5 {
		r6 = 5
	}
	if g6 > 5 {
		g6 = 5
	}
	if b6 > 5 {
		b6 = 5
	}

	// Calculate ANSI color code
	return 16 + (36 * r6) + (6 * g6) + b6
}
