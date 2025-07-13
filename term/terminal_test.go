package term_test

import (
	"os"
	"testing"

	"github.com/mattn/go-isatty"
	. "github.com/pseudomuto/gooey/term"
	"github.com/stretchr/testify/require"
)

const defaultTerminalWidth = 120

func TestWidth(t *testing.T) {
	width := Width()

	// In a CI environment, this will likely not be a TTY
	if isatty.IsTerminal(os.Stdin.Fd()) {
		require.Positive(t, width, "expected to get a terminal width")
	} else {
		require.Equal(t, defaultTerminalWidth, width, "expected to get default width when not in a TTY")
	}
}

func TestPrintableWidth(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "simple string",
			input:    "hello",
			expected: 5,
		},
		{
			name:     "string with color codes",
			input:    "\033[31mhello\033[0m",
			expected: 5,
		},
		{
			name:     "string with unicode",
			input:    "✓ hello",
			expected: 7,
		},
		{
			name:     "string with color and unicode",
			input:    "\033[32m✓ success\033[0m",
			expected: 9,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "string with only ansi codes",
			input:    "\033[31m\033[1m\033[0m",
			expected: 0,
		},
		{
			name:     "string with emoji",
			input:    "hello 👋",
			expected: 8,
		},
		{
			name:     "string with wide characters",
			input:    "\u4f60\u597d", // "你好"
			expected: 4,
		},
		{
			name:     "string with ZWJ emoji",
			input:    "👩‍👩‍👧‍👦",
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, PrintableWidth(tt.input))
		})
	}
}

func TestStripCodes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "string with color codes",
			input:    "\033[31mhello\033[0m",
			expected: "hello",
		},
		{
			name:     "string with unicode",
			input:    "✓ hello",
			expected: "✓ hello",
		},
		{
			name:     "string with color and unicode",
			input:    "\033[32m✓ success\033[0m",
			expected: "✓ success",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "string with only ansi codes",
			input:    "\033[31m\033[1m\033[0m",
			expected: "",
		},
		{
			name:     "string with emoji",
			input:    "hello 👋",
			expected: "hello 👋",
		},
		{
			name:     "string with wide characters",
			input:    "\u4f60\u597d", // "你好"
			expected: "\u4f60\u597d",
		},
		{
			name:     "string with ZWJ emoji",
			input:    "👩‍👩‍👧‍👦",
			expected: "👩‍👩‍👧‍👦",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, StripCodes(tt.input))
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxWidth int
		expected string
	}{
		{
			name:     "simple string under limit",
			input:    "hello",
			maxWidth: 10,
			expected: "hello",
		},
		{
			name:     "simple string at limit",
			input:    "hello",
			maxWidth: 5,
			expected: "hello",
		},
		{
			name:     "simple string over limit",
			input:    "hello world",
			maxWidth: 5,
			expected: "hello",
		},
		{
			name:     "zero width returns empty",
			input:    "hello",
			maxWidth: 0,
			expected: "",
		},
		{
			name:     "negative width returns empty",
			input:    "hello",
			maxWidth: -1,
			expected: "",
		},
		{
			name:     "empty string",
			input:    "",
			maxWidth: 5,
			expected: "",
		},
		{
			name:     "string with ANSI colors under limit",
			input:    "\033[31mhello\033[0m",
			maxWidth: 10,
			expected: "\033[31mhello\033[0m",
		},
		{
			name:     "string with ANSI colors at limit",
			input:    "\033[31mhello\033[0m",
			maxWidth: 5,
			expected: "\033[31mhello\033[0m",
		},
		{
			name:     "string with ANSI colors over limit",
			input:    "\033[31mhello world\033[0m",
			maxWidth: 5,
			expected: "\033[31mhello",
		},
		{
			name:     "truncate in middle of ANSI sequence preserves escape",
			input:    "\033[31;1mhello world\033[0m",
			maxWidth: 7,
			expected: "\033[31;1mhello w",
		},
		{
			name:     "string with unicode characters",
			input:    "\u4f60\u597d\u4e16\u754c", // "你好世界"
			maxWidth: 6,
			expected: "\u4f60\u597d\u4e16", // "你好世"
		},
		{
			name:     "string with unicode and ANSI",
			input:    "\033[32m\u4f60\u597d\u4e16\u754c\033[0m", // "\033[32m你好世界\033[0m"
			maxWidth: 6,
			expected: "\033[32m\u4f60\u597d\u4e16", // "\033[32m你好世"
		},
		{
			name:     "string with emoji",
			input:    "hello 👋 world",
			maxWidth: 8,
			expected: "hello 👋",
		},
		{
			name:     "string with wide emoji",
			input:    "👩‍👩‍👧‍👦 family",
			maxWidth: 4,
			expected: "👩‍👩‍",
		},
		{
			name:     "only ANSI codes",
			input:    "\033[31m\033[1m\033[0m",
			maxWidth: 5,
			expected: "\033[31m\033[1m\033[0m",
		},
		{
			name:     "mixed content with various elements",
			input:    "\033[31mHello\033[0m \u4e16\u754c 👋 \033[32mworld\033[0m", // "\033[31mHello\033[0m 世界 👋 \033[32mworld\033[0m"
			maxWidth: 10,
			expected: "\033[31mHello\033[0m \u4e16\u754c", // "\033[31mHello\033[0m 世界"
		},
		{
			name:     "truncate exactly at character boundary",
			input:    "abcdef",
			maxWidth: 3,
			expected: "abc",
		},
		{
			name:     "complex ANSI with multiple codes",
			input:    "\033[31;1;4munderlined bold red\033[0m",
			maxWidth: 10,
			expected: "\033[31;1;4munderlined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateString(tt.input, tt.maxWidth)
			require.Equal(t, tt.expected, result)

			// Verify the result doesn't exceed the max width (only for positive widths)
			if tt.maxWidth > 0 {
				actualWidth := PrintableWidth(result)
				require.LessOrEqual(t, actualWidth, tt.maxWidth, "truncated string exceeds max width")
			}
		})
	}
}
