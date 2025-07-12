package ansi

import "fmt"

const (
	Reset         Color = iota
	Black         Color = iota
	Red           Color = iota
	Green         Color = iota
	Yellow        Color = iota
	Blue          Color = iota
	Magenta       Color = iota
	Cyan          Color = iota
	White         Color = iota
	BrightBlack   Color = iota
	BrightRed     Color = iota
	BrightGreen   Color = iota
	BrightYellow  Color = iota
	BrightBlue    Color = iota
	BrightMagenta Color = iota
	BrightCyan    Color = iota
	BrightWhite   Color = iota
)

type (
	// Color represents ANSI color codes for terminal text formatting.
	// Colors can be used standalone or combined with styles using the Combine function.
	//
	// Available colors include standard colors (Red, Green, Blue, etc.) and
	// bright variants (BrightRed, BrightGreen, BrightBlue, etc.).
	//
	// Example:
	//
	//	fmt.Print(ansi.Red.Sprint("Error message"))
	//	fmt.Printf("%sWarning%s\n", ansi.Yellow, ansi.Reset)
	Color int
)

// String returns the ANSI escape sequence for the color
func (c Color) String() string {
	switch c {
	case Reset:
		return "\033[0m"
	case Black:
		return "\033[30m"
	case Red:
		return "\033[31m"
	case Green:
		return "\033[32m"
	case Yellow:
		return "\033[33m"
	case Blue:
		return "\033[34m"
	case Magenta:
		return "\033[35m"
	case Cyan:
		return "\033[36m"
	case White:
		return "\033[37m"
	case BrightBlack:
		return "\033[90m"
	case BrightRed:
		return "\033[91m"
	case BrightGreen:
		return "\033[92m"
	case BrightYellow:
		return "\033[93m"
	case BrightBlue:
		return "\033[94m"
	case BrightMagenta:
		return "\033[95m"
	case BrightCyan:
		return "\033[96m"
	case BrightWhite:
		return "\033[97m"
	default:
		return "\033[0m"
	}
}

// Colorize wraps text with the color escape sequence and reset code.
// This ensures the color is applied only to the specified text and doesn't affect subsequent output.
//
// Example:
//
//	redText := ansi.Red.Colorize("Error")
//	fmt.Println(redText + " occurred")  // Only "Error" is red
func (c Color) Colorize(text string) string {
	return fmt.Sprintf("%s%s%s", c.String(), text, Reset.String())
}

// Sprint returns a colored string using fmt.Sprint formatting.
// Equivalent to calling c.Colorize(fmt.Sprint(a...)).
//
// Example:
//
//	message := ansi.Green.Sprint("Operation", " ", "completed")
//	fmt.Println(message)  // Prints "Operation completed" in green
func (c Color) Sprint(a ...any) string {
	return c.Colorize(fmt.Sprint(a...))
}

// Sprintf returns a colored formatted string using fmt.Sprintf formatting.
// Equivalent to calling c.Colorize(fmt.Sprintf(format, a...)).
//
// Example:
//
//	message := ansi.Red.Sprintf("Error %d: %s", 404, "Not found")
//	fmt.Println(message)  // Prints "Error 404: Not found" in red
func (c Color) Sprintf(format string, a ...any) string {
	return c.Colorize(fmt.Sprintf(format, a...))
}
