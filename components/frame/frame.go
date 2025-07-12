package frame

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pseudomuto/gooey/ansi"
	"github.com/pseudomuto/gooey/term"
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

	frameStack         = new(FrameStack)
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

	frameStack.Push(frame)

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
	if frameStack.Current() != f {
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
	frameStack.Pop()
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
	frameDepth := frameStack.FrameDepth(f)

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

func prefix() string {
	return frameStack.Prefix()
}

// contentPrefix returns the prefix that should be used for content inside the current frame
func contentPrefix() string {
	return frameStack.ContentPrefix()
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
//	// Creates: [ Minimal ]
//	//          │  Content goes here
//	//          [ ]
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
