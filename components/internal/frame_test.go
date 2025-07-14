package internal

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/pseudomuto/gooey/ansi"
	"github.com/pseudomuto/gooey/components/frame"
	"github.com/stretchr/testify/require"
)

// mockFrameReplacer implements both io.Writer and FrameReplacer for testing
type mockFrameReplacer struct {
	*bytes.Buffer
	replaceLineCalls []replaceLineCall
}

type replaceLineCall struct {
	format string
	args   []interface{}
}

func newMockFrameReplacer() *mockFrameReplacer {
	return &mockFrameReplacer{
		Buffer:           &bytes.Buffer{},
		replaceLineCalls: make([]replaceLineCall, 0),
	}
}

func (m *mockFrameReplacer) ReplaceLine(format string, a ...interface{}) {
	m.replaceLineCalls = append(m.replaceLineCalls, replaceLineCall{
		format: format,
		args:   a,
	})
	// Also write to buffer for verification
	fmt.Fprintf(m.Buffer, format, a...)
}

func TestNewFrameAware(t *testing.T) {
	tests := []struct {
		name            string
		output          io.Writer
		expectedInFrame bool
	}{
		{
			name:            "regular buffer writer",
			output:          &bytes.Buffer{},
			expectedInFrame: false,
		},
		{
			name:            "frame writer",
			output:          frame.Open("test", frame.WithOutput(&bytes.Buffer{})),
			expectedInFrame: true,
		},
		{
			name:            "mock frame replacer",
			output:          newMockFrameReplacer(),
			expectedInFrame: false, // mock doesn't inherit from frame.Frame
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fa := NewFrameAware(tt.output)

			require.NotNil(t, fa)
			require.Equal(t, tt.output, fa.Output())
			require.Equal(t, tt.expectedInFrame, fa.InFrame())
			require.True(t, fa.FirstRender())
		})
	}
}

