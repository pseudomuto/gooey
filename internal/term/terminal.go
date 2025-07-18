package term

import (
	"os"
	"regexp"
	"strings"
	"syscall"
	"unsafe"

	"github.com/mattn/go-runewidth"
)

// Default terminal width if detection fails
const defaultTerminalWidth = 120

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

// Width returns the current terminal width in columns.
// If the terminal width cannot be detected (e.g., when not running in a TTY),
// it returns a default width of 120 columns.
//
// Example:
//
//	width := term.Width()
//	if width < 80 {
//		fmt.Println("Terminal is too narrow for optimal display")
//	}
func Width() int {
	var ws winsize
	retCode, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(&ws)))

	if int(retCode) == -1 || errno != 0 {
		return defaultTerminalWidth
	}

	return int(ws.Col)
}

// IsTTY returns true if the current environment supports TTY operations and ANSI escape sequences.
// This is used to determine whether cursor positioning and other terminal control sequences will work.
//
// Example:
//
//	if term.IsTTY() {
//		// Use ANSI escape sequences for cursor control
//		fmt.Print("\033[1A\033[K")
//	} else {
//		// Fall back to simpler output without cursor control
//		fmt.Println("Updated content")
//	}
func IsTTY() bool {
	// Check if stdout is a terminal
	if fileInfo, err := os.Stdout.Stat(); err == nil {
		return (fileInfo.Mode() & os.ModeCharDevice) != 0
	}
	return false
}

// PrintableWidth returns the width of printable characters in a string, excluding ANSI escape sequences.
// This function correctly handles Unicode characters, emojis, and wide characters.
//
// Examples:
//
//	PrintableWidth("hello")                    // Returns: 5
//	PrintableWidth("\033[31mhello\033[0m")     // Returns: 5 (ignores ANSI codes)
//	PrintableWidth("你好")                      // Returns: 4 (wide characters)
//	PrintableWidth("hello 👋")                 // Returns: 8 (emoji counts as 2)
func PrintableWidth(s string) int {
	cleanString := ansiRegex.ReplaceAllString(s, "")
	return runewidth.StringWidth(cleanString)
}

// StripCodes returns the printable characters in a string, excluding ANSI escape sequences.
// This function correctly handles Unicode characters, emojis, and wide characters.
//
// Examples:
//
//	StripCodes("hello")                    // Returns: "hello"
//	StripCodes("\033[31mhello\033[0m")     // Returns: "hello"
//	StripCodes("你好")                      // Returns: "你好"
//	StripCodes("hello 👋")                 // Returns: "hello 👋"
func StripCodes(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// TruncateString truncates a string to the specified printable width while preserving ANSI escape sequences.
// The function correctly handles Unicode characters, emojis, and ANSI color codes.
// If maxWidth is 0 or negative, it returns an empty string.
//
// Examples:
//
//	TruncateString("hello world", 5)                      // Returns: "hello"
//	TruncateString("\033[31mhello world\033[0m", 5)       // Returns: "\033[31mhello"
//	TruncateString("你好世界", 6)                          // Returns: "你好世" (wide chars)
//	TruncateString("hello 👋 world", 8)                   // Returns: "hello 👋"
func TruncateString(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	currentWidth := 0
	var result strings.Builder

	// Track if we're inside an ANSI escape sequence
	inEscape := false

	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			result.WriteRune(r)
			continue
		}

		if inEscape {
			result.WriteRune(r)
			// End of escape sequence (letter)
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}

		// Calculate the width this character would add
		charWidth := PrintableWidth(string(r))

		// If adding this character would exceed the max width, stop
		if currentWidth+charWidth > maxWidth {
			break
		}

		result.WriteRune(r)
		currentWidth += charWidth
	}

	return result.String()
}
