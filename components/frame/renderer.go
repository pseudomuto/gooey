package frame

import (
	"strings"
	"time"

	"github.com/pseudomuto/gooey/ansi"
	"github.com/pseudomuto/gooey/term"
)

const (
	// Box drawing characters
	boxTopLeft     = "┌"
	boxTopRight    = "┐"
	boxBottomLeft  = "└"
	boxBottomRight = "┘"
	boxHorizontal  = "─"
	boxVertical    = "│"
	boxTee         = "├"
	boxTeeRight    = "┤"
)

// frameDimensions holds calculated dimensions for frame rendering
type frameDimensions struct {
	depth              int
	parentPrefixWidth  int
	rightBorderWidth   int
	outerFrameBorders  int
	parentBorderSpaces int
	availableWidth     int
}

// calculateFrameDimensions computes the dimensions needed for frame rendering
func calculateFrameDimensions(termWidth int, depth int) frameDimensions {
	parentPrefixWidth := 0
	if depth > 1 {
		parentPrefixWidth = (depth - 1) * strlen(frameVerticalPrefix)
	}

	rightBorderWidth := strlen(boxVertical)
	outerFrameBorders := max(depth-1, 0) * rightBorderWidth
	parentBorderSpaces := max(depth-1, 0) * 1 // 1 space per parent border

	availableWidth := max(termWidth-parentPrefixWidth-outerFrameBorders-parentBorderSpaces, 10)

	return frameDimensions{
		depth:              depth,
		parentPrefixWidth:  parentPrefixWidth,
		rightBorderWidth:   rightBorderWidth,
		outerFrameBorders:  outerFrameBorders,
		parentBorderSpaces: parentBorderSpaces,
		availableWidth:     availableWidth,
	}
}

// renderParentPrefixes renders the prefixes for parent frames
func renderParentPrefixes(result *strings.Builder, depth int) {
	if depth > 1 {
		parentFrameColors := stack.frameColors(depth - 1)
		for i := 0; i < depth-1; i++ {
			var prefixColor ansi.Color
			if i < len(parentFrameColors) {
				prefixColor = parentFrameColors[i]
			} else {
				prefixColor = defaultFrameColor // fallback
			}
			result.WriteString(prefixColor.Sprint(frameVerticalPrefix))
		}
	}
}

// renderParentBorders renders the right borders for parent frames
func renderParentBorders(result *strings.Builder, depth int) {
	parentFrameColors := stack.frameColors(depth - 1)
	for i := depth - 2; i >= 0; i-- {
		var borderColor ansi.Color
		if i < len(parentFrameColors) {
			borderColor = parentFrameColors[i]
		} else {
			borderColor = defaultFrameColor // fallback
		}
		result.WriteString(" " + borderColor.Sprint(boxVertical))
	}
}

type (
	// frameRenderer defines the interface for different frame rendering styles
	frameRenderer interface {
		formatContentLine(content string, color ansi.Color) string
		formatContentLineWithDepth(content string, color ansi.Color, frameDepth int) string
		openFrame(title string, color ansi.Color) string
		closeFrame(elapsed time.Duration, color ansi.Color) string
		createDivider(text string, color ansi.Color) string
	}

	// boxRenderer implements frameRenderer for Box style frames
	boxRenderer struct {
		termWidth int
	}

	// bracketRenderer implements frameRenderer for Bracket style frames
	bracketRenderer struct {
		termWidth int
	}
)

func (r *boxRenderer) formatContentLine(content string, color ansi.Color) string {
	// Use the current stack depth for backward compatibility
	return r.formatContentLineWithDepth(content, color, stack.depth())
}

