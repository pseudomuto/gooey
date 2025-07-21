// Package progress provides interactive progress bars with extensible renderers and frame integration.
// Progress bars support real-time updates, completion states, and customizable visual styles. They implement
// the TaskComponent interface for seamless integration with SpinGroup workflows.
package progress

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pseudomuto/gooey/ansi"
	"github.com/pseudomuto/gooey/internal/frame"
	"github.com/pseudomuto/gooey/internal/term"
)

const (
	defaultProgressWidth = 40
	defaultProgressColor = ansi.Cyan
)

var defaultProgressOutput io.Writer = os.Stdout

type (
	Progress struct {
		title                  string
		total                  int
		current                int
		color                  ansi.Color
		width                  int
		frameAware             *frame.FrameAware
		startTime              time.Time
		message                string
		completed              bool
		failed                 bool    // tracks if progress ended in failure
		lastRenderedPercentage float64 // tracks last rendered percentage for frame mode
		renderer               ProgressRenderer
	}

	ProgressOption func(*Progress)
)

// New creates a new progress bar with the given title and total steps.
// The progress bar can be customized using functional options.
//
// Basic usage:
//
//	p := progress.New("Downloading", 100)
//	p.Update(25, "Downloaded 25 files")
//	p.Complete("Download finished!")
//
// With options:
//
//	p := progress.New("Processing", 100,
//		progress.WithColor(ansi.Green),
//		progress.WithStyle(progress.Bar),
//		progress.WithWidth(60))
func New(title string, total int, options ...ProgressOption) *Progress {
	p := &Progress{
		title:                  title,
		total:                  total,
		current:                0,
		color:                  defaultProgressColor,
		width:                  defaultProgressWidth,
		frameAware:             frame.NewFrameAware(defaultProgressOutput),
		startTime:              time.Now(),
		message:                "",
		completed:              false,
		lastRenderedPercentage: -1, // Initialize to -1 to ensure first render
		renderer:               Bar,
	}

	for _, option := range options {
		option(p)
	}

	return p
}

// Update sets the current progress value and optional message, then re-renders the progress bar.
// The current value should be between 0 and the total value set during creation.
//
// Example:
//
//	p.Update(50, "Processing item 50 of 100")
func (p *Progress) Update(current int, message string) {
	if p.completed {
		return
	}

	p.current = current
	p.message = message
	p.render()
}

// Increment increases the current progress by 1 and optionally updates the message.
// This is a convenience method equivalent to calling Update(current+1, message).
//
// Example:
//
//	p.Increment("Processed another item")
func (p *Progress) Increment(message string) {
	if p.completed {
		return
	}

	p.current++
	p.message = message
	p.render()
}

// Complete marks the progress as finished, shows 100% completion, and displays the final message.
// After calling Complete, further Update/Increment calls will be ignored.
//
// Example:
//
//	p.Complete("All tasks completed successfully!")
func (p *Progress) Complete(message string) {
	if p.completed {
		return
	}

	p.current = p.total
	if message != "" {
		p.message = message
	}
	p.completed = true
	p.render()

	// Show success symbol and add newline
	p.frameAware.RenderFinal(func() string {
		return fmt.Sprintf("%s %s", ansi.CheckMark.Colorize(ansi.Green), p.message)
	})

	// Add newline if not in frame
	if !p.frameAware.InFrame() {
		fmt.Fprint(p.frameAware.Output(), "\n")
	}
}

// Start begins showing the progress bar. This renders the initial state of the progress bar.
// This method is provided for TaskComponent interface compatibility and matches the
// behavior of Spinner components.
//
// Example:
//
//	p := progress.New("Upload", 100)
//	p.Start() // Shows the initial progress bar
func (p *Progress) Start() {
	p.render()
}

// Fail marks the progress as failed with an error message. The progress bar will
// show an error state (typically with red coloring) and the provided message.
// After calling Fail, further Update/Increment/Complete calls will be ignored.
//
// Example:
//
//	p := progress.New("Upload", 100)
//	p.Update(50, "Uploading...")
//	if err := uploadFiles(); err != nil {
//		p.Fail("Upload failed: " + err.Error())
//	}
func (p *Progress) Fail(message string) {
	if p.completed || p.failed {
		return
	}

	if message != "" {
		p.message = message
	}
	p.failed = true
	p.completed = true // Prevent further updates
	p.render()

	// Show failure symbol and add newline
	p.frameAware.RenderFinal(func() string {
		return fmt.Sprintf("%s %s", ansi.CrossMark.Colorize(ansi.Red), p.message)
	})

	// Add newline if not in frame
	if !p.frameAware.InFrame() {
		fmt.Fprint(p.frameAware.Output(), "\n")
	}
}

