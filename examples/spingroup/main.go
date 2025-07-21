package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/pkg/errors"
	"github.com/pseudomuto/gooey/ansi"
	"github.com/pseudomuto/gooey/frame"
	"github.com/pseudomuto/gooey/progress"
	"github.com/pseudomuto/gooey/spinner"
)

func main() {
	fmt.Println(ansi.Bold.Apply("SpinGroup Examples"))
	fmt.Println("Demonstrating sequential task execution with TaskComponent interface")
	fmt.Println()

	// Example 1: Basic usage
	basicExample()
	fmt.Println()

	// Example 2: Custom spinner configurations
	customSpinnersExample()
	fmt.Println()

	// Example 3: Mixed Spinner and Progress components
	mixedComponentsExample()
	fmt.Println()

	// Example 4: Frame integration
	frameExample()
	fmt.Println()

	// Example 5: Real-world deployment with mixed components
	realWorldExample()
	fmt.Println()

	// Example 6: Progress failure demonstration
	progressFailureExample()
	fmt.Println()

	// Example 7: Nested frames
	nestedFrameExample()
}

func basicExample() {
	fmt.Println(ansi.Cyan.Colorize("1. Basic SpinGroup Usage"))

	sg := spinner.NewSpinGroup("Basic Tasks")

	// Add tasks with default spinners
	sg.AddTask("Initializing", spinner.New("Starting up..."), func(component spinner.TaskComponent) error {
		time.Sleep(randomDuration(800, 1200))
		return nil
	})

	sg.AddTask("Processing", spinner.New("Processing data..."), func(component spinner.TaskComponent) error {
		time.Sleep(randomDuration(1000, 1500))
		return nil
	})

	sg.AddTask("Finalizing", spinner.New("Cleaning up..."), func(component spinner.TaskComponent) error {
		time.Sleep(randomDuration(600, 900))
		return nil
	})

	// Run all tasks sequentially
	err := sg.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func customSpinnersExample() {
	fmt.Println(ansi.Cyan.Colorize("2. Custom Spinner Configurations"))

	sg := spinner.NewSpinGroup("Custom Tasks")

	// Each task can have its own spinner configuration
	sg.AddTask("Building",
		spinner.New("Building application...",
			spinner.WithColor(ansi.Blue),
			spinner.WithRenderer(spinner.Dots)),
		func(component spinner.TaskComponent) error {
			time.Sleep(randomDuration(1200, 1800))
			return nil
		})

	sg.AddTask("Testing",
		spinner.New("Running tests...",
			spinner.WithColor(ansi.Yellow),
			spinner.WithRenderer(spinner.Clock),
			spinner.WithShowElapsed(true)),
		func(component spinner.TaskComponent) error {
			time.Sleep(randomDuration(800, 1200))
			return nil
		})

	sg.AddTask("Deploying",
		spinner.New("Deploying to production...",
			spinner.WithColor(ansi.Green),
			spinner.WithRenderer(spinner.Arrow),
			spinner.WithInterval(200*time.Millisecond)),
		func(component spinner.TaskComponent) error {
			time.Sleep(randomDuration(1000, 1400))
			return nil
		})

	err := sg.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func frameExample() {
	fmt.Println(ansi.Cyan.Colorize("3. Frame Integration"))

	sg := spinner.NewSpinGroup("Deployment Pipeline")

	// Add tasks
	sg.AddTask("Database Migration",
		spinner.New("Migrating database schema...",
			spinner.WithColor(ansi.BrightBlue)),
		func(component spinner.TaskComponent) error {
			time.Sleep(randomDuration(1000, 1500))
			return nil
		})

	sg.AddTask("Service Update",
		spinner.New("Updating services...",
			spinner.WithColor(ansi.BrightGreen)),
		func(component spinner.TaskComponent) error {
			time.Sleep(randomDuration(800, 1200))
			return nil
		})

	sg.AddTask("Health Check",
		spinner.New("Verifying system health...",
			spinner.WithColor(ansi.BrightYellow)),
		func(component spinner.TaskComponent) error {
			time.Sleep(randomDuration(600, 900))
			return nil
		})

	// Run within a frame for organized display
	err := sg.RunInFrame()
	if err != nil {
		fmt.Printf("Deployment failed: %v\n", err)
	} else {
		fmt.Println("✅ Deployment completed successfully!")
	}
}

func nestedFrameExample() {
	fmt.Println(ansi.Cyan.Colorize("4. Nested Frame Example"))

	// Outer frame for the entire application deployment
	appFrame := frame.Open("Complete Application Deployment", frame.WithColor(ansi.Blue))
	appFrame.Println("Starting comprehensive deployment process...")

	// Database operations nested frame
	dbFrame := frame.Open("Database Operations", frame.WithColor(ansi.Yellow))

	dbGroup := spinner.NewSpinGroup("Database Tasks", spinner.WithSpinGroupOutput(dbFrame))
	dbGroup.AddTask("Backup",
		spinner.New("Creating database backup...", spinner.WithColor(ansi.BrightYellow)),
		func(component spinner.TaskComponent) error {
			time.Sleep(randomDuration(800, 1200))
			return nil
		})

	dbGroup.AddTask("Migration",
		spinner.New("Running schema migrations...", spinner.WithColor(ansi.BrightBlue)),
		func(component spinner.TaskComponent) error {
			time.Sleep(randomDuration(1000, 1500))
			return nil
		})

	err := dbGroup.Run()
	if err != nil {
		dbFrame.Println("❌ Database operations failed: %v", err)
		dbFrame.Close()
		appFrame.Close()
		return
	}

	dbFrame.Println("✅ Database operations completed successfully!")
	dbFrame.Close()

	// Service deployment nested frame
	serviceFrame := frame.Open("Service Deployment", frame.WithColor(ansi.Green))

	serviceGroup := spinner.NewSpinGroup("Service Tasks", spinner.WithSpinGroupOutput(serviceFrame))
	serviceGroup.AddTask("Build",
		spinner.New("Building Docker images...",
			spinner.WithColor(ansi.BrightGreen),
			spinner.WithRenderer(spinner.Dots)),
		func(component spinner.TaskComponent) error {
			time.Sleep(randomDuration(1200, 1800))
			return nil
		})

	serviceGroup.AddTask("Deploy",
		spinner.New("Deploying to cluster...",
			spinner.WithColor(ansi.BrightCyan),
			spinner.WithRenderer(spinner.Clock)),
		func(component spinner.TaskComponent) error {
			time.Sleep(randomDuration(900, 1300))
			return nil
		})

	serviceGroup.AddTask("Verify",
		spinner.New("Running health checks...",
			spinner.WithColor(ansi.BrightMagenta),
			spinner.WithShowElapsed(true)),
		func(component spinner.TaskComponent) error {
			time.Sleep(randomDuration(600, 900))
			return nil
		})

	err = serviceGroup.Run()
	if err != nil {
		serviceFrame.Println("❌ Service deployment failed: %v", err)
		serviceFrame.Close()
		appFrame.Close()
		return
	}

	serviceFrame.Println("✅ Service deployment completed successfully!")
	serviceFrame.Close()

	// Final status in the outer frame
	appFrame.Println("")
	appFrame.Divider("Deployment Summary")
	appFrame.Println("🚀 Complete application deployment finished!")
	appFrame.Println("📊 Results:")
	appFrame.Println("   • Database Operations: Success ✅")
	appFrame.Println("   • Service Deployment: Success ✅")
	appFrame.Println("   • Total Tasks: %d", dbGroup.TaskCount()+serviceGroup.TaskCount())
	appFrame.Close()
}

func mixedComponentsExample() {
	fmt.Println(ansi.Cyan.Colorize("3. Mixed Spinner and Progress Components"))

	sg := spinner.NewSpinGroup("Mixed Component Tasks")

	// Indefinite task with spinner (connection has no known duration)
	sg.AddTask("Connect",
		spinner.New("Connecting to server...", spinner.WithColor(ansi.Yellow)),
		func(component spinner.TaskComponent) error {
			// Demonstrate updating spinner messages dynamically
			if s, ok := component.(*spinner.Spinner); ok {
				time.Sleep(randomDuration(300, 500))
				s.UpdateMessage("Authenticating...")
				time.Sleep(randomDuration(300, 500))
				s.UpdateMessage("Connection established")
				time.Sleep(randomDuration(200, 400))
			}
			return nil
		})

	// Definite task with progress bar (file download has known size)
	downloadProgress := progress.New("Download", 100,
		progress.WithColor(ansi.Green),
		progress.WithRenderer(progress.Bar))
	sg.AddTask("Download", downloadProgress, func(component spinner.TaskComponent) error {
		// Demonstrate using the component parameter for progress updates
		if p, ok := component.(*progress.Progress); ok {
			for i := 0; i <= 100; i += 10 {
				p.Update(i, fmt.Sprintf("Downloaded %d%%", i))
				time.Sleep(50 * time.Millisecond)
			}
		}
		return nil
	})

	// Another indefinite task with spinner (processing has unknown duration)
	sg.AddTask("Process",
		spinner.New("Processing files...",
			spinner.WithColor(ansi.Blue),
			spinner.WithRenderer(spinner.Dots)),
		func(component spinner.TaskComponent) error {
			time.Sleep(randomDuration(1000, 1500))
			return nil
		})

	// Definite task with progress bar using dots renderer
	uploadProgress := progress.New("Upload", 50,
		progress.WithColor(ansi.Magenta),
		progress.WithRenderer(progress.Dots))
	sg.AddTask("Upload", uploadProgress, func(component spinner.TaskComponent) error {
		for i := 0; i <= 50; i += 5 {
			uploadProgress.Update(i, fmt.Sprintf("Uploading... %d files", i))
			time.Sleep(80 * time.Millisecond)
		}
		return nil
	})

	// Final indefinite task with spinner
	sg.AddTask("Cleanup",
		spinner.New("Cleaning up temporary files...", spinner.WithColor(ansi.Cyan)),
		func(component spinner.TaskComponent) error {
			time.Sleep(randomDuration(600, 800))
			return nil
		})

	err := sg.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func realWorldExample() {
	fmt.Println(ansi.Cyan.Colorize("5. Real-World Deployment Example"))

	sg := spinner.NewSpinGroup("Application Deployment")

	// Pre-deployment checks (indefinite - we don't know how long validation takes)
	sg.AddTask("Validate",
		spinner.New("Validating configuration...",
			spinner.WithColor(ansi.Yellow),
			spinner.WithRenderer(spinner.Clock)),
		func(component spinner.TaskComponent) error {
			time.Sleep(randomDuration(800, 1200))
			return nil
		})

	// Build process (definite - we know the build steps)
	buildProgress := progress.New("Build", 5,
		progress.WithColor(ansi.Blue),
		progress.WithRenderer(progress.Bar))
	sg.AddTask("Build", buildProgress, func(component spinner.TaskComponent) error {
		steps := []string{"Installing deps", "Compiling", "Running tests", "Creating artifacts", "Packaging"}
		for i, step := range steps {
			buildProgress.Update(i+1, step)
			time.Sleep(300 * time.Millisecond)
		}
		return nil
	})

	// Database migration (definite - we know the number of migrations)
	migrationProgress := progress.New("Migrate", 12,
		progress.WithColor(ansi.Green),
		progress.WithRenderer(progress.Minimal))
	sg.AddTask("Migrate", migrationProgress, func(component spinner.TaskComponent) error {
		for i := 0; i <= 12; i++ {
			migrationProgress.Update(i, fmt.Sprintf("Applied migration %d", i))
			time.Sleep(100 * time.Millisecond)
		}
		return nil
	})

	// Deployment (indefinite - network operations have unpredictable timing)
	sg.AddTask("Deploy",
		spinner.New("Deploying to production...",
			spinner.WithColor(ansi.Red),
			spinner.WithRenderer(spinner.Arrow)),
		func(component spinner.TaskComponent) error {
			time.Sleep(randomDuration(1500, 2000))
			return nil
		})

	// Health check (indefinite - service startup time varies)
	sg.AddTask("Health Check",
		spinner.New("Waiting for service to be healthy...",
			spinner.WithColor(ansi.Green),
			spinner.WithShowElapsed(true)),
		func(component spinner.TaskComponent) error {
			time.Sleep(randomDuration(1000, 1500))
			return nil
		})

	err := sg.RunInFrame()
	if err != nil {
		fmt.Printf("Deployment failed: %v\n", err)
	} else {
		fmt.Println("🚀 Deployment completed successfully!")
	}
}

func randomDuration(minMs, maxMs int) time.Duration {
	// G404: Using weak random for demo timing purposes only
	//nolint:gosec
	duration := minMs + rand.Intn(maxMs-minMs)
	return time.Duration(duration) * time.Millisecond
}

func progressFailureExample() {
	fmt.Println(ansi.Cyan.Colorize("6. Progress Success and Failure Demonstration"))

	sg := spinner.NewSpinGroup("Progress Failure Demo")

	// Success case: Progress that completes successfully
	successProgress := progress.New("Successful Task", 10,
		progress.WithColor(ansi.Green),
		progress.WithRenderer(progress.Bar))
	sg.AddTask("Success", successProgress, func(component spinner.TaskComponent) error {
		for i := 0; i <= 10; i++ {
			successProgress.Update(i, fmt.Sprintf("Processing step %d/10", i))
			time.Sleep(100 * time.Millisecond)
		}
		return nil // Success
	})

	// Failure case: Progress that fails partway through
	failureProgress := progress.New("Task That Fails", 20,
		progress.WithColor(ansi.Yellow),
		progress.WithRenderer(progress.Bar))
	sg.AddTask("Failure", failureProgress, func(component spinner.TaskComponent) error {
		for i := 0; i <= 12; i++ {
			failureProgress.Update(i, fmt.Sprintf("Processing item %d/20", i))
			time.Sleep(80 * time.Millisecond)
			// Simulate failure at 60% completion
			if i == 12 {
				return errors.New("network connection lost")
			}
		}
		return nil
	})

	// This task won't run because the previous one failed
	neverRunProgress := progress.New("Never Executed", 5,
		progress.WithColor(ansi.Blue),
		progress.WithRenderer(progress.Dots))
	sg.AddTask("Skipped", neverRunProgress, func(component spinner.TaskComponent) error {
		for i := 0; i <= 5; i++ {
			neverRunProgress.Update(i, fmt.Sprintf("Step %d", i))
			time.Sleep(100 * time.Millisecond)
		}
		return nil
	})

	err := sg.RunInFrame()
	if err != nil {
		fmt.Printf("❌ SpinGroup failed as expected: %v\n", err)
		fmt.Println("Notice how the progress bar shows a red ✗ when it fails!")
	} else {
		fmt.Println("✅ All tasks completed successfully!")
	}
}
