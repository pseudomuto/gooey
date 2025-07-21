// Package frame provides bordered content areas with nesting support and multiple rendering styles.
// Frames can contain other UI components and support automatic color inheritance, proper indentation,
// and single-line updates for dynamic content like progress bars and spinners.
package frame

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pseudomuto/gooey/ansi"
	"github.com/pseudomuto/gooey/internal/term"
)

const (
	Box     FrameStyle = iota
	Bracket FrameStyle = iota
)

const (
	// Frame prefix constants
	frameBranch         = "├─ "
	frameVerticalPrefix = "│  "
)

var (
	defaultFrameColor            = ansi.Cyan
	defaultFrameStyle            = Box
	defaultFrameOutput io.Writer = os.Stdout

	stack              = new(frameStack)
	frameColorOverride *ansi.Color
	frameColorMutex    sync.RWMutex
)

type (
	Frame struct {
		title        string
		color        ansi.Color
		startTime    time.Time
		output       io.Writer
		needsNewline bool // tracks if the last write ended without a newline
		renderer     frameRenderer
	}

	FrameOption func(*Frame)

	FrameStyle int
)

// Open creates and renders a new frame with the given title.
// Frames provide bordered content areas that can be nested and styled.
// The frame will automatically detect nesting and handle proper indentation and color inheritance.
//
// Basic usage:
//
//	frame := frame.Open("My Task")
//	frame.Println("Task is running...")
//	frame.Close()
//
// With options:
//
//	frame := frame.Open("Error Log",
//		frame.WithColor(ansi.Red),
//		frame.WithStyle(frame.Box),
//		frame.WithOutput(logWriter))
//	frame.Println("Error: %s", err.Error())
//	frame.Close()
//
// Nested frames:
//
//	outer := frame.Open("Deployment", frame.WithColor(ansi.Blue))
//	outer.Println("Starting deployment...")
//
//	inner := frame.Open("Database", frame.WithColor(ansi.Green))
//	inner.Println("Migrating database...")
//	inner.Close()
//
//	outer.Println("Deployment complete!")
//	outer.Close()
//
// The returned frame implements io.Writer and provides Print/Println methods for content.
func Open(title string, options ...FrameOption) *Frame {
	frame := &Frame{
		title:     title,
		color:     defaultFrameColor,
		startTime: time.Now(),
		output:    defaultFrameOutput,
		renderer:  defaultRenderer(),
	}

	for _, option := range options {
		option(frame)
	}

	frameColorMutex.RLock()
	if frameColorOverride != nil {
		frame.color = *frameColorOverride
	}
	frameColorMutex.RUnlock()

	stack.push(frame)

	fmt.Fprint(frame.output, frame.renderer.openFrame(frame.title, frame.color))
	return frame
}

// Close closes the current frame and renders the closing border with elapsed time.
// This method should always be called to properly close frames and maintain the frame stack.
//
// Example:
//
//	frame := frame.Open("Task")
//	defer frame.Close()  // Good practice: ensure frame is always closed
//
//	// Do work...
//	time.Sleep(100 * time.Millisecond)
//	frame.Println("Work completed")
//	// When Close() is called, it will show: └─────────── (100ms) ┘
func (f *Frame) Close() {
	if stack.current() != f {
		return
	}

	elapsed := time.Since(f.startTime)
	frameColorMutex.RLock()
	color := f.color
	if frameColorOverride != nil {
		color = *frameColorOverride
	}
	frameColorMutex.RUnlock()

	closeOutput := f.renderer.closeFrame(elapsed, color)
	stack.pop()
	fmt.Fprint(f.output, closeOutput)
}

// Write implements io.Writer, automatically adding the colored content prefix to each line
func (f *Frame) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	content := string(p)

	// Split content into lines for processing
	lines := strings.Split(content, "\n")
	endsWithNewline := strings.HasSuffix(content, "\n")

	var output strings.Builder

	// Process each line
	for i, line := range lines {
		// Skip the last empty line if content ended with newline
		if i == len(lines)-1 && line == "" && endsWithNewline {
			break
		}

		// Add newline before each line except the first (unless we need one from previous write)
		if i > 0 || f.needsNewline {
			output.WriteString("\n")
		}

		// Add the formatted line
		formattedLine := f.formatContentLine(line)
		output.WriteString(formattedLine)
	}

	// Add final newline if original content ended with one
	if endsWithNewline {
		output.WriteString("\n")
		f.needsNewline = false
	} else {
		f.needsNewline = true
	}

	written, err := f.output.Write([]byte(output.String()))
	if err != nil {
		return written, err
	}

	return len(p), nil
}

