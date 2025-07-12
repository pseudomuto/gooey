package ansi

import "fmt"

// Control sequences for terminal manipulation
const (
	// Cursor control
	ClearScreen    = "\033[2J"
	CursorHome     = "\033[H"
	ClearLine      = "\033[K"
	CursorUp       = "\033[A"
	CursorDown     = "\033[B"
	CursorForward  = "\033[C"
	CursorBackward = "\033[D"
	SaveCursor     = "\033[s"
	RestoreCursor  = "\033[u"
	HideCursor     = "\033[?25l"
	ShowCursor     = "\033[?25h"
)

// ClearScreenAndHome clears the screen and moves cursor to home position
func ClearScreenAndHome() string {
	return ClearScreen + CursorHome
}

// MoveCursor moves the cursor to the specified position (1-indexed)
func MoveCursor(row, col int) string {
	return fmt.Sprintf("\033[%d;%dH", row, col)
}

// MoveCursorUp moves the cursor up by n lines
func MoveCursorUp(n int) string {
	return fmt.Sprintf("\033[%dA", n)
}

// MoveCursorDown moves the cursor down by n lines
func MoveCursorDown(n int) string {
	return fmt.Sprintf("\033[%dB", n)
}

// MoveCursorForward moves the cursor forward by n columns
func MoveCursorForward(n int) string {
	return fmt.Sprintf("\033[%dC", n)
}

// MoveCursorBackward moves the cursor backward by n columns
func MoveCursorBackward(n int) string {
	return fmt.Sprintf("\033[%dD", n)
}
