package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/pseudomuto/gooey/ansi"
	"github.com/pseudomuto/gooey/frame"
	"github.com/pseudomuto/gooey/progress"
)

// fancyRenderer implements a more elaborate custom renderer
type fancyRenderer struct{}

func main() {
	demos := []func(){
		colorsAndWidths,
		builtinRenderers,
		incrementingProgress,
		nestedInFrame,
		customRenderFunc,
		customRenderer,
	}

	for _, fn := range demos {
		fmt.Println()
		fn()
	}

	fmt.Println("\n🎉 All examples completed!")
	fmt.Println("\nThe progress component supports:")
	fmt.Println("- Multiple visual styles (Bar, Minimal, Dots)")
	fmt.Println("- Customizable colors and widths")
	fmt.Println("- Real-time updates with messages")
	fmt.Println("- Increment and complete methods")
	fmt.Println("- Integration with other components")
	fmt.Println("- Custom renderers for unlimited styling possibilities")
}

func nestedInFrame() {
	f := frame.Open("Deployment Process", frame.WithColor(ansi.Cyan))
	f.Println("Simulating deployment pipeline...")

	deploymentSteps := []struct {
		name  string
		count int
		color ansi.Color
	}{
		{"Building application", 15, ansi.Blue},
		{"Running tests", 8, ansi.Green},
		{"Deploying to staging", 12, ansi.Yellow},
		{"Deploying to production", 20, ansi.Red},
	}

	for _, step := range deploymentSteps {
		p := progress.New(
			step.name,
			step.count,
			progress.WithRenderer(progress.NewChar("=", "*")),
			progress.WithColor(step.color),
			progress.WithOutput(f),
		)

		for i := 0; i <= step.count; i++ {
			p.Update(i, fmt.Sprintf("Step %d of %d", i, step.count))
			time.Sleep(100 * time.Millisecond)
		}
		p.Complete("✓ Completed")
		time.Sleep(100 * time.Millisecond)
	}

	f.Println("\n🎉 Deployment pipeline completed successfully!")
	f.Close()
}

func colorsAndWidths() {
	colors := []ansi.Color{ansi.Red, ansi.Green, ansi.Blue, ansi.Yellow}
	widths := []int{20, 30, 40, 50}

	for i, color := range colors {
		width := widths[i]
		p := progress.New(
			"Task",
			10,
			progress.WithColor(color),
			progress.WithWidth(width),
		)

		for j := 0; j <= 10; j += 2 {
			p.Update(j, fmt.Sprintf("Step %d", j))
			time.Sleep(150 * time.Millisecond)
		}
		p.Complete("Done!")
		time.Sleep(100 * time.Millisecond)
	}
}

func incrementingProgress() {
	p := progress.New("Building", 8, progress.WithColor(ansi.Magenta))
	tasks := []string{
		"Compiling main.go",
		"Compiling utils.go",
		"Compiling handlers.go",
		"Running tests",
		"Generating docs",
		"Optimizing binary",
		"Creating package",
		"Finalizing build",
	}

	for _, task := range tasks {
		time.Sleep(250 * time.Millisecond)
		p.Increment(task)
	}
	p.Complete("Build completed successfully!")
}

func builtinRenderers() {
	fmt.Println("Built-in renderers:")
	p := progress.New("Bar", 20, progress.WithColor(ansi.Green))
	for i := 0; i <= 20; i += 4 {
		p.Update(i, fmt.Sprintf("Step %d", i))
		time.Sleep(200 * time.Millisecond)
	}
	p.Complete("Processing complete!")

	time.Sleep(100 * time.Millisecond)

	// Dots style
	p = progress.New("Dots", 12, progress.WithRenderer(progress.Dots), progress.WithColor(ansi.Yellow), progress.WithWidth(24))
	for i := 0; i <= 12; i += 2 {
		p.Update(i, fmt.Sprintf("Package %d installed", i))
		time.Sleep(200 * time.Millisecond)
	}
	p.Complete("All packages installed!")

	time.Sleep(100 * time.Millisecond)

	// Minimal style
	p = progress.New("Minimal", 15, progress.WithRenderer(progress.Minimal), progress.WithColor(ansi.Blue))
	for i := 0; i <= 15; i += 3 {
		p.Update(i, fmt.Sprintf("Uploaded %d items", i))
		time.Sleep(200 * time.Millisecond)
	}
	p.Complete("Upload finished!")
}

func customRenderFunc() {
	p := progress.New(
		"Processing",
		10,
		progress.WithRenderer(progress.RenderFunc(func(p *progress.Progress, w io.Writer) {
			percentage := p.Percentage()
			fmt.Fprintf(w, "%s: ", p.Title())

			// Use different symbols based on percentage
			if percentage < 25 {
				fmt.Fprint(w, "🔴")
			} else if percentage < 50 {
				fmt.Fprint(w, "🟡")
			} else if percentage < 75 {
				fmt.Fprint(w, "🟠")
			} else if percentage < 100 {
				fmt.Fprint(w, "🟢")
			} else {
				fmt.Fprint(w, "✅")
			}

			fmt.Fprintf(w, " %.1f%% (%d/%d)", percentage, p.Current(), p.Total())
			if p.Message() != "" {
				fmt.Fprintf(w, " - %s", p.Message())
			}
		})),
	)

	for i := 0; i <= 10; i += 2 {
		p.Update(i, fmt.Sprintf("Step %d", i))
		time.Sleep(300 * time.Millisecond)
	}
	p.Complete("All done!")
}

func customRenderer() {
	f := frame.Open("Custom Progress Demo", frame.WithColor(ansi.Magenta))
	f.Println("Testing fancy custom renderer...")

	p := progress.New("Deployment", 8,
		progress.WithRenderer(new(fancyRenderer)),
		progress.WithColor(ansi.Green),
		progress.WithOutput(f),
	)

	tasks := []string{"Build", "Test", "Package", "Deploy", "Verify", "Cleanup", "Notify", "Complete"}
	for i, task := range tasks {
		p.Update(i+1, task)
		time.Sleep(250 * time.Millisecond)
	}

	p.Complete("Deployment successful!")
	f.ReplaceLine("{{success+green:Deployment successful!}}")
	f.Close()
}

func (r *fancyRenderer) Render(p *progress.Progress, w io.Writer) {
	var sb strings.Builder
	sb.WriteString(p.Title())
	sb.WriteRune(' ')

	// Custom bracket style with colors
	sb.WriteString(ansi.BoldText("["))
	sb.WriteString(p.Color().Sprintf("%.1f%%", p.Percentage()))
	sb.WriteString(ansi.BoldText("]"))

	// Add elapsed time
	sb.WriteString(fmt.Sprintf(" {{clock:}}  %v", p.Elapsed().Round(time.Millisecond)))

	if p.Message() != "" {
		sb.WriteString(" | ")
		sb.WriteString(p.Message())
	}

	fmt.Fprint(w, sb.String())
}