// SetOutput sets the output writer for the progress bar, allowing redirection
// for frame integration or custom output destinations.
//
// Example:
//
//	var buf bytes.Buffer
//	p := progress.New("Task", 100)
//	p.SetOutput(&buf) // Redirect to buffer
func (p *Progress) SetOutput(output io.Writer) {
	p.frameAware.SetOutput(output)
}

// render draws the current progress bar state to the output writer.
// This method handles cursor positioning to update the progress bar in-place.
func (p *Progress) render() {
	p.frameAware.RenderWithStringBuilder(func(w io.Writer) {
		p.renderer.Render(p, w)
	})
}

// WithColor sets the color for the progress bar.
// The color applies to the filled portion of the progress indicator.
//
// Example:
//
//	p := progress.New("Task", 100, progress.WithColor(ansi.Green))
func WithColor(color ansi.Color) ProgressOption {
	return func(p *Progress) {
		p.color = color
	}
}

// WithWidth sets the width of the progress bar in characters.
// This only affects Bar and Dots styles. Default width is 40 characters.
//
// Example:
//
//	p := progress.New("Task", 100, progress.WithWidth(60))
func WithWidth(width int) ProgressOption {
	return func(p *Progress) {
		if width > 0 {
			p.width = width
		}
	}
}

// WithOutput sets the output writer for the progress bar.
// By default, progress bars write to os.Stdout.
//
// Example:
//
//	var buf bytes.Buffer
//	p := progress.New("Task", 100, progress.WithOutput(&buf))
func WithOutput(output io.Writer) ProgressOption {
	return func(p *Progress) {
		p.frameAware.SetOutput(output)
	}
}

// WithRenderer sets a custom renderer for the progress bar.
// This allows for completely custom progress bar styles beyond the built-in options.
//
// Example:
//
//	type customRenderer struct{}
//	func (r *customRenderer) Render(p *Progress, w io.Writer) {
//		fmt.Fprintf(w, "Custom: %.1f%% [%s]", p.GetPercentage(), p.GetMessage())
//	}
//	p := progress.New("Task", 100, progress.WithRenderer(&customRenderer{}))
func WithRenderer(renderer ProgressRenderer) ProgressOption {
	return func(p *Progress) {
		p.renderer = renderer
	}
}

// Current returns the current progress value.
func (p *Progress) Current() int {
	return p.current
}

// Total returns the total progress value.
func (p *Progress) Total() int {
	return p.total
}

// IsCompleted returns true if the progress has been marked as complete.
func (p *Progress) IsCompleted() bool {
	return p.completed
}

// IsFailed returns true if the progress has been marked as failed.
func (p *Progress) IsFailed() bool {
	return p.failed
}

// Percentage returns the current completion percentage as a float64.
func (p *Progress) Percentage() float64 {
	if p.total == 0 {
		return 0
	}
	return float64(p.current) / float64(p.total) * 100
}

// Elapsed returns the time elapsed since the progress bar was created.
func (p *Progress) Elapsed() time.Duration {
	return time.Since(p.startTime)
}

// Message returns the current progress message.
func (p *Progress) Message() string {
	return p.message
}

// Title returns the progress bar title.
func (p *Progress) Title() string {
	return p.title
}

// Color returns the progress bar color.
func (p *Progress) Color() ansi.Color {
	return p.color
}

// Width returns the progress bar width.
func (p *Progress) Width() int {
	return p.width
}

// AvailableWidth calculates the available width for the progress section (60% of total).
// This matches the three-section layout used by charRenderer.
func (p *Progress) AvailableWidth() int {
	totalWidth := term.Width()
	if p.frameAware.InFrame() {
		totalWidth = totalWidth - 6 // Account for frame borders and padding
	}

	// Progress section gets 60% of total width (matching the renderer layout)
	progressWidth := max(totalWidth*60/100, 20) // At least 20 chars for progress section
	return progressWidth
}
