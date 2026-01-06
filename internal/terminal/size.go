package terminal

// This file contains utilities for terminal size detection.
// The main GetTerminalSize function is implemented in capability.go
// to keep all terminal detection logic in one place.

// Note: For full functionality, this would use golang.org/x/term
// which requires CGO on some platforms. The implementation above
// provides a working fallback that uses environment variables.
//
// Full implementation would look like:
/*
import "golang.org/x/term"

func GetTerminalSize() (int, int, error) {
	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		// Fallback to environment variables
		if w := os.Getenv("COLUMNS"); w != "" {
			if width, err := strconv.Atoi(w); err == nil {
				if h := os.Getenv("LINES"); h != "" {
					if height, err := strconv.Atoi(h); err == nil {
						return width, height, nil
					}
				}
			}
		}
		return 80, 24, err
	}
	return width, height, nil
}
*/
