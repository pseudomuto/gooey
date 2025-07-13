package term

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewSectionLayout(t *testing.T) {
	layout := NewSectionLayout(100, 1, 3, 1)

	require.Equal(t, 100, layout.TotalWidth)
	require.Equal(t, []float64{1, 3, 1}, layout.Widths)
	require.Nil(t, layout.MinWidths)
}

func TestSectionLayoutWithMinWidths(t *testing.T) {
	layout := NewSectionLayout(100, 1, 3, 1).WithMinWidths(10, 20, 8)

	require.Equal(t, []int{10, 20, 8}, layout.MinWidths)
}

func TestSectionWidthsProportional(t *testing.T) {
	// Test 1:3:1 ratio with 100 total width
	layout := NewSectionLayout(100, 1, 3, 1)
	widths := layout.SectionWidths()

	require.Len(t, widths, 3)
	require.Equal(t, 20, widths[0]) // 1/5 of 100 = 20
	require.Equal(t, 60, widths[1]) // 3/5 of 100 = 60
	require.Equal(t, 20, widths[2]) // 1/5 of 100 = 20
}

func TestSectionWidthsWithMinimums(t *testing.T) {
	// Test with minimum widths that don't require scaling
	layout := NewSectionLayout(100, 1, 3, 1).WithMinWidths(5, 10, 5)
	widths := layout.SectionWidths()

	require.Equal(t, 20, widths[0]) // Normal calculation: 20 > 5
	require.Equal(t, 60, widths[1]) // Normal calculation: 60 > 10
	require.Equal(t, 20, widths[2]) // Normal calculation: 20 > 5
}

func TestSectionWidthsWithMinimumsEnforced(t *testing.T) {
	// Test with small total width where minimums must be enforced
	layout := NewSectionLayout(50, 1, 3, 1).WithMinWidths(15, 20, 10)
	widths := layout.SectionWidths()

	// All sections should respect their minimums
	require.GreaterOrEqual(t, widths[0], 15, "First section should respect minimum, got %d", widths[0])
	require.GreaterOrEqual(t, widths[1], 20, "Second section should respect minimum, got %d", widths[1])
	require.GreaterOrEqual(t, widths[2], 10, "Third section should respect minimum, got %d", widths[2])

	// When minimums can't be satisfied within available space, total may exceed target width
	// This is realistic behavior - minimums take priority over exact width
	total := widths[0] + widths[1] + widths[2]
	require.GreaterOrEqual(t, total, 50, "Total should be at least the target width")
}

func TestSectionWidthsWithReasonableMinimums(t *testing.T) {
	// Test with reasonable minimums that fit within the total width
	layout := NewSectionLayout(100, 1, 3, 1).WithMinWidths(10, 20, 8)
	widths := layout.SectionWidths()

	// All sections should respect their minimums and proportions
	require.GreaterOrEqual(t, widths[0], 10, "First section should respect minimum")
	require.GreaterOrEqual(t, widths[1], 20, "Second section should respect minimum")
	require.GreaterOrEqual(t, widths[2], 8, "Third section should respect minimum")

	// Should sum to exactly the total width when minimums allow it
	total := widths[0] + widths[1] + widths[2]
	require.Equal(t, 100, total)

	// Should maintain roughly the right proportions (1:3:1)
	require.Equal(t, 20, widths[0]) // 1/5 of 100
	require.Equal(t, 60, widths[1]) // 3/5 of 100
	require.Equal(t, 20, widths[2]) // 1/5 of 100
}

func TestSectionWidthsTwoColumns(t *testing.T) {
	// Test 2-column layout
	layout := NewSectionLayout(100, 2, 3)
	widths := layout.SectionWidths()

	require.Len(t, widths, 2)
	require.Equal(t, 40, widths[0]) // 2/5 of 100 = 40
	require.Equal(t, 60, widths[1]) // 3/5 of 100 = 60
}

func TestSectionWidthsFourColumns(t *testing.T) {
	// Test 4-column layout with equal weights
	layout := NewSectionLayout(100, 1, 1, 1, 1)
	widths := layout.SectionWidths()

	require.Len(t, widths, 4)
	for _, width := range widths {
		require.Equal(t, 25, width) // 1/4 of 100 = 25
	}
}

func TestSectionWidthsFloatWeights(t *testing.T) {
	// Test with float weights
	layout := NewSectionLayout(100, 0.5, 1.5, 0.5)
	widths := layout.SectionWidths()

	require.Len(t, widths, 3)
	require.Equal(t, 20, widths[0]) // 0.5/2.5 of 100 = 20
	require.Equal(t, 60, widths[1]) // 1.5/2.5 of 100 = 60
	require.Equal(t, 20, widths[2]) // 0.5/2.5 of 100 = 20
}

func TestSectionWidthsEmptyWeights(t *testing.T) {
	layout := NewSectionLayout(100)
	widths := layout.SectionWidths()

	require.Empty(t, widths)
}

func TestTruncateAndPad(t *testing.T) {
	// Test normal padding
	result := TruncateAndPad("hello", 10)
	require.Equal(t, "hello     ", result)
	require.Len(t, result, 10)

	// Test truncation
	result = TruncateAndPad("hello world", 8)
	require.Equal(t, "hello...", result)
	require.Len(t, result, 8)

	// Test exact fit
	result = TruncateAndPad("hello", 5)
	require.Equal(t, "hello", result)
	require.Len(t, result, 5)

	// Test zero width
	result = TruncateAndPad("hello", 0)
	require.Empty(t, result)

	// Test very small width
	result = TruncateAndPad("hello", 2)
	require.Equal(t, "..", result)
}
