package spinner

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/pseudomuto/gooey/frame"
	"github.com/pseudomuto/gooey/internal/writer"
)

var defaultSpinGroupOutput io.Writer = os.Stdout

type (
	// SpinGroup manages multiple sequential tasks using actual Spinner instances
	SpinGroup struct {
		title         string
		tasks         []SpinGroupTask
		mutex         sync.RWMutex
		output        io.Writer
		running       bool
		startTime     time.Time
		currentIndex  int // Track the currently executing task for dynamic insertion
		subtaskOffset int // Track how many subtasks have been added during current task execution
	}

	// SpinGroupTask represents a single task with its component and function
	SpinGroupTask struct {
		name      string
		component TaskComponent
		taskFunc  func(TaskComponent, *SpinGroup) error
		depth     int // Track nesting depth for indentation (0 = root task, 1+ = subtask)
	}

	// SpinGroupOption is a function type for configuring spin groups
	SpinGroupOption func(*SpinGroup)
)

// NewSpinGroup creates a new spin group for managing sequential tasks with TaskComponents.
// Each task can use either a Spinner (for indefinite tasks) or Progress (for definite tasks),
// providing flexible visual feedback that matches the task characteristics.
//
// Example with Spinners:
//
//	sg := spinner.NewSpinGroup("Deployment Tasks")
//
//	// Indefinite tasks with spinners
//	sg.AddTask("Building", spinner.New("Building application..."), func() error {
//		return buildApp()
//	})
//
// Example with Mixed Components:
//
//	// Mix indefinite and definite tasks in the same workflow
//	sg.AddTask("Connect", spinner.New("Connecting to server..."),
//		func(component TaskComponent, sg *SpinGroup) error {
//			// Component is passed to task function for dynamic updates
//			if s, ok := component.(*spinner.Spinner); ok {
//				s.UpdateMessage("Establishing connection...")
//				time.Sleep(1 * time.Second)
//				s.UpdateMessage("Authentication successful")
//			}
//			return nil
//		})
//
//	sg.AddTask("Download", progress.New("Download", 100),
//		func(component TaskComponent, sg *SpinGroup) error {
//			if p, ok := component.(*progress.Progress); ok {
//				for i := 0; i <= 100; i += 10 {
//					p.Update(i, fmt.Sprintf("Downloaded %d%%", i))
//					time.Sleep(50 * time.Millisecond)
//				}
//			}
//			return nil
//		})
//
//	// Run all tasks sequentially
//	sg.Run()
//
// Each task gets its own fully-configured component with custom colors,
// renderers, intervals, and other options.
func NewSpinGroup(title string, options ...SpinGroupOption) *SpinGroup {
	sg := &SpinGroup{
		title:  title,
		tasks:  make([]SpinGroupTask, 0),
		output: defaultSpinGroupOutput,
	}

	for _, option := range options {
		option(sg)
	}

	return sg
}

// AddTask adds a new task to the spin group with its associated component and task function.
// The component can be either a Spinner (for indefinite tasks) or Progress (for definite tasks).
// The task function receives the component and SpinGroup as arguments, allowing for dynamic updates
// and the ability to add subtasks during execution.
//
// Example with Spinner:
//
//	sg.AddTask("Compile", spinner.New("Compiling...", spinner.WithColor(ansi.Blue)),
//		func(s TaskComponent, sg *SpinGroup) error {
//			if spinner, ok := s.(*spinner.Spinner); ok {
//				spinner.UpdateMessage("Compiling source files...")
//				time.Sleep(2 * time.Second)
//				spinner.UpdateMessage("Linking binaries...")
//				time.Sleep(1 * time.Second)
//			}
//			return nil
//		})
//
// Example with Dynamic Subtasks:
//
//	sg.AddTask("Deploy", spinner.New("Deploying application..."),
//		func(component TaskComponent, sg *SpinGroup) error {
//			// Discover services to deploy
//			services := []string{"web", "api", "worker"}
//			for _, service := range services {
//				// Add subtasks dynamically
//				sg.AddSubtask(fmt.Sprintf("Deploy %s", service),
//					spinner.New(fmt.Sprintf("Deploying %s service...", service)),
//					func(c TaskComponent, _ *SpinGroup) error {
//						time.Sleep(1 * time.Second)
//						return deployService(service)
//					})
//			}
//			return nil
//		})
//
// Example with Progress:
//
//	sg.AddTask("Download", progress.New("Downloading", 0, progress.WithColor(ansi.Green)), // Unknown total initially
//		func(p TaskComponent, sg *SpinGroup) error {
//			if progress, ok := p.(*progress.Progress); ok {
//				// Simulate discovering file size from HTTP headers
//				time.Sleep(100 * time.Millisecond)
//				fileSize := 1024 * 1024 // 1MB
//				progress.SetTotal(fileSize)
//
//				for downloaded := 0; downloaded <= fileSize; downloaded += 102400 {
//					progress.Update(downloaded, fmt.Sprintf("Downloaded %d bytes", downloaded))
//					time.Sleep(10 * time.Millisecond)
//				}
//			}
//			return nil
//		})

