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

	// Example 6: Dynamic subtasks demonstration
	dynamicSubtasksExample()
	fmt.Println()

	// Example 7: Progress failure demonstration
	progressFailureExample()
	fmt.Println()

	// Example 8: Nested frames
	nestedFrameExample()
}

func basicExample() {
	fmt.Println(ansi.Cyan.Colorize("1. Basic SpinGroup Usage"))

	sg := spinner.NewSpinGroup("Basic Tasks")

	// Add tasks with default spinners
	sg.AddTask("Initializing", spinner.New("Starting up..."), func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
		time.Sleep(randomDuration(800, 1200))
		return nil
	})

	sg.AddTask("Processing", spinner.New("Processing data..."), func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
		time.Sleep(randomDuration(1000, 1500))
		return nil
	})

	sg.AddTask("Finalizing", spinner.New("Cleaning up..."), func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
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
		func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
			time.Sleep(randomDuration(1200, 1800))
			return nil
		})

	sg.AddTask("Testing",
		spinner.New("Running tests...",
			spinner.WithColor(ansi.Yellow),
			spinner.WithRenderer(spinner.Clock),
			spinner.WithShowElapsed(true)),
		func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
			time.Sleep(randomDuration(800, 1200))
			return nil
		})

	sg.AddTask("Deploying",
		spinner.New("Deploying to production...",
			spinner.WithColor(ansi.Green),
			spinner.WithRenderer(spinner.Arrow),
			spinner.WithInterval(200*time.Millisecond)),
		func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
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
		func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
			time.Sleep(randomDuration(1000, 1500))
			return nil
		})

	sg.AddTask("Service Update",
		spinner.New("Updating services...",
			spinner.WithColor(ansi.BrightGreen)),
		func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
			time.Sleep(randomDuration(800, 1200))
			return nil
		})

	sg.AddTask("Health Check",
		spinner.New("Verifying system health...",
			spinner.WithColor(ansi.BrightYellow)),
		func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
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
	fmt.Println(ansi.Cyan.Colorize("8. Nested Frame Example"))

	// Outer frame for the entire application deployment
	appFrame := frame.Open("Complete Application Deployment", frame.WithColor(ansi.Blue))
	appFrame.Println("Starting comprehensive deployment process...")

	// Database operations nested frame
	dbFrame := frame.Open("Database Operations", frame.WithColor(ansi.Yellow))

	dbGroup := spinner.NewSpinGroup("Database Tasks", spinner.WithSpinGroupOutput(dbFrame))
	dbGroup.AddTask("Backup",
		spinner.New("Creating database backup...", spinner.WithColor(ansi.BrightYellow)),
		func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
			time.Sleep(randomDuration(800, 1200))
			return nil
		})

	dbGroup.AddTask("Migration",
		spinner.New("Running schema migrations...", spinner.WithColor(ansi.BrightBlue)),
		func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
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
		func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
			time.Sleep(randomDuration(1200, 1800))
			return nil
		})

	serviceGroup.AddTask("Deploy",
		spinner.New("Deploying to cluster...",
			spinner.WithColor(ansi.BrightCyan),
			spinner.WithRenderer(spinner.Clock)),
		func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
			time.Sleep(randomDuration(900, 1300))
			return nil
		})

	serviceGroup.AddTask("Verify",
		spinner.New("Running health checks...",
			spinner.WithColor(ansi.BrightMagenta),
			spinner.WithShowElapsed(true)),
		func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
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
		func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
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
	sg.AddTask("Download", downloadProgress, func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
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
		func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
			time.Sleep(randomDuration(1000, 1500))
			return nil
		})

	// Definite task with progress bar using dots renderer
	uploadProgress := progress.New("Upload", 50,
		progress.WithColor(ansi.Magenta),
		progress.WithRenderer(progress.Dots))
	sg.AddTask("Upload", uploadProgress, func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
		for i := 0; i <= 50; i += 5 {
			uploadProgress.Update(i, fmt.Sprintf("Uploading... %d files", i))
			time.Sleep(80 * time.Millisecond)
		}
		return nil
	})

	// Final indefinite task with spinner
	sg.AddTask("Cleanup",
		spinner.New("Cleaning up temporary files...", spinner.WithColor(ansi.Cyan)),
		func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
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
		func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
			time.Sleep(randomDuration(800, 1200))
			return nil
		})

	// Build process (definite - but total discovered during execution)
	buildProgress := progress.New("Build", 0, // Unknown total initially
		progress.WithColor(ansi.Blue),
		progress.WithRenderer(progress.Bar))
	sg.AddTask("Build", buildProgress, func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
		if p, ok := component.(*progress.Progress); ok {
			// Simulate discovering build steps dynamically (e.g., from build manifest)
			time.Sleep(100 * time.Millisecond)
			steps := []string{"Installing deps", "Compiling", "Running tests", "Creating artifacts", "Packaging"}
			p.SetTotal(len(steps)) // Set total after discovering build steps

			for i, step := range steps {
				p.Update(i+1, step)
				time.Sleep(300 * time.Millisecond)
			}
		}
		return nil
	})

	// Database migration (definite - we know the number of migrations)
	migrationProgress := progress.New("Migrate", 12,
		progress.WithColor(ansi.Green),
		progress.WithRenderer(progress.Minimal))
	sg.AddTask("Migrate", migrationProgress, func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
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
		func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
			time.Sleep(randomDuration(1500, 2000))
			return nil
		})

	// Health check (indefinite - service startup time varies)
	sg.AddTask("Health Check",
		spinner.New("Waiting for service to be healthy...",
			spinner.WithColor(ansi.Green),
			spinner.WithShowElapsed(true)),
		func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
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

func dynamicSubtasksExample() {
	fmt.Println(ansi.Cyan.Colorize("6. Dynamic Subtasks Demonstration"))

	sg := spinner.NewSpinGroup("Dynamic Deployment")

	// Main deployment task that discovers services and adds subtasks dynamically
	sg.AddTask("Discover Services", spinner.New("Scanning deployment manifest..."),
		func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
			if s, ok := component.(*spinner.Spinner); ok {
				time.Sleep(500 * time.Millisecond)
				s.UpdateMessage("Found microservices to deploy...")
				time.Sleep(200 * time.Millisecond)

				// Simulate discovering services from a manifest
				services := []string{"auth-service", "api-gateway", "user-service", "notification-service"}

				// Dynamically add subtasks for each discovered service
				for _, service := range services {
					sg.AddSubtask("Deploy "+service,
						spinner.New("Deploying "+service+"...",
							spinner.WithColor(ansi.Green),
							spinner.WithRenderer(spinner.Dots)),
						func(c spinner.TaskComponent, _ *spinner.SpinGroup) error {
							serviceName := service // Capture in closure
							if spinner, ok := c.(*spinner.Spinner); ok {
								spinner.UpdateMessage("Starting " + serviceName + " deployment...")
								time.Sleep(randomDuration(300, 600))
								spinner.UpdateMessage("Configuring " + serviceName + "...")
								time.Sleep(randomDuration(200, 400))
								spinner.UpdateMessage(serviceName + " deployed successfully")
								time.Sleep(randomDuration(100, 200))
							}
							return nil
						})
				}

				s.UpdateMessage(fmt.Sprintf("Discovery complete - %d services found", len(services)))
			}
			return nil
		})

	// This task runs after all the dynamically added subtasks complete
	sg.AddTask("Health Check", spinner.New("Running post-deployment health checks..."),
		func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
			time.Sleep(randomDuration(800, 1200))
			return nil
		})

	err := sg.RunInFrame()
	if err != nil {
		fmt.Printf("❌ Dynamic deployment failed: %v\n", err)
	} else {
		fmt.Println("🚀 Dynamic deployment completed successfully!")
		fmt.Printf("📊 Total tasks executed: %d\n", sg.TaskCount())
	}
}

