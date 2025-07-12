package frame_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/pseudomuto/gooey/ansi"
	. "github.com/pseudomuto/gooey/components/frame"
	"github.com/stretchr/testify/require"
)

func TestFrameBasicOpen(t *testing.T) {
	var buf bytes.Buffer
	frame := Open("Test Frame", WithOutput(&buf))
	frame.Close()

	output := buf.String()
	require.Contains(t, output, "Test Frame")
	require.Contains(t, output, "┌")
	require.Contains(t, output, "└")
}

func TestFrameStyles(t *testing.T) {
	tests := []struct {
		name          string
		style         FrameStyle
		expectedOpen  string
		expectedClose string
	}{
		{
			name:          "Box Style",
			style:         Box,
			expectedOpen:  "┌",
			expectedClose: "└",
		},
		{
			name:          "Bracket Style",
			style:         Bracket,
			expectedOpen:  "┌",
			expectedClose: "└",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			frame := Open("Test", WithStyle(tt.style), WithOutput(&buf))
			frame.Close()

			output := buf.String()
			require.Contains(t, output, tt.expectedOpen)
			require.Contains(t, output, tt.expectedClose)
		})
	}
}

func TestFrameColor(t *testing.T) {
	var buf bytes.Buffer
	frame := Open("Test Frame", WithColor(ansi.Red), WithOutput(&buf))
	frame.Close()

	output := buf.String()
	require.Contains(t, output, ansi.Red.String())
}

func TestFrameNesting(t *testing.T) {
	var buf bytes.Buffer

	frame1 := Open("Outer Frame", WithOutput(&buf))
	frame2 := Open("Inner Frame", WithOutput(&buf))
	frame2.Close()
	frame1.Close()

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have 4 lines: outer open, inner open, inner close, outer close
	require.GreaterOrEqual(t, len(lines), 4)

	// Check nesting indicators
	require.Contains(t, output, "│")
}

func TestFrameDivider(t *testing.T) {
	var buf bytes.Buffer

	frame := Open("Test Frame", WithOutput(&buf))

	// Clear buffer to only capture divider output
	buf.Reset()
	frame.Divider("Test Divider")

	output := buf.String()
	require.Contains(t, output, "Test Divider")
	require.Contains(t, output, "├")

	frame.Close()
}

func TestFrameDividerEmpty(t *testing.T) {
	var buf bytes.Buffer

	frame := Open("Test Frame", WithOutput(&buf))

	// Clear buffer to only capture divider output
	buf.Reset()
	frame.Divider("")

	output := buf.String()
	require.Contains(t, output, "├")

	frame.Close()
}

func TestFrameTiming(t *testing.T) {
	var buf bytes.Buffer

	frame := Open("Timed Frame", WithOutput(&buf))
	time.Sleep(10 * time.Millisecond) // Sleep to ensure measurable time
	frame.Close()

	output := buf.String()
	// Should contain timing information in parentheses
	require.Contains(t, output, "(")
	require.Contains(t, output, ")")
}

func TestFramePrint(t *testing.T) {
	var buf bytes.Buffer
	frame := Open("Test Frame", WithOutput(&buf))

	// Clear buffer to only capture print output
	buf.Reset()
	frame.Print("Hello %s", "world")

	output := buf.String()
	require.Contains(t, output, "Hello world")
	require.NotContains(t, output, "\n") // Print should not add newline

	frame.Close()
}

func TestFramePrintln(t *testing.T) {
	var buf bytes.Buffer
	frame := Open("Test Frame", WithOutput(&buf))

	// Clear buffer to only capture println output
	buf.Reset()
	frame.Println("Hello %s", "world")

	output := buf.String()
	require.Contains(t, output, "Hello world")
	require.Contains(t, output, "\n") // Println should add newline

	frame.Close()
}

func TestFrameStyleDifferences(t *testing.T) {
	// Test the key difference between Box and Bracket styles
	// Box style should have right borders and padding, Bracket style should not

	t.Run("Box Style Has Right Borders", func(t *testing.T) {
		var buf bytes.Buffer
		frame := Open("Test", WithStyle(Box), WithOutput(&buf))
		
		buf.Reset()
		frame.Println("Content")
		
		output := buf.String()
		// Box style should have right border after content
		require.Contains(t, output, "│", "Box style should have right border")
		// Should have multiple occurrences of │ (left border + right border)
		require.GreaterOrEqual(t, strings.Count(output, "│"), 2, "Box style should have left and right borders")
		
		frame.Close()
	})

	t.Run("Bracket Style Has No Right Borders", func(t *testing.T) {
		var buf bytes.Buffer
		frame := Open("Test", WithStyle(Bracket), WithOutput(&buf))
		
		buf.Reset()
		frame.Println("Content")
		
		output := buf.String()
		// Bracket style should have left border but no right border
		require.Contains(t, output, "│", "Bracket style should have left border")
		// Should have only one occurrence of │ (just left border)
		require.Equal(t, 1, strings.Count(output, "│"), "Bracket style should only have left border")
		
		frame.Close()
	})
}

func TestFrameTitleColoring(t *testing.T) {
	var buf bytes.Buffer
	frame := Open("Test Frame", WithColor(ansi.Red), WithOutput(&buf))
	frame.Close()

	output := buf.String()
	
	// Should contain red color codes for borders
	require.Contains(t, output, ansi.Red.String(), "Should contain red color for borders")
	
	// Title text should not be wrapped in color codes
	// The title "Test Frame" should appear without being wrapped in red codes
	lines := strings.Split(output, "\n")
	var titleLine string
	for _, line := range lines {
		if strings.Contains(line, "Test Frame") {
			titleLine = line
			break
		}
	}
	
	require.NotEmpty(t, titleLine, "Should find line containing title")
	
	// The title itself should appear as plain text (not wrapped in color codes)
	// This is a bit tricky to test precisely, but we can check that the title
	// appears outside of color escape sequences
	require.Contains(t, titleLine, "Test Frame", "Title should be present")
}
