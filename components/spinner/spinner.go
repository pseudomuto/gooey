package spinner

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/pseudomuto/gooey/ansi"
	"github.com/pseudomuto/gooey/components/internal"
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
	Spinner struct {
		message     string
		color       ansi.Color
		customColor bool // tracks if color was explicitly set via WithColor
		showElapsed bool // tracks if elapsed time should be shown on completion
		frameAware  *internal.FrameAware
		interval    time.Duration
		running     bool
		startTime   time.Time
		stopChan    chan bool
		mutex       sync.RWMutex
		renderer    SpinnerRenderer
	}

	SpinnerOption func(*Spinner)
)

func New(message string, options ...SpinnerOption) *Spinner {
	s := &Spinner{
		message:     message,
		color:       spinnerColors[0], // Use first color as default
		customColor: false,            // Default is to use rotation
		showElapsed: true,             // Default is to show elapsed time
		frameAware:  internal.NewFrameAware(defaultSpinnerOutput),
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
	if !s.running {
		return
	}

	s.frameAware.RenderWithStringBuilder(func(w io.Writer) {
		s.renderer.Render(s, frame, w)
	})
}

func (s *Spinner) renderFinal() {
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

func WithColor(color ansi.Color) SpinnerOption {
	return func(s *Spinner) {
		s.color = color
		s.customColor = true
	}
}

func WithInterval(interval time.Duration) SpinnerOption {
	return func(s *Spinner) {
		if interval > 0 {
			s.interval = interval
		}
	}
}

func WithOutput(output io.Writer) SpinnerOption {
	return func(s *Spinner) {
		s.frameAware.SetOutput(output)
	}
}

func WithRenderer(renderer SpinnerRenderer) SpinnerOption {
	return func(s *Spinner) {
		s.renderer = renderer
	}
}

func WithShowElapsed(show bool) SpinnerOption {
	return func(s *Spinner) {
		s.showElapsed = show
	}
}

func (s *Spinner) IsRunning() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.running
}

func (s *Spinner) Message() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.message
}

func (s *Spinner) Color() ansi.Color {
	return s.color
}

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

func (s *Spinner) Elapsed() time.Duration {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if !s.running {
		return 0
	}
	return time.Since(s.startTime)
}