func TestIsFrameWriter(t *testing.T) {
	tests := []struct {
		name     string
		writer   io.Writer
		expected bool
	}{
		{
			name:     "bytes.Buffer",
			writer:   &bytes.Buffer{},
			expected: false,
		},
		{
			name:     "frame.Frame",
			writer:   frame.Open("test", frame.WithOutput(&bytes.Buffer{})),
			expected: true,
		},
		{
			name:     "nil writer",
			writer:   nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsFrameWriter(tt.writer)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestFrameAware_SetOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	fa := NewFrameAware(buf)

	require.False(t, fa.InFrame())
	require.Equal(t, buf, fa.Output())

	// Change to frame writer
	frameWriter := frame.Open("test", frame.WithOutput(&bytes.Buffer{}))
	fa.SetOutput(frameWriter)

	require.True(t, fa.InFrame())
	require.Equal(t, frameWriter, fa.Output())

	// Change back to regular writer
	newBuf := &bytes.Buffer{}
	fa.SetOutput(newBuf)

	require.False(t, fa.InFrame())
	require.Equal(t, newBuf, fa.Output())
}

func TestFrameAware_MarkRendered(t *testing.T) {
	fa := NewFrameAware(&bytes.Buffer{})

	require.True(t, fa.FirstRender())

	fa.MarkRendered()

	require.False(t, fa.FirstRender())
}

func TestFrameAware_RenderContent_Standalone(t *testing.T) {
	buf := &bytes.Buffer{}
	fa := NewFrameAware(buf)

	// First render
	fa.RenderContent(func() string {
		return "first content"
	})

	require.Equal(t, "first content", buf.String())
	require.False(t, fa.FirstRender())

	// Second render (should use carriage return and clear line)
	fa.RenderContent(func() string {
		return "second content"
	})

	expected := "first content\r" + ansi.ClearLine + "second content"
	require.Equal(t, expected, buf.String())
}

func TestFrameAware_RenderContent_Frame(t *testing.T) {
	mock := newMockFrameReplacer()
	fa := NewFrameAware(mock)
	fa.inFrame = true // Force frame mode for the mock

	// First render
	fa.RenderContent(func() string {
		return "first content"
	})

	require.Equal(t, "first content\n", mock.String())
	require.False(t, fa.FirstRender())
	require.Empty(t, mock.replaceLineCalls) // No ReplaceLine calls on first render

	// Second render (should use ReplaceLine)
	fa.RenderContent(func() string {
		return "second content"
	})

	require.Len(t, mock.replaceLineCalls, 1)
	require.Equal(t, "%s", mock.replaceLineCalls[0].format)
	require.Equal(t, []interface{}{"second content"}, mock.replaceLineCalls[0].args)
}

func TestFrameAware_RenderFinal_Standalone(t *testing.T) {
	buf := &bytes.Buffer{}
	fa := NewFrameAware(buf)

	fa.RenderFinal(func() string {
		return "final content"
	})

	expected := "\r" + ansi.ClearLine + "final content"
	require.Equal(t, expected, buf.String())
}

func TestFrameAware_RenderFinal_Frame(t *testing.T) {
	mock := newMockFrameReplacer()
	fa := NewFrameAware(mock)
	fa.inFrame = true // Force frame mode for the mock

	fa.RenderFinal(func() string {
		return "final content"
	})

	require.Len(t, mock.replaceLineCalls, 1)
	require.Equal(t, "%s", mock.replaceLineCalls[0].format)
	require.Equal(t, []interface{}{"final content"}, mock.replaceLineCalls[0].args)
}

func TestFrameAware_RenderWithStringBuilder_Standalone(t *testing.T) {
	buf := &bytes.Buffer{}
	fa := NewFrameAware(buf)

	// First render
	fa.RenderWithStringBuilder(func(w io.Writer) {
		fmt.Fprint(w, "builder content 1")
	})

	require.Equal(t, "builder content 1", buf.String())
	require.False(t, fa.FirstRender())

	// Second render
	fa.RenderWithStringBuilder(func(w io.Writer) {
		fmt.Fprint(w, "builder content 2")
	})

	expected := "builder content 1\r" + ansi.ClearLine + "builder content 2"
	require.Equal(t, expected, buf.String())
}

func TestFrameAware_RenderWithStringBuilder_Frame(t *testing.T) {
	mock := newMockFrameReplacer()
	fa := NewFrameAware(mock)
	fa.inFrame = true // Force frame mode for the mock

	// First render
	fa.RenderWithStringBuilder(func(w io.Writer) {
		fmt.Fprint(w, "builder content 1")
	})

	require.Equal(t, "builder content 1\n", mock.String())
	require.False(t, fa.FirstRender())
	require.Empty(t, mock.replaceLineCalls)

	// Second render
	fa.RenderWithStringBuilder(func(w io.Writer) {
		fmt.Fprint(w, "builder content 2")
	})

	require.Len(t, mock.replaceLineCalls, 1)
	require.Equal(t, "%s", mock.replaceLineCalls[0].format)
	require.Equal(t, []interface{}{"builder content 2"}, mock.replaceLineCalls[0].args)
}

func TestFrameAware_RenderContent_FrameWithoutReplacer(t *testing.T) {
	// Test frame writer that doesn't implement FrameReplacer
	frameWithoutReplacer := frame.Open("test", frame.WithOutput(&bytes.Buffer{}))
	fa := NewFrameAware(frameWithoutReplacer)

	fa.RenderContent(func() string {
		return "content"
	})

	// Should fallback to normal frame printing (through the frame's output)
	require.False(t, fa.FirstRender())
}

func TestFrameAware_ComplexRenderingScenario(t *testing.T) {
	// Test a complex scenario with multiple renders and state changes
	buf := &bytes.Buffer{}
	fa := NewFrameAware(buf)

	// Initial render
	fa.RenderContent(func() string {
		return "Loading..."
	})

	require.Equal(t, "Loading...", buf.String())

	// Multiple updates
	fa.RenderContent(func() string {
		return "Progress: 25%"
	})

	fa.RenderContent(func() string {
		return "Progress: 50%"
	})

	fa.RenderContent(func() string {
		return "Progress: 75%"
	})

	// Final render
	fa.RenderFinal(func() string {
		return "Complete!"
	})

	// Verify the complete sequence
	expected := strings.Join([]string{
		"Loading...",
		"\r" + ansi.ClearLine + "Progress: 25%",
		"\r" + ansi.ClearLine + "Progress: 50%",
		"\r" + ansi.ClearLine + "Progress: 75%",
		"\r" + ansi.ClearLine + "Complete!",
	}, "")

	require.Equal(t, expected, buf.String())
}

func TestFrameAware_RenderWithStringBuilder_MultipleWrites(t *testing.T) {
	// Test that string builder properly accumulates multiple writes
	buf := &bytes.Buffer{}
	fa := NewFrameAware(buf)

	fa.RenderWithStringBuilder(func(w io.Writer) {
		fmt.Fprint(w, "Part 1")
		fmt.Fprint(w, " - ")
		fmt.Fprint(w, "Part 2")
		fmt.Fprint(w, " - ")
		fmt.Fprint(w, "Part 3")
	})

	require.Equal(t, "Part 1 - Part 2 - Part 3", buf.String())
}

func TestFrameAware_StateConsistency(t *testing.T) {
	// Test that state changes are properly tracked
	buf := &bytes.Buffer{}
	fa := NewFrameAware(buf)

	require.True(t, fa.FirstRender())
	require.False(t, fa.InFrame())

	// After first render, FirstRender should be false
	fa.RenderContent(func() string { return "test" })
	require.False(t, fa.FirstRender())

	// Changing output should update frame status but not affect FirstRender
	frameWriter := frame.Open("test", frame.WithOutput(&bytes.Buffer{}))
	fa.SetOutput(frameWriter)
	require.True(t, fa.InFrame())
	require.False(t, fa.FirstRender()) // Should still be false

	// Manual mark should also work
	fa.firstRender = true // Reset for test
	fa.MarkRendered()
	require.False(t, fa.FirstRender())
}
