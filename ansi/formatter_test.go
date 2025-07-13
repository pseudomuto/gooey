package ansi_test

import (
	"bytes"
	"strings"
	"testing"

	. "github.com/pseudomuto/gooey/ansi"
	"github.com/stretchr/testify/require"
)

func TestFormatterBasic(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf)

	// Test basic color formatting
	result := f.Format("{{red:Hello}} {{green:World}}")

	require.Contains(t, result, "\033[31mHello\033[0m", "Should contain red Hello")
	require.Contains(t, result, "\033[32mWorld\033[0m", "Should contain green World")
}

func TestFormatterStyles(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf)

	// Test style formatting
	result := f.Format("{{bold:Bold text}} {{italic:Italic text}}")

	require.Contains(t, result, "\033[1mBold text\033[0m", "Should contain bold text")
	require.Contains(t, result, "\033[3mItalic text\033[0m", "Should contain italic text")
}

func TestFormatterCombinations(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf)

	// Test color + style combinations
	result := f.Format("{{red+bold:Important}} {{green+italic:Emphasized}}")

	// Should contain both red and bold codes
	require.Contains(t, result, "\033[31m", "Should contain red color code")
	require.Contains(t, result, "\033[1m", "Should contain bold style code")
	require.Contains(t, result, "Important", "Should contain 'Important' text")
}

func TestFormatterSprintf(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf)

	// Test sprintf functionality
	result := f.Sprintf("{{red:Error}}: %s occurred at line %d", "ParseError", 42)

	require.Contains(t, result, "\033[31mError\033[0m", "Should contain colored Error")
	require.Contains(t, result, "ParseError occurred at line 42", "Should contain formatted string")
}

func TestFormatterWrite(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf)

	// Test io.Writer interface
	n, err := f.Write([]byte("{{blue:Testing}} write method"))
	require.NoError(t, err, "Write should not fail")
	require.NotZero(t, n, "Write should return non-zero bytes written")

	result := buf.String()
	require.Contains(t, result, "\033[34mTesting\033[0m", "Should contain blue Testing")
}

func TestFormatterPrint(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf)

	// Test Print method
	f.Print("{{yellow:Warning}}: ", "Something happened")

	result := buf.String()
	require.Contains(t, result, "\033[33mWarning\033[0m", "Should contain yellow Warning")
	require.Contains(t, result, "Something happened", "Should contain full message")
}

func TestFormatterPrintln(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf)

	// Test Println method
	f.Println("{{green:Success}}: Operation completed")

	result := buf.String()
	require.Contains(t, result, "\033[32mSuccess\033[0m", "Should contain green Success")
	require.True(t, strings.HasSuffix(result, "\n"), "Should end with newline")
}

func TestFormatterPrintf(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf)

	// Test Printf method
	f.Printf("{{red:Error}} %d: %s\n", 404, "Not Found")

	result := buf.String()
	require.Contains(t, result, "\033[31mError\033[0m", "Should contain red Error")
	require.Contains(t, result, "404: Not Found", "Should contain formatted error message")
}

func TestFormatterCustomColors(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf)

	// Add custom color
	f.AddColor("orange", BrightRed) // Using bright red as orange

	result := f.Format("{{orange:Custom color}}")

	require.Contains(t, result, "\033[91mCustom color\033[0m", "Should contain bright red (orange)")
}

func TestFormatterCustomStyles(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf)

	// Add custom style
	f.AddStyle("important", Bold)

	result := f.Format("{{important:Custom style}}")

	require.Contains(t, result, "\033[1mCustom style\033[0m", "Should contain bold (important)")
}

func TestFormatterInvalidModifiers(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf)

	// Test with invalid modifier
	result := f.Format("{{invalid:Text}} {{red:Valid}}")

	// Invalid modifier should be left as-is
	require.Contains(t, result, "{{invalid:Text}}", "Invalid modifier should remain unchanged")

	// Valid modifier should be processed
	require.Contains(t, result, "\033[31mValid\033[0m", "Valid modifier should be processed")
}

func TestFormatterEmptyTemplate(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf)

	result := f.Format("")
	require.Empty(t, result, "Empty template should return empty string")

	result = f.Format("No templates here")
	require.Equal(t, "No templates here", result, "Text without templates should remain unchanged")
}

func TestFormatterNestedTemplates(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf)

	// Test that nested templates don't break parsing
	result := f.Format("{{red:Error: message}} {{blue:normal}} text")

	// Should process both templates
	require.Contains(t, result, "\033[31m", "Should contain red color code")
	require.Contains(t, result, "\033[34m", "Should contain blue color code")
	require.Contains(t, result, "Error: message", "Should contain 'Error: message' text")
}

func TestFormatterMultipleStyles(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf)

	// Test multiple style combinations
	result := f.Format("{{red+bold+italic:Complex formatting}}")

	// Should contain all codes
	require.Contains(t, result, "\033[31m", "Should contain red color code")
	require.Contains(t, result, "\033[1m", "Should contain bold style code")
	require.Contains(t, result, "\033[3m", "Should contain italic style code")
}

func TestFormat(t *testing.T) {
	result := Format("{{green:Quick}} format test")

	require.Contains(t, result, "\033[32mQuick\033[0m", "Should contain green Quick")
	require.Contains(t, result, "format test", "Should contain full text")
}

