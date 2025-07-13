package progress

import (
	"fmt"
	"io"
	"strings"

	"github.com/pseudomuto/gooey/term"
)

var (
	// Bar is a ProgressRenderer that renders a progress bar.
	Bar = NewChar("█", "░")

	// Dots is a ProgressRenderer that renderes dots for showing progress.
	Dots = NewChar("●", "○")

	// Minimal is a ProgressRenderer that only uses percentage completed to show progress.
	Minimal = new(minimalRenderer)
)

type (
	// ProgressRenderer interface for different progress bar styles
	// Custom renderers can be implemented to create new progress bar styles
	ProgressRenderer interface {
		Render(p *Progress, w io.Writer)
	}

	// charRenderer implements ProgressRenderer for character based progress bars. (e.g. =/-)
	charRenderer struct {
		completed string
		pending   string
	}

	// minimalRenderer implements ProgressRenderer for Minimal style progress bars
	minimalRenderer struct{}

	rendererFunc func(*Progress, io.Writer)
)

func NewChar(completed, pending string) ProgressRenderer {
	return &charRenderer{
		completed: completed,
		pending:   pending,
	}
}

func RenderFunc(fn func(*Progress, io.Writer)) ProgressRenderer {
	return rendererFunc(fn)
}

func (r rendererFunc) Render(p *Progress, w io.Writer) {
	r(p, w)
}

func (r *charRenderer) Render(p *Progress, w io.Writer) {
	// Calculate total available width (depends on context)
	totalWidth := term.Width()
	if p.inFrame {
		totalWidth = totalWidth - 6 // Account for frame borders and padding
	}

	// Create three-section layout: (20%, 70%, 10%)
	layout := term.
		NewSectionLayout(totalWidth, 2, 7, 1).
		WithMinWidths(10, 20, 8)

	widths := layout.SectionWidths()
	titleWidth, progressWidth, updateWidth := widths[0], widths[1], widths[2]

	// Section 1: Title (left)
	titleSection := term.TruncateAndPad(p.Title(), titleWidth)

	// Section 2: Progress Bar (middle)
	progressSection := r.buildProgressSection(p, progressWidth)

	// Section 3: Update Text (right)
	updateText := p.Message()
	updateSection := term.TruncateAndPad(updateText, updateWidth)

	// Combine all sections
	fmt.Fprint(w, titleSection+progressSection+updateSection)
}

// buildProgressSection creates the formatted progress bar section
func (r *charRenderer) buildProgressSection(p *Progress, sectionWidth int) string {
	// Build the non-bar parts first to calculate exact space needed
	percentage := fmt.Sprintf(" %5.1f%%", p.Percentage())
	count := fmt.Sprintf(" (%02d/%02d) ", p.Current(), p.Total())
	brackets := "[]" // 2 characters

	// Calculate actual bar width (excluding brackets, percentage, and count)
	nonBarWidth := len(brackets) + term.PrintableWidth(percentage) + term.PrintableWidth(count)
	barWidth := max(sectionWidth-nonBarWidth, 5)

	// Build the progress bar
	var bar strings.Builder
	filled := int(float64(barWidth) * float64(p.current) / float64(p.total))
	for i := range barWidth {
		if i < filled {
			bar.WriteString(r.completed)
		} else {
			bar.WriteString(r.pending)
		}
	}

	// Build progress section with bar, percentage, and count
	var progressSection strings.Builder
	progressSection.WriteString("[")
	progressSection.WriteString(p.Color().Sprint(bar.String()))
	progressSection.WriteString("]")
	progressSection.WriteString(percentage)
	progressSection.WriteString(count)

	// Pad progress section to exact width (don't truncate the essential parts)
	currentWidth := term.PrintableWidth(progressSection.String())
	if currentWidth < sectionWidth {
		progressSection.WriteString(strings.Repeat(" ", sectionWidth-currentWidth))
	}

	return progressSection.String()
}

func (r *minimalRenderer) Render(p *Progress, w io.Writer) {
	percentage := float64(p.current) / float64(p.total) * 100

	var result strings.Builder

	// Title
	result.WriteString(p.title)
	result.WriteString(": ")

	// Colored percentage
	result.WriteString(p.color.Sprintf("%.1f%%", percentage))

	// Message if provided
	if p.message != "" {
		result.WriteString(" - ")
		result.WriteString(p.message)
	}

	fmt.Fprint(w, result.String())
}
