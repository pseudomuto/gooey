package progress_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/pseudomuto/gooey/ansi"
	"github.com/pseudomuto/gooey/internal/term"
	. "github.com/pseudomuto/gooey/progress"
	"github.com/stretchr/testify/require"
)

func TestProgressBasic(t *testing.T) {
	var buf bytes.Buffer
	p := New("Test Progress", 100, WithOutput(&buf))
	p.Update(0, "")

	// Initial state should show 0%
	output := buf.String()
	require.Contains(t, output, "Test Progress")
	require.Contains(t, output, "0.0%")
	require.Contains(t, output, "(00/100)")
}

func TestProgressUpdate(t *testing.T) {
	var buf bytes.Buffer
	p := New("Test Progress", 100, WithOutput(&buf))

	// Clear initial output
	buf.Reset()

	// Update to 50%
	p.Update(50, "Half done")

	output := buf.String()
	require.Contains(t, output, "50.0%")
	require.Contains(t, output, "(50/100)")
	require.Contains(t, output, "Half done")
}

func TestProgressIncrement(t *testing.T) {
	var buf bytes.Buffer
	p := New("Test Progress", 10, WithOutput(&buf))

	// Test increment
	buf.Reset()
	p.Increment("Step 1")

	output := buf.String()
	require.Contains(t, output, "10.0%") // 1/10 = 10%
	require.Contains(t, output, "(01/10)")
	require.Contains(t, output, "Step 1")

	// Test another increment
	buf.Reset()
	p.Increment("Step 2")

	output = buf.String()
	require.Contains(t, output, "20.0%") // 2/10 = 20%
	require.Contains(t, output, "(02/10)")
}

func TestProgressComplete(t *testing.T) {
	var buf bytes.Buffer
	p := New("Test Progress", 100, WithOutput(&buf))

	// Complete the progress
	buf.Reset()
	p.Complete("All done!")

	output := buf.String()
	require.Contains(t, output, "100.0%")
	require.Contains(t, output, "(100/100)")
	require.Contains(t, output, "All done!")
	require.True(t, strings.HasSuffix(output, "\n"), "Complete should add newline")
}

func TestProgressCompletePreventsUpdates(t *testing.T) {
	var buf bytes.Buffer
	p := New("Test Progress", 100, WithOutput(&buf))

	// Complete the progress
	p.Complete("Done")
	buf.Reset()

	// Try to update after completion
	p.Update(50, "Should be ignored")

	// Buffer should be empty since update was ignored
	output := buf.String()
	require.Empty(t, output, "Updates after completion should be ignored")
}

func TestProgressStyles(t *testing.T) {
	tests := []struct {
		name          string
		renderer      ProgressRenderer
		expectedChars []string
	}{
		{
			name:          "Bar Style",
			renderer:      Bar,
			expectedChars: []string{"█", "░", "[", "]"},
		},
		{
			name:          "Minimal Style",
			renderer:      Minimal,
			expectedChars: []string{":"},
		},
		{
			name:          "Dots Style",
			renderer:      Dots,
			expectedChars: []string{"●", "○"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			p := New("Test", 10, WithRenderer(tt.renderer), WithOutput(&buf))

			buf.Reset()
			p.Update(5, "Half way")

			output := buf.String()
			for _, char := range tt.expectedChars {
				require.Contains(t, output, char)
			}
		})
	}
}

func TestProgressColors(t *testing.T) {
	var buf bytes.Buffer
	p := New("Test Progress", 100, WithColor(ansi.Red), WithOutput(&buf))

	buf.Reset()
	p.Update(50, "Colored")

	output := buf.String()
	require.Contains(t, output, ansi.Red.String(), "Should contain red color code")
}

func TestProgressWidth(t *testing.T) {
	var buf bytes.Buffer
	p := New("Test", 100, WithWidth(20), WithOutput(&buf))

	buf.Reset()
	p.Update(50, "")

	output := buf.String()

	// Count the bar characters (█ and ░) - calculated using flexible layout system
	barChars := strings.Count(output, "█") + strings.Count(output, "░")
	require.Equal(t, 65, barChars, "Progress bar should use three-section layout with calculated bar width (53 characters)")
}

func TestProgressZeroWidth(t *testing.T) {
	var buf bytes.Buffer
	// Zero width should be ignored and use default
	p := New("Test", 100, WithWidth(0), WithOutput(&buf))

	buf.Reset()
	p.Update(50, "")

	output := buf.String()

	// Should use three-section layout (53 bar characters) when width is 0
	barChars := strings.Count(output, "█") + strings.Count(output, "░")
	require.Equal(t, 65, barChars, "Zero width should use three-section layout with calculated bar width (53 characters)")
}