func TestColorize(t *testing.T) {
	result := Colorize("{{red:Error}} %d: %s", 500, "Server Error")

	require.Contains(t, result, "\033[31mError\033[0m", "Should contain red Error")
	require.Contains(t, result, "500: Server Error", "Should contain formatted message")
}

func TestFormatterCaseSensitivity(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf)

	// Test case insensitive color names
	result := f.Format("{{RED:Upper}} {{Green:Mixed}} {{blue:Lower}}")

	require.Contains(t, result, "\033[31mUpper\033[0m", "Should contain red Upper")
	require.Contains(t, result, "\033[32mMixed\033[0m", "Should contain green Mixed")
	require.Contains(t, result, "\033[34mLower\033[0m", "Should contain blue Lower")
}

func TestFormatterIcons(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf)

	// Test basic icon formatting
	result := f.Format("{{check:}} Task completed")

	require.Contains(t, result, "✓ Task completed", "Should contain check icon with text")
}

func TestFormatterIconsWithText(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf)

	// Test icons with descriptive text
	result := f.Format("{{warning:Important}} {{error:Failed}} {{success:Done}}")

	require.Contains(t, result, "⚠ Important", "Should contain warning icon with text")
	require.Contains(t, result, "✗ Failed", "Should contain error icon with text")
	require.Contains(t, result, "✓ Done", "Should contain success icon with text")
}

func TestFormatterIconsWithoutText(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf)

	// Test icons without text
	result := f.Format("{{check:}} {{cross:}} {{info:}}")

	require.Contains(t, result, "✓", "Should contain check icon")
	require.Contains(t, result, "✗", "Should contain cross icon")
	require.Contains(t, result, "ℹ", "Should contain info icon")
}

func TestFormatterIconsWithColors(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf)

	// Test icons combined with colors
	result := f.Format("{{green+check:Success}} {{red+error:Failed}}")

	// Should contain both icon and color codes
	require.Contains(t, result, "✓", "Should contain check icon")
	require.Contains(t, result, "\033[32m", "Should contain green color code")
	require.Contains(t, result, "Success", "Should contain success text")
	require.Contains(t, result, "✗", "Should contain error icon")
	require.Contains(t, result, "\033[31m", "Should contain red color code")
	require.Contains(t, result, "Failed", "Should contain failed text")
}

func TestFormatterIconsWithStyles(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf)

	// Test icons combined with styles
	result := f.Format("{{bold+rocket:Launch}} {{italic+star:Special}}")

	// Should contain both icon and style codes
	require.Contains(t, result, "🚀", "Should contain rocket icon")
	require.Contains(t, result, "\033[1m", "Should contain bold style code")
	require.Contains(t, result, "Launch", "Should contain launch text")
	require.Contains(t, result, "★", "Should contain star icon")
	require.Contains(t, result, "\033[3m", "Should contain italic style code")
}

func TestFormatterCustomIcons(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf)

	// Add custom icon
	f.AddIcon("custom", "🎯")

	result := f.Format("{{custom:Target achieved}}")

	require.Contains(t, result, "🎯 Target achieved", "Should contain custom icon with text")
}

func TestFormatterSpinnerIcons(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf)

	// Test spinner icons
	result := f.Format("{{spinner1:}} {{spinner2:}} {{spinner3:}}")

	require.Contains(t, result, "⠋", "Should contain spinner1 icon")
	require.Contains(t, result, "⠙", "Should contain spinner2 icon")
	require.Contains(t, result, "⠹", "Should contain spinner3 icon")
}

func TestFormatterCheckboxIcons(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf)

	// Test checkbox icons
	result := f.Format("{{checkbox-empty:Todo}} {{checkbox-checked:Done}} {{checkbox-crossed:Failed}}")

	require.Contains(t, result, "☐ Todo", "Should contain empty checkbox with text")
	require.Contains(t, result, "☑ Done", "Should contain checked checkbox with text")
	require.Contains(t, result, "☒ Failed", "Should contain crossed checkbox with text")
}

func TestFormatterComplexIconCombinations(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf)

	// Test complex combinations with icons, colors, and styles
	result := f.Format("{{bold+green+check:Success}} {{red+italic+error:Critical Error}}")

	// Should contain all elements
	require.Contains(t, result, "✓", "Should contain check icon")
	require.Contains(t, result, "\033[1m", "Should contain bold style")
	require.Contains(t, result, "\033[32m", "Should contain green color")
	require.Contains(t, result, "Success", "Should contain success text")
	require.Contains(t, result, "✗", "Should contain error icon")
	require.Contains(t, result, "\033[31m", "Should contain red color")
	require.Contains(t, result, "\033[3m", "Should contain italic style")
}

func TestFormatterAllAvailableIcons(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf)

	// Test newly added icons
	result := f.Format("{{star-empty:}} {{success-icon:}} {{error-icon:}} {{bug:}} {{fix:}} {{unicorn:}}")

	require.Contains(t, result, "☆", "Should contain star-empty icon")
	require.Contains(t, result, "🎉", "Should contain success-icon")
	require.Contains(t, result, "💥", "Should contain error-icon")
	require.Contains(t, result, "🐛", "Should contain bug icon")
	require.Contains(t, result, "🔧", "Should contain fix icon")
	require.Contains(t, result, "🦄", "Should contain unicorn icon")
}
