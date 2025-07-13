package ansi

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

type (
	// Formatter provides template-based text formatting with color and style support
	Formatter struct {
		writer io.Writer
		colors map[string]Color
		styles map[string]Style
		icons  map[string]Icon
	}
)

// NewFormatter creates a new formatter that writes to the given writer.
// The formatter supports template-based text formatting with color and style support.
//
// Example:
//
//	formatter := ansi.NewFormatter(os.Stdout)
//	formatter.Printf("{{bold+red:Error}}: File not found\n")
//	formatter.Print("{{green:Success}}: Operation completed\n")
//
// Template syntax supports:
//   - Colors: {{red:text}}, {{blue:text}}, {{brightgreen:text}}
//   - Styles: {{bold:text}}, {{italic:text}}, {{underline:text}}
//   - Combinations: {{bold+red:text}}, {{italic+blue:text}}
//   - Icons: {{check:text}}, {{cross:text}}, {{warning:text}}
func NewFormatter(w io.Writer) *Formatter {
	f := &Formatter{
		writer: w,
		colors: make(map[string]Color),
		styles: make(map[string]Style),
		icons:  make(map[string]Icon),
	}

	f.initializeDefaults()
	return f
}

// NewFormatterTo creates a formatter with automatic color/style detection
func NewFormatterTo(w io.Writer) *Formatter {
	return NewFormatter(w)
}

// Format is a convenience function to format a template string without creating a formatter. This is useful for one-off
// formatting operations where you don't need to write to an io.Writer.
//
// Example:
//
//	formatted := ansi.Format("{{bold+green:SUCCESS}}: Operation completed")
//	fmt.Println(formatted)
//
// This is equivalent to creating a formatter and calling Format():
//
//	formatter := ansi.NewFormatter(nil)
//	formatted := formatter.Format("{{bold+green:SUCCESS}}: Operation completed")
func Format(template string) string {
	f := NewFormatter(nil)
	return f.Format(template)
}

// Colorize applies template formatting to a string
func Colorize(template string, args ...any) string {
	f := NewFormatter(nil)
	return f.Sprintf(template, args...)
}

// Write implements io.Writer interface
func (f *Formatter) Write(p []byte) (n int, err error) {
	formatted := f.Format(string(p))
	return f.writer.Write([]byte(formatted))
}

// Format processes template strings and applies formatting.
// This is the core method that transforms template strings into ANSI-formatted text.
//
// Supported syntax:
//   - {{color:text}} - Apply a single color
//   - {{style:text}} - Apply a single style
//   - {{color+style:text}} - Combine multiple modifiers
//   - {{icon:text}} - Add an icon before text
//
// Examples:
//
//	formatter.Format("{{red:Error}}: Something went wrong")
//	// Returns: "\033[31mError\033[0m: Something went wrong"
//
//	formatter.Format("{{bold+blue:Important}} message")
//	// Returns: "\033[1;34mImportant\033[0m message"
//
//	formatter.Format("{{check:}} Task completed")
//	// Returns: "✓ Task completed"
//
// Invalid templates are returned unchanged. Colors and styles are case-insensitive.
func (f *Formatter) Format(template string) string {
	// Regex to match {{modifier:text}} patterns (non-greedy matching)
	re := regexp.MustCompile(`\{\{([^:}]+):([^}]*)\}\}`)

	return re.ReplaceAllStringFunc(template, func(match string) string {
		// Extract modifier and text
		parts := re.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match // Return original if parsing fails
		}

		modifier := strings.ToLower(strings.TrimSpace(parts[1]))
		text := parts[2]

		result := f.applyModifier(modifier, text)

		// If no formatting was applied (invalid modifier), return original
		if result == text {
			return match
		}

		return result
	})
}

// Sprintf formats and processes template strings like fmt.Sprintf
func (f *Formatter) Sprintf(format string, args ...any) string {
	// First apply sprintf formatting
	formatted := fmt.Sprintf(format, args...)
	// Then apply template processing
	return f.Format(formatted)
}

// Printf formats and writes to the underlying writer
func (f *Formatter) Printf(format string, args ...any) (n int, err error) {
	formatted := f.Sprintf(format, args...)
	return f.writer.Write([]byte(formatted))
}

// Println formats and writes with a newline
func (f *Formatter) Println(args ...any) (n int, err error) {
	text := fmt.Sprintln(args...)
	formatted := f.Format(text)
	return f.writer.Write([]byte(formatted))
}

// Print formats and writes without a newline
func (f *Formatter) Print(args ...any) (n int, err error) {
	text := fmt.Sprint(args...)
	formatted := f.Format(text)
	return f.writer.Write([]byte(formatted))
}

// AddColor adds a custom color mapping
func (f *Formatter) AddColor(name string, color Color) {
	f.colors[strings.ToLower(name)] = color
}

// AddStyle adds a custom style mapping
func (f *Formatter) AddStyle(name string, style Style) {
	f.styles[strings.ToLower(name)] = style
}

// AddIcon adds a custom icon mapping
func (f *Formatter) AddIcon(name string, icon Icon) {
	f.icons[strings.ToLower(name)] = icon
}

// SetWriter changes the underlying writer
func (f *Formatter) SetWriter(w io.Writer) {
	f.writer = w
}

// initializeDefaults sets up default color and style mappings
func (f *Formatter) initializeDefaults() {
	f.initializeColors()
	f.initializeStyles()
	f.initializeIcons()
}

