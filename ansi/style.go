package ansi

import "fmt"

const (
	StyleReset    Style = iota
	Bold          Style = iota
	Dim           Style = iota
	Italic        Style = iota
	Underline     Style = iota
	Blink         Style = iota
	Reverse       Style = iota
	Strikethrough Style = iota
)

type (
	// Style represents ANSI text formatting styles such as bold, italic, underline, etc.
	// Styles can be applied individually or combined with colors using the Combine function.
	//
	// Example:
	//
	//	fmt.Print(ansi.Bold.Apply("Important text"))
	//	fmt.Print(ansi.Combine("Styled text", ansi.Bold, ansi.Red))
	Style int
)

// String returns the ANSI escape sequence for the style
func (s Style) String() string {
	switch s {
	case StyleReset:
		return "\033[0m"
	case Bold:
		return "\033[1m"
	case Dim:
		return "\033[2m"
	case Italic:
		return "\033[3m"
	case Underline:
		return "\033[4m"
	case Blink:
		return "\033[5m"
	case Reverse:
		return "\033[7m"
	case Strikethrough:
		return "\033[9m"
	default:
		return ""
	}
}

// Apply applies the style to text
func (s Style) Apply(text string) string {
	return fmt.Sprintf("%s%s%s", s.String(), text, StyleReset.String())
}

// BoldText applies bold formatting to text.
//
// Example:
//
//	fmt.Println(ansi.BoldText("Important message"))
func BoldText(text string) string {
	return Bold.Apply(text)
}

// ItalicText applies italic formatting to text.
//
// Example:
//
//	fmt.Println(ansi.ItalicText("Emphasized text"))
func ItalicText(text string) string {
	return Italic.Apply(text)
}

// UnderlineText applies underline formatting to text.
//
// Example:
//
//	fmt.Println(ansi.UnderlineText("Underlined text"))
func UnderlineText(text string) string {
	return Underline.Apply(text)
}

// StrikethroughText applies strikethrough formatting to text.
//
// Example:
//
//	fmt.Println(ansi.StrikethroughText("Cancelled text"))
func StrikethroughText(text string) string {
	return Strikethrough.Apply(text)
}

// DimText applies dim formatting to text.
//
// Example:
//
//	fmt.Println(ansi.DimText("Secondary information"))
func DimText(text string) string {
	return Dim.Apply(text)
}

// Combine combines multiple styles and colors into a single formatted string.
// This function allows you to apply multiple formatting options at once.
//
// Supported types:
//   - ansi.Color: Apply color formatting
//   - ansi.Style: Apply style formatting (bold, italic, etc.)
//
// Examples:
//
//	// Bold red text
//	text := ansi.Combine("Error", ansi.Bold, ansi.Red)
//
//	// Italic blue underlined text
//	text := ansi.Combine("Link", ansi.Italic, ansi.Blue, ansi.Underline)
//
//	// Mixed order (colors and styles can be in any order)
//	text := ansi.Combine("Warning", ansi.Yellow, ansi.Bold, ansi.Italic)
func Combine(text string, styles ...any) string {
	var codes []string

	for _, style := range styles {
		switch v := style.(type) {
		case Style:
			if code := v.String(); code != "" {
				codes = append(codes, code)
			}
		case Color:
			if code := v.String(); code != "" {
				codes = append(codes, code)
			}
		}
	}

	if len(codes) == 0 {
		return text
	}

	var combined string
	for _, code := range codes {
		combined += code
	}

	return fmt.Sprintf("%s%s%s", combined, text, StyleReset.String())
}
