package writer

import (
	"bytes"
	"io"
	"testing"

	"github.com/pseudomuto/gooey/ansi"
	"github.com/pseudomuto/gooey/frame"
	"github.com/stretchr/testify/require"
)

func TestNewIndentedWriter(t *testing.T) {
	tests := []struct {
		name     string
		depth    int
		wantSame bool // true if should return the same writer
	}{
		{"zero depth returns original", 0, true},
		{"negative depth returns original", -1, true},
		{"positive depth creates wrapper", 1, false},
		{"large depth creates wrapper", 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := &bytes.Buffer{}
			result := NewIndentedWriter(original, tt.depth)

			if tt.wantSame {
				require.Same(t, original, result, "should return original writer")
			} else {
				require.NotSame(t, original, result, "should return wrapped writer")
				require.IsType(t, &IndentedWriter{}, result, "should return IndentedWriter")
			}
		})
	}
}

func TestIndentedWriter_Write(t *testing.T) {
	tests := []struct {
		name     string
		depth    int
		input    string
		expected string
	}{
		{
			name:     "single line with depth 1",
			depth:    1,
			input:    "Hello World",
			expected: "  Hello World",
		},
		{
			name:     "single line with depth 2",
			depth:    2,
			input:    "Hello World",
			expected: "    Hello World",
		},
		{
			name:     "multiline content",
			depth:    1,
			input:    "Line 1\nLine 2\nLine 3",
			expected: "  Line 1\n  Line 2\n  Line 3",
		},
		{
			name:     "content with empty lines",
			depth:    1,
			input:    "Line 1\n\nLine 3",
			expected: "  Line 1\n\n  Line 3",
		},
		{
			name:     "standalone newline",
			depth:    1,
			input:    "\n",
			expected: "\n",
		},
		{
			name:     "ANSI clear line sequence",
			depth:    1,
			input:    "\r\x1b[K",
			expected: "\r\x1b[K",
		},
		{
			name:     "content with ANSI escape sequences",
			depth:    1,
			input:    ansi.Red.Colorize("Red Text"),
			expected: "  " + ansi.Red.Colorize("Red Text"),
		},
		{
			name:     "empty string",
			depth:    1,
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only",
			depth:    1,
			input:    "   ",
			expected: "   ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := NewIndentedWriter(&buf, tt.depth)

			n, err := writer.Write([]byte(tt.input))
			require.NoError(t, err)
			require.Equal(t, len(tt.input), n, "should return original byte count")
			require.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestIndentedWriter_IsFrameWriter(t *testing.T) {
	tests := []struct {
		name     string
		writer   func() io.Writer
		expected bool
	}{
		{
			name: "with frame writer",
			writer: func() io.Writer {
				return frame.Open("test", frame.WithOutput(&bytes.Buffer{}))
			},
			expected: true,
		},
		{
			name: "with non-frame writer",
			writer: func() io.Writer {
				return &bytes.Buffer{}
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := tt.writer()
			if f, ok := base.(*frame.Frame); ok {
				defer f.Close()
			}

			indented := NewIndentedWriter(base, 1).(*IndentedWriter)
			result := indented.IsFrameWriter()
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestIndentedWriter_ReplaceLine(t *testing.T) {
	// Create a mock frame replacer for testing
	mock := &mockFrameReplacer{Buffer: &bytes.Buffer{}}
	indented := NewIndentedWriter(mock, 2).(*IndentedWriter)

	indented.ReplaceLine("Test content: %s", "hello")

	require.Len(t, mock.replaceLineCalls, 1)
	require.Equal(t, "%s", mock.replaceLineCalls[0].format)
	require.Len(t, mock.replaceLineCalls[0].args, 1)
	require.Equal(t, "    Test content: hello", mock.replaceLineCalls[0].args[0])
}

func TestIndentedWriter_ReplaceLineN(t *testing.T) {
	mock := &mockFrameReplacer{Buffer: &bytes.Buffer{}}
	indented := NewIndentedWriter(mock, 1).(*IndentedWriter)

	indented.ReplaceLineN(2, "Line content: %d", 42)

	require.Len(t, mock.replaceLineNCalls, 1)
	require.Equal(t, 2, mock.replaceLineNCalls[0].linePosition)
	require.Equal(t, "%s", mock.replaceLineNCalls[0].format)
	require.Equal(t, "  Line content: 42", mock.replaceLineNCalls[0].args[0])
}

func TestIndentedWriter_ReplaceBlock(t *testing.T) {
	mock := &mockFrameReplacer{Buffer: &bytes.Buffer{}}
	indented := NewIndentedWriter(mock, 1).(*IndentedWriter)

	lines := []string{"Line 1", "", "Line 3"}
	indented.ReplaceBlock(3, lines)

	require.Len(t, mock.replaceBlockCalls, 1)
	require.Equal(t, 3, mock.replaceBlockCalls[0].lineCount)
	expected := []string{"  Line 1", "", "  Line 3"}
	require.Equal(t, expected, mock.replaceBlockCalls[0].lines)
}

// mockFrameReplacer implements FrameReplacer for testing
type mockFrameReplacer struct {
	*bytes.Buffer
	replaceLineCalls  []replaceLineCall
	replaceLineNCalls []replaceLineNCall
	replaceBlockCalls []replaceBlockCall
}

type replaceLineCall struct {
	format string
	args   []any
}

type replaceLineNCall struct {
	linePosition int
	format       string
	args         []any
}

type replaceBlockCall struct {
	lineCount int
	lines     []string
}

func (m *mockFrameReplacer) ReplaceLine(format string, a ...any) {
	m.replaceLineCalls = append(m.replaceLineCalls, replaceLineCall{
		format: format,
		args:   a,
	})
}

func (m *mockFrameReplacer) ReplaceLineN(linePosition int, format string, a ...any) {
	m.replaceLineNCalls = append(m.replaceLineNCalls, replaceLineNCall{
		linePosition: linePosition,
		format:       format,
		args:         a,
	})
}

func (m *mockFrameReplacer) ReplaceBlock(lineCount int, lines []string) {
	m.replaceBlockCalls = append(m.replaceBlockCalls, replaceBlockCall{
		lineCount: lineCount,
		lines:     lines,
	})
}
