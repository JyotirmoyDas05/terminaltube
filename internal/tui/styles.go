package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette - modern gradient purples/cyans for media player aesthetic
var (
	// Primary gradient colors
	PrimaryColor   = lipgloss.Color("#BD93F9")
	SecondaryColor = lipgloss.Color("#8BE9FD")
	AccentColor    = lipgloss.Color("#FF79C6")
	SuccessColor   = lipgloss.Color("#50FA7B")
	WarningColor   = lipgloss.Color("#FFB86C")
	ErrorColor     = lipgloss.Color("#FF5555")

	// Text colors
	TextColor      = lipgloss.Color("#F8F8F2")
	SubtleColor    = lipgloss.Color("#6272A4")
	HighlightColor = lipgloss.Color("#F1FA8C")

	// Background colors
	BackgroundColor = lipgloss.Color("#282A36")
	SurfaceColor    = lipgloss.Color("#44475A")
)

// Styles for the TUI

// Title style for the main header
var TitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(PrimaryColor)

// Gradient title with animation effect
var GradientTitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#BD93F9"))

// Subtitle style
var SubtitleStyle = lipgloss.NewStyle().
	Foreground(SubtleColor).
	Italic(true)

// Menu item styles
var (
	MenuItemStyle = lipgloss.NewStyle().
			Foreground(TextColor).
			PaddingLeft(2)

	SelectedMenuItemStyle = lipgloss.NewStyle().
				Foreground(HighlightColor).
				Bold(true).
				PaddingLeft(2)

	MenuCursorStyle = lipgloss.NewStyle().
			Foreground(AccentColor).
			Bold(true)
)

// Badge styles for capability indicators
var (
	BadgeEnabledStyle = lipgloss.NewStyle().
				Foreground(SuccessColor).
				Bold(true)

	BadgeDisabledStyle = lipgloss.NewStyle().
				Foreground(SubtleColor)
)

// Box styles
var (
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor).
			Padding(1, 2)

	HeaderBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(PrimaryColor).
			Padding(0, 2).
			Align(lipgloss.Center)

	FooterStyle = lipgloss.NewStyle().
			Foreground(SubtleColor).
			Align(lipgloss.Center)
)

// Status bar style
var StatusBarStyle = lipgloss.NewStyle().
	Foreground(TextColor).
	Background(SurfaceColor).
	Padding(0, 1)

// Input styles
var (
	InputStyle = lipgloss.NewStyle().
			Foreground(TextColor).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(SecondaryColor).
			Padding(0, 1)

	InputPromptStyle = lipgloss.NewStyle().
				Foreground(SecondaryColor).
				Bold(true)
)

// Progress bar colors
var (
	ProgressFullColor  = SecondaryColor
	ProgressEmptyColor = SurfaceColor
)

// Helper function to create a badge
func Badge(text string, enabled bool) string {
	if enabled {
		return BadgeEnabledStyle.Render(text + " ✓")
	}
	return BadgeDisabledStyle.Render(text + " ✗")
}

// Helper function to apply gradient to text (character by character)
func GradientText(text string, colors []lipgloss.Color) string {
	if len(colors) == 0 || len(text) == 0 {
		return text
	}

	result := ""
	for i, char := range text {
		colorIdx := i % len(colors)
		style := lipgloss.NewStyle().Foreground(colors[colorIdx])
		result += style.Render(string(char))
	}
	return result
}

// Rainbow gradient colors for animated effects
var RainbowGradient = []lipgloss.Color{
	lipgloss.Color("#FF6B6B"),
	lipgloss.Color("#FFE66D"),
	lipgloss.Color("#4ECDC4"),
	lipgloss.Color("#45B7D1"),
	lipgloss.Color("#96E6A1"),
	lipgloss.Color("#BD93F9"),
	lipgloss.Color("#FF79C6"),
}

// Media player gradient (purple to cyan)
var MediaGradient = []lipgloss.Color{
	lipgloss.Color("#BD93F9"),
	lipgloss.Color("#A78BFA"),
	lipgloss.Color("#8B5CF6"),
	lipgloss.Color("#7C3AED"),
	lipgloss.Color("#6D28D9"),
	lipgloss.Color("#5B21B6"),
	lipgloss.Color("#4C1D95"),
	lipgloss.Color("#8BE9FD"),
}
