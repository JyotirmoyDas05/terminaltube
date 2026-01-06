package types

import "image"

// RenderMode defines the different rendering modes available
type RenderMode int

const (
	// SIXEL uses high-quality SIXEL graphics for terminals that support it
	SIXEL RenderMode = iota
	// ASCII_COLOR uses block-based color rendering with ANSI codes
	ASCII_COLOR
	// ASCII_GRAY uses character-based monochrome rendering
	ASCII_GRAY
	// EXACT uses true-color terminal rendering with Unicode blocks
	EXACT
)

// String returns the string representation of the render mode
func (r RenderMode) String() string {
	switch r {
	case SIXEL:
		return "SIXEL"
	case ASCII_COLOR:
		return "ASCII_COLOR"
	case ASCII_GRAY:
		return "ASCII_GRAY"
	case EXACT:
		return "EXACT"
	default:
		return "UNKNOWN"
	}
}

// RenderOptions contains configuration for media rendering
type RenderOptions struct {
	// Width and Height of the output (0 = auto-detect from terminal)
	Width  int
	Height int

	// Mode specifies the rendering mode to use
	Mode RenderMode

	// PaletteSize for SIXEL rendering (default 256)
	PaletteSize int

	// AspectRatio preservation flag
	PreserveAspectRatio bool

	// Image adjustments
	Contrast   float64 // Contrast adjustment (1.0 = no change)
	Brightness float64 // Brightness adjustment (0.0 = no change)

	// TerminalAspectRatio accounts for character cell dimensions (default 0.5)
	TerminalAspectRatio float64
}

// DefaultRenderOptions returns sensible default render options
func DefaultRenderOptions() RenderOptions {
	return RenderOptions{
		Width:               0,           // Auto-detect
		Height:              0,           // Auto-detect
		Mode:                ASCII_COLOR, // Safe fallback
		PaletteSize:         256,
		PreserveAspectRatio: true,
		Contrast:            1.0,
		Brightness:          0.0,
		TerminalAspectRatio: 0.5,
	}
}

// MediaType represents the type of media being processed
type MediaType int

const (
	IMAGE MediaType = iota
	GIF
	VIDEO
)

// MediaInfo contains information about a media file
type MediaInfo struct {
	Type       MediaType
	Width      int
	Height     int
	FPS        float64 // For video/GIF
	Duration   float64 // In seconds
	HasAudio   bool
	AudioCodec string
	VideoCodec string
	FrameCount int
}

// Frame represents a single frame of media content
type Frame struct {
	Image     image.Image
	Timestamp float64 // Time in seconds
	Duration  float64 // Frame duration in seconds
}

// PlaybackStats tracks playback performance
type PlaybackStats struct {
	FramesRendered int
	FramesDropped  int
	DropRate       float64
	FPS            float64
	StartTime      int64 // Unix timestamp in nanoseconds
}

// TerminalCapabilities represents what the terminal supports
type TerminalCapabilities struct {
	SixelSupport   bool
	TrueColor      bool
	Color256       bool
	Width          int
	Height         int
	UnicodeSupport bool
}
