package terminal

import (
	"fmt"
	"os"
)

// Control provides terminal control functions using ANSI escape sequences
type Control struct {
	cursorHidden bool
}

// NewControl creates a new terminal control instance
func NewControl() *Control {
	return &Control{}
}

// ClearScreen clears the entire terminal screen and moves cursor to home
func (c *Control) ClearScreen() error {
	_, err := fmt.Print("\033[2J\033[H")
	return err
}

// MoveCursorHome moves the cursor to the top-left corner (1,1)
func (c *Control) MoveCursorHome() error {
	_, err := fmt.Print("\033[H")
	return err
}

// MoveCursor moves the cursor to the specified position (1-based)
func (c *Control) MoveCursor(row, col int) error {
	_, err := fmt.Printf("\033[%d;%dH", row, col)
	return err
}

// HideCursor hides the terminal cursor
func (c *Control) HideCursor() error {
	if !c.cursorHidden {
		_, err := fmt.Print("\033[?25l")
		if err == nil {
			c.cursorHidden = true
		}
		return err
	}
	return nil
}

// ShowCursor shows the terminal cursor
func (c *Control) ShowCursor() error {
	if c.cursorHidden {
		_, err := fmt.Print("\033[?25h")
		if err == nil {
			c.cursorHidden = false
		}
		return err
	}
	return nil
}

// SaveCursorPosition saves the current cursor position
func (c *Control) SaveCursorPosition() error {
	_, err := fmt.Print("\033[s")
	return err
}

// RestoreCursorPosition restores the previously saved cursor position
func (c *Control) RestoreCursorPosition() error {
	_, err := fmt.Print("\033[u")
	return err
}

// SetTitle sets the terminal window title
func (c *Control) SetTitle(title string) error {
	_, err := fmt.Printf("\033]0;%s\007", title)
	return err
}

// EnableAlternateScreen switches to the alternate screen buffer
func (c *Control) EnableAlternateScreen() error {
	_, err := fmt.Print("\033[?1049h")
	return err
}

// DisableAlternateScreen switches back to the main screen buffer
func (c *Control) DisableAlternateScreen() error {
	_, err := fmt.Print("\033[?1049l")
	return err
}

// Reset resets all terminal attributes and clears the screen
func (c *Control) Reset() error {
	// Reset attributes, show cursor, clear screen
	_, err := fmt.Print("\033[0m\033[?25h\033[2J\033[H")
	c.cursorHidden = false
	return err
}

// Flush flushes the output buffer
func (c *Control) Flush() error {
	return nil // fmt.Print automatically flushes
}

// GetTerminalInfo returns basic terminal information
func (c *Control) GetTerminalInfo() map[string]string {
	info := make(map[string]string)

	info["TERM"] = os.Getenv("TERM")
	info["TERM_PROGRAM"] = os.Getenv("TERM_PROGRAM")
	info["TERM_PROGRAM_VERSION"] = os.Getenv("TERM_PROGRAM_VERSION")
	info["COLORTERM"] = os.Getenv("COLORTERM")
	info["LANG"] = os.Getenv("LANG")
	info["LC_ALL"] = os.Getenv("LC_ALL")

	return info
}

// QueryTerminalSize queries the terminal for its current size using ANSI escape sequences
func (c *Control) QueryTerminalSize() (int, int, error) {
	// This is a more advanced method that queries the terminal directly
	// For now, we'll return an error to indicate it's not implemented
	// A full implementation would:
	// 1. Send cursor position query: "\033[999;999H\033[6n"
	// 2. Read the response from stdin
	// 3. Parse the response to get terminal dimensions
	return 0, 0, fmt.Errorf("terminal size query not implemented")
}
