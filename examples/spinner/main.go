package main

import (
	"time"

	"github.com/pseudomuto/gooey/ansi"
	"github.com/pseudomuto/gooey/frame"
	"github.com/pseudomuto/gooey/spinner"
)

func main() {
	basicSpinnerDemo()
	customSpinnerDemo()
	frameIntegrationDemo()
	multipleSpinnersDemo()
	failureDemo()
}

func basicSpinnerDemo() {
	s := spinner.New("Loading data...")
	s.Start()
	time.Sleep(3 * time.Second)
	s.Stop()
}

func customSpinnerDemo() {
	s := spinner.New("Processing files...",
		spinner.WithColor(ansi.Green),
		spinner.WithRenderer(spinner.Clock),
		spinner.WithInterval(200*time.Millisecond),
		spinner.WithShowElapsed(false)) // Disable elapsed time display

	s.Start()
	time.Sleep(2 * time.Second)

	s.UpdateMessage("Almost done...")
	time.Sleep(1 * time.Second)

	s.Stop()
}

func frameIntegrationDemo() {
	frame := frame.Open("Task Execution", frame.WithColor(ansi.Blue))

	s := spinner.New("Initializing...",
		spinner.WithOutput(frame),
		spinner.WithColor(ansi.Yellow))

	s.Start()
	time.Sleep(1500 * time.Millisecond)

	s.UpdateMessage("Configuring settings...")
	time.Sleep(1500 * time.Millisecond)

	s.UpdateMessage("Finalizing...")
	time.Sleep(1 * time.Second)

	s.Stop()

	frame.Println("Task completed successfully!")
	frame.Close()
}

func multipleSpinnersDemo() {
	outer := frame.Open("Deployment Pipeline", frame.WithColor(ansi.Magenta))

	s1 := spinner.New("Building application...",
		spinner.WithOutput(outer),
		spinner.WithRenderer(spinner.Dots))
	s1.Start()
	time.Sleep(2 * time.Second)
	s1.Stop()

	inner := frame.Open("Database Migration", frame.WithColor(ansi.Green))

	s2 := spinner.New("Running migrations...",
		spinner.WithOutput(inner),
		spinner.WithRenderer(spinner.Arrow),
		spinner.WithColor(ansi.Green))
	s2.Start()
	time.Sleep(1500 * time.Millisecond)
	s2.Stop()

	inner.Close()

	s3 := spinner.New("Deploying to production...",
		spinner.WithOutput(outer),
		spinner.WithRenderer(spinner.Clock),
		spinner.WithColor(ansi.Red))
	s3.Start()
	time.Sleep(2 * time.Second)
	s3.Stop()

	outer.Println("Deployment completed!")
	outer.Close()
}

func failureDemo() {
	frame := frame.Open("Error Handling Demo", frame.WithColor(ansi.Red))

	// Successful task
	s1 := spinner.New("Validating configuration...", spinner.WithOutput(frame))
	s1.Start()
	time.Sleep(1 * time.Second)
	s1.Stop()

	// Failed task
	s2 := spinner.New("Connecting to database...",
		spinner.WithOutput(frame),
		spinner.WithColor(ansi.Yellow))
	s2.Start()
	time.Sleep(1500 * time.Millisecond)
	s2.UpdateMessage("Connection timeout occurred")
	time.Sleep(500 * time.Millisecond)
	s2.Fail("") // This will show a red crossmark instead of green checkmark

	// Another failed task with elapsed time disabled
	s3 := spinner.New("Deploying service...",
		spinner.WithOutput(frame),
		spinner.WithShowElapsed(false))
	s3.Start()
	time.Sleep(1 * time.Second)
	s3.UpdateMessage("Permission denied")
	time.Sleep(500 * time.Millisecond)
	s3.Fail("")

	frame.Println("Some tasks failed - check logs for details")
	frame.Close()
}
