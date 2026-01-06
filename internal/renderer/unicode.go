package renderer

import (
	"fmt"
	"image"
	"image/color"
	"strings"
	"terminaltube/pkg/types"

	"github.com/nfnt/resize"
)

// UnicodeRenderer implements true-color Unicode block rendering
type UnicodeRenderer struct {
	initialized bool
}

// NewUnicodeRenderer creates a new Unicode renderer
func NewUnicodeRenderer() *UnicodeRenderer {
	return &UnicodeRenderer{}
}

// Name returns the renderer name
func (r *UnicodeRenderer) Name() string {
	return "Unicode_TrueColor"
}

// SupportsMode checks if this renderer supports the given mode
func (r *UnicodeRenderer) SupportsMode(mode types.RenderMode) bool {
	return mode == types.EXACT
}

// Initialize sets up the renderer
func (r *UnicodeRenderer) Initialize() error {
	r.initialized = true
	return nil
}

// Cleanup performs cleanup
func (r *UnicodeRenderer) Cleanup() error {
	r.initialized = false
	return nil
}

// Render converts an image to Unicode block representation with true color
func (r *UnicodeRenderer) Render(img image.Image, options types.RenderOptions) (string, error) {
	if !r.initialized {
		return "", fmt.Errorf("renderer not initialized")
	}

	width, height := options.Width, options.Height

	if width == 0 || height == 0 {
		return "", fmt.Errorf("invalid dimensions: width=%d, height=%d", width, height)
	}

	// For Unicode half-blocks (▀), each character displays 2 vertical pixels
	// So we need height*2 pixels vertically to get 'height' rows of output
	pixelHeight := height * 2

	// Resize the image to target dimensions
	// Width stays the same (1 pixel = 1 character width)
	// Height is doubled because each character is 2 pixels tall
	resizedImg := resize.Resize(uint(width), uint(pixelHeight), img, resize.Lanczos3)

	return r.renderTrueColorUnicode(resizedImg, width, height)
}

// renderTrueColorUnicode renders using Unicode half-blocks with true color
func (r *UnicodeRenderer) renderTrueColorUnicode(img image.Image, targetWidth, targetHeight int) (string, error) {
	bounds := img.Bounds()
	var result strings.Builder

	// Pre-allocate approximate size for efficiency
	result.Grow(targetWidth * targetHeight * 40)

	// Process pairs of rows (top and bottom half-blocks)
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 2 {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Get colors for top and bottom pixels
			topColor := img.At(x, y)
			var bottomColor color.Color
			if y+1 < bounds.Max.Y {
				bottomColor = img.At(x, y+1)
			} else {
				bottomColor = color.RGBA{0, 0, 0, 255}
			}

			// Convert to RGB
			topR, topG, topB := r.colorToRGB(topColor)
			bottomR, bottomG, bottomB := r.colorToRGB(bottomColor)

			// Use upper half block (▀) with foreground as top color and background as bottom color
			result.WriteString(fmt.Sprintf("\033[38;2;%d;%d;%dm\033[48;2;%d;%d;%dm▀",
				topR, topG, topB, bottomR, bottomG, bottomB))
		}
		// Reset colors and newline
		result.WriteString("\033[0m\n")
	}

	return result.String(), nil
}

// colorToRGB converts a color to RGB values
func (r *UnicodeRenderer) colorToRGB(c color.Color) (uint8, uint8, uint8) {
	red, green, blue, _ := c.RGBA()
	return uint8(red >> 8), uint8(green >> 8), uint8(blue >> 8)
}
