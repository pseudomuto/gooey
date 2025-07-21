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

	sg.AddTask("Task 1", spinner.New("Task 1 message"), func(spinner.TaskComponent) error {
		return nil
	})

	require.Equal(t, 1, sg.TaskCount())

	// Add another task
	sg.AddTask("Task 2", spinner.New("Task 2 message"), func(spinner.TaskComponent) error {
		return nil
	})

	require.Equal(t, 2, sg.TaskCount())
}

func TestSpinGroup_Run(t *testing.T) {
	buf := &bytes.Buffer{}
	sg := spinner.NewSpinGroup("Test Group", spinner.WithSpinGroupOutput(buf))

	var executed bool
	sg.AddTask("Task 1", spinner.New("Executing..."), func(spinner.TaskComponent) error {
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
	sg.AddTask("Task 1", spinner.New("Task 1 executing..."), func(spinner.TaskComponent) error {
		mu.Lock()
		executionOrder = append(executionOrder, 1)
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	sg.AddTask("Task 2", spinner.New("Task 2 executing..."), func(spinner.TaskComponent) error {
		mu.Lock()
		executionOrder = append(executionOrder, 2)
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	sg.AddTask("Task 3", spinner.New("Task 3 executing..."), func(spinner.TaskComponent) error {
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
	sg.AddTask("Task 1", spinner.New("Task 1 executing..."), func(spinner.TaskComponent) error {
		mu.Lock()
		executionOrder = append(executionOrder, 1)
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	sg.AddTask("Task 2", spinner.New("Task 2 executing..."), func(spinner.TaskComponent) error {
		mu.Lock()
		executionOrder = append(executionOrder, 2)
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		return errors.New("task failed")
	})

	sg.AddTask("Task 3", spinner.New("Task 3 executing..."), func(spinner.TaskComponent) error {
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

	sg.AddTask("Frame Task", spinner.New("Running in frame..."), func(spinner.TaskComponent) error {
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

	sg.AddTask("Custom Task", customSpinner, func(spinner.TaskComponent) error {
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
			sg.AddTask("Task", spinner.New("Running..."), func(spinner.TaskComponent) error {
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
	sg.AddTask("Connect", spinner.New("Connecting to server..."), func(spinner.TaskComponent) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	// Add definite task with progress
	progressBar := progress.New("Download", 3)
	sg.AddTask("Download", progressBar, func(spinner.TaskComponent) error {
		for i := 0; i <= 3; i++ {
			progressBar.Update(i, fmt.Sprintf("Downloaded %d files", i))
			time.Sleep(5 * time.Millisecond)
		}
		return nil
	})

	// Add another indefinite task
	sg.AddTask("Cleanup", spinner.New("Cleaning up..."), func(spinner.TaskComponent) error {
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
	sg.AddTask("Connect", spinner.New("Connecting..."), func(spinner.TaskComponent) error {
		time.Sleep(5 * time.Millisecond)
		return nil
	})

	// Failed task with progress
	progressBar := progress.New("Upload", 10)
	sg.AddTask("Upload", progressBar, func(spinner.TaskComponent) error {
		progressBar.Update(5, "Uploading...")
		time.Sleep(5 * time.Millisecond)
		return errors.New("upload failed")
	})

	// This task should not execute
	sg.AddTask("Finalize", spinner.New("Finalizing..."), func(spinner.TaskComponent) error {
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
