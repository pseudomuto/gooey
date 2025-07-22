// Package writer provides specialized io.Writer implementations for terminal UI components.
// It includes utilities for indentation, frame integration, and ANSI sequence handling.
package writer

import (
	"fmt"
	"io"
	"strings"

	"github.com/pseudomuto/gooey/ansi"
	internalframe "github.com/pseudomuto/gooey/internal/frame"
)

// IndentedWriter wraps an io.Writer to add consistent indentation to output.
// It intelligently handles ANSI escape sequences, frame integration, and provides
// seamless ReplaceLine functionality for animated components like spinners and progress bars.
//
// The writer automatically detects and handles special content types:
//   - ANSI control sequences (passed through without indentation)
//   - Empty lines (preserved without indentation)
//   - Frame replacement operations (delegated with proper indentation)
//
// Example usage:
//
//	baseWriter := os.Stdout
//	indentedWriter := writer.NewIndentedWriter(baseWriter, 2) // 4 spaces of indentation
//	fmt.Fprintln(indentedWriter, "This line will be indented")
//
// Frame integration example:
//
//	frame := frame.Open("My Frame")
//	indentedFrame := writer.NewIndentedWriter(frame, 1) // 2 spaces of indentation
//	spinner := spinner.New("Loading...", spinner.WithOutput(indentedFrame))
//	spinner.Start() // Will animate with proper indentation within the frame
type IndentedWriter struct {
	writer io.Writer
	indent string
}

// NewIndentedWriter creates a writer that indents all output by the specified depth.
// Each depth level adds 2 spaces of indentation. A depth of 0 returns the original writer unchanged.
//
// The returned writer implements both io.Writer and, when appropriate, the FrameReplacer interface
// for seamless integration with frame-based components.
//
// Parameters:
//   - writer: The underlying writer to wrap
//   - depth: The indentation depth (0 = no indentation, 1 = 2 spaces, 2 = 4 spaces, etc.)
//
// Returns:
//   - An io.Writer that adds indentation to all output
//
// Example:
//
//	stdout := os.Stdout
//	indented := writer.NewIndentedWriter(stdout, 2)
//	fmt.Fprintln(indented, "Hello") // Outputs: "    Hello"
func NewIndentedWriter(writer io.Writer, depth int) io.Writer {
	if depth <= 0 {
		return writer
	}
	return &IndentedWriter{
		writer: writer,
		indent: strings.Repeat("  ", depth), // 2 spaces per depth level
	}
}

// Write implements io.Writer, adding indentation to each line of output.
// It intelligently handles special content types to ensure proper terminal behavior:
//
//   - ANSI control sequences are passed through without modification
//   - Empty lines are preserved without adding unnecessary indentation
//   - Multi-line content has indentation applied to each non-empty line
//
// The method preserves the original byte count in its return value for compatibility.
func (iw *IndentedWriter) Write(p []byte) (n int, err error) {
	// Convert bytes to string for processing
	content := string(p)

	// Handle special cases:
	// 1. ANSI control sequences - don't indent these
	// 2. Standalone newlines - these are completion newlines that shouldn't be indented
	// 3. Empty content - pass through without modification
	if strings.HasPrefix(content, "\r") ||
		strings.Contains(content, ansi.ClearLine) ||
		content == "\n" || // Standalone newline from spinner completion
		strings.TrimSpace(content) == "" {
		// Pass through control sequences, standalone newlines, and empty content without indentation
		_, err = iw.writer.Write(p)
		if err != nil {
			return 0, err
		}
		return len(p), nil
	}

	// Add indentation to each line, but skip empty lines to avoid extra blank lines
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		// Only add indentation to non-empty lines
		if line != "" {
			lines[i] = iw.indent + line
		}
		// Empty lines stay empty (no indentation)
	}

	indentedContent := strings.Join(lines, "\n")

	// Write the indented content
	_, err = iw.writer.Write([]byte(indentedContent))
	if err != nil {
		return 0, err
	}

	// Return the number of bytes from the original input
	return len(p), nil
}

// IsFrameWriter checks if this indented writer wraps a frame writer.
// This enables proper frame detection even through the indentation wrapper,
// allowing components like spinners to correctly identify frame contexts.
//
// Returns true if the underlying writer is a frame or frame-capable writer.
func (iw *IndentedWriter) IsFrameWriter() bool {
	return internalframe.IsFrameWriter(iw.writer)
}

// ReplaceLine implements the FrameReplacer interface for indented writers.
// This method enables proper in-place line updates for animated components
// like spinners and progress bars, while maintaining correct indentation.
//
// The indentation is applied to the replacement content before delegating
// to the underlying writer's ReplaceLine implementation.
//
// Parameters:
//   - format: Printf-style format string
//   - a: Arguments for the format string
//
// Example:
//
//	// This will replace the current line with indented content
//	indentedWriter.ReplaceLine("Progress: %d%%", 75)
func (iw *IndentedWriter) ReplaceLine(format string, a ...any) {
	if replacer, ok := iw.writer.(internalframe.FrameReplacer); ok {
		// Apply indentation to the content
		content := fmt.Sprintf(format, a...)
		if content != "" {
			content = iw.indent + content
		}
		replacer.ReplaceLine("%s", content)
	}
}

// ReplaceLineN implements the FrameReplacer interface for indented writers.
// This method enables replacement of specific lines by position while maintaining indentation.
//
// Parameters:
//   - linePosition: Number of lines back from current position (1 = previous line)
//   - format: Printf-style format string
//   - a: Arguments for the format string
func (iw *IndentedWriter) ReplaceLineN(linePosition int, format string, a ...any) {
	if replacer, ok := iw.writer.(internalframe.FrameReplacer); ok {
		// Apply indentation to the content
		content := fmt.Sprintf(format, a...)
		if content != "" {
			content = iw.indent + content
		}
		replacer.ReplaceLineN(linePosition, "%s", content)
	}
}

// ReplaceBlock implements the FrameReplacer interface for indented writers.
// This method enables replacement of multiple lines while maintaining consistent indentation.
//
// Parameters:
//   - lineCount: Number of lines to replace
//   - lines: Slice of replacement line content
func (iw *IndentedWriter) ReplaceBlock(lineCount int, lines []string) {
	if replacer, ok := iw.writer.(internalframe.FrameReplacer); ok {
		// Apply indentation to each non-empty line
		indentedLines := make([]string, len(lines))
		for i, line := range lines {
			if line != "" {
				indentedLines[i] = iw.indent + line
			} else {
				indentedLines[i] = line
			}
		}
		replacer.ReplaceBlock(lineCount, indentedLines)
	}
}