// formatContentLine formats a single line of content with proper prefix and suffix
func (f *Frame) formatContentLine(content string) string {
	// Get frame color
	frameColorMutex.RLock()
	color := f.color
	if frameColorOverride != nil {
		color = *frameColorOverride
	}
	frameColorMutex.RUnlock()

	// Get this frame's depth in the stack
	frameDepth := stack.frameDepth(f)

	return f.renderer.formatContentLineWithDepth(content, color, frameDepth)
}

// Print formats according to a format specifier and writes to the frame without adding a newline.
// This method provides printf-style formatting while automatically handling frame prefixes and styling.
//
// Example:
//
//	frame := frame.Open("Progress")
//	frame.Print("Processing item %d of %d", current, total)
//	frame.Print("... ")
//	frame.Println("done!")
//	frame.Close()
func (f *Frame) Print(format string, a ...any) {
	fmt.Fprintf(f, format, a...)
}

// Println formats according to a format specifier and writes to the frame, adding a newline.
// This method provides printf-style formatting while automatically handling frame prefixes and styling.
//
// Example:
//
//	frame := frame.Open("Status")
//	frame.Println("Starting process...")
//	frame.Println("Progress: %d%%", percentage)
//	frame.Println("Status: %s", status)
//	frame.Close()
func (f *Frame) Println(format string, a ...any) {
	fmt.Fprintln(f, fmt.Sprintf(format, a...))
}

// Divider renders a horizontal divider line within the current frame.
// Dividers help organize content into logical sections. The heading parameter is optional.
//
// Examples:
//
//	frame := frame.Open("Report")
//	frame.Println("Header information...")
//	frame.Divider("Main Content")  // Creates: ├── Main Content ─────┤
//	frame.Println("Body content...")
//	frame.Divider("")              // Creates: ├─────────────────────┤
//	frame.Println("Footer content...")
//	frame.Close()
func (f *Frame) Divider(heading string) {
	frameColorMutex.RLock()
	color := f.color
	if frameColorOverride != nil {
		color = *frameColorOverride
	}
	frameColorMutex.RUnlock()

	fmt.Fprint(f.output, f.renderer.createDivider(heading, color))
}

// ReplaceLine replaces the last line written to the frame with new content
// This allows components like progress bars to update in place while maintaining frame formatting
func (f *Frame) ReplaceLine(format string, a ...any) {
	content := fmt.Sprintf(format, a...)

	// Format the content with proper frame styling
	formattedLine := f.formatContentLine(content)

	// Check if we're in a TTY environment that supports ANSI escape sequences
	if term.IsTTY() {
		// Write cursor control directly to the underlying output to bypass frame processing
		// This ensures ANSI sequences are interpreted as control commands, not text
		fmt.Fprint(f.output, ansi.MoveCursorUp(1)+ansi.ClearLine+formattedLine+"\n")
	} else {
		// Non-TTY environment: just append the update as a new line
		fmt.Fprintf(f.output, "%s\n", formattedLine)
	}
}

// ReplaceLineN replaces the Nth line from the current cursor position with new content
// linePosition 1 means the line directly above, 2 means two lines above, etc.
// This allows components to update specific lines by their position
func (f *Frame) ReplaceLineN(linePosition int, format string, a ...any) {
	if linePosition < 1 {
		// Invalid position, fall back to ReplaceLine
		f.ReplaceLine(format, a...)
		return
	}

	content := fmt.Sprintf(format, a...)
	formattedLine := f.formatContentLine(content)

	// Check if we're in a TTY environment that supports ANSI escape sequences
	if term.IsTTY() {
		// Move up, clear line, write content, then move back down to original position
		moveUp := ansi.MoveCursorUp(linePosition)
		moveDown := ansi.MoveCursorDown(linePosition)
		fmt.Fprint(f.output, moveUp+ansi.ClearLine+formattedLine+moveDown)
	} else {
		// Non-TTY environment: just append the update as a new line
		fmt.Fprintf(f.output, "%s\n", formattedLine)
	}
}

