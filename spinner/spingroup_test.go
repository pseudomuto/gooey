package spinner_test

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/pseudomuto/gooey/ansi"
	"github.com/pseudomuto/gooey/progress"
	"github.com/pseudomuto/gooey/spinner"
	"github.com/stretchr/testify/require"
)

func TestNewSpinGroup(t *testing.T) {
	sg := spinner.NewSpinGroup("Test Group")

	require.NotNil(t, sg)
	require.Equal(t, 0, sg.TaskCount())
	require.Equal(t, "Test Group", sg.Title())
}

func TestNewSpinGroupWithOptions(t *testing.T) {
	buf := &bytes.Buffer{}
	sg := spinner.NewSpinGroup("Test Group", spinner.WithSpinGroupOutput(buf))

	require.NotNil(t, sg)
	require.Equal(t, 0, sg.TaskCount())
	require.Equal(t, "Test Group", sg.Title())
}

func TestSpinGroup_AddTask(t *testing.T) {
	sg := spinner.NewSpinGroup("Test Group")

	sg.AddTask("Task 1", spinner.New("Task 1 message"), func(spinner.TaskComponent, *spinner.SpinGroup) error {
		return nil
	})

	require.Equal(t, 1, sg.TaskCount())

	// Add another task
	sg.AddTask("Task 2", spinner.New("Task 2 message"), func(spinner.TaskComponent, *spinner.SpinGroup) error {
		return nil
	})

	require.Equal(t, 2, sg.TaskCount())
}

func TestSpinGroup_Run(t *testing.T) {
	buf := &bytes.Buffer{}
	sg := spinner.NewSpinGroup("Test Group", spinner.WithSpinGroupOutput(buf))

	var executed bool
	sg.AddTask("Task 1", spinner.New("Executing..."), func(spinner.TaskComponent, *spinner.SpinGroup) error {
		time.Sleep(10 * time.Millisecond)
		executed = true
		return nil
	})

	err := sg.Run()
	require.NoError(t, err)
	require.True(t, executed)

	output := buf.String()
	require.NotEmpty(t, output)
	require.Contains(t, output, "✓") // Should show success indicator
}

