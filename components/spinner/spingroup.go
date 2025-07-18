package spinner

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/pseudomuto/gooey/ansi"
	"github.com/pseudomuto/gooey/components/frame"
	frameinternal "github.com/pseudomuto/gooey/internal/frame"
)

const (
	// TaskPending represents a task that has not yet started
	TaskPending TaskStatus = iota
	// TaskRunning represents a task that is currently executing
	TaskRunning TaskStatus = iota
	// TaskCompleted represents a task that has finished successfully
	TaskCompleted TaskStatus = iota
	// TaskFailed represents a task that has failed with an error
	TaskFailed TaskStatus = iota
)

var defaultSpinGroupOutput io.Writer = os.Stdout

type (
	// TaskStatus represents the current state of a task
	TaskStatus int

	// Task represents a single task within the spin group
	Task struct {
		id        int
		name      string
		status    TaskStatus
		result    error
		taskFunc  func() error
		mutex     sync.RWMutex
		startTime time.Time
	}

	// SpinGroup manages multiple sequential tasks with individual spinners
	SpinGroup struct {
		title       string
		tasks       []*Task
		tasksMutex  sync.RWMutex
		frameAware  *frameinternal.FrameAware
		running     bool
		ctx         context.Context
		cancel      context.CancelFunc
		startTime   time.Time
		renderMutex sync.Mutex
	}

	// SpinGroupOption is a function type for configuring spin groups
	SpinGroupOption func(*SpinGroup)
)

