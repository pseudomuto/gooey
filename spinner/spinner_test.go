package spinner_test

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/pseudomuto/gooey/ansi"
	. "github.com/pseudomuto/gooey/spinner"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	s := New("test message")

	require.Equal(t, "test message", s.Message())
	require.Equal(t, ansi.Red, s.Color()) // Default is now first color in rotation
	require.False(t, s.IsRunning())
	require.True(t, s.ShowElapsed()) // Default is to show elapsed time
}

func TestNewWithOptions(t *testing.T) {
	var buf bytes.Buffer

	s := New("test",
		WithColor(ansi.Red),
		WithInterval(50*time.Millisecond),
		WithOutput(&buf),
		WithRenderer(Clock),
		WithShowElapsed(false))

	require.Equal(t, ansi.Red, s.Color())
	require.Equal(t, "test", s.Message())
	require.False(t, s.IsRunning())
	require.False(t, s.ShowElapsed())
}

func TestStartStop(t *testing.T) {
	var buf bytes.Buffer
	s := New("test", WithOutput(&buf))

	require.False(t, s.IsRunning())

	s.Start()
	require.True(t, s.IsRunning())

	time.Sleep(10 * time.Millisecond)

	s.Stop()
	require.False(t, s.IsRunning())

	output := buf.String()
	require.Contains(t, output, "test")
	require.Contains(t, output, ansi.CheckMark.String())
}

func TestUpdateMessage(t *testing.T) {
	s := New("initial")

	require.Equal(t, "initial", s.Message())

	s.UpdateMessage("updated")
	require.Equal(t, "updated", s.Message())
}

func TestDoubleStart(t *testing.T) {
	var buf bytes.Buffer
	s := New("test", WithOutput(&buf))

	s.Start()
	require.True(t, s.IsRunning())

	s.Start()
	require.True(t, s.IsRunning())

	s.Stop()
	require.False(t, s.IsRunning())
}

func TestDoubleStop(t *testing.T) {
	var buf bytes.Buffer
	s := New("test", WithOutput(&buf))

	s.Start()
	s.Stop()
	require.False(t, s.IsRunning())

	s.Stop()
	require.False(t, s.IsRunning())
}

func TestElapsed(t *testing.T) {
	s := New("test")

	require.Equal(t, time.Duration(0), s.Elapsed())

	s.Start()
	time.Sleep(10 * time.Millisecond)

	elapsed := s.Elapsed()
	require.GreaterOrEqual(t, elapsed, 10*time.Millisecond)

	s.Stop()
	require.Equal(t, time.Duration(0), s.Elapsed())
}

func TestRenderers(t *testing.T) {
	tests := []struct {
		name     string
		renderer SpinnerRenderer
	}{
		{"Dots", Dots},
		{"Clock", Clock},
		{"Arrow", Arrow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			s := New("test message",
				WithRenderer(tt.renderer),
				WithOutput(&buf),
				WithInterval(10*time.Millisecond))

			s.Start()
			time.Sleep(50 * time.Millisecond)
			s.Stop()

			output := buf.String()
			require.Contains(t, output, "test message")
			require.NotEmpty(t, output)
		})
	}
}

func TestCustomRenderer(t *testing.T) {
	var buf bytes.Buffer

	customRenderer := RenderFunc(func(s *Spinner, frame int, w io.Writer) {
		fmt.Fprint(w, "CUSTOM: "+s.Message())
	})

	s := New("test",
		WithRenderer(customRenderer),
		WithOutput(&buf),
		WithInterval(10*time.Millisecond))

	s.Start()
	time.Sleep(20 * time.Millisecond)
	s.Stop()

	output := buf.String()
	require.Contains(t, output, "CUSTOM: test")
}

func TestSpinnerAnimation(t *testing.T) {
	var buf bytes.Buffer
	s := New("animating",
		WithRenderer(Dots),
		WithOutput(&buf),
		WithInterval(5*time.Millisecond))

	s.Start()
	time.Sleep(25 * time.Millisecond)
	s.Stop()

	output := buf.String()

	require.Contains(t, output, "animating")

	spinnerCount := strings.Count(output, ansi.Spinner1.String()) +
		strings.Count(output, ansi.Spinner2.String()) +
		strings.Count(output, ansi.Spinner3.String()) +
		strings.Count(output, ansi.Spinner4.String())

	require.Positive(t, spinnerCount, "Should contain spinner icons")
}

func TestMessageUpdate(t *testing.T) {
	var buf bytes.Buffer
	s := New("initial message",
		WithOutput(&buf),
		WithInterval(5*time.Millisecond))

	s.Start()
	time.Sleep(10 * time.Millisecond)

	s.UpdateMessage("updated message")
	time.Sleep(10 * time.Millisecond)

	s.Stop()

	output := buf.String()
	require.Contains(t, output, "updated message")
}