// ReplaceBlock replaces the last N lines with new content lines
// This is more reliable than individual line replacements for multi-line content
func (f *Frame) ReplaceBlock(lineCount int, lines []string) {
	if lineCount < 1 {
		return
	}

	// If no new lines, just clear the old content
	if len(lines) == 0 {
		if lineCount > 1 {
			fmt.Fprint(f.output, ansi.MoveCursorUp(lineCount-1))
		}
		for i := 0; i < lineCount; i++ {
			fmt.Fprint(f.output, ansi.ClearLine)
			if i < lineCount-1 {
				fmt.Fprint(f.output, "\n")
			}
		}
		return
	}

	// Move cursor to the beginning of the block we want to replace
	if lineCount > 1 {
		fmt.Fprint(f.output, ansi.MoveCursorUp(lineCount-1))
	}

	// Clear all old lines first
	for i := 0; i < lineCount; i++ {
		fmt.Fprint(f.output, ansi.ClearLine)
		if i < lineCount-1 {
			fmt.Fprint(f.output, "\n")
		}
	}

	// Move cursor back to the beginning of the cleared area
	if lineCount > 1 {
		fmt.Fprint(f.output, ansi.MoveCursorUp(lineCount-1))
	}

	// Write new content
	for i, line := range lines {
		formattedLine := f.formatContentLine(line)
		fmt.Fprint(f.output, formattedLine)

		// Add newline except for the last line
		if i < len(lines)-1 {
			fmt.Fprint(f.output, "\n")
		}
	}

	// If new content has more lines than old content, add them
	if len(lines) > lineCount {
		for i := lineCount; i < len(lines); i++ {
			formattedLine := f.formatContentLine(lines[i])
			fmt.Fprint(f.output, "\n"+formattedLine)
		}
	}
}

// WithColor sets the color for the frame's border and content prefixes.
// The color applies to all frame elements including borders, dividers, and nested frame prefixes.
//
// Example:
//
//	redFrame := frame.Open("Error", frame.WithColor(ansi.Red))
//	blueFrame := frame.Open("Info", frame.WithColor(ansi.Blue))
//	greenFrame := frame.Open("Success", frame.WithColor(ansi.Green))
func WithColor(color ansi.Color) FrameOption {
	return func(f *Frame) {
		f.color = color
	}
}

// WithStyle sets the frame's rendering style.
//
// Two styles are available:
//   - frame.Box: Full box borders with complete enclosure (default)
//   - frame.Bracket: Simple bracket-style markers without full borders
//
// Examples:
//
//	// Box style (full borders)
//	boxFrame := frame.Open("Full Border", frame.WithStyle(frame.Box))
//	// Creates: ┌── Full Border ──────┐
//	//          │ Content goes here   │
//	//          └─────────────────────┘
//
//	// Bracket style (minimal markers)
//	bracketFrame := frame.Open("Minimal", frame.WithStyle(frame.Bracket))
//	// Creates: ┌── Minimal
//	//          │  Content goes here
//	//          └──
func WithStyle(style FrameStyle) FrameOption {
	return func(f *Frame) {
		width := term.Width()
		f.renderer = &boxRenderer{termWidth: width}
		if style == Bracket {
			f.renderer = &bracketRenderer{termWidth: width}
		}
	}
}

// WithOutput overrides the frame's output writer.
// By default, frames write to os.Stdout, but this option allows directing output elsewhere.
//
// Examples:
//
//	// Write to a file
//	logFile, _ := os.Create("frame.log")
//	defer logFile.Close()
//	frame := frame.Open("Log", frame.WithOutput(logFile))
//
//	// Write to a buffer for testing
//	var buf bytes.Buffer
//	frame := frame.Open("Test", frame.WithOutput(&buf))
//
//	// Write with ANSI formatting
//	formatter := ansi.NewFormatter(os.Stdout)
//	frame := frame.Open("Colored", frame.WithOutput(formatter))
func WithOutput(output io.Writer) FrameOption {
	return func(f *Frame) {
		f.output = output
	}
}

func defaultRenderer() frameRenderer {
	if defaultFrameStyle == Box {
		return &boxRenderer{termWidth: term.Width()}
	}

	return &bracketRenderer{termWidth: term.Width()}
}
