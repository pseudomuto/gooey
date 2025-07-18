package spinner_test

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/pseudomuto/gooey/components/frame"
	"github.com/pseudomuto/gooey/components/spinner"
	"github.com/stretchr/testify/require"
)

func TestTaskStatus_String(t *testing.T) {
	tests := []struct {
		name     string
		status   spinner.TaskStatus
		expected string
	}{
		{
			name:     "TaskPending",
			status:   spinner.TaskPending,
			expected: "pending",
		},
		{
			name:     "TaskRunning",
			status:   spinner.TaskRunning,
			expected: "running",
		},
		{
			name:     "TaskCompleted",
			status:   spinner.TaskCompleted,
			expected: "completed",
		},
		{
			name:     "TaskFailed",
			status:   spinner.TaskFailed,
			expected: "failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.String()
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestNewSpinGroup(t *testing.T) {
	sg := spinner.NewSpinGroup("Test Group")

	require.NotNil(t, sg)
	require.Equal(t, 0, sg.TaskCount())
	require.False(t, sg.IsRunning())
}

func TestNewSpinGroupWithOptions(t *testing.T) {
	buf := &bytes.Buffer{}
	sg := spinner.NewSpinGroup("Test Group", spinner.WithSpinGroupOutput(buf))

	require.NotNil(t, sg)
	require.Equal(t, 0, sg.TaskCount())
	require.False(t, sg.IsRunning())
}

func TestSpinGroup_AddTask(t *testing.T) {
	sg := spinner.NewSpinGroup("Test Group")

	taskID := sg.AddTask("Task 1", func() error {
		return nil
	})

	require.Equal(t, 0, taskID)
	require.Equal(t, 1, sg.TaskCount())

	// Add another task
	taskID2 := sg.AddTask("Task 2", func() error {
		return nil
	})

	require.Equal(t, 1, taskID2)
	require.Equal(t, 2, sg.TaskCount())
}

func TestSpinGroup_StartStop(t *testing.T) {
	sg := spinner.NewSpinGroup("Test Group")

	require.False(t, sg.IsRunning())

	sg.Start()
	require.True(t, sg.IsRunning())

	sg.Stop()
	require.False(t, sg.IsRunning())
}

func TestSpinGroup_DoubleStart(t *testing.T) {
	sg := spinner.NewSpinGroup("Test Group")

	sg.Start()
	require.True(t, sg.IsRunning())

	// Second start should be a no-op
	sg.Start()
	require.True(t, sg.IsRunning())

	sg.Stop()
}

func TestSpinGroup_DoubleStop(t *testing.T) {
	sg := spinner.NewSpinGroup("Test Group")

	sg.Start()
	require.True(t, sg.IsRunning())

	sg.Stop()
	require.False(t, sg.IsRunning())

	// Second stop should be a no-op
	sg.Stop()
	require.False(t, sg.IsRunning())
}

func TestSpinGroup_ExecuteTasksSequentially(t *testing.T) {
	buf := &bytes.Buffer{}
	sg := spinner.NewSpinGroup("Test Group", spinner.WithSpinGroupOutput(buf))

	var executionOrder []int
	var mu sync.Mutex

	// Add tasks that record execution order
	sg.AddTask("Task 1", func() error {
		mu.Lock()
		executionOrder = append(executionOrder, 1)
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	sg.AddTask("Task 2", func() error {
		mu.Lock()
		executionOrder = append(executionOrder, 2)
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	sg.AddTask("Task 3", func() error {
		mu.Lock()
		executionOrder = append(executionOrder, 3)
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	sg.Start()
	sg.Wait()

	// After all tasks complete, SpinGroup should still be running waiting for more tasks
	require.True(t, sg.IsRunning())
	require.Equal(t, []int{1, 2, 3}, executionOrder)

	// Stop the SpinGroup
	sg.Stop()
	require.False(t, sg.IsRunning())
}

func TestSpinGroup_StopOnFirstFailure(t *testing.T) {
	buf := &bytes.Buffer{}
	sg := spinner.NewSpinGroup("Test Group", spinner.WithSpinGroupOutput(buf))

	var executionOrder []int
	var mu sync.Mutex

	// Add tasks where second task fails
	sg.AddTask("Task 1", func() error {
		mu.Lock()
		executionOrder = append(executionOrder, 1)
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	sg.AddTask("Task 2", func() error {
		mu.Lock()
		executionOrder = append(executionOrder, 2)
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		return errors.New("task failed")
	})

	sg.AddTask("Task 3", func() error {
		mu.Lock()
		executionOrder = append(executionOrder, 3)
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	sg.Start()
	sg.Wait()

	// Give a small amount of time for cleanup after failure
	time.Sleep(5 * time.Millisecond)

	// When a task fails, SpinGroup should stop automatically
	require.False(t, sg.IsRunning())
	// Task 3 should not execute because task 2 failed
	require.Equal(t, []int{1, 2}, executionOrder)
}

func TestSpinGroup_Wait(t *testing.T) {
	buf := &bytes.Buffer{}
	sg := spinner.NewSpinGroup("Test Group", spinner.WithSpinGroupOutput(buf))

	var completed bool

	sg.AddTask("Task 1", func() error {
		time.Sleep(50 * time.Millisecond)
		completed = true
		return nil
	})

	sg.Start()

	start := time.Now()
	sg.Wait()
	elapsed := time.Since(start)

	require.True(t, completed)
	require.GreaterOrEqual(t, elapsed, 50*time.Millisecond)
	// After all tasks complete, SpinGroup should still be running waiting for more tasks
	require.True(t, sg.IsRunning())

	// Stop the SpinGroup
	sg.Stop()
	require.False(t, sg.IsRunning())
}

func TestSpinGroup_Elapsed(t *testing.T) {
	sg := spinner.NewSpinGroup("Test Group")

	// Elapsed should be > 0 after some time
	sg.Start()
	time.Sleep(10 * time.Millisecond)
	elapsed := sg.Elapsed()
	sg.Stop()

	require.Greater(t, elapsed, time.Duration(0))
	require.Greater(t, elapsed, 5*time.Millisecond)
}

func TestSpinGroup_DynamicTaskAddition(t *testing.T) {
	buf := &bytes.Buffer{}
	sg := spinner.NewSpinGroup("Test Group", spinner.WithSpinGroupOutput(buf))

	var executionOrder []int
	var mu sync.Mutex

	// Add initial task
	sg.AddTask("Task 1", func() error {
		mu.Lock()
		executionOrder = append(executionOrder, 1)
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	sg.Start()

	// Add task after starting
	time.Sleep(5 * time.Millisecond)
	sg.AddTask("Task 2", func() error {
		mu.Lock()
		executionOrder = append(executionOrder, 2)
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	sg.Wait()

	// After all tasks complete, SpinGroup should still be running waiting for more tasks
	require.True(t, sg.IsRunning())
	require.Equal(t, []int{1, 2}, executionOrder)

	// Stop the SpinGroup
	sg.Stop()
	require.False(t, sg.IsRunning())
}

func TestSpinGroup_OutputStandalone(t *testing.T) {
	buf := &bytes.Buffer{}
	sg := spinner.NewSpinGroup("Test Group", spinner.WithSpinGroupOutput(buf))

	sg.AddTask("Test Task", func() error {
		time.Sleep(20 * time.Millisecond)
		return nil
	})

	sg.Start()
	sg.Wait()

	output := buf.String()
	require.NotEmpty(t, output)
	require.Contains(t, output, "Test Task")
}

func TestSpinGroup_OutputWithFrame(t *testing.T) {
	buf := &bytes.Buffer{}
	frameWriter := frame.Open("Test Frame", frame.WithOutput(buf))
	sg := spinner.NewSpinGroup("Test Group", spinner.WithSpinGroupOutput(frameWriter))

	sg.AddTask("Test Task", func() error {
		time.Sleep(20 * time.Millisecond)
		return nil
	})

	sg.Start()
	sg.Wait()
	frameWriter.Close()

	output := buf.String()
	require.NotEmpty(t, output)
	require.Contains(t, output, "Test Task")
}

func TestSpinGroup_TaskCancellation(t *testing.T) {
	buf := &bytes.Buffer{}
	sg := spinner.NewSpinGroup("Test Group", spinner.WithSpinGroupOutput(buf))

	var taskStarted bool
	var taskCompleted bool

	sg.AddTask("Long Task", func() error {
		taskStarted = true
		time.Sleep(100 * time.Millisecond)
		taskCompleted = true
		return nil
	})

	sg.Start()

	// Wait for task to start
	time.Sleep(10 * time.Millisecond)
	require.True(t, taskStarted)

	// Stop the spin group (should cancel the task)
	sg.Stop()

	// Wait a bit more
	time.Sleep(20 * time.Millisecond)

	require.False(t, sg.IsRunning())
	// Task should have been cancelled, so it shouldn't complete
	require.False(t, taskCompleted)
}

func TestSpinGroup_FormatTaskLine(t *testing.T) {
	buf := &bytes.Buffer{}
	sg := spinner.NewSpinGroup("Test Group", spinner.WithSpinGroupOutput(buf))

	sg.AddTask("Format Test", func() error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	sg.Start()
	sg.Wait()

	output := buf.String()
	require.NotEmpty(t, output)
	require.Contains(t, output, "Format Test")
}

func TestSpinGroup_ErrorHandling(t *testing.T) {
	buf := &bytes.Buffer{}
	sg := spinner.NewSpinGroup("Test Group", spinner.WithSpinGroupOutput(buf))

	testError := errors.New("test error")

	sg.AddTask("Failing Task", func() error {
		time.Sleep(10 * time.Millisecond)
		return testError
	})

	sg.Start()
	sg.Wait()

	// Give a small amount of time for cleanup after failure
	time.Sleep(5 * time.Millisecond)

	// When a task fails, SpinGroup should stop automatically
	require.False(t, sg.IsRunning())
	output := buf.String()
	require.NotEmpty(t, output)
	require.Contains(t, output, "Failing Task")
}

func TestSpinGroup_MultipleTasksWithMixedResults(t *testing.T) {
	buf := &bytes.Buffer{}
	sg := spinner.NewSpinGroup("Test Group", spinner.WithSpinGroupOutput(buf))

	var results []string
	var mu sync.Mutex

	sg.AddTask("Success Task", func() error {
		mu.Lock()
		results = append(results, "success")
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	sg.AddTask("Failure Task", func() error {
		mu.Lock()
		results = append(results, "failure")
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		return errors.New("failure")
	})

	sg.AddTask("Never Executed", func() error {
		mu.Lock()
		results = append(results, "never")
		mu.Unlock()
		return nil
	})

	sg.Start()
	sg.Wait()

	// Give a small amount of time for cleanup after failure
	time.Sleep(5 * time.Millisecond)

	// When a task fails, SpinGroup should stop automatically
	require.False(t, sg.IsRunning())
	require.Equal(t, []string{"success", "failure"}, results)
}

func TestWithSpinGroupOutput(t *testing.T) {
	buf := &bytes.Buffer{}

	sg := spinner.NewSpinGroup("Test Group", spinner.WithSpinGroupOutput(buf))

	sg.AddTask("Output Test", func() error {
		return nil
	})

	sg.Start()
	sg.Wait()

	output := buf.String()
	require.NotEmpty(t, output)
	require.Contains(t, output, "Output Test")
}

func TestWithMaxConcurrent(t *testing.T) {
	// This function is deprecated but should not panic
	sg := spinner.NewSpinGroup("Test Group", spinner.WithMaxConcurrent(5))

	require.NotNil(t, sg)
	require.Equal(t, 0, sg.TaskCount())
}

func TestSpinGroup_ConcurrentAccess(t *testing.T) {
	buf := &bytes.Buffer{}
	sg := spinner.NewSpinGroup("Test Group", spinner.WithSpinGroupOutput(buf))

	var wg sync.WaitGroup

	// Add tasks concurrently
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			sg.AddTask(fmt.Sprintf("Task %d", id), func() error {
				time.Sleep(5 * time.Millisecond)
				return nil
			})
		}(i)
	}

	wg.Wait()

	require.Equal(t, 5, sg.TaskCount())

	sg.Start()
	sg.Wait()

	// After all tasks complete, SpinGroup should still be running waiting for more tasks
	require.True(t, sg.IsRunning())

	// Stop the SpinGroup
	sg.Stop()
	require.False(t, sg.IsRunning())
}

func TestSpinGroup_NoTasksExecution(t *testing.T) {
	buf := &bytes.Buffer{}
	sg := spinner.NewSpinGroup("Test Group", spinner.WithSpinGroupOutput(buf))

	// Start without any tasks
	sg.Start()

	// Should complete quickly since there are no tasks
	done := make(chan struct{})
	go func() {
		sg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Good - completed as expected
	case <-time.After(1 * time.Second):
		t.Fatal("Wait() took too long with no tasks")
	}

	// For no tasks, the SpinGroup should still be running because it's waiting for tasks
	// This is the expected behavior for dynamic task addition
	require.True(t, sg.IsRunning())
	require.Equal(t, 0, sg.TaskCount())

	// Clean up
	sg.Stop()
}

func TestSpinGroup_RapidStartStop(t *testing.T) {
	buf := &bytes.Buffer{}
	sg := spinner.NewSpinGroup("Test Group", spinner.WithSpinGroupOutput(buf))

	sg.AddTask("Quick Task", func() error {
		time.Sleep(50 * time.Millisecond)
		return nil
	})

	// Start and immediately stop
	sg.Start()
	sg.Stop()

	require.False(t, sg.IsRunning())
}

func TestSpinGroup_StatusTransitions(t *testing.T) {
	buf := &bytes.Buffer{}
	sg := spinner.NewSpinGroup("Test Group", spinner.WithSpinGroupOutput(buf))

	var states []string
	var mu sync.Mutex

	sg.AddTask("State Task", func() error {
		mu.Lock()
		states = append(states, "executing")
		mu.Unlock()
		time.Sleep(20 * time.Millisecond)
		return nil
	})

	sg.Start()

	// Check running state
	require.True(t, sg.IsRunning())

	sg.Wait()

	// After all tasks complete, SpinGroup should still be running waiting for more tasks
	require.True(t, sg.IsRunning())
	require.Equal(t, []string{"executing"}, states)

	// Stop the SpinGroup
	sg.Stop()
	require.False(t, sg.IsRunning())
}
