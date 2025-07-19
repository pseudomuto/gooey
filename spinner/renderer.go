package spinner

import (
	"fmt"
	"io"

	"github.com/pseudomuto/gooey/ansi"
)

var (
	// Dots renderer uses spinning dots for the spinner animation
	Dots = &dotsRenderer{}
	// Clock renderer uses clock icons for the spinner animation
	Clock = &clockRenderer{}
	// Arrow renderer uses arrow icons for the spinner animation
	Arrow = &arrowRenderer{}
)

type (
	// SpinnerRenderer defines the interface for rendering spinner animations
	SpinnerRenderer interface {
		Render(s *Spinner, frame int, w io.Writer)
	}

	// RenderFunc is a function type that implements SpinnerRenderer
	RenderFunc func(s *Spinner, frame int, w io.Writer)

	dotsRenderer  struct{}
	clockRenderer struct{}
	arrowRenderer struct{}
)

// Render executes the render function for the spinner
func (f RenderFunc) Render(s *Spinner, frame int, w io.Writer) {
	f(s, frame, w)
}

// Render implements SpinnerRenderer for the dots renderer.
// Uses 8-frame braille spinner icons with automatic color rotation.
func (r *dotsRenderer) Render(s *Spinner, frame int, w io.Writer) {
	icons := []ansi.Icon{
		ansi.Spinner1, ansi.Spinner2, ansi.Spinner3, ansi.Spinner4,
		ansi.Spinner5, ansi.Spinner6, ansi.Spinner7, ansi.Spinner8,
	}

	icon := icons[frame%len(icons)]
	color := s.CurrentColor(frame)
	coloredIcon := icon.Colorize(color)

	fmt.Fprintf(w, "%s %s", coloredIcon, s.message)
}

// Render implements SpinnerRenderer for the clock renderer.
// Uses a 4-frame subset of braille icons for slower animation.
func (r *clockRenderer) Render(s *Spinner, frame int, w io.Writer) {
	icons := []ansi.Icon{
		ansi.Spinner1, ansi.Spinner3, ansi.Spinner5, ansi.Spinner7,
	}

	icon := icons[frame%len(icons)]
	color := s.CurrentColor(frame)
	coloredIcon := icon.Colorize(color)

	fmt.Fprintf(w, "%s %s", coloredIcon, s.message)
}

// Render implements SpinnerRenderer for the arrow renderer.
// Uses directional arrows rotating clockwise (→↓←↑) with color cycling.
func (r *arrowRenderer) Render(s *Spinner, frame int, w io.Writer) {
	icons := []ansi.Icon{
		ansi.ArrowRight, ansi.ArrowDown, ansi.ArrowLeft, ansi.ArrowUp,
	}

	icon := icons[frame%len(icons)]
	color := s.CurrentColor(frame)
	coloredIcon := icon.Colorize(color)

	fmt.Fprintf(w, "%s %s", coloredIcon, s.message)
}
