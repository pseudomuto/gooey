package spinner

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/pseudomuto/gooey/frame"
)

var defaultSpinGroupOutput io.Writer = os.Stdout

type (
	// SpinGroup manages multiple sequential tasks using actual Spinner instances
	SpinGroup struct {
		title     string
		tasks     []SpinGroupTask
		mutex     sync.RWMutex
		output    io.Writer
		running   bool
		startTime time.Time
	}

	// SpinGroupTask represents a single task with its component and function
	SpinGroupTask struct {
		name      string
		component TaskComponent
		taskFunc  func(TaskComponent) error
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
//		func(component TaskComponent) error {
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
//		func(component TaskComponent) error {
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
// The task function receives the component as an argument, allowing for dynamic updates during execution.
//
// Example with Spinner:
//
//	sg.AddTask("Compile", spinner.New("Compiling...", spinner.WithColor(ansi.Blue)),
//		func(s TaskComponent) error {
//			if spinner, ok := s.(*spinner.Spinner); ok {
//				spinner.UpdateMessage("Compiling source files...")
//				time.Sleep(2 * time.Second)
//				spinner.UpdateMessage("Linking binaries...")
//				time.Sleep(1 * time.Second)
//			}
//			return nil
//		})
//
// Example with Progress:
//
//	sg.AddTask("Download", progress.New("Downloading", 0, progress.WithColor(ansi.Green)), // Unknown total initially
//		func(p TaskComponent) error {
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
func (sg *SpinGroup) AddTask(name string, component TaskComponent, taskFunc func(TaskComponent) error) {
	sg.mutex.Lock()
	defer sg.mutex.Unlock()

	sg.tasks = append(sg.tasks, SpinGroupTask{
		name:      name,
		component: component,
		taskFunc:  taskFunc,
	})
}

// Run executes all tasks sequentially, using each task's associated component.
// If any task fails, execution stops and the error is returned.
func (sg *SpinGroup) Run() error {
	sg.mutex.Lock()
	sg.running = true
	sg.startTime = time.Now()
	sg.mutex.Unlock()

	defer func() {
		sg.mutex.Lock()
		sg.running = false
		sg.mutex.Unlock()
	}()

	for _, task := range sg.tasks {
		// Set component output to match SpinGroup output
		task.component.SetOutput(sg.output)

		// Start the component (spinners animate, progress shows)
		task.component.Start()

		// Execute the task, passing the component for dynamic updates
		err := task.taskFunc(task.component)

		// Complete the component with appropriate status
		if err != nil {
			task.component.Fail(err.Error())
			return err
		} else {
			task.component.Complete("")
		}
	}

	return nil
}

// RunInFrame runs all tasks within a frame for organized display
func (sg *SpinGroup) RunInFrame() error {
	f := frame.Open(sg.title, frame.WithOutput(sg.output))
	defer f.Close()

	// Temporarily set output to the frame
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
