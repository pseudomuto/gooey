package spinner_test

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/pseudomuto/gooey/ansi"
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

	sg.AddTask("Task 1", spinner.New("Task 1 message"), func() error {
		return nil
	})

	require.Equal(t, 1, sg.TaskCount())

	// Add another task
	sg.AddTask("Task 2", spinner.New("Task 2 message"), func() error {
		return nil
	})

	require.Equal(t, 2, sg.TaskCount())
}

func TestSpinGroup_Run(t *testing.T) {
	buf := &bytes.Buffer{}
	sg := spinner.NewSpinGroup("Test Group", spinner.WithSpinGroupOutput(buf))

	var executed bool
	sg.AddTask("Task 1", spinner.New("Executing..."), func() error {
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
	sg.AddTask("Task 1", spinner.New("Task 1 executing..."), func() error {
		mu.Lock()
		executionOrder = append(executionOrder, 1)
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	sg.AddTask("Task 2", spinner.New("Task 2 executing..."), func() error {
		mu.Lock()
		executionOrder = append(executionOrder, 2)
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	sg.AddTask("Task 3", spinner.New("Task 3 executing..."), func() error {
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
	sg.AddTask("Task 1", spinner.New("Task 1 executing..."), func() error {
		mu.Lock()
		executionOrder = append(executionOrder, 1)
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	sg.AddTask("Task 2", spinner.New("Task 2 executing..."), func() error {
		mu.Lock()
		executionOrder = append(executionOrder, 2)
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		return errors.New("task failed")
	})

	sg.AddTask("Task 3", spinner.New("Task 3 executing..."), func() error {
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

	sg.AddTask("Frame Task", spinner.New("Running in frame..."), func() error {
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

	sg.AddTask("Custom Task", customSpinner, func() error {
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
			sg.AddTask("Task", spinner.New("Running..."), func() error {
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