func TestProgressGetters(t *testing.T) {
	p := New("Test", 100)
	eps := 0.0001

	// Test initial values
	require.Equal(t, 0, p.Current())
	require.Equal(t, 100, p.Total())
	require.Zero(t, p.Percentage())
	require.False(t, p.IsCompleted())

	// Test after update
	p.Update(25, "")
	require.Equal(t, 25, p.Current())
	require.InEpsilon(t, 25.0, p.Percentage(), eps)
	require.False(t, p.IsCompleted())

	// Test after completion
	p.Complete("Done")
	require.Equal(t, 100, p.Current())
	require.InEpsilon(t, 100.0, p.Percentage(), eps)
	require.True(t, p.IsCompleted())
}

func TestProgressElapsed(t *testing.T) {
	p := New("Test", 100)

	// Sleep a bit to ensure elapsed time
	time.Sleep(10 * time.Millisecond)

	elapsed := p.Elapsed()
	require.Positive(t, elapsed, "Elapsed time should be positive")
	require.GreaterOrEqual(t, elapsed, 10*time.Millisecond, "Elapsed time should be at least 10ms")
}

func TestProgressZeroTotal(t *testing.T) {
	p := New("Test", 0)

	// Should handle zero total gracefully
	require.Zero(t, p.Percentage())
}

func TestProgressBarRendering(t *testing.T) {
	var buf bytes.Buffer
	p := New("Download", 4, WithWidth(8), WithOutput(&buf))

	// Test 0% (no filled bars)
	buf.Reset()
	p.Update(0, "Starting")
	output := buf.String()
	require.Contains(t, output, "░░░░░░░░", "Should show all empty bars")
	require.Contains(t, output, "0.0%")

	// Test 50% (half filled)
	buf.Reset()
	p.Update(2, "Half done")
	output = buf.String()
	require.Contains(t, output, "████░░░░", "Should show half filled bars")
	require.Contains(t, output, "50.0%")

	// Test 100% (all filled)
	buf.Reset()
	p.Complete("Finished")
	output = buf.String()
	require.Contains(t, output, "████████", "Should show all filled bars")
	require.Contains(t, output, "100.0%")
}

func TestProgressDotsRendering(t *testing.T) {
	var buf bytes.Buffer
	p := New("Process", 4, WithRenderer(Dots), WithWidth(8), WithOutput(&buf))

	// Test 50% dots
	buf.Reset()
	p.Update(2, "Half done")
	output := buf.String()
	require.Contains(t, output, "●●●●○○○○", "Should show half filled dots")
	require.Contains(t, output, "50.0%")
}

func TestProgressMinimalRendering(t *testing.T) {
	var buf bytes.Buffer
	p := New("Task", 100, WithRenderer(Minimal), WithOutput(&buf))

	buf.Reset()
	p.Update(75, "Almost there")
	output := term.StripCodes(buf.String())

	require.Contains(t, output, "Task:")
	require.Contains(t, output, "75.0%")
	require.Contains(t, output, "Almost there")
	// Minimal style should not contain bar characters
	require.NotContains(t, output, "█")
	require.NotContains(t, output, "[")
}

func TestProgressTaskComponentInterface(t *testing.T) {
	var buf bytes.Buffer
	p := New("Upload", 10, WithOutput(&buf))

	// Test Start method (no-op for Progress)
	p.Start() // Should not panic or cause issues

	// Test normal progress updates
	p.Update(5, "Halfway")
	require.False(t, p.IsFailed())
	require.False(t, p.IsCompleted())

	// Test Fail method
	p.Fail("Failed")
	require.True(t, p.IsFailed())
	require.True(t, p.IsCompleted()) // Failed also means completed (no more updates)

	output := buf.String()
	require.Contains(t, output, "Failed")
}

func TestProgressFailRendering(t *testing.T) {
	var buf bytes.Buffer
	p := New("Test", 100, WithColor(ansi.Green), WithOutput(&buf))

	p.Update(50, "In progress")
	p.Fail("Error")

	output := buf.String()
	require.Contains(t, output, "Error")
	// Should contain red color for failed state (overriding green)
	require.Contains(t, output, ansi.Red.String())
}

func TestProgressSetOutput(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	p := New("Test", 100, WithOutput(&buf1))

	p.Update(25, "First")
	output1 := buf1.String()
	require.Contains(t, output1, "First")

	// Change output
	p.SetOutput(&buf2)
	p.Update(75, "Second")

	output2 := buf2.String()
	require.Contains(t, output2, "Second")
	require.NotContains(t, buf1.String(), "Second")
}
