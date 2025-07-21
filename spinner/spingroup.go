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

	// SpinGroupTask represents a single task with its spinner and function
	SpinGroupTask struct {
		name     string
		spinner  *Spinner
		taskFunc func() error
	}

	// SpinGroupOption is a function type for configuring spin groups
	SpinGroupOption func(*SpinGroup)
)

// NewSpinGroup creates a new spin group for managing sequential tasks with individual spinners.
// Each task uses a real Spinner instance, making the implementation much simpler and more consistent.
//
// Example:
//
//	// Create a spin group
//	sg := spinner.NewSpinGroup("Deployment Tasks")
//
//	// Add tasks with their own spinners
//	sg.AddTask("Building application", spinner.New("Building..."), func() error {
//		time.Sleep(2 * time.Second)
//		return nil
//	})
//
//	sg.AddTask("Running tests", spinner.New("Testing..."), func() error {
//		time.Sleep(3 * time.Second)
//		return nil
//	})
//
//	// Run all tasks sequentially
//	sg.Run()
//
// Each task gets its own fully-configured Spinner instance with custom colors,
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

// AddTask adds a new task to the spin group with its associated spinner and task function.
// The spinner will be used to show progress while the task function executes.
//
// Example:
//
//	// Custom spinner with specific options
//	s := spinner.New("Compiling...",
//		spinner.WithColor(ansi.Blue),
//		spinner.WithRenderer(spinner.Dots))
//
//	sg.AddTask("Compile", s, func() error {
//		// Your task logic here
//		return buildProject()
//	})
func (sg *SpinGroup) AddTask(name string, spinner *Spinner, taskFunc func() error) {
	sg.mutex.Lock()
	defer sg.mutex.Unlock()

	sg.tasks = append(sg.tasks, SpinGroupTask{
		name:     name,
		spinner:  spinner,
		taskFunc: taskFunc,
	})
}

// Run executes all tasks sequentially, using each task's associated spinner.
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
		// Set spinner output to match SpinGroup output
		task.spinner.frameAware.SetOutput(sg.output)

		// Start the spinner
		task.spinner.Start()

		// Execute the task
		err := task.taskFunc()

		// Stop the spinner with appropriate status
		if err != nil {
			task.spinner.Fail()
			return err
		} else {
			task.spinner.Stop()
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
