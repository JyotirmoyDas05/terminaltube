package renderer

import (
	"fmt"
	"image"
	"strings"
	"terminaltube/pkg/types"

	"github.com/nfnt/resize"
)

// SixelRenderer implements SIXEL graphics rendering with optimized color palette
// Uses 8-8-4 RGB distribution (256 colors) for better color accuracy
// Human eyes are more sensitive to red and green, less to blue
type SixelRenderer struct {
	initialized bool
	palette     string // Pre-generated palette string
}

// NewSixelRenderer creates a new SIXEL renderer
func NewSixelRenderer() *SixelRenderer {
	return &SixelRenderer{}
}

// Name returns the renderer name
func (r *SixelRenderer) Name() string {
	return "SIXEL"
}

// SupportsMode checks if this renderer supports the given mode
func (r *SixelRenderer) SupportsMode(mode types.RenderMode) bool {
	return mode == types.SIXEL
}

// Initialize sets up the renderer with pre-computed palette
func (r *SixelRenderer) Initialize() error {
	r.palette = r.generatePalette()
	r.initialized = true
	return nil
}

// Cleanup performs cleanup
func (r *SixelRenderer) Cleanup() error {
	r.initialized = false
	return nil
}

// generatePalette creates an 8-8-4 RGB palette (256 colors)
// 8 levels for R, 8 levels for G, 4 levels for B
// This matches human vision sensitivity (more sensitive to R/G than B)
func (r *SixelRenderer) generatePalette() string {
	var sb strings.Builder

	// 8-8-4 RGB color cube = 256 colors
	for rLevel := 0; rLevel < 8; rLevel++ {
		for gLevel := 0; gLevel < 8; gLevel++ {
			for bLevel := 0; bLevel < 4; bLevel++ {
				idx := rLevel*32 + gLevel*4 + bLevel
				// Convert levels to 0-100 percentage with gamma correction
				rPct := gammaCorrect(rLevel, 7)
				gPct := gammaCorrect(gLevel, 7)
				bPct := gammaCorrect(bLevel, 3)
				sb.WriteString(fmt.Sprintf("#%d;2;%d;%d;%d", idx, rPct, gPct, bPct))
			}
		}
	}

	return sb.String()
}

// gammaCorrect applies gamma correction for better perceptual color distribution
func gammaCorrect(level, maxLevel int) int {
	// Use gamma 2.2 for more perceptually uniform distribution
	normalized := float64(level) / float64(maxLevel)
	// Apply gamma and convert to percentage
	corrected := normalized * normalized // Simple gamma approximation
	return int(corrected * 100)
}

// colorToPalette converts RGB to 8-8-4 palette index with improved accuracy
func colorToPalette(r, g, b uint8) uint8 {
	// Apply inverse gamma for better color matching
	rNorm := float64(r) / 255.0
	gNorm := float64(g) / 255.0
	bNorm := float64(b) / 255.0

	// Take square root (inverse of gamma 2.0) for perceptual mapping
	rLevel := int(sqrt(rNorm) * 7.99)
	gLevel := int(sqrt(gNorm) * 7.99)
	bLevel := int(sqrt(bNorm) * 3.99)

	// Clamp to valid ranges
	if rLevel > 7 {
		rLevel = 7
	}
	if gLevel > 7 {
		gLevel = 7
	}
	if bLevel > 3 {
		bLevel = 3
	}

	return uint8(rLevel*32 + gLevel*4 + bLevel)
}

// sqrt computes square root for color correction
func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	// Newton-Raphson square root
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

// Render converts an image to SIXEL representation
func (r *SixelRenderer) Render(img image.Image, options types.RenderOptions) (string, error) {
	if !r.initialized {
		return "", fmt.Errorf("renderer not initialized")
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// If image is very large, resize it
	maxPixels := 1200 * 800
	if width*height > maxPixels {
		targetWidth := options.Width * 8
		targetHeight := options.Height * 16
		img = resize.Resize(uint(targetWidth), uint(targetHeight), img, resize.Bilinear)
		bounds = img.Bounds()
		width = bounds.Dx()
		height = bounds.Dy()
	}

	// Ensure height is divisible by 6 (required for SIXEL)
	height6 := ((height + 5) / 6) * 6

	// Convert image to palette indices
	pixels := make([]uint8, width*height6)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := img.At(bounds.Min.X+x, bounds.Min.Y+y)
			rv, gv, bv, _ := c.RGBA()
			pixels[y*width+x] = colorToPalette(uint8(rv>>8), uint8(gv>>8), uint8(bv>>8))
		}
	}
	// Fill remaining rows with black (index 0)
	for y := height; y < height6; y++ {
		for x := 0; x < width; x++ {
			pixels[y*width+x] = 0
		}
	}

	return r.encodeSixel(pixels, width, height6)
}

// encodeSixel creates the SIXEL escape sequence
func (r *SixelRenderer) encodeSixel(pixels []uint8, width, height int) (string, error) {
	var sb strings.Builder
	sb.Grow(width * height / 2) // Rough estimate

	// SIXEL header: ESC P 7;1;q (7=800dpi aspect, 1=transparent background)
	sb.WriteString("\x1bP7;1;q")

	// Write pre-computed palette
	sb.WriteString(r.palette)

	// Process in 6-row bands
	usedColors := make([]bool, 256)

	for y6 := 0; y6 < height; y6 += 6 {
		// Find colors used in this band
		for i := range usedColors {
			usedColors[i] = false
		}
		for dy := 0; dy < 6 && y6+dy < height; dy++ {
			rowOff := (y6 + dy) * width
			for x := 0; x < width; x++ {
				usedColors[pixels[rowOff+x]] = true
			}
		}

		// Process each used color
		for colorIdx := 0; colorIdx < 256; colorIdx++ {
			if !usedColors[colorIdx] {
				continue
			}

			// Select color
			sb.WriteString(fmt.Sprintf("#%d", colorIdx))

			// Generate sixel data for this color
			var sixelReps int
			var repeatedSixel byte = 63 // '?' = empty sixel

			for x := 0; x < width; x++ {
				// Build 6-bit sixel value
				var sixel byte = 0
				for dy := 0; dy < 6; dy++ {
					if y6+dy < height {
						if pixels[(y6+dy)*width+x] == uint8(colorIdx) {
							sixel |= 1 << dy
						}
					}
				}
				sixelChar := byte(63 + sixel)

				// Run-length encoding
				if sixelChar == repeatedSixel {
					sixelReps++
				} else {
					// Output previous run
					r.writeRLE(&sb, repeatedSixel, sixelReps)
					repeatedSixel = sixelChar
					sixelReps = 1
				}
			}

			// Output final run
			r.writeRLE(&sb, repeatedSixel, sixelReps)

			// Carriage return (return to start of band)
			sb.WriteByte('$')
		}

		// Line feed (move to next 6-row band)
		sb.WriteByte('-')
	}

	// SIXEL terminator
	sb.WriteString("\x1b\\")

	return sb.String(), nil
}

// writeRLE writes run-length encoded data
func (r *SixelRenderer) writeRLE(sb *strings.Builder, char byte, count int) {
	if count == 0 {
		return
	}
	if count > 3 {
		sb.WriteString(fmt.Sprintf("!%d%c", count, char))
	} else {
		for i := 0; i < count; i++ {
			sb.WriteByte(char)
		}
	}
}