// String returns the string representation of the task status.
// This method provides human-readable names for task statuses.
func (ts TaskStatus) String() string {
	switch ts {
	case TaskPending:
		return "pending"
	case TaskRunning:
		return "running"
	case TaskCompleted:
		return "completed"
	case TaskFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// NewSpinGroup creates a new spin group for managing sequential tasks with individual spinners.
// The spin group executes tasks sequentially, showing a spinner for each task and displaying
// their status in a structured format.
//
// Example:
//
//	// Create a spin group
//	sg := spinner.NewSpinGroup("Deployment Tasks")
//
//	// Add tasks that will be executed sequentially
//	sg.AddTask("Building application", func() error {
//		time.Sleep(2 * time.Second)
//		return nil
//	})
//
//	sg.AddTask("Running tests", func() error {
//		time.Sleep(3 * time.Second)
//		return nil
//	})
//
//	sg.AddTask("Deploying to production", func() error {
//		time.Sleep(1 * time.Second)
//		return nil
//	})
//
//	// Start execution and wait for completion
//	sg.Start()
//	sg.Wait()
//
//	// Custom output destination
//	f := frame.Open("Operations", frame.WithColor(ansi.Blue))
//	sg := spinner.NewSpinGroup("Tasks", spinner.WithSpinGroupOutput(f))
//
// Each task runs with its own spinner animation and shows completion status.
// Tasks are executed sequentially, and the group shows overall progress.
func NewSpinGroup(title string, options ...SpinGroupOption) *SpinGroup {
	ctx, cancel := context.WithCancel(context.Background())

	sg := &SpinGroup{
		title:      title,
		tasks:      make([]*Task, 0),
		frameAware: frameinternal.NewFrameAware(defaultSpinGroupOutput),
		running:    false,
		ctx:        ctx,
		cancel:     cancel,
	}

	for _, option := range options {
		option(sg)
	}

	return sg
}

// AddTask adds a new task to the spin group and returns its ID.
// The task function will be executed sequentially when Start() is called.
// Returns the task ID which can be used to track the task's progress.
func (sg *SpinGroup) AddTask(name string, taskFunc func() error) int {
	sg.tasksMutex.Lock()
	defer sg.tasksMutex.Unlock()

	taskID := len(sg.tasks)

	task := &Task{
		id:       taskID,
		name:     name,
		status:   TaskPending,
		taskFunc: taskFunc,
	}

	sg.tasks = append(sg.tasks, task)

	return taskID
}

// Start begins the spin group execution
func (sg *SpinGroup) Start() {
	sg.tasksMutex.Lock()
	if sg.running {
		sg.tasksMutex.Unlock()
		return
	}

	sg.running = true
	sg.startTime = time.Now()
	sg.tasksMutex.Unlock()

	// Start sequential task execution
	go sg.executeTasksSequentially()
}

// Stop stops the spin group execution
func (sg *SpinGroup) Stop() {
	sg.tasksMutex.Lock()
	if !sg.running {
		sg.tasksMutex.Unlock()
		return
	}

	sg.running = false
	sg.tasksMutex.Unlock()

	// Cancel all running tasks
	sg.cancel()
}

// Wait waits for all tasks to complete or for execution to stop
func (sg *SpinGroup) Wait() {
	for {
		sg.tasksMutex.RLock()
		allDone := true
		for _, task := range sg.tasks {
			task.mutex.RLock()
			if task.status == TaskPending || task.status == TaskRunning {
				allDone = false
				task.mutex.RUnlock()
				break
			}
			task.mutex.RUnlock()
		}
		sg.tasksMutex.RUnlock()

		if allDone || !sg.IsRunning() {
			break
		}

		time.Sleep(10 * time.Millisecond)
	}
}

// executeTasksSequentially runs all tasks one by one, stopping on first failure
func (sg *SpinGroup) executeTasksSequentially() {
	defer sg.Stop()

	for {
		// Check if we should continue running
		sg.tasksMutex.RLock()
		running := sg.running
		sg.tasksMutex.RUnlock()

		if !running {
			break
		}

		// Find next pending task
		sg.tasksMutex.RLock()
		var nextTask *Task
		for _, task := range sg.tasks {
			task.mutex.RLock()
			if task.status == TaskPending {
				nextTask = task
				task.mutex.RUnlock()
				break
			}
			task.mutex.RUnlock()
		}
		sg.tasksMutex.RUnlock()

		if nextTask == nil {
			// No pending tasks, wait a bit for new tasks to be added
			time.Sleep(50 * time.Millisecond)
			continue
		}

		// Execute the task
		err := sg.runTask(nextTask)
		// If task failed, stop processing subsequent tasks
		if err != nil {
			break
		}
	}
}

// runTask executes a single task and returns error if failed
func (sg *SpinGroup) runTask(task *Task) error {
	// Update task status to running
	task.mutex.Lock()
	task.status = TaskRunning
	task.startTime = time.Now()
	task.mutex.Unlock()

	// Render the task when it starts running (for non-frame cases)
	sg.renderTaskStart(task)

	// Execute task with context
	done := make(chan error, 1)
	go func() {
		done <- task.taskFunc()
	}()

	var result error
	select {
	case result = <-done:
		// Task completed normally
	case <-sg.ctx.Done():
		// Task was cancelled
		result = sg.ctx.Err()
	}

	// Update task status based on result
	task.mutex.Lock()
	task.result = result
	if result != nil {
		task.status = TaskFailed
	} else {
		task.status = TaskCompleted
	}
	task.mutex.Unlock()

	// Render the completion immediately
	sg.renderTaskCompletion(task)

	return result
}

// renderTaskStart renders a task when it starts running
func (sg *SpinGroup) renderTaskStart(task *Task) {
	sg.renderMutex.Lock()
	defer sg.renderMutex.Unlock()

	output := sg.frameAware.Output()
	task.mutex.RLock()
	line := sg.formatTaskLine(TaskRunning, task.name, task.startTime)
	task.mutex.RUnlock()

	if sg.frameAware.InFrame() {
		if frameWriter, ok := output.(*frame.Frame); ok {
			frameWriter.Println("%s", line)
		}
	} else {
		fmt.Fprintf(output, "%s\n", line)
	}
}

// renderTaskCompletion renders a single task completion
func (sg *SpinGroup) renderTaskCompletion(task *Task) {
	sg.renderMutex.Lock()
	defer sg.renderMutex.Unlock()

	output := sg.frameAware.Output()
	task.mutex.RLock()
	line := sg.formatTaskLine(task.status, task.name, task.startTime)
	task.mutex.RUnlock()

	if sg.frameAware.InFrame() {
		if frameWriter, ok := output.(*frame.Frame); ok {
			// For frames, use ReplaceLine to update the running task line with completion
			frameWriter.ReplaceLine("%s", line)
		}
	} else {
		// For non-frame, replace the last line (the running state) with completion state
		fmt.Fprint(output, ansi.MoveCursorUp(1)+ansi.ClearLine+line+"\n")
	}
}

// formatTaskLine formats a single task line
func (sg *SpinGroup) formatTaskLine(status TaskStatus, message string, taskStartTime time.Time) string {
	switch status {
	case TaskPending:
		return "  ⏳ " + ansi.Yellow.Colorize(message)
	case TaskRunning:
		// Show spinning indicator
		icons := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧"}
		frame := int(time.Since(taskStartTime).Milliseconds() / 100) // Use 100ms interval
		icon := ansi.Red.Colorize(icons[frame%len(icons)])
		return fmt.Sprintf("  %s %s", icon, message)
	case TaskCompleted:
		return fmt.Sprintf("  %s %s", ansi.CheckMark.Colorize(ansi.Green), ansi.Green.Colorize(message))
	case TaskFailed:
		return fmt.Sprintf("  %s %s", ansi.CrossMark.Colorize(ansi.Red), ansi.Red.Colorize(message))
	default:
		return "  " + message
	}
}

// WithSpinGroupOutput sets the output writer for the spin group
func WithSpinGroupOutput(output io.Writer) SpinGroupOption {
	return func(sg *SpinGroup) {
		sg.frameAware.SetOutput(output)
	}
}

// WithMaxConcurrent is deprecated - SpinGroup now runs tasks sequentially
func WithMaxConcurrent(max int) SpinGroupOption {
	return func(sg *SpinGroup) {
		// No-op: sequential execution only
	}
}

// TaskCount returns the number of tasks in the spin group
func (sg *SpinGroup) TaskCount() int {
	sg.tasksMutex.RLock()
	defer sg.tasksMutex.RUnlock()
	return len(sg.tasks)
}

// IsRunning returns true if the spin group is currently executing tasks
func (sg *SpinGroup) IsRunning() bool {
	sg.tasksMutex.RLock()
	defer sg.tasksMutex.RUnlock()
	return sg.running
}

// Elapsed returns the duration since the spin group started
func (sg *SpinGroup) Elapsed() time.Duration {
	return time.Since(sg.startTime)
}
