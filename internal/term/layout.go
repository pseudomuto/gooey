package term

import "strings"

// SectionLayout represents a flexible layout with variable number of columns and their relative widths.
// The Widths slice contains relative proportions (they don't need to sum to 1.0 or 100%).
// MinWidths contains the minimum width for each section (optional, can be nil).
type SectionLayout struct {
	TotalWidth int       // Total available width
	Widths     []float64 // Relative widths for each section (e.g., [0.2, 0.6, 0.2] or [1, 3, 1])
	MinWidths  []int     // Minimum widths for each section (optional, can be nil)
}

// NewSectionLayout creates a new SectionLayout with the given total width and relative proportions.
// Example: NewSectionLayout(100, 1, 3, 1) creates a 3-column layout with 20%, 60%, 20% proportions.
func NewSectionLayout(totalWidth int, weights ...float64) SectionLayout {
	return SectionLayout{
		TotalWidth: totalWidth,
		Widths:     weights,
		MinWidths:  nil,
	}
}

// TruncateAndPad truncates text to fit within maxWidth and pads it to exactly that width.
// If the text is longer than maxWidth, it's truncated with "..." suffix.
// If shorter, it's padded with spaces to reach exactly maxWidth.
func TruncateAndPad(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	// Truncate if too long
	if PrintableWidth(text) > maxWidth {
		if maxWidth <= 3 {
			return strings.Repeat(".", min(maxWidth, 3))
		}
		text = TruncateString(text, maxWidth-3) + "..."
	}

	// Pad to exact width
	currentWidth := PrintableWidth(text)
	padding := max(maxWidth-currentWidth, 0)
	return text + strings.Repeat(" ", padding)
}

// SectionWidths calculates the actual widths for each section, applying minimums
// and proportional reduction if the total exceeds available width.
func (l SectionLayout) SectionWidths() []int {
	if len(l.Widths) == 0 {
		return []int{}
	}

	// Calculate the sum of relative widths to normalize proportions
	totalWeight := 0.0
	for _, weight := range l.Widths {
		totalWeight += weight
	}

	// Calculate initial widths based on proportions
	widths := make([]int, len(l.Widths))
	for i, weight := range l.Widths {
		proportion := weight / totalWeight
		widths[i] = int(float64(l.TotalWidth) * proportion)

		// Apply minimum width if specified
		if l.MinWidths != nil && i < len(l.MinWidths) {
			widths[i] = max(widths[i], l.MinWidths[i])
		}
	}

	// Check if total exceeds available width and adjust proportionally
	totalUsed := 0
	for _, width := range widths {
		totalUsed += width
	}

	if totalUsed > l.TotalWidth {
		// Proportionally reduce sections to fit
		ratio := float64(l.TotalWidth) / float64(totalUsed)

		// Apply ratio reduction while respecting minimums
		remainingWidth := l.TotalWidth
		for i := range widths {
			minWidth := 1 // Default minimum
			if l.MinWidths != nil && i < len(l.MinWidths) {
				minWidth = l.MinWidths[i]
			}

			if i == len(widths)-1 {
				// Last section gets remaining width, but respects minimum
				widths[i] = max(remainingWidth, minWidth)
			} else {
				widths[i] = max(int(float64(widths[i])*ratio), minWidth)
				remainingWidth -= widths[i]
			}
		}
	}

	return widths
}

// WithMinWidths sets minimum widths for each section and returns the updated layout.
// This method ensures that no section will be smaller than its minimum width,
// even when proportional scaling would make it smaller.
//
// Example:
//
//	// Create layout with minimum constraints
//	layout := term.NewSectionLayout(100, 1, 3, 1).WithMinWidths(15, 20, 10)
//	widths := layout.SectionWidths() // Ensures sections are at least 15, 20, 10 wide
//
//	// Without minimums: might get [20, 60, 20]
//	// With minimums: might get [20, 60, 20] or [15, 65, 20] depending on constraints
//
// If the sum of minimum widths exceeds the total width, the layout will
// still attempt to fit all sections by proportionally reducing non-minimum space.
func (l SectionLayout) WithMinWidths(minWidths ...int) SectionLayout {
	l.MinWidths = minWidths
	return l
}
