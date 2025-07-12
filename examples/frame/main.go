package main

import (
	"os"
	"time"

	"github.com/pseudomuto/gooey/ansi"
	"github.com/pseudomuto/gooey/components/frame"
)

func main() {
	// Basic frame example
	f := frame.Open("Basic Frame Example")
	f.Println("This is content inside the frame")
	f.Println("Multiple lines of content")
	f.Close()

	// Nested frames example
	outer := frame.Open("Outer Frame", frame.WithColor(ansi.Blue))
	outer.Println("Content in outer frame")

	inner := frame.Open("Inner Frame", frame.WithColor(ansi.Green))
	inner.Println("Content in inner frame")
	inner.Close()

	outer.Println("Back to outer frame content")
	outer.Close()

	// Different styles example
	// Box style (default)
	f = frame.Open("Box Style Frame", frame.WithStyle(frame.Box))
	f.Println("This uses box style borders")
	f.Close()

	// Bracket style
	f = frame.Open("Bracket Style Frame", frame.WithStyle(frame.Bracket))
	f.Println("This uses bracket style borders")
	f.Close()

	// Dividers example
	f = frame.Open("Frame with Dividers")
	f.Println("Content before divider")
	f.Divider("Section Break")
	f.Println("Content after divider")
	f.Divider("")
	f.Print("Content after empty divider")
	f.Println(" (using Print + Println)")
	f.Close()

	// Timing example (with artificial delay)
	fmtr := ansi.NewFormatter(os.Stdout)
	f = frame.Open("{{unicorn:}} Timed Operation", frame.WithColor(ansi.Yellow), frame.WithOutput(fmtr))
	f.Println("Simulating some {{bold+cyan:work}}")
	time.Sleep(100 * time.Millisecond)
	f.Println("Work {{green:completed}}!")
	f.Println("Progress: %d%% complete", 100)
	f.Close()

	colors := []ansi.Color{
		ansi.Red,
		ansi.BrightRed,
		ansi.Yellow,
		ansi.Green,
		ansi.Blue,
		ansi.BrightBlue,
	}
	frames := make([]*frame.Frame, len(colors))
	for i, c := range colors {
		frames[i] = frame.Open(c.Sprint("Color"), frame.WithColor(c))
	}

	frames[len(frames)-1].Println("Colors of the rainbow...")
	for i := len(frames) - 1; i >= 0; i-- {
		frames[i].Close()
	}
}