func progressFailureExample() {
	fmt.Println(ansi.Cyan.Colorize("7. Mixed Components with Indented Subtasks"))

	sg := spinner.NewSpinGroup("Advanced Mixed Components")

	// Main deployment task that creates both spinner and progress subtasks
	sg.AddTask("Deploy Application", spinner.New("Preparing application deployment..."),
		func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
			if s, ok := component.(*spinner.Spinner); ok {
				s.UpdateMessage("Initializing deployment environment...")
				time.Sleep(300 * time.Millisecond)

				// Add a progress subtask for file copying (definite task)
				sg.AddSubtask("Copy Files",
					progress.New("File Copy", 25,
						progress.WithColor(ansi.Blue),
						progress.WithRenderer(progress.Bar)),
					func(c spinner.TaskComponent, _ *spinner.SpinGroup) error {
						if p, ok := c.(*progress.Progress); ok {
							files := []string{"config.yaml", "app.jar", "static/", "templates/", "lib/"}
							for i, file := range files {
								for step := i * 5; step <= (i+1)*5; step++ {
									p.Update(step, fmt.Sprintf("Copying %s (%d/%d files)", file, i+1, len(files)))
									time.Sleep(60 * time.Millisecond)
								}
							}
						}
						return nil
					})

				// Add a spinner subtask for service configuration (indefinite task)
				sg.AddSubtask("Configure Services", spinner.New("Configuring application services..."),
					func(c spinner.TaskComponent, sg *spinner.SpinGroup) error {
						if s, ok := c.(*spinner.Spinner); ok {
							s.UpdateMessage("Configuring database connection...")
							time.Sleep(200 * time.Millisecond)
							s.UpdateMessage("Setting up Redis cache...")
							time.Sleep(200 * time.Millisecond)
							s.UpdateMessage("Configuring message queues...")
							time.Sleep(200 * time.Millisecond)

							// Add a nested progress subtask for cache warming (definite sub-task)
							sg.AddSubtask("Warm Cache",
								progress.New("Cache Warmup", 15,
									progress.WithColor(ansi.Magenta),
									progress.WithRenderer(progress.Dots)),
								func(c spinner.TaskComponent, _ *spinner.SpinGroup) error {
									if p, ok := c.(*progress.Progress); ok {
										cacheItems := []string{"users", "products", "categories", "settings", "templates"}
										for i, item := range cacheItems {
											for step := i * 3; step <= (i+1)*3; step++ {
												p.Update(step, fmt.Sprintf("Warming %s cache", item))
												time.Sleep(40 * time.Millisecond)
											}
										}
									}
									return nil
								})

							s.UpdateMessage("Services configured successfully")
							time.Sleep(100 * time.Millisecond)
						}
						return nil
					})

				s.UpdateMessage("Environment preparation complete")
			}
			return nil
		})

	// Second main task demonstrating failure with mixed components
	sg.AddTask("Run Health Checks", spinner.New("Starting comprehensive health checks..."),
		func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
			if s, ok := component.(*spinner.Spinner); ok {
				// Add a successful progress subtask
				sg.AddSubtask("Database Health",
					progress.New("DB Check", 5,
						progress.WithColor(ansi.Green),
						progress.WithRenderer(progress.Minimal)),
					func(c spinner.TaskComponent, _ *spinner.SpinGroup) error {
						if p, ok := c.(*progress.Progress); ok {
							checks := []string{"Connection", "Schema", "Indexes", "Constraints", "Performance"}
							for i, check := range checks {
								p.Update(i+1, "Checking "+check)
								time.Sleep(100 * time.Millisecond)
							}
						}
						return nil
					})

				// Add a failing progress subtask to demonstrate error handling
				sg.AddSubtask("Network Health",
					progress.New("Network Check", 8,
						progress.WithColor(ansi.Yellow),
						progress.WithRenderer(progress.Bar)),
					func(c spinner.TaskComponent, _ *spinner.SpinGroup) error {
						if p, ok := c.(*progress.Progress); ok {
							endpoints := []string{"API Gateway", "Load Balancer", "CDN", "External APIs"}
							for i, endpoint := range endpoints {
								p.Update(i*2+1, fmt.Sprintf("Testing %s connectivity", endpoint))
								time.Sleep(80 * time.Millisecond)
								// Simulate failure on external APIs
								if i == 3 {
									return errors.New("external API connection timeout")
								}
								p.Update(i*2+2, endpoint+" - OK")
								time.Sleep(50 * time.Millisecond)
							}
						}
						return nil
					})

				s.UpdateMessage("Health checks completed")
			}
			return nil
		})

	// This task demonstrates that execution stops on error
	sg.AddTask("Finalize Deployment",
		progress.New("Finalization", 3,
			progress.WithColor(ansi.BrightGreen),
			progress.WithRenderer(progress.Bar)),
		func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
			// This won't run because the previous task failed
			return nil
		})

	err := sg.RunInFrame()
	if err != nil {
		fmt.Printf("❌ Deployment failed as expected: %v\n", err)
		fmt.Println("💡 Notice the hierarchical indentation:")
		fmt.Println("   • Main tasks (no indentation)")
		fmt.Println("   • Subtasks (2-space indentation)")
		fmt.Println("   • Nested subtasks (4-space indentation)")
		fmt.Println("   • Mixed spinners and progress bars with proper icons!")
	} else {
		fmt.Println("✅ All tasks completed successfully!")
	}
}
