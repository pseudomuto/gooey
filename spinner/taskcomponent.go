package spinner

import "io"

// TaskComponent represents a component that can show progress for a task.
// Both Spinner and Progress components implement this interface, allowing
// SpinGroup to work with either type of component seamlessly.
//
// The interface provides a consistent API for task visualization:
// - Start() begins the visual indication (spinners animate, progress shows)
// - Complete() marks successful completion with optional message
// - Fail() marks failure with optional error message
// - SetOutput() allows output redirection for frame integration
type TaskComponent interface {
	// Start begins showing the component. For spinners, this starts animation.
	// For progress bars, this could be a no-op since they show immediately.
	Start()

	// Complete marks the task as successfully finished with an optional message.
	// Spinners show a green checkmark, progress bars show 100% completion.
	Complete(message string)

	// Fail marks the task as failed with an optional error message.
	// Spinners show a red crossmark, progress bars show error state.
	Fail(message string)

	// SetOutput allows redirecting output for frame integration.
	// This enables components to render within frames or custom writers.
	SetOutput(output io.Writer)
}
