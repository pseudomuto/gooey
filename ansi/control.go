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

// ClearScreenAndHome clears the screen and moves cursor to home position.
//
// Example:
//
//	fmt.Print(ansi.ClearScreenAndHome())
//	fmt.Println("Screen cleared!")
func ClearScreenAndHome() string {
	return ClearScreen + CursorHome
}

// MoveCursor moves the cursor to the specified position (1-indexed).
//
// Example:
//
//	// Move cursor to row 5, column 10
//	fmt.Print(ansi.MoveCursor(5, 10))
//	fmt.Print("Hello at (5,10)")
func MoveCursor(row, col int) string {
	return fmt.Sprintf("\033[%d;%dH", row, col)
}

// MoveCursorUp moves the cursor up by n lines.
//
// Example:
//
//	fmt.Println("Line 1")
//	fmt.Println("Line 2")
//	fmt.Print(ansi.MoveCursorUp(1))
//	fmt.Print("Replacing Line 2")
func MoveCursorUp(n int) string {
	return fmt.Sprintf("\033[%dA", n)
}

// MoveCursorDown moves the cursor down by n lines.
//
// Example:
//
//	fmt.Print("Current line")
//	fmt.Print(ansi.MoveCursorDown(2))
//	fmt.Print("Two lines below")
func MoveCursorDown(n int) string {
	return fmt.Sprintf("\033[%dB", n)
}

// MoveCursorForward moves the cursor forward by n columns.
//
// Example:
//
//	fmt.Print("Hello")
//	fmt.Print(ansi.MoveCursorForward(5))
//	fmt.Print("World") // Prints "Hello     World"
func MoveCursorForward(n int) string {
	return fmt.Sprintf("\033[%dC", n)
}

// MoveCursorBackward moves the cursor backward by n columns.
//
// Example:
//
//	fmt.Print("Hello World")
//	fmt.Print(ansi.MoveCursorBackward(6))
//	fmt.Print("Gooey") // Overwrites "World" with "Gooey"
func MoveCursorBackward(n int) string {
	return fmt.Sprintf("\033[%dD", n)
}