func (sg *SpinGroup) AddTask(name string, component TaskComponent, taskFunc func(TaskComponent, *SpinGroup) error) {
	sg.mutex.Lock()
	defer sg.mutex.Unlock()

	sg.tasks = append(sg.tasks, SpinGroupTask{
		name:      name,
		component: component,
		taskFunc:  taskFunc,
		depth:     0, // Root tasks have depth 0
	})
}

// AddSubtask dynamically adds a new task to the spin group during execution, inserting it
// immediately after the currently executing task. This allows for hierarchical task structures
// where a parent task can discover and add subtasks during its execution.
//
// This method is safe to call from within task functions and will cause the subtasks to be
// executed in the order they were added, immediately after the current task completes.
//
// Example:
//
//	sg.AddTask("Deploy Services", spinner.New("Discovering services..."),
//		func(component TaskComponent, sg *SpinGroup) error {
//			services := discoverServices() // Returns []string{"web", "api", "worker"}
//
//			// Add subtasks for each discovered service
//			for _, service := range services {
//				sg.AddSubtask(fmt.Sprintf("Deploy %s", service),
//					spinner.New(fmt.Sprintf("Deploying %s...", service)),
//					func(c TaskComponent, _ *SpinGroup) error {
//						return deployService(service)
//					})
//			}
//			return nil
//		})
func (sg *SpinGroup) AddSubtask(name string, component TaskComponent, taskFunc func(TaskComponent, *SpinGroup) error) {
	sg.mutex.Lock()
	defer sg.mutex.Unlock()

	// Insert subtask after current task and any previously added subtasks
	insertIndex := sg.currentIndex + 1 + sg.subtaskOffset

	// Get the current task's depth to determine subtask depth
	currentTaskDepth := 0
	if sg.currentIndex < len(sg.tasks) {
		currentTaskDepth = sg.tasks[sg.currentIndex].depth
	}

	newTask := SpinGroupTask{
		name:      name,
		component: component,
		taskFunc:  taskFunc,
		depth:     currentTaskDepth + 1, // Subtasks are one level deeper
	}

	// Insert at the correct position
	if insertIndex >= len(sg.tasks) {
		sg.tasks = append(sg.tasks, newTask)
	} else {
		// Insert in the middle by creating a new slice
		sg.tasks = append(sg.tasks[:insertIndex], append([]SpinGroupTask{newTask}, sg.tasks[insertIndex:]...)...)
	}

	// Increment the subtask offset to ensure next subtask is inserted after this one
	sg.subtaskOffset++
}

// Run executes all tasks sequentially, using each task's associated component.
// If any task fails, execution stops and the error is returned.
func (sg *SpinGroup) Run() error {
	if err := sg.validate(); err != nil {
		return err
	}

	sg.initializeExecution()
	defer sg.finalizeExecution()

	return sg.executeTasksSequentially()
}