func TestSpinGroup_ExecuteTasksSequentially(t *testing.T) {
	buf := &bytes.Buffer{}
	sg := spinner.NewSpinGroup("Test Group", spinner.WithSpinGroupOutput(buf))

	var executionOrder []int
	var mu sync.Mutex

	// Add tasks that record execution order
	sg.AddTask("Task 1", spinner.New("Task 1 executing..."), func(spinner.TaskComponent, *spinner.SpinGroup) error {
		mu.Lock()
		executionOrder = append(executionOrder, 1)
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	sg.AddTask("Task 2", spinner.New("Task 2 executing..."), func(spinner.TaskComponent, *spinner.SpinGroup) error {
		mu.Lock()
		executionOrder = append(executionOrder, 2)
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	sg.AddTask("Task 3", spinner.New("Task 3 executing..."), func(spinner.TaskComponent, *spinner.SpinGroup) error {
		mu.Lock()
		executionOrder = append(executionOrder, 3)
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	err := sg.Run()
	require.NoError(t, err)
	require.Equal(t, []int{1, 2, 3}, executionOrder)
}

func TestSpinGroup_StopOnFirstFailure(t *testing.T) {
	buf := &bytes.Buffer{}
	sg := spinner.NewSpinGroup("Test Group", spinner.WithSpinGroupOutput(buf))

	var executionOrder []int
	var mu sync.Mutex

	// Add tasks where second task fails
	sg.AddTask("Task 1", spinner.New("Task 1 executing..."), func(spinner.TaskComponent, *spinner.SpinGroup) error {
		mu.Lock()
		executionOrder = append(executionOrder, 1)
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	sg.AddTask("Task 2", spinner.New("Task 2 executing..."), func(spinner.TaskComponent, *spinner.SpinGroup) error {
		mu.Lock()
		executionOrder = append(executionOrder, 2)
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		return errors.New("task failed")
	})

	sg.AddTask("Task 3", spinner.New("Task 3 executing..."), func(spinner.TaskComponent, *spinner.SpinGroup) error {
		mu.Lock()
		executionOrder = append(executionOrder, 3)
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	err := sg.Run()
	require.Error(t, err)
	require.Contains(t, err.Error(), "task failed")
	// Task 3 should not execute because task 2 failed
	require.Equal(t, []int{1, 2}, executionOrder)

	output := buf.String()
	require.Contains(t, output, "✗") // Should show failure indicator
}

func TestSpinGroup_RunInFrame(t *testing.T) {
	buf := &bytes.Buffer{}
	sg := spinner.NewSpinGroup("Test Group", spinner.WithSpinGroupOutput(buf))

	sg.AddTask("Frame Task", spinner.New("Running in frame..."), func(spinner.TaskComponent, *spinner.SpinGroup) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	err := sg.RunInFrame()
	require.NoError(t, err)

	output := buf.String()
	require.NotEmpty(t, output)
	require.Contains(t, output, "Test Group") // Frame title should contain the SpinGroup title
	require.Contains(t, output, "✓")          // Should show success indicator
}

func TestSpinGroup_WithCustomSpinners(t *testing.T) {
	buf := &bytes.Buffer{}
	sg := spinner.NewSpinGroup("Custom Spinners Test", spinner.WithSpinGroupOutput(buf))

	// Add task with custom spinner configuration
	customSpinner := spinner.New("Custom spinner message",
		spinner.WithColor(ansi.Blue),
		spinner.WithRenderer(spinner.Dots))

	sg.AddTask("Custom Task", customSpinner, func(spinner.TaskComponent, *spinner.SpinGroup) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	err := sg.Run()
	require.NoError(t, err)

	output := buf.String()
	require.NotEmpty(t, output)
	require.Contains(t, output, "✓") // Should show success indicator
}

func TestSpinGroup_ConcurrentTaskAddition(t *testing.T) {
	sg := spinner.NewSpinGroup("Concurrent Test")
	var wg sync.WaitGroup

	// Add tasks concurrently to test thread safety
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sg.AddTask("Task", spinner.New("Running..."), func(spinner.TaskComponent, *spinner.SpinGroup) error {
				time.Sleep(5 * time.Millisecond)
				return nil
			})
		}()
	}

	wg.Wait()
	require.Equal(t, 5, sg.TaskCount())

	err := sg.Run()
	require.NoError(t, err)
}

func TestSpinGroup_MixedComponents(t *testing.T) {
	buf := &bytes.Buffer{}
	sg := spinner.NewSpinGroup("Mixed Components Test", spinner.WithSpinGroupOutput(buf))

	// Add indefinite task with spinner
	sg.AddTask("Connect", spinner.New("Connecting to server..."), func(spinner.TaskComponent, *spinner.SpinGroup) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	// Add definite task with progress
	progressBar := progress.New("Download", 3)
	sg.AddTask("Download", progressBar, func(spinner.TaskComponent, *spinner.SpinGroup) error {
		for i := 0; i <= 3; i++ {
			progressBar.Update(i, fmt.Sprintf("Downloaded %d files", i))
			time.Sleep(5 * time.Millisecond)
		}
		return nil
	})

	// Add another indefinite task
	sg.AddTask("Cleanup", spinner.New("Cleaning up..."), func(spinner.TaskComponent, *spinner.SpinGroup) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	err := sg.Run()
	require.NoError(t, err)

	output := buf.String()
	require.Contains(t, output, "Connecting to server...")
	require.Contains(t, output, "Download")
	require.Contains(t, output, "Cleaning up...")
	require.Contains(t, output, "✓") // Should show success indicators
}

func TestSpinGroup_MixedComponentsWithFailure(t *testing.T) {
	buf := &bytes.Buffer{}
	sg := spinner.NewSpinGroup("Mixed Failure Test", spinner.WithSpinGroupOutput(buf))

	// Success task with spinner
	sg.AddTask("Connect", spinner.New("Connecting..."), func(spinner.TaskComponent, *spinner.SpinGroup) error {
		time.Sleep(5 * time.Millisecond)
		return nil
	})

	// Failed task with progress
	progressBar := progress.New("Upload", 10)
	sg.AddTask("Upload", progressBar, func(spinner.TaskComponent, *spinner.SpinGroup) error {
		progressBar.Update(5, "Uploading...")
		time.Sleep(5 * time.Millisecond)
		return errors.New("upload failed")
	})

	// This task should not execute
	sg.AddTask("Finalize", spinner.New("Finalizing..."), func(spinner.TaskComponent, *spinner.SpinGroup) error {
		time.Sleep(5 * time.Millisecond)
		return nil
	})

	err := sg.Run()
	require.Error(t, err)
	require.Contains(t, err.Error(), "upload failed")

	output := buf.String()
	require.Contains(t, output, "✓") // Should show success for first task
	require.Contains(t, output, "✗") // Should show failure for failed task
}

func TestSpinGroup_AddSubtask(t *testing.T) {
	buf := &bytes.Buffer{}
	sg := spinner.NewSpinGroup("Dynamic Test", spinner.WithSpinGroupOutput(buf))

	var executionOrder []string
	var mu sync.Mutex

	// Main task that adds subtasks
	sg.AddTask("Main Task", spinner.New("Main task executing..."), func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
		mu.Lock()
		executionOrder = append(executionOrder, "main")
		mu.Unlock()

		// Add subtasks dynamically
		sg.AddSubtask("Subtask 1", spinner.New("Subtask 1 executing..."), func(spinner.TaskComponent, *spinner.SpinGroup) error {
			mu.Lock()
			executionOrder = append(executionOrder, "sub1")
			mu.Unlock()
			return nil
		})

		sg.AddSubtask("Subtask 2", spinner.New("Subtask 2 executing..."), func(spinner.TaskComponent, *spinner.SpinGroup) error {
			mu.Lock()
			executionOrder = append(executionOrder, "sub2")
			mu.Unlock()
			return nil
		})

		return nil
	})

	// Final task that should run after all subtasks
	sg.AddTask("Final Task", spinner.New("Final task executing..."), func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
		mu.Lock()
		executionOrder = append(executionOrder, "final")
		mu.Unlock()
		return nil
	})

	err := sg.Run()
	require.NoError(t, err)

	// Should execute: main, sub1, sub2, final
	require.Equal(t, []string{"main", "sub1", "sub2", "final"}, executionOrder)
	require.Equal(t, 4, sg.TaskCount()) // Original 2 tasks + 2 subtasks
}

func TestSpinGroup_NestedSubtasks(t *testing.T) {
	buf := &bytes.Buffer{}
	sg := spinner.NewSpinGroup("Nested Dynamic Test", spinner.WithSpinGroupOutput(buf))

	var executionOrder []string
	var mu sync.Mutex

	// Main task that adds subtasks, and subtasks add their own subtasks
	sg.AddTask("Main Task", spinner.New("Main task executing..."), func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
		mu.Lock()
		executionOrder = append(executionOrder, "main")
		mu.Unlock()

		// Add subtasks that will also add subtasks
		sg.AddSubtask("Parent Subtask", spinner.New("Parent subtask executing..."), func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
			mu.Lock()
			executionOrder = append(executionOrder, "parent")
			mu.Unlock()

			// Add nested subtasks
			sg.AddSubtask("Nested Subtask 1", spinner.New("Nested 1 executing..."), func(spinner.TaskComponent, *spinner.SpinGroup) error {
				mu.Lock()
				executionOrder = append(executionOrder, "nested1")
				mu.Unlock()
				return nil
			})

			sg.AddSubtask("Nested Subtask 2", spinner.New("Nested 2 executing..."), func(spinner.TaskComponent, *spinner.SpinGroup) error {
				mu.Lock()
				executionOrder = append(executionOrder, "nested2")
				mu.Unlock()
				return nil
			})

			return nil
		})

		return nil
	})

	err := sg.Run()
	require.NoError(t, err)

	// Should execute: main, parent, nested1, nested2
	require.Equal(t, []string{"main", "parent", "nested1", "nested2"}, executionOrder)
	require.Equal(t, 4, sg.TaskCount()) // Original 1 task + 3 dynamically added
}

func TestSpinGroup_SubtaskWithFailure(t *testing.T) {
	buf := &bytes.Buffer{}
	sg := spinner.NewSpinGroup("Failure Test", spinner.WithSpinGroupOutput(buf))

	var executionOrder []string
	var mu sync.Mutex

	// Main task that adds subtasks, where one subtask fails
	sg.AddTask("Main Task", spinner.New("Main task executing..."), func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
		mu.Lock()
		executionOrder = append(executionOrder, "main")
		mu.Unlock()

		// Add subtasks where second one fails
		sg.AddSubtask("Success Subtask", spinner.New("Success subtask executing..."), func(spinner.TaskComponent, *spinner.SpinGroup) error {
			mu.Lock()
			executionOrder = append(executionOrder, "success")
			mu.Unlock()
			return nil
		})

		sg.AddSubtask("Failing Subtask", spinner.New("Failing subtask executing..."), func(spinner.TaskComponent, *spinner.SpinGroup) error {
			mu.Lock()
			executionOrder = append(executionOrder, "failing")
			mu.Unlock()
			return errors.New("subtask failed")
		})

		// This shouldn't be reached because the parent task adds subtasks dynamically
		sg.AddSubtask("Never Reached", spinner.New("Never reached..."), func(spinner.TaskComponent, *spinner.SpinGroup) error {
			mu.Lock()
			executionOrder = append(executionOrder, "never")
			mu.Unlock()
			return nil
		})

		return nil
	})

	// This task should not execute because subtask fails
	sg.AddTask("Final Task", spinner.New("Final task executing..."), func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
		mu.Lock()
		executionOrder = append(executionOrder, "final")
		mu.Unlock()
		return nil
	})

	err := sg.Run()
	require.Error(t, err)
	require.Contains(t, err.Error(), "subtask failed")

	// Should execute: main, success, failing, never (all subtasks are added before execution)
	require.Equal(t, []string{"main", "success", "failing"}, executionOrder)
}

func TestSpinGroup_SubtaskConcurrentAddition(t *testing.T) {
	buf := &bytes.Buffer{}
	sg := spinner.NewSpinGroup("Concurrent Subtask Test", spinner.WithSpinGroupOutput(buf))

	var executionCount int
	var mu sync.Mutex

	// Main task that adds multiple subtasks concurrently (simulating real-world discovery)
	sg.AddTask("Discovery Task", spinner.New("Discovering services..."), func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
		var wg sync.WaitGroup

		// Simulate discovering and adding subtasks concurrently
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				sg.AddSubtask(fmt.Sprintf("Service %d", id), spinner.New(fmt.Sprintf("Deploying service %d...", id)), func(spinner.TaskComponent, *spinner.SpinGroup) error {
					mu.Lock()
					executionCount++
					mu.Unlock()
					time.Sleep(5 * time.Millisecond)
					return nil
				})
			}(i)
		}

		wg.Wait() // Wait for all subtasks to be added
		return nil
	})

	err := sg.Run()
	require.NoError(t, err)
	require.Equal(t, 3, executionCount) // All 3 subtasks should execute
	require.Equal(t, 4, sg.TaskCount()) // Original 1 task + 3 subtasks
}

func TestSpinGroup_SubtaskWithProgress(t *testing.T) {
	buf := &bytes.Buffer{}
	sg := spinner.NewSpinGroup("Progress Subtask Test", spinner.WithSpinGroupOutput(buf))

	var executionOrder []string
	var mu sync.Mutex

	// Main task that adds progress subtasks
	sg.AddTask("Setup", spinner.New("Setting up..."), func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
		mu.Lock()
		executionOrder = append(executionOrder, "setup")
		mu.Unlock()

		// Add progress subtasks
		progressBar := progress.New("Download Files", 3)
		sg.AddSubtask("Download", progressBar, func(component spinner.TaskComponent, sg *spinner.SpinGroup) error {
			if p, ok := component.(*progress.Progress); ok {
				for i := 0; i <= 3; i++ {
					p.Update(i, fmt.Sprintf("Downloaded %d files", i))
					time.Sleep(2 * time.Millisecond)
				}
			}
			mu.Lock()
			executionOrder = append(executionOrder, "download")
			mu.Unlock()
			return nil
		})

		return nil
	})

	err := sg.Run()
	require.NoError(t, err)
	require.Equal(t, []string{"setup", "download"}, executionOrder)
	require.Equal(t, 2, sg.TaskCount()) // Original 1 task + 1 progress subtask

	output := buf.String()
	require.Contains(t, output, "Download Files")
	require.Contains(t, output, "Downloaded")
}
