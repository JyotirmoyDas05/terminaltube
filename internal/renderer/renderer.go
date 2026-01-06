package renderer

import (
	"image"
	"terminaltube/pkg/types"
)

// Renderer defines the interface for different rendering backends
type Renderer interface {
	// Render converts an image to a string representation for terminal display
	Render(img image.Image, options types.RenderOptions) (string, error)

	// SupportsMode returns true if this renderer can handle the specified mode
	SupportsMode(mode types.RenderMode) bool

	// Name returns the name of this renderer
	Name() string

	// Initialize performs any necessary setup for the renderer
	Initialize() error

	// Cleanup performs any necessary cleanup
	Cleanup() error
}

// RendererManager manages different rendering backends
type RendererManager struct {
	renderers    map[types.RenderMode]Renderer
	capabilities types.TerminalCapabilities
}

// NewRendererManager creates a new renderer manager
func NewRendererManager(capabilities types.TerminalCapabilities) *RendererManager {
	rm := &RendererManager{
		renderers:    make(map[types.RenderMode]Renderer),
		capabilities: capabilities,
	}

	// Register available renderers
	rm.registerRenderers()

	return rm
}

// registerRenderers registers all available renderers based on terminal capabilities
func (rm *RendererManager) registerRenderers() {
	// Always register ASCII renderers as fallbacks
	rm.renderers[types.ASCII_GRAY] = NewASCIIRenderer(types.ASCII_GRAY)
	rm.renderers[types.ASCII_COLOR] = NewASCIIRenderer(types.ASCII_COLOR)

	// Register true-color renderer if supported
	if rm.capabilities.TrueColor {
		rm.renderers[types.EXACT] = NewUnicodeRenderer()
	}

	// Register SIXEL renderer if supported
	if rm.capabilities.SixelSupport {
		rm.renderers[types.SIXEL] = NewSixelRenderer()
	}
}

// GetRenderer returns the appropriate renderer for the given mode
func (rm *RendererManager) GetRenderer(mode types.RenderMode) (Renderer, bool) {
	renderer, exists := rm.renderers[mode]
	return renderer, exists
}

// GetBestRenderer returns the best available renderer for the terminal
func (rm *RendererManager) GetBestRenderer() Renderer {
	// Preference order: SIXEL > EXACT > ASCII_COLOR > ASCII_GRAY
	modes := []types.RenderMode{types.SIXEL, types.EXACT, types.ASCII_COLOR, types.ASCII_GRAY}

	for _, mode := range modes {
		if renderer, exists := rm.renderers[mode]; exists {
			return renderer
		}
	}

	// Fallback to ASCII_GRAY (should always be available)
	return rm.renderers[types.ASCII_GRAY]
}

// GetAvailableModes returns all available rendering modes
func (rm *RendererManager) GetAvailableModes() []types.RenderMode {
	var modes []types.RenderMode
	for mode := range rm.renderers {
		modes = append(modes, mode)
	}
	return modes
}