func (r *boxRenderer) formatContentLineWithDepth(content string, color ansi.Color, frameDepth int) string {
	// Use the specific frame depth instead of total stack depth
	depth := frameDepth

	// Calculate total prefix width: parent prefixes + current frame left border
	// Parent frames use frameVerticalPrefix ("│  "), current frame uses boxVertical + " " ("│ ")
	var totalPrefixWidth int
	if depth > 1 {
		totalPrefixWidth = (depth-1)*strlen(frameVerticalPrefix) + strlen(boxVertical+" ")
	} else {
		totalPrefixWidth = strlen(boxVertical + " ")
	}

	// Right border is just the current frame's right border
	rightBorderWidth := strlen(boxVertical)

	// For nested frames, we need to account for the outer frame's right border
	// Each level of nesting reduces available width by the outer frame's right border
	outerFrameBorders := max(depth-1, 0) * rightBorderWidth

	// Account for spaces before parent borders - each parent border gets a space prefix
	parentBorderSpaces := max(depth-1, 0) * 1 // 1 space per parent border

	// Available content width accounts for all prefixes, right borders, and border spaces
	availableContentWidth := max(r.termWidth-totalPrefixWidth-rightBorderWidth-outerFrameBorders-parentBorderSpaces, 1)

	// Pre-process content to handle ANSI template syntax (e.g., {{bold+cyan:work}})
	// This ensures width calculations are based on the final rendered content
	processedContent := content

	if strings.Contains(content, "{{") && strings.Contains(content, "}}") {
		// Use QuickFormat to process template syntax and get the actual ANSI result
		processedContent = ansi.Format(content)
	}

	// Truncate content if too long - need to handle ANSI sequences properly
	if strlen(processedContent) > availableContentWidth {
		// For content with ANSI codes, we need to truncate based on printable width
		processedContent = term.TruncateString(processedContent, availableContentWidth-3) + "..."
	}

	// Pad content to exactly fill the available width
	// We need to be careful with ANSI sequences - the padding should be based on printable width
	contentPrintableWidth := strlen(processedContent)
	padding := max(availableContentWidth-contentPrintableWidth, 0)
	paddedContent := processedContent + strings.Repeat(" ", padding)

	// Get the colors of all frames in the stack up to current depth
	frameColors := stack.frameColors(depth)

	// Build the full line
	var result strings.Builder

	// Add all frame prefixes (including current frame's left border)
	// The last prefix becomes the current frame's left border
	for i := range depth {
		// Use the appropriate frame's color for each prefix
		var prefixColor ansi.Color
		if i < len(frameColors) {
			prefixColor = frameColors[i]
		} else {
			prefixColor = color // fallback to current frame color
		}

		if i == depth-1 {
			// Current frame - use left border
			result.WriteString(prefixColor.Sprint(boxVertical + " "))
		} else {
			// Parent frame - use vertical continuation
			result.WriteString(prefixColor.Sprint(frameVerticalPrefix))
		}
	}

	// Add the padded content
	result.WriteString(paddedContent)

	// Add right borders for current frame and all parent frames
	// Each frame's content should show ALL right borders from innermost to outermost
	// We only need colors up to this frame's depth
	allFrameColors := stack.frameColors(depth)

	// Show current frame border first
	if depth > 0 {
		currentFrameColor := color
		if depth-1 < len(allFrameColors) {
			currentFrameColor = allFrameColors[depth-1]
		}
		result.WriteString(currentFrameColor.Sprint(boxVertical))
	}

	// Then show parent frames from immediate parent to outermost parent
	// For frame at depth 3: immediate parent (index 1), then outermost parent (index 0)
	for i := depth - 2; i >= 0; i-- {
		// Current logic goes: 1, 0 (immediate parent, outermost) - this is CORRECT
		// But somehow it's still showing reverse order...
		var borderColor ansi.Color
		if i < len(allFrameColors) {
			borderColor = allFrameColors[i]
		} else {
			borderColor = color // fallback
		}
		result.WriteString(" " + borderColor.Sprint(boxVertical))
	}

	return result.String()
}