func TestColorRotation(t *testing.T) {
	s := New("test")

	// Test that colors rotate through the expected sequence
	require.Equal(t, ansi.Red, s.CurrentColor(0))
	require.Equal(t, ansi.Blue, s.CurrentColor(1))
	require.Equal(t, ansi.Cyan, s.CurrentColor(2))
	require.Equal(t, ansi.Magenta, s.CurrentColor(3))
	require.Equal(t, ansi.Red, s.CurrentColor(4)) // Should wrap around

	// Test that custom color overrides rotation
	s = New("test", WithColor(ansi.Green))
	require.Equal(t, ansi.Green, s.CurrentColor(0))
	require.Equal(t, ansi.Green, s.CurrentColor(1))
	require.Equal(t, ansi.Green, s.CurrentColor(100))

	// Test that setting a rotation color via WithColor still uses that fixed color
	s = New("test", WithColor(ansi.Red))
	require.Equal(t, ansi.Red, s.CurrentColor(0))
	require.Equal(t, ansi.Red, s.CurrentColor(1))
	require.Equal(t, ansi.Red, s.CurrentColor(2))
}

func TestElapsedTimeOption(t *testing.T) {
	var buf bytes.Buffer

	// Test with elapsed time enabled (default)
	s1 := New("test", WithOutput(&buf))
	s1.Start()
	time.Sleep(10 * time.Millisecond)
	s1.Stop()

	output1 := buf.String()
	require.Contains(t, output1, "test")
	require.Contains(t, output1, ansi.CheckMark.String())
	require.Contains(t, output1, "(") // Should contain elapsed time in parentheses

	// Test with elapsed time disabled
	buf.Reset()
	s2 := New("test2", WithOutput(&buf), WithShowElapsed(false))
	s2.Start()
	time.Sleep(10 * time.Millisecond)
	s2.Stop()

	output2 := buf.String()
	require.Contains(t, output2, "test2")
	require.Contains(t, output2, ansi.CheckMark.String())
	require.NotContains(t, output2, "(") // Should not contain elapsed time
}

func TestSpinnerFailure(t *testing.T) {
	var buf bytes.Buffer
	s := New("test task", WithOutput(&buf))

	require.False(t, s.IsRunning())
	require.Equal(t, SpinnerCompleted, s.State()) // Default state

	s.Start()
	require.True(t, s.IsRunning())

	time.Sleep(10 * time.Millisecond)

	s.Fail("")
	require.False(t, s.IsRunning())
	require.Equal(t, SpinnerFailed, s.State())

	output := buf.String()
	require.Contains(t, output, "test task")
	require.Contains(t, output, ansi.CrossMark.String())
	require.NotContains(t, output, ansi.CheckMark.String())
}

func TestSpinnerSuccess(t *testing.T) {
	var buf bytes.Buffer
	s := New("test task", WithOutput(&buf))

	s.Start()
	time.Sleep(10 * time.Millisecond)
	s.Stop()

	require.Equal(t, SpinnerCompleted, s.State())

	output := buf.String()
	require.Contains(t, output, "test task")
	require.Contains(t, output, ansi.CheckMark.String())
	require.NotContains(t, output, ansi.CrossMark.String())
}

func TestDoubleFailure(t *testing.T) {
	var buf bytes.Buffer
	s := New("test", WithOutput(&buf))

	s.Start()
	s.Fail("")
	require.False(t, s.IsRunning())
	require.Equal(t, SpinnerFailed, s.State())

	s.Fail("") // Should not panic or cause issues
	require.False(t, s.IsRunning())
	require.Equal(t, SpinnerFailed, s.State())
}

func TestFailureWithElapsedTime(t *testing.T) {
	var buf bytes.Buffer
	s := New("test task", WithOutput(&buf))

	s.Start()
	time.Sleep(10 * time.Millisecond)
	s.Fail("")

	output := buf.String()
	require.Contains(t, output, "test task")
	require.Contains(t, output, ansi.CrossMark.String())
	require.Contains(t, output, "(") // Should contain elapsed time in parentheses
}

func TestFailureWithoutElapsedTime(t *testing.T) {
	var buf bytes.Buffer
	s := New("test task", WithOutput(&buf), WithShowElapsed(false))

	s.Start()
	time.Sleep(10 * time.Millisecond)
	s.Fail("")

	output := buf.String()
	require.Contains(t, output, "test task")
	require.Contains(t, output, ansi.CrossMark.String())
	require.NotContains(t, output, "(") // Should not contain elapsed time
}
