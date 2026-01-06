package terminal

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"terminaltube/pkg/types"

	"golang.org/x/term"
)

// DetectCapabilities detects what the current terminal supports
func DetectCapabilities() (types.TerminalCapabilities, error) {
	capabilities := types.TerminalCapabilities{}

	// Detect terminal size using golang.org/x/term
	width, height, err := GetTerminalSize()
	if err != nil {
		// Default fallback size
		width, height = 80, 24
	}
	capabilities.Width = width
	capabilities.Height = height

	// Detect SIXEL support
	capabilities.SixelSupport = detectSixelSupport()

	// Detect color support
	capabilities.TrueColor = detectTrueColorSupport()
	capabilities.Color256 = detectColor256Support()

	// Detect Unicode support
	capabilities.UnicodeSupport = detectUnicodeSupport()

	return capabilities, nil
}

// GetTerminalSize returns the current terminal dimensions
func GetTerminalSize() (int, int, error) {
	// Try to get terminal size from stdout
	if term.IsTerminal(int(os.Stdout.Fd())) {
		width, height, err := term.GetSize(int(os.Stdout.Fd()))
		if err == nil && width > 0 && height > 0 {
			return width, height, nil
		}
	}

	// Try stdin as fallback
	if term.IsTerminal(int(os.Stdin.Fd())) {
		width, height, err := term.GetSize(int(os.Stdin.Fd()))
		if err == nil && width > 0 && height > 0 {
			return width, height, nil
		}
	}

	// Try stderr as fallback
	if term.IsTerminal(int(os.Stderr.Fd())) {
		width, height, err := term.GetSize(int(os.Stderr.Fd()))
		if err == nil && width > 0 && height > 0 {
			return width, height, nil
		}
	}

	// Try environment variables (cross-platform fallback)
	if w := os.Getenv("COLUMNS"); w != "" {
		if width, err := strconv.Atoi(w); err == nil {
			if h := os.Getenv("LINES"); h != "" {
				if height, err := strconv.Atoi(h); err == nil {
					return width, height, nil
				}
			}
		}
	}

	// Default fallback size
	return 120, 30, nil
}

// detectSixelSupport checks if the terminal supports SIXEL graphics
func detectSixelSupport() bool {
	termType := strings.ToLower(os.Getenv("TERM"))
	termProgram := strings.ToLower(os.Getenv("TERM_PROGRAM"))

	// Known SIXEL-supporting terminals
	sixelTerminals := []string{
		"xterm-256color",
		"xterm-color",
		"mlterm",
		"wezterm",
		"foot",
		"mintty",
		"xterm",
	}

	for _, supportedTerm := range sixelTerminals {
		if strings.Contains(termType, supportedTerm) {
			return true
		}
	}

	// Check TERM_PROGRAM for some terminals
	sixelPrograms := []string{
		"iterm.app",
		"wezterm",
		"mintty",
		"foot",
	}

	for _, program := range sixelPrograms {
		if strings.Contains(termProgram, program) {
			return true
		}
	}

	// Check for Windows Terminal with SIXEL support (v1.22+)
	// Windows Terminal 1.22+ supports SIXEL graphics
	if os.Getenv("WT_SESSION") != "" {
		// Windows Terminal is running - check version
		if checkWindowsTerminalSixelSupport() {
			return true
		}
	}

	// Additional check: look for SIXEL in terminal capabilities
	if strings.Contains(termType, "sixel") {
		return true
	}

	return false
}

// checkWindowsTerminalSixelSupport checks if Windows Terminal version supports SIXEL
func checkWindowsTerminalSixelSupport() bool {
	// Try to get Windows Terminal version via PowerShell
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		"(Get-AppxPackage Microsoft.WindowsTerminal).Version")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	version := strings.TrimSpace(string(output))
	if version == "" {
		return false
	}

	// Parse version (format: 1.22.xxxx.x or similar)
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return false
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return false
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return false
	}

	// SIXEL support was added in Windows Terminal 1.22
	if major > 1 || (major == 1 && minor >= 22) {
		return true
	}

	return false
}

// detectTrueColorSupport checks if the terminal supports 24-bit true color
func detectTrueColorSupport() bool {
	// Check COLORTERM environment variable
	colorTerm := strings.ToLower(os.Getenv("COLORTERM"))
	if colorTerm == "truecolor" || colorTerm == "24bit" {
		return true
	}

	// Check TERM for true color indicators
	termType := strings.ToLower(os.Getenv("TERM"))
	trueColorTerms := []string{
		"xterm-256color",
		"screen-256color",
		"tmux-256color",
		"xterm-direct",
		"alacritty",
	}

	for _, term := range trueColorTerms {
		if strings.Contains(termType, term) {
			return true
		}
	}

	// Check terminal program
	termProgram := strings.ToLower(os.Getenv("TERM_PROGRAM"))
	trueColorPrograms := []string{
		"iterm.app",
		"wezterm",
		"alacritty",
		"hyper",
		"vscode",
	}

	for _, program := range trueColorPrograms {
		if strings.Contains(termProgram, program) {
			return true
		}
	}

	// Windows Terminal supports true color
	if os.Getenv("WT_SESSION") != "" {
		return true
	}

	// ConEmu supports true color
	if os.Getenv("ConEmuANSI") == "ON" {
		return true
	}

	// For Windows, assume modern terminals support true color
	if os.Getenv("OS") == "Windows_NT" {
		// PowerShell in Windows Terminal, Windows Terminal, and modern cmd support it
		if len(os.Getenv("PSModulePath")) > 0 {
			return true
		}
	}

	return false
}

// detectColor256Support checks if the terminal supports 256 colors
func detectColor256Support() bool {
	termType := strings.ToLower(os.Getenv("TERM"))

	// Most modern terminals support 256 colors
	if strings.Contains(termType, "256color") ||
		strings.Contains(termType, "xterm") ||
		strings.Contains(termType, "screen") ||
		strings.Contains(termType, "tmux") {
		return true
	}

	// For Windows, assume modern terminals support 256 colors
	if os.Getenv("OS") == "Windows_NT" {
		return true
	}

	// Most modern terminals support at least 256 colors by default
	return true
}

// detectUnicodeSupport checks if the terminal supports Unicode
func detectUnicodeSupport() bool {
	// Check locale settings
	locale := os.Getenv("LC_ALL")
	if locale == "" {
		locale = os.Getenv("LC_CTYPE")
	}
	if locale == "" {
		locale = os.Getenv("LANG")
	}

	// Most UTF-8 locales support Unicode
	if strings.Contains(strings.ToUpper(locale), "UTF-8") ||
		strings.Contains(strings.ToUpper(locale), "UTF8") {
		return true
	}

	// Windows typically supports Unicode in modern terminals
	if os.Getenv("OS") == "Windows_NT" {
		return true
	}

	// Most modern terminals support Unicode by default
	return true
}