func (r *boxRenderer) openFrame(title string, color ansi.Color) string {
	depth := stack.depth()
	dims := calculateFrameDimensions(r.termWidth, depth)

	// Pre-process title to handle ANSI template syntax (e.g., {{unicorn:}})
	// This ensures width calculations are based on the final rendered content
	processedTitle := title
	if strings.Contains(title, "{{") && strings.Contains(title, "}}") {
		processedTitle = ansi.Format(title)
	}

	// Create title with spaces but without color applied yet
	titleWithSpaces := " " + processedTitle + " "
	if strlen(titleWithSpaces) > dims.availableWidth-4 {
		// Truncate title if too long
		maxTitleLen := dims.availableWidth - 6 // Space for borders and spaces
		if maxTitleLen > 0 {
			processedTitle = processedTitle[:maxTitleLen] + "..."
			titleWithSpaces = " " + processedTitle + " "
		} else {
			titleWithSpaces = ""
		}
	}

	// Build the line
	var result strings.Builder

	// Add parent frame prefixes with their original colors
	renderParentPrefixes(&result, depth)

	// Add current frame's top border with title in default color and borders in frame color
	horizontalFill := max(dims.availableWidth-4-strlen(titleWithSpaces), 0)
	leftBorder := color.Sprint(boxTopLeft + strings.Repeat(boxHorizontal, 2))
	rightBorder := color.Sprint(strings.Repeat(boxHorizontal, horizontalFill) + boxTopRight)

	result.WriteString(leftBorder)
	result.WriteString(titleWithSpaces) // Title in default color
	result.WriteString(rightBorder)

	// Add parent frame right borders
	renderParentBorders(&result, depth)

	result.WriteString("\n")
	return result.String()
}

func (r *boxRenderer) closeFrame(elapsed time.Duration, color ansi.Color) string {
	depth := stack.depth()
	dims := calculateFrameDimensions(r.termWidth, depth)

	// Create bottom border with timing
	var timingText string
	if elapsed > time.Millisecond {
		timingText = " (" + elapsed.Round(time.Millisecond).String() + ") "
	}

	// Build the line
	var result strings.Builder

	// Add parent frame prefixes with their original colors
	renderParentPrefixes(&result, depth)

	// Add current frame's bottom border
	horizontalFill := max(dims.availableWidth-4-strlen(timingText), 0)
	bottomBorder := boxBottomLeft + strings.Repeat(boxHorizontal, 2) + strings.Repeat(boxHorizontal, horizontalFill) + timingText + boxBottomRight
	result.WriteString(color.Sprint(bottomBorder))

	// Add parent frame right borders
	renderParentBorders(&result, depth)

	result.WriteString("\n")
	return result.String()
}

func (r *boxRenderer) createDivider(text string, color ansi.Color) string {
	depth := stack.depth()
	dims := calculateFrameDimensions(r.termWidth, depth)

	// Create divider text with spaces but without color applied yet
	var textWithSpaces string
	if text != "" {
		textWithSpaces = " " + text + " "
	}

	if strlen(textWithSpaces) > dims.availableWidth-4 {
		// Truncate text if too long
		maxTextLen := dims.availableWidth - 6 // Space for borders and spaces
		if maxTextLen > 0 {
			text = text[:maxTextLen] + "..."
			textWithSpaces = " " + text + " "
		} else {
			textWithSpaces = ""
		}
	}

	// Build the line
	var result strings.Builder

	// Add parent frame prefixes with their original colors
	renderParentPrefixes(&result, depth)

	// Add current frame's divider with text in default color and borders in frame color
	rightFill := max(dims.availableWidth-4-strlen(textWithSpaces), 0)
	leftBorder := color.Sprint(boxTee + strings.Repeat(boxHorizontal, 2))
	rightBorder := color.Sprint(strings.Repeat(boxHorizontal, rightFill) + boxTeeRight)

	result.WriteString(leftBorder)
	result.WriteString(textWithSpaces) // Text in default color
	result.WriteString(rightBorder)

	// Add parent frame right borders
	renderParentBorders(&result, depth)

	result.WriteString("\n")
	return result.String()
}

func (r *bracketRenderer) formatContentLine(content string, color ansi.Color) string {
	// Use the current stack depth for backward compatibility
	return r.formatContentLineWithDepth(content, color, stack.depth())
}

