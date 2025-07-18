package spinner

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/pseudomuto/gooey/ansi"
	"github.com/pseudomuto/gooey/internal/frame"
)

const (
	defaultSpinnerInterval = 100 * time.Millisecond
)

var (
	spinnerColors = []ansi.Color{
		ansi.Red,
		ansi.Blue,
		ansi.Cyan,
		ansi.Magenta,
	}

	defaultSpinnerOutput io.Writer = os.Stdout
)

type (
	// Spinner represents an animated spinner that can display progress
	Spinner struct {
		message        string
		color          ansi.Color
		customColor    bool // tracks if color was explicitly set via WithColor
		showElapsed    bool // tracks if elapsed time should be shown on completion
		suppressRender bool // prevents rendering when used in groups
		frameAware     *frame.FrameAware
		interval       time.Duration
		running        bool
		startTime      time.Time
		stopChan       chan bool
		mutex          sync.RWMutex
		renderer       SpinnerRenderer
	}

	// SpinnerOption is a function type for configuring spinners
	SpinnerOption func(*Spinner)
)

// New creates a new animated spinner with the given message and options.
// The spinner provides a visual indicator for long-running operations with
// automatic color rotation, customizable animation styles, and frame integration.
//
// Example:
//
//	// Basic spinner with automatic color rotation
//	s := spinner.New("Loading data...")
//	s.Start()
//	time.Sleep(3 * time.Second)
//	s.Stop()
//
//	// Custom spinner with fixed color and different animation
//	s := spinner.New("Processing files...",
//		spinner.WithColor(ansi.Green),
//		spinner.WithRenderer(spinner.Clock),
//		spinner.WithInterval(200*time.Millisecond))
//	s.Start()
//	time.Sleep(2 * time.Second)
//	s.UpdateMessage("Almost done...")
//	s.Stop()
//
//	// Spinner within a frame
//	f := frame.Open("Deployment", frame.WithColor(ansi.Cyan))
//	s := spinner.New("Building...", spinner.WithOutput(f))
//	s.Start()
//	time.Sleep(2 * time.Second)
//	s.Stop()
//	f.Close()
//
// The spinner automatically rotates through colors (Red→Blue→Cyan→Magenta)
// unless WithColor() is used to set a fixed color. On completion, it shows
// a green checkmark and elapsed time (unless disabled with WithShowElapsed(false)).
func New(message string, options ...SpinnerOption) *Spinner {
	s := &Spinner{
		message:     message,
		color:       spinnerColors[0], // Use first color as default
		customColor: false,            // Default is to use rotation
		showElapsed: true,             // Default is to show elapsed time
		frameAware:  frame.NewFrameAware(defaultSpinnerOutput),
		interval:    defaultSpinnerInterval,
		running:     false,
		stopChan:    make(chan bool),
		renderer:    Dots,
	}

	for _, option := range options {
		option(s)
	}

	return s
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	s.mutex.Lock()
	if s.running {
		s.mutex.Unlock()
		return
	}

	s.running = true
	s.startTime = time.Now()
	s.mutex.Unlock()

	go s.animate()
}

// Stop ends the spinner animation and renders the final state
func (s *Spinner) Stop() {
	s.mutex.Lock()
	if !s.running {
		s.mutex.Unlock()
		return
	}

	s.running = false
	s.mutex.Unlock()

	s.stopChan <- true
	s.renderFinal()

	if !s.frameAware.InFrame() {
		fmt.Fprintln(s.frameAware.Output())
	}
}

// UpdateMessage changes the spinner message while it's running
func (s *Spinner) UpdateMessage(message string) {
	s.mutex.Lock()
	s.message = message
	s.mutex.Unlock()
}

func (s *Spinner) animate() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	frame := 0
	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.render(frame)
			frame++
		}
	}
}

func (s *Spinner) render(frame int) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if !s.running || s.suppressRender {
		return
	}

	s.frameAware.RenderWithStringBuilder(func(w io.Writer) {
		s.renderer.Render(s, frame, w)
	})
}

func (s *Spinner) renderFinal() {
	if s.suppressRender {
		return
	}

	checkmark := ansi.CheckMark.Colorize(ansi.Green)
	message := s.message

	var elapsedText string
	if s.showElapsed {
		elapsed := time.Since(s.startTime)
		elapsedText = " " + ansi.Cyan.Colorize(fmt.Sprintf("(%v)", elapsed.Truncate(time.Millisecond)))
	}

	s.frameAware.RenderFinal(func() string {
		return fmt.Sprintf("%s %s%s", checkmark, message, elapsedText)
	})
}

// WithColor sets a custom color for the spinner
func WithColor(color ansi.Color) SpinnerOption {
	return func(s *Spinner) {
		s.color = color
		s.customColor = true
	}
}

// WithInterval sets the animation interval for the spinner
func WithInterval(interval time.Duration) SpinnerOption {
	return func(s *Spinner) {
		if interval > 0 {
			s.interval = interval
		}
	}
}

// WithOutput sets the output writer for the spinner
func WithOutput(output io.Writer) SpinnerOption {
	return func(s *Spinner) {
		s.frameAware.SetOutput(output)
	}
}

// WithRenderer sets a custom renderer for the spinner
func WithRenderer(renderer SpinnerRenderer) SpinnerOption {
	return func(s *Spinner) {
		s.renderer = renderer
	}
}

// WithShowElapsed controls whether elapsed time is shown on completion
func WithShowElapsed(show bool) SpinnerOption {
	return func(s *Spinner) {
		s.showElapsed = show
	}
}

// WithSuppressRender controls whether the spinner renders output (useful for testing)
func WithSuppressRender(suppress bool) SpinnerOption {
	return func(s *Spinner) {
		s.suppressRender = suppress
	}
}

// IsRunning returns true if the spinner is currently animating
func (s *Spinner) IsRunning() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.running
}

// Message returns the current spinner message
func (s *Spinner) Message() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.message
}

// Color returns the spinner's color
func (s *Spinner) Color() ansi.Color {
	return s.color
}

// ShowElapsed returns whether elapsed time is shown on completion
func (s *Spinner) ShowElapsed() bool {
	return s.showElapsed
}

// CurrentColor returns the color for the current frame, rotating through spinnerColors
func (s *Spinner) CurrentColor(frame int) ansi.Color {
	if s.customColor {
		// Custom color was explicitly set via WithColor, use it instead of rotating
		return s.color
	}
	// Use rotating colors
	return spinnerColors[frame%len(spinnerColors)]
}

// Elapsed returns the duration since the spinner started
func (s *Spinner) Elapsed() time.Duration {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if !s.running {
		return 0
	}
	return time.Since(s.startTime)
}
