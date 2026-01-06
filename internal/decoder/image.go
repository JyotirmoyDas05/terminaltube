package decoder

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"terminaltube/pkg/types"
)

// ImageDecoder handles static image decoding
type ImageDecoder struct {
	supportedFormats []string
}

// NewImageDecoder creates a new image decoder
func NewImageDecoder() *ImageDecoder {
	return &ImageDecoder{
		supportedFormats: []string{".jpg", ".jpeg", ".png", ".bmp", ".gif"},
	}
}

// IsSupported checks if the file extension is supported
func (d *ImageDecoder) IsSupported(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, format := range d.supportedFormats {
		if ext == format {
			return true
		}
	}
	return false
}

// DecodeImage decodes a static image file
func (d *ImageDecoder) DecodeImage(filename string) (image.Image, *types.MediaInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	// Decode the image
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Get image dimensions
	bounds := img.Bounds()

	// Create media info
	mediaInfo := &types.MediaInfo{
		Type:       types.IMAGE,
		Width:      bounds.Dx(),
		Height:     bounds.Dy(),
		FPS:        0, // Static image
		Duration:   0, // Static image
		HasAudio:   false,
		AudioCodec: "",
		VideoCodec: format,
		FrameCount: 1,
	}

	return img, mediaInfo, nil
}

// GetSupportedFormats returns the list of supported image formats
func (d *ImageDecoder) GetSupportedFormats() []string {
	return d.supportedFormats
}