func (r *bracketRenderer) formatContentLineWithDepth(content string, color ansi.Color, frameDepth int) string {
	// Use the specific frame depth instead of total stack depth
	depth := frameDepth

	// Pre-process content to handle ANSI template syntax (e.g., {{bold+cyan:work}})
	// This ensures width calculations are based on the final rendered content
	processedContent := content
	if strings.Contains(content, "{{") && strings.Contains(content, "}}") {
		// Use QuickFormat to process template syntax and get the actual ANSI result
		processedContent = ansi.Format(content)
	}

	// Get the colors of all frames in the stack up to current depth
	frameColors := stack.frameColors(depth)

	// Build the full line
	var result strings.Builder

	// Add all frame prefixes (including current frame's left border)
	// The last prefix becomes the current frame's left border
	for i := range depth {
		// Use the appropriate frame's color for each prefix
		var prefixColor ansi.Color
		if i < len(frameColors) {
			prefixColor = frameColors[i]
		} else {
			prefixColor = color // fallback to current frame color
		}

		if i == depth-1 {
			// Current frame - use left border
			result.WriteString(prefixColor.Sprint(boxVertical + " "))
		} else {
			// Parent frame - use vertical continuation
			result.WriteString(prefixColor.Sprint(frameVerticalPrefix))
		}
	}

	// Add the content without any padding or right borders
	result.WriteString(processedContent)

	return result.String()
}

func (r *bracketRenderer) openFrame(title string, color ansi.Color) string {
	depth := stack.depth()
	dims := calculateFrameDimensions(r.termWidth, depth)

	// Pre-process title to handle ANSI template syntax (e.g., {{unicorn:}})
	// This ensures width calculations are based on the final rendered content
	processedTitle := title
	if strings.Contains(title, "{{") && strings.Contains(title, "}}") {
		processedTitle = ansi.Format(title)
	}

	// Create title with spaces but without color applied yet
	titleWithSpaces := " " + processedTitle + " "
	if strlen(titleWithSpaces) > dims.availableWidth-4 {
		// Truncate title if too long
		maxTitleLen := dims.availableWidth - 6 // Space for borders and spaces
		if maxTitleLen > 0 {
			processedTitle = processedTitle[:maxTitleLen] + "..."
			titleWithSpaces = " " + processedTitle + " "
		} else {
			titleWithSpaces = ""
		}
	}

	// Build the line
	var result strings.Builder

	// Add parent frame prefixes with their original colors
	renderParentPrefixes(&result, depth)

	// Add current frame's top border with title in default color and borders in frame color
	// For bracket style, we only add the left border without horizontal fill or right border
	leftBorder := color.Sprint(boxTopLeft + strings.Repeat(boxHorizontal, 2))

	result.WriteString(leftBorder)
	result.WriteString(titleWithSpaces) // Title in default color

	result.WriteString("\n")
	return result.String()
}

func (r *bracketRenderer) closeFrame(elapsed time.Duration, color ansi.Color) string {
	depth := stack.depth()

	// Create bottom border with timing
	var timingText string
	if elapsed > time.Millisecond {
		timingText = " (" + elapsed.Round(time.Millisecond).String() + ") "
	}

	// Build the line
	var result strings.Builder

	// Add parent frame prefixes with their original colors
	renderParentPrefixes(&result, depth)

	// Add current frame's bottom border without horizontal fill or right border
	bottomBorder := color.Sprint(boxBottomLeft + strings.Repeat(boxHorizontal, 2))
	result.WriteString(bottomBorder)
	result.WriteString(timingText) // Timing in default color

	result.WriteString("\n")
	return result.String()
}

func (r *bracketRenderer) createDivider(text string, color ansi.Color) string {
	// Create divider text with spaces but without color applied yet
	var textWithSpaces string
	if text != "" {
		textWithSpaces = " " + text + " "
	}

	// Build the line
	var result strings.Builder

	// Add parent frame prefixes with their original colors
	renderParentPrefixes(&result, stack.depth())

	// Add current frame's divider without horizontal fill or right border
	leftBorder := color.Sprint(boxTee + strings.Repeat(boxHorizontal, 2))

	result.WriteString(leftBorder)
	result.WriteString(textWithSpaces) // Text in default color

	result.WriteString("\n")
	return result.String()
}

func strlen(s string) int {
	return term.PrintableWidth(s)
}
