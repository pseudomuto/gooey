package ansi_test

import (
	"testing"

	. "github.com/pseudomuto/gooey/ansi"
	"github.com/stretchr/testify/require"
)

func TestColorString(t *testing.T) {
	tests := []struct {
		color    Color
		expected string
	}{
		{Reset, "\033[0m"},
		{Red, "\033[31m"},
		{Green, "\033[32m"},
		{Yellow, "\033[33m"},
		{Blue, "\033[34m"},
		{White, "\033[37m"},
		{BrightRed, "\033[91m"},
		{BrightGreen, "\033[92m"},
	}

	for _, tt := range tests {
		got := tt.color.String()
		require.Equal(t, tt.expected, got, "Color.String() should return correct ANSI code")
	}
}

func TestColorColorize(t *testing.T) {
	text := "Hello, World!"
	colored := Red.Colorize(text)
	expected := "\033[31mHello, World!\033[0m"

	require.Equal(t, expected, colored, "Colorize() should wrap text with color codes")
}

func TestColorSprint(t *testing.T) {
	result := Green.Sprint("Test", " ", "message")
	expected := "\033[32mTest message\033[0m"

	require.Equal(t, expected, result, "Sprint() should format and colorize text")
}

func TestColorSprintf(t *testing.T) {
	result := Blue.Sprintf("Hello %s!", "world")
	expected := "\033[34mHello world!\033[0m"

	require.Equal(t, expected, result, "Sprintf() should format and colorize text")
}

func TestControlSequences(t *testing.T) {
	tests := []struct {
		name     string
		sequence string
		expected string
	}{
		{"ClearScreen", ClearScreen, "\033[2J"},
		{"CursorHome", CursorHome, "\033[H"},
		{"ClearLine", ClearLine, "\033[K"},
		{"HideCursor", HideCursor, "\033[?25l"},
		{"ShowCursor", ShowCursor, "\033[?25h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.sequence, "Control sequence should match expected value")
		})
	}
}

func TestClearScreenAndHome(t *testing.T) {
	result := ClearScreenAndHome()
	expected := "\033[2J\033[H"

	require.Equal(t, expected, result, "ClearScreenAndHome() should return combined clear and home sequence")
}

func TestMoveCursor(t *testing.T) {
	result := MoveCursor(10, 20)
	expected := "\033[10;20H"

	require.Equal(t, expected, result, "MoveCursor() should return correct position sequence")
}

func TestMoveCursorDirectional(t *testing.T) {
	tests := []struct {
		name     string
		function func(int) string
		input    int
		expected string
	}{
		{"MoveCursorUp", MoveCursorUp, 5, "\033[5A"},
		{"MoveCursorDown", MoveCursorDown, 3, "\033[3B"},
		{"MoveCursorForward", MoveCursorForward, 7, "\033[7C"},
		{"MoveCursorBackward", MoveCursorBackward, 2, "\033[2D"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function(tt.input)
			require.Equal(t, tt.expected, result, "Directional cursor movement should return correct sequence")
		})
	}
}
