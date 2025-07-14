package spinner

import (
	"fmt"
	"io"

	"github.com/pseudomuto/gooey/ansi"
)

var (
	Dots  = &dotsRenderer{}
	Clock = &clockRenderer{}
	Arrow = &arrowRenderer{}
)

type (
	SpinnerRenderer interface {
		Render(s *Spinner, frame int, w io.Writer)
	}

	RenderFunc func(s *Spinner, frame int, w io.Writer)

	dotsRenderer  struct{}
	clockRenderer struct{}
	arrowRenderer struct{}
)

func (f RenderFunc) Render(s *Spinner, frame int, w io.Writer) {
	f(s, frame, w)
}

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

func (r *clockRenderer) Render(s *Spinner, frame int, w io.Writer) {
	icons := []ansi.Icon{
		ansi.Spinner1, ansi.Spinner3, ansi.Spinner5, ansi.Spinner7,
	}

	icon := icons[frame%len(icons)]
	color := s.CurrentColor(frame)
	coloredIcon := icon.Colorize(color)

	fmt.Fprintf(w, "%s %s", coloredIcon, s.message)
}

func (r *arrowRenderer) Render(s *Spinner, frame int, w io.Writer) {
	icons := []ansi.Icon{
		ansi.ArrowRight, ansi.ArrowDown, ansi.ArrowLeft, ansi.ArrowUp,
	}

	icon := icons[frame%len(icons)]
	color := s.CurrentColor(frame)
	coloredIcon := icon.Colorize(color)

	fmt.Fprintf(w, "%s %s", coloredIcon, s.message)
}
