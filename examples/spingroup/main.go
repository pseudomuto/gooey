package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/pseudomuto/gooey/ansi"
	"github.com/pseudomuto/gooey/frame"
	"github.com/pseudomuto/gooey/spinner"
)

func main() {
	fmt.Println(ansi.Bold.Apply("SpinGroup Examples"))
	fmt.Println("Demonstrating sequential task execution with real Spinner instances")
	fmt.Println()

	// Example 1: Basic usage
	basicExample()
	fmt.Println()

	// Example 2: Custom spinner configurations
	customSpinnersExample()
	fmt.Println()

	// Example 3: Frame integration
	frameExample()
	fmt.Println()

	// Example 4: Nested frames
	nestedFrameExample()
}

func basicExample() {
	fmt.Println(ansi.Cyan.Colorize("1. Basic SpinGroup Usage"))

	sg := spinner.NewSpinGroup("Basic Tasks")

	// Add tasks with default spinners
	sg.AddTask("Initializing", spinner.New("Starting up..."), func() error {
		time.Sleep(randomDuration(800, 1200))
		return nil
	})

	sg.AddTask("Processing", spinner.New("Processing data..."), func() error {
		time.Sleep(randomDuration(1000, 1500))
		return nil
	})

	sg.AddTask("Finalizing", spinner.New("Cleaning up..."), func() error {
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
		func() error {
			time.Sleep(randomDuration(1200, 1800))
			return nil
		})

	sg.AddTask("Testing",
		spinner.New("Running tests...",
			spinner.WithColor(ansi.Yellow),
			spinner.WithRenderer(spinner.Clock),
			spinner.WithShowElapsed(true)),
		func() error {
			time.Sleep(randomDuration(800, 1200))
			return nil
		})

	sg.AddTask("Deploying",
		spinner.New("Deploying to production...",
			spinner.WithColor(ansi.Green),
			spinner.WithRenderer(spinner.Arrow),
			spinner.WithInterval(200*time.Millisecond)),
		func() error {
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
		func() error {
			time.Sleep(randomDuration(1000, 1500))
			return nil
		})

	sg.AddTask("Service Update",
		spinner.New("Updating services...",
			spinner.WithColor(ansi.BrightGreen)),
		func() error {
			time.Sleep(randomDuration(800, 1200))
			return nil
		})

	sg.AddTask("Health Check",
		spinner.New("Verifying system health...",
			spinner.WithColor(ansi.BrightYellow)),
		func() error {
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
		func() error {
			time.Sleep(randomDuration(800, 1200))
			return nil
		})

	dbGroup.AddTask("Migration",
		spinner.New("Running schema migrations...", spinner.WithColor(ansi.BrightBlue)),
		func() error {
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
		func() error {
			time.Sleep(randomDuration(1200, 1800))
			return nil
		})

	serviceGroup.AddTask("Deploy",
		spinner.New("Deploying to cluster...",
			spinner.WithColor(ansi.BrightCyan),
			spinner.WithRenderer(spinner.Clock)),
		func() error {
			time.Sleep(randomDuration(900, 1300))
			return nil
		})

	serviceGroup.AddTask("Verify",
		spinner.New("Running health checks...",
			spinner.WithColor(ansi.BrightMagenta),
			spinner.WithShowElapsed(true)),
		func() error {
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

func randomDuration(minMs, maxMs int) time.Duration {
	// G404: Using weak random for demo timing purposes only
	//nolint:gosec
	duration := minMs + rand.Intn(maxMs-minMs)
	return time.Duration(duration) * time.Millisecond
}
