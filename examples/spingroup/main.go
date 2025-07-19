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
	fmt.Println("Demonstrating dynamic task management with sequential spinners")
	fmt.Println()

	// Example 1: Dynamic task addition
	dynamicExample()
	fmt.Println()

	// Example 2: Real-world deployment simulation
	deploymentExample()
	fmt.Println()

	// Example 3: Nested frame example
	nestedFrameExample()
}

func dynamicExample() {
	fmt.Println(ansi.Cyan.Colorize("1. Dynamic Task Addition"))

	sg := spinner.NewSpinGroup("Dynamic Processing")

	// Add initial task
	sg.AddTask("Initializing", func() error {
		time.Sleep(randomDuration(300, 500))
		return nil
	})

	sg.Start()

	// Add tasks dynamically while running
	go func() {
		time.Sleep(200 * time.Millisecond)
		sg.AddTask("Loading module A", func() error {
			time.Sleep(randomDuration(400, 600))
			return nil
		})

		time.Sleep(300 * time.Millisecond)
		sg.AddTask("Loading module B", func() error {
			time.Sleep(randomDuration(500, 700))
			return nil
		})

		time.Sleep(200 * time.Millisecond)
		sg.AddTask("Starting services", func() error {
			time.Sleep(randomDuration(600, 800))
			return nil
		})

		time.Sleep(400 * time.Millisecond)
		sg.AddTask("Final validation", func() error {
			time.Sleep(randomDuration(300, 500))
			return nil
		})
	}()

	sg.Wait()
	// Small delay to ensure cleanup
	time.Sleep(50 * time.Millisecond)
}

func deploymentExample() {
	fmt.Println(ansi.Cyan.Colorize("2. Real-world Deployment Simulation"))

	f := frame.Open("Production Deployment", frame.WithColor(ansi.Green))

	sg := spinner.NewSpinGroup("Deployment Pipeline",
		spinner.WithSpinGroupOutput(f))

	// Pre-deployment tasks
	sg.AddTask("Backing up database", func() error {
		time.Sleep(randomDuration(1000, 1500))
		return nil
	})

	sg.AddTask("Stopping services", func() error {
		time.Sleep(randomDuration(500, 800))
		return nil
	})

	sg.Start()

	// Add deployment tasks dynamically
	go func() {
		time.Sleep(600 * time.Millisecond)

		sg.AddTask("Deploying application v2.1.0", func() error {
			time.Sleep(randomDuration(1200, 1800))
			return nil
		})

		sg.AddTask("Updating configuration", func() error {
			time.Sleep(randomDuration(400, 600))
			return nil
		})

		sg.AddTask("Running database migrations", func() error {
			time.Sleep(randomDuration(800, 1200))
			return nil
		})

		time.Sleep(800 * time.Millisecond)

		sg.AddTask("Starting services", func() error {
			time.Sleep(randomDuration(600, 900))
			return nil
		})

		sg.AddTask("Running health checks", func() error {
			time.Sleep(randomDuration(500, 700))
			return nil
		})

		sg.AddTask("Warming up caches", func() error {
			time.Sleep(randomDuration(400, 600))
			return nil
		})
	}()

	sg.Wait()
	// Small delay to ensure cleanup
	time.Sleep(50 * time.Millisecond)

	f.Println("")
	f.Println("🚀 Deployment completed successfully!")
	f.Println("📊 Summary:")
	fmt.Fprintf(f, "   • Total tasks: %d\n", sg.TaskCount())
	fmt.Fprintf(f, "   • Duration: %v\n", sg.Elapsed())
	f.Close()
}

func nestedFrameExample() {
	fmt.Println(ansi.Cyan.Colorize("3. Nested Frame Example"))

	// Outer frame for the entire application deployment
	appFrame := frame.Open("Application Deployment", frame.WithColor(ansi.Blue))
	appFrame.Println("  Starting complete application deployment...")

	// Database migration frame
	dbFrame := frame.Open("Database Migration", frame.WithColor(ansi.Yellow))
	dbSg := spinner.NewSpinGroup("Migration Tasks", spinner.WithSpinGroupOutput(dbFrame))

	dbSg.AddTask("Backing up current database", func() error {
		time.Sleep(randomDuration(800, 1200))
		return nil
	})

	dbSg.AddTask("Running migration scripts", func() error {
		time.Sleep(randomDuration(1000, 1500))
		return nil
	})

	dbSg.AddTask("Validating schema changes", func() error {
		time.Sleep(randomDuration(600, 900))
		return nil
	})

	dbSg.Start()
	dbSg.Wait()
	// Small delay to ensure cleanup
	time.Sleep(50 * time.Millisecond)

	dbFrame.Println("✅ Database migration completed successfully!")
	dbFrame.Close()

	// Application services frame
	serviceFrame := frame.Open("Service Deployment", frame.WithColor(ansi.Green))
	serviceSg := spinner.NewSpinGroup("Service Tasks", spinner.WithSpinGroupOutput(serviceFrame))

	serviceSg.AddTask("Building Docker images", func() error {
		time.Sleep(randomDuration(1200, 1800))
		return nil
	})

	serviceSg.AddTask("Deploying to staging", func() error {
		time.Sleep(randomDuration(800, 1200))
		return nil
	})

	serviceSg.AddTask("Running health checks", func() error {
		time.Sleep(randomDuration(400, 600))
		return nil
	})

	serviceSg.AddTask("Promoting to production", func() error {
		time.Sleep(randomDuration(600, 800))
		return nil
	})

	serviceSg.Start()
	serviceSg.Wait()
	// Small delay to ensure cleanup
	time.Sleep(50 * time.Millisecond)

	serviceFrame.Println("✅ Service deployment completed successfully!")
	serviceFrame.Close()

	// Final status in the outer frame
	appFrame.Println("")
	appFrame.Println("🚀 Complete application deployment finished!")
	appFrame.Println("📊 Summary:")
	appFrame.Println("   • Database migration: Success")
	appFrame.Println("   • Service deployment: Success")
	appFrame.Println("   • Total deployment time: ~15 seconds")
	appFrame.Close()
}

func randomDuration(minMs, maxMs int) time.Duration {
	// G404: Using weak random for demo timing purposes only
	//nolint:gosec
	duration := minMs + rand.Intn(maxMs-minMs)
	return time.Duration(duration) * time.Millisecond
}
