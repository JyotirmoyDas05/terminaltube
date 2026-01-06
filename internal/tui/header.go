package tui

import (
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ASCII Art placeholder - user will provide the actual art
// This is a sample that can be replaced
var ASCIIArt = `

████████╗███████╗██████╗ ███╗   ███╗██╗███╗   ██╗ █████╗ ██╗  ████████╗██╗   ██╗██████╗ ███████╗
╚══██╔══╝██╔════╝██╔══██╗████╗ ████║██║████╗  ██║██╔══██╗██║  ╚══██╔══╝██║   ██║██╔══██╗██╔════╝
   ██║   █████╗  ██████╔╝██╔████╔██║██║██╔██╗ ██║███████║██║     ██║   ██║   ██║██████╔╝█████╗  
   ██║   ██╔══╝  ██╔══██╗██║╚██╔╝██║██║██║╚██╗██║██╔══██║██║     ██║   ██║   ██║██╔══██╗██╔══╝  
   ██║   ███████╗██║  ██║██║ ╚═╝ ██║██║██║ ╚████║██║  ██║███████╗██║   ╚██████╔╝██████╔╝███████╗
   ╚═╝   ╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝╚══════╝╚═╝    ╚═════╝ ╚═════╝ ╚══════╝
`

// SetASCIIArt allows updating the ASCII art header
func SetASCIIArt(art string) {
	ASCIIArt = art
}

// Animation characters for scramble effect
const scrambleChars = "!@#$%^&*()_+-=[]{}|;':\",./<>?ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789█▓▒░"

// HeaderAnimation holds the state for the animated header
type HeaderAnimation struct {
	text           string
	lines          []string
	revealedChars  int
	totalChars     int
	scrambleOffset int
	done           bool
	colorOffset    int

	// Animation speed settings
	charsPerTick   int
	scrambleFrames int
}

// NewHeaderAnimation creates a new header animation with the ASCII art
func NewHeaderAnimation(art string) HeaderAnimation {
	lines := strings.Split(strings.TrimSpace(art), "\n")
	totalChars := 0
	for _, line := range lines {
		totalChars += len([]rune(line))
	}

	return HeaderAnimation{
		text:           art,
		lines:          lines,
		revealedChars:  0,
		totalChars:     totalChars,
		scrambleOffset: 0,
		done:           false,
		colorOffset:    0,
		charsPerTick:   8, // Characters revealed per tick
		scrambleFrames: 3, // Scramble frames before revealing
	}
}

// HeaderAnimationTickMsg is sent to update the animation
type HeaderAnimationTickMsg struct{}

// HeaderAnimationCmd returns a command that ticks the animation
func HeaderAnimationCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*30, func(t time.Time) tea.Msg {
		return HeaderAnimationTickMsg{}
	})
}

// ColorCycleTickMsg is sent to cycle colors after animation completes
type ColorCycleTickMsg struct{}

// ColorCycleCmd returns a command that cycles colors for the gradient effect
func ColorCycleCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return ColorCycleTickMsg{}
	})
}

// Update advances the animation state
func (h *HeaderAnimation) Update(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case HeaderAnimationTickMsg:
		if !h.done {
			h.revealedChars += h.charsPerTick
			h.scrambleOffset++
			if h.revealedChars >= h.totalChars {
				h.revealedChars = h.totalChars
				h.done = true
				return ColorCycleCmd()
			}
			return HeaderAnimationCmd()
		}
	case ColorCycleTickMsg:
		if h.done {
			h.colorOffset++
			return ColorCycleCmd()
		}
	}
	return nil
}

// View renders the current animation frame
func (h *HeaderAnimation) View() string {
	if h.done {
		// Animation complete - render with gradient colors
		return h.renderWithGradient()
	}

	// Animation in progress - render with scramble effect
	return h.renderWithScramble()
}

// renderWithScramble renders the text with scramble effect on unrevealed characters
func (h *HeaderAnimation) renderWithScramble() string {
	var result strings.Builder
	charCount := 0

	for lineIdx, line := range h.lines {
		runes := []rune(line)
		for i, r := range runes {
			if charCount < h.revealedChars {
				// Revealed character - apply gradient color
				colorIdx := (charCount + h.colorOffset) % len(MediaGradient)
				style := lipgloss.NewStyle().Foreground(MediaGradient[colorIdx]).Bold(true)
				result.WriteString(style.Render(string(r)))
			} else if charCount < h.revealedChars+10 {
				// Scramble zone - show random characters
				if r == ' ' || r == '\n' {
					result.WriteRune(r)
				} else {
					scrambleIdx := (charCount + h.scrambleOffset) % len(scrambleChars)
					style := lipgloss.NewStyle().Foreground(SubtleColor)
					result.WriteString(style.Render(string(scrambleChars[scrambleIdx])))
				}
			} else {
				// Not yet reached - show dim placeholder
				if r == ' ' || r == '\n' {
					result.WriteRune(r)
				} else {
					style := lipgloss.NewStyle().Foreground(lipgloss.Color("#333333"))
					result.WriteString(style.Render("░"))
				}
			}
			charCount++
			_ = i // suppress unused variable warning
		}
		if lineIdx < len(h.lines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// renderWithGradient renders the fully revealed text with cycling gradient
func (h *HeaderAnimation) renderWithGradient() string {
	var result strings.Builder
	charCount := 0

	for lineIdx, line := range h.lines {
		runes := []rune(line)
		for _, r := range runes {
			if r == ' ' || r == '\n' {
				result.WriteRune(r)
			} else {
				// Cycling gradient effect
				colorIdx := (charCount + h.colorOffset) % len(MediaGradient)
				style := lipgloss.NewStyle().Foreground(MediaGradient[colorIdx]).Bold(true)
				result.WriteString(style.Render(string(r)))
			}
			charCount++
		}
		if lineIdx < len(h.lines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// IsDone returns whether the animation has completed
func (h *HeaderAnimation) IsDone() bool {
	return h.done
}

// Reset restarts the animation
func (h *HeaderAnimation) Reset() {
	h.revealedChars = 0
	h.scrambleOffset = 0
	h.done = false
	h.colorOffset = 0
}

// SkipAnimation immediately completes the animation
func (h *HeaderAnimation) SkipAnimation() {
	h.revealedChars = h.totalChars
	h.done = true
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