// initializeExecution sets up the execution state
func (sg *SpinGroup) initializeExecution() {
	sg.mutex.Lock()
	defer sg.mutex.Unlock()

	sg.running = true
	sg.startTime = time.Now()
}

// finalizeExecution cleans up the execution state
func (sg *SpinGroup) finalizeExecution() {
	sg.mutex.Lock()
	defer sg.mutex.Unlock()

	sg.running = false
	sg.currentIndex = 0  // Reset current index when done
	sg.subtaskOffset = 0 // Reset subtask offset when done
}

// executeTasksSequentially runs all tasks in sequence, handling dynamic task insertion
func (sg *SpinGroup) executeTasksSequentially() error {
	// Use index-based iteration to handle dynamic task insertion
	for i := 0; i < len(sg.tasks); i++ {
		if err := sg.executeTask(i); err != nil {
			return err
		}
	}
	return nil
}

// executeTask runs a single task with proper setup and cleanup
func (sg *SpinGroup) executeTask(taskIndex int) error {
	task := sg.prepareTaskForExecution(taskIndex)
	if task == nil {
		return nil // Task was modified during execution, skip
	}

	// Set component output with appropriate indentation based on task depth
	taskOutput := writer.NewIndentedWriter(sg.output, task.depth)
	task.component.SetOutput(taskOutput)

	// Start the component (spinners animate, progress shows)
	task.component.Start()

	// Execute the task, passing both component and SpinGroup for dynamic updates
	err := task.taskFunc(task.component, sg)

	// Complete the component with appropriate status
	if err != nil {
		task.component.Fail(err.Error())
		return err
	} else {
		task.component.Complete("")
	}

	return nil
}

// prepareTaskForExecution sets up task execution state and returns the task to execute
func (sg *SpinGroup) prepareTaskForExecution(taskIndex int) *SpinGroupTask {
	sg.mutex.Lock()
	defer sg.mutex.Unlock()

	sg.currentIndex = taskIndex
	sg.subtaskOffset = 0 // Reset subtask offset for each new task

	if taskIndex >= len(sg.tasks) {
		// Tasks array was modified during execution, skip if index is out of bounds
		return nil
	}

	return &sg.tasks[taskIndex]
}

// validate ensures the SpinGroup is properly configured before execution
func (sg *SpinGroup) validate() error {
	sg.mutex.RLock()
	defer sg.mutex.RUnlock()

	if sg.title == "" {
		return errors.New("spingroup title cannot be empty")
	}

	if len(sg.tasks) == 0 {
		return errors.New("spingroup must have at least one task")
	}

	// Validate each task
	for _, task := range sg.tasks {
		if task.name == "" {
			return errors.New("task name cannot be empty")
		}
		if task.component == nil {
			return errors.New("task component cannot be nil")
		}
		if task.taskFunc == nil {
			return errors.New("task function cannot be nil")
		}
		if task.depth < 0 {
			return errors.New("task depth cannot be negative")
		}
	}

	return nil
}

// RunInFrame runs all tasks within a frame for organized display
func (sg *SpinGroup) RunInFrame() error {
	f := frame.Open(sg.title, frame.WithOutput(sg.output))
	defer f.Close()

	// Instead of setting the frame as the output (which causes nesting issues),
	// we'll run tasks with their indented writers outputting directly to the frame
	// This avoids the nested frame problem while preserving the frame border
	originalOutput := sg.output
	sg.output = f

	err := sg.Run()

	// Restore original output
	sg.output = originalOutput

	return err
}

// TaskCount returns the number of tasks in the spin group
func (sg *SpinGroup) TaskCount() int {
	sg.mutex.RLock()
	defer sg.mutex.RUnlock()
	return len(sg.tasks)
}

// Title returns the title of the spin group
func (sg *SpinGroup) Title() string {
	return sg.title
}

// WithSpinGroupOutput sets the output writer for the spin group
func WithSpinGroupOutput(output io.Writer) SpinGroupOption {
	return func(sg *SpinGroup) {
		sg.output = output
	}
}