// applyModifier applies color and/or style modifiers to text
func (f *Formatter) applyModifier(modifier, text string) string {
	// Split by + to support combinations like "red+bold"
	parts := strings.Split(modifier, "+")

	var colors []Color
	var styles []Style
	var icons []Icon

	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Check if it's a color
		if color, exists := f.colors[part]; exists {
			colors = append(colors, color)
		}

		// Check if it's a style
		if style, exists := f.styles[part]; exists {
			styles = append(styles, style)
		}

		// Check if it's an icon
		if icon, exists := f.icons[part]; exists {
			icons = append(icons, icon)
		}
	}

	// Apply formatting
	if len(colors) == 0 && len(styles) == 0 && len(icons) == 0 {
		return text // No valid modifiers found
	}

	// Handle icons specially - they replace the text
	if len(icons) > 0 {
		iconText := ""
		for _, icon := range icons {
			iconText += icon.String()
		}
		if text != "" {
			iconText += " " + text
		}
		text = iconText
	}

	// Apply color and style formatting if any
	if len(colors) > 0 || len(styles) > 0 {
		var modifiers []any
		for _, color := range colors {
			modifiers = append(modifiers, color)
		}
		for _, style := range styles {
			modifiers = append(modifiers, style)
		}
		return Combine(text, modifiers...)
	}

	return text
}

// initializeColors sets up default color mappings
func (f *Formatter) initializeColors() {
	f.colors["reset"] = Reset
	f.colors["black"] = Black
	f.colors["red"] = Red
	f.colors["green"] = Green
	f.colors["yellow"] = Yellow
	f.colors["blue"] = Blue
	f.colors["magenta"] = Magenta
	f.colors["cyan"] = Cyan
	f.colors["white"] = White
	f.colors["brightblack"] = BrightBlack
	f.colors["brightred"] = BrightRed
	f.colors["brightgreen"] = BrightGreen
	f.colors["brightyellow"] = BrightYellow
	f.colors["brightblue"] = BrightBlue
	f.colors["brightmagenta"] = BrightMagenta
	f.colors["brightcyan"] = BrightCyan
	f.colors["brightwhite"] = BrightWhite
}

// initializeStyles sets up default style mappings
func (f *Formatter) initializeStyles() {
	f.styles["reset"] = StyleReset
	f.styles["bold"] = Bold
	f.styles["dim"] = Dim
	f.styles["italic"] = Italic
	f.styles["underline"] = Underline
	f.styles["blink"] = Blink
	f.styles["reverse"] = Reverse
	f.styles["strikethrough"] = Strikethrough
}

// initializeIcons sets up default icon mappings
func (f *Formatter) initializeIcons() { // nolint: funlen
	// Status icons
	f.icons["check"] = CheckMark
	f.icons["cross"] = CrossMark
	f.icons["warning"] = Warning
	f.icons["info"] = Info
	f.icons["error"] = CrossMark
	f.icons["success"] = CheckMark
	f.icons["question"] = Question
	f.icons["exclamation"] = Exclamation

	// Shape icons
	f.icons["circle"] = Circle
	f.icons["filled-circle"] = FilledCircle
	f.icons["square"] = Square
	f.icons["star"] = Star
	f.icons["star-empty"] = StarEmpty
	f.icons["diamond"] = Diamond
	f.icons["triangle"] = Triangle
	f.icons["heart"] = Heart

	// Checkbox icons
	f.icons["checkbox-empty"] = CheckboxEmpty
	f.icons["checkbox-checked"] = CheckboxChecked
	f.icons["checkbox-crossed"] = CheckboxCrossed
	f.icons["checkbox-progress"] = CheckboxProgress

	// Arrow icons
	f.icons["arrow-up"] = ArrowUp
	f.icons["arrow-down"] = ArrowDown
	f.icons["arrow-left"] = ArrowLeft
	f.icons["arrow-right"] = ArrowRight

	// Progress icons
	f.icons["spinner1"] = Spinner1
	f.icons["spinner2"] = Spinner2
	f.icons["spinner3"] = Spinner3
	f.icons["spinner4"] = Spinner4
	f.icons["spinner5"] = Spinner5
	f.icons["spinner6"] = Spinner6
	f.icons["spinner7"] = Spinner7
	f.icons["spinner8"] = Spinner8

	// Common icons
	f.icons["play"] = Play
	f.icons["pause"] = Pause
	f.icons["stop"] = Stop
	f.icons["record"] = Record
	f.icons["fast"] = Fast
	f.icons["slow"] = Slow
	f.icons["fire"] = Fire
	f.icons["rocket"] = Rocket
	f.icons["gear"] = Gear
	f.icons["lock"] = Lock
	f.icons["key"] = Key
	f.icons["shield"] = Shield
	f.icons["target"] = Target
	f.icons["unicorn"] = Unicorn

	// Success/Error indicator icons
	f.icons["success-icon"] = Success
	f.icons["error-icon"] = Error
	f.icons["bug"] = Bug
	f.icons["fix"] = Fix

	// File icons
	f.icons["file"] = File
	f.icons["folder"] = Folder
	f.icons["folder-open"] = FolderOpen

	// Network icons
	f.icons["download"] = Download
	f.icons["upload"] = Upload
	f.icons["link"] = Link

	// Time icons
	f.icons["clock"] = Clock
	f.icons["hourglass"] = Hourglass

	// Separator icons
	f.icons["dash"] = Dash
	f.icons["double-dash"] = DoubleDash
	f.icons["dot"] = Dot
	f.icons["bullet"] = Bullet
}
