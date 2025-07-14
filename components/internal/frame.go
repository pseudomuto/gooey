package internal

import (
	"fmt"
	"io"
	"strings"

	"github.com/pseudomuto/gooey/ansi"
	"github.com/pseudomuto/gooey/components/frame"
)

// FrameReplacer interface allows frames to update lines in place
type FrameReplacer interface {
	ReplaceLine(format string, a ...any)
}

// FrameAware provides common frame integration functionality for components
type FrameAware struct {
	output      io.Writer
	inFrame     bool
	firstRender bool
}

// NewFrameAware creates a new frame-aware utility for the given output writer
func NewFrameAware(output io.Writer) *FrameAware {
	return &FrameAware{
		output:      output,
		inFrame:     IsFrameWriter(output),
		firstRender: true,
	}
}

// IsFrameWriter checks if the writer is a frame by examining its type
func IsFrameWriter(w io.Writer) bool {
	_, ok := w.(*frame.Frame)
	return ok
}

// Output returns the current output writer
func (fa *FrameAware) Output() io.Writer {
	return fa.output
}

// InFrame returns true if the output is a frame
func (fa *FrameAware) InFrame() bool {
	return fa.inFrame
}

// FirstRender returns true if this is the first render call
func (fa *FrameAware) FirstRender() bool {
	return fa.firstRender
}

// MarkRendered marks that rendering has occurred (clears firstRender flag)
func (fa *FrameAware) MarkRendered() {
	fa.firstRender = false
}

// SetOutput updates the output writer and recalculates frame status
func (fa *FrameAware) SetOutput(output io.Writer) {
	fa.output = output
	fa.inFrame = IsFrameWriter(output)
}

// RenderContent renders content appropriately for frame or non-frame context
func (fa *FrameAware) RenderContent(renderFunc func() string) {
	content := renderFunc()

	if fa.inFrame {
		fa.renderInFrame(content)
	} else {
		fa.renderStandalone(content)
	}
}

// renderInFrame renders content within a frame context using ReplaceLine
func (fa *FrameAware) renderInFrame(content string) {
	if frameReplacer, ok := fa.output.(FrameReplacer); ok {
		if fa.firstRender {
			// First render: use normal Println
			fmt.Fprintln(fa.output, content)
			fa.firstRender = false
		} else {
			// Subsequent renders: use ReplaceLine for single-line updates
			frameReplacer.ReplaceLine("%s", content)
		}
	} else {
		// Fallback if frame doesn't support ReplaceLine
		fmt.Fprintln(fa.output, content)
	}
}

// renderStandalone renders content for non-frame context with cursor control
func (fa *FrameAware) renderStandalone(content string) {
	if fa.firstRender {
		// First render: just print the content
		fmt.Fprint(fa.output, content)
		fa.firstRender = false
	} else {
		// Subsequent renders: use carriage return and clear line for in-place update
		fmt.Fprint(fa.output, "\r"+ansi.ClearLine+content)
	}
}

// RenderFinal renders final content with completion handling
func (fa *FrameAware) RenderFinal(renderFunc func() string) {
	content := renderFunc()

	if fa.inFrame {
		if frameReplacer, ok := fa.output.(FrameReplacer); ok {
			frameReplacer.ReplaceLine("%s", content)
		}
	} else {
		fmt.Fprint(fa.output, "\r"+ansi.ClearLine+content)
	}
}

// RenderWithStringBuilder is a utility for components that need to render to a string first
func (fa *FrameAware) RenderWithStringBuilder(renderFunc func(w io.Writer)) {
	if fa.inFrame {
		fa.renderInFrameWithBuilder(renderFunc)
	} else {
		fa.renderStandaloneWithFunc(renderFunc)
	}
}

// renderInFrameWithBuilder handles frame rendering using string builder
func (fa *FrameAware) renderInFrameWithBuilder(renderFunc func(w io.Writer)) {
	var contentBuilder strings.Builder
	renderFunc(&contentBuilder)
	content := contentBuilder.String()

	if fa.firstRender {
		fmt.Fprintln(fa.output, content)
		fa.firstRender = false
	} else {
		fa.replaceLineInFrame(content)
	}
}

// renderStandaloneWithFunc handles standalone rendering with cursor control
func (fa *FrameAware) renderStandaloneWithFunc(renderFunc func(w io.Writer)) {
	if fa.firstRender {
		renderFunc(fa.output)
		fa.firstRender = false
	} else {
		fmt.Fprint(fa.output, "\r"+ansi.ClearLine)
		renderFunc(fa.output)
	}
}

// replaceLineInFrame attempts to use ReplaceLine or falls back to normal printing
func (fa *FrameAware) replaceLineInFrame(content string) {
	if frameReplacer, ok := fa.output.(FrameReplacer); ok {
		frameReplacer.ReplaceLine("%s", content)
	} else {
		fmt.Fprintln(fa.output, content)
	}
}
