# Gooey

[![CI](https://github.com/pseudomuto/gooey/workflows/CI/badge.svg)](https://github.com/pseudomuto/gooey/actions)
[![codecov](https://codecov.io/gh/pseudomuto/gooey/branch/main/graph/badge.svg)](https://codecov.io/gh/pseudomuto/gooey)
[![Go Reference](https://pkg.go.dev/badge/github.com/pseudomuto/gooey.svg)](https://pkg.go.dev/github.com/pseudomuto/gooey)
[![Go Report Card](https://goreportcard.com/badge/github.com/pseudomuto/gooey)](https://goreportcard.com/report/github.com/pseudomuto/gooey)

A Go CLI UI library inspired by [Shopify's cli-ui](https://github.com/Shopify/cli-ui), providing beautiful terminal interfaces for command-line applications.

## Features

- **Frame Components**: Create bordered content areas with nested frame support
- **Progress Components**: Interactive progress bars with extensible renderers, adaptive width calculation, real-time updates, and seamless frame integration
- **Spinner Components**: Animated loading indicators with automatic color rotation (Red→Blue→Cyan→Magenta), multiple animation styles, and real-time message updates
- **SpinGroup Components**: Coordinated execution of multiple sequential tasks with mixed Spinner and Progress components using the TaskComponent interface
- **ANSI Color Support**: Rich color and styling with template-based formatting
- **Multiple Frame Styles**: Box and bracket frame styles
- **Automatic Formatting**: Smart content alignment and border management
- **Terminal Width Detection**: Responsive layouts that adapt to terminal size
- **Template Processing**: Enhanced syntax supporting `{{bold+cyan:text}}`, `{{check:text}}`, and `{{icon+color:text}}` combinations
- **Icon System**: Comprehensive icon sets for status, tasks, checklists, and spinners
- **Terminal Control**: Cursor movement, screen clearing, and visibility controls

## Installation

```bash
go get github.com/pseudomuto/gooey
```

## Quick Start

### Frame Example

```go
package main

import (
    "github.com/pseudomuto/gooey/ansi"
    "github.com/pseudomuto/gooey/frame"
)

func main() {
    // Create a frame with custom styling
    f := frame.Open("Deployment Status", frame.WithColor(ansi.Blue))
    f.Println("Starting deployment process...")
    f.Divider("Progress")
    f.Println("✅ Database migration completed")
    f.Println("✅ Services deployed successfully")
    f.Close() // Shows total elapsed time
}
```

### Progress Bar Example

```go
package main

import (
    "fmt"
    "time"
    "github.com/pseudomuto/gooey/ansi"
    "github.com/pseudomuto/gooey/progress"
)

func main() {
    // Create a progress bar with unknown total initially
    p := progress.New("Download", 0, // Total unknown at creation
        progress.WithColor(ansi.Green),
        progress.WithRenderer(progress.Bar))
    
    p.Start()
    
    // Simulate discovering the actual file size
    time.Sleep(200 * time.Millisecond)
    fileSize := 1024 * 1024 // 1MB discovered from HTTP headers
    p.SetTotal(fileSize)
    
    // Now update with actual progress
    for downloaded := 0; downloaded <= fileSize; downloaded += 102400 {
        percentage := float64(downloaded) / float64(fileSize) * 100
        p.Update(downloaded, fmt.Sprintf("Downloaded %.1f%% (%d bytes)", percentage, downloaded))
        time.Sleep(50 * time.Millisecond)
    }
    p.Complete("Download finished!")
}
```

### Spinner Example

```go
package main

import (
    "time"
    "github.com/pseudomuto/gooey/ansi"
    "github.com/pseudomuto/gooey/spinner"
)

func main() {
    // Create an animated spinner
    s := spinner.New("Connecting to server...",
        spinner.WithColor(ansi.Yellow),
        spinner.WithRenderer(spinner.Dots))
    
    // Start animation and simulate work
    s.Start()
    time.Sleep(3 * time.Second)
    s.UpdateMessage("Authentication successful")
    time.Sleep(1 * time.Second)
    s.Stop() // Shows green checkmark
}
```

### SpinGroup Example (Mixed Components)

```go
package main

import (
    "fmt"
    "time"
    "github.com/pseudomuto/gooey/ansi"
    "github.com/pseudomuto/gooey/progress"
    "github.com/pseudomuto/gooey/spinner"
)

func main() {
    sg := spinner.NewSpinGroup("Deployment Pipeline")
    
    // Add indefinite task with spinner
    sg.AddTask("Connect", 
        spinner.New("Connecting to server..."), 
        func(component spinner.TaskComponent) error {
            // Access spinner for dynamic updates
            if s, ok := component.(*spinner.Spinner); ok {
                time.Sleep(1 * time.Second)
                s.UpdateMessage("Authenticating...")
                time.Sleep(1 * time.Second)
            }
            return nil
        })
    
    // Add definite task with progress bar
    sg.AddTask("Download", progress.New("Download", 50), 
        func(component spinner.TaskComponent) error {
            // Access progress bar for updates
            if p, ok := component.(*progress.Progress); ok {
                for i := 0; i <= 50; i += 5 {
                    p.Update(i, fmt.Sprintf("Downloaded %d files", i))
                    time.Sleep(100 * time.Millisecond)
                }
            }
            return nil
        })
    
    // Run all tasks in a frame
    sg.RunInFrame()
}
```

See the [examples directory](./examples) for more comprehensive examples and advanced usage patterns.

## API Reference

### Frame Methods

- `frame.Open(title string, options ...FrameOption) *Frame` - Create and open a new frame
- `frame.Close()` - Close the current frame with timing information
- `frame.Print(format string, args ...any)` - Print formatted content without newline
- `frame.Println(format string, args ...any)` - Print formatted content with newline
- `frame.Divider(text string)` - Add a divider line with optional text
- `frame.ReplaceLine(format string, args ...any)` - Replace the last line with new content (enables single-line updates)

### Frame Options

- `frame.WithColor(color ansi.Color)` - Set frame border color
- `frame.WithStyle(style FrameStyle)` - Set frame style (Box or Bracket)
- `frame.WithOutput(w io.Writer)` - Set custom output writer

### Frame Styles

- `frame.Box` - Full box borders with complete enclosure
- `frame.Bracket` - Simple bracket-style markers

### Progress Methods

- `progress.New(title string, total int, options ...ProgressOption) *Progress` - Create and initialize a new progress bar
- `progress.Start()` - Begin showing the progress bar (TaskComponent interface method)
- `progress.Update(current int, message string)` - Update progress value and optional message
- `progress.Increment(message string)` - Increment progress by 1 with optional message
- `progress.Complete(message string)` - Mark progress as 100% complete with final message
- `progress.Fail(message string)` - Mark progress as failed with error message (TaskComponent interface method)

#### Progress Getters

- `progress.Current() int` - Get current progress value
- `progress.Total() int` - Get total progress value
- `progress.SetTotal(total int)` - Update the total progress value (useful when total is unknown at creation)
- `progress.Percentage() float64` - Get completion percentage
- `progress.IsCompleted() bool` - Check if progress is completed
- `progress.Elapsed() time.Duration` - Get elapsed time since creation
- `progress.Message() string` - Get current progress message
- `progress.Title() string` - Get progress bar title
- `progress.Color() ansi.Color` - Get progress bar color
- `progress.Width() int` - Get progress bar width
- `progress.AvailableWidth() int` - Get calculated available width for progress bar

### Progress Options

- `progress.WithColor(color ansi.Color)` - Set progress bar color
- `progress.WithRenderer(renderer ProgressRenderer)` - Set custom renderer for unlimited styling
- `progress.WithWidth(width int)` - Set progress bar width in characters
- `progress.WithOutput(w io.Writer)` - Set custom output writer

### Built-in Progress Renderers

- `progress.Bar` - Full progress bar with filled/empty characters (default)
- `progress.Minimal` - Minimal display with just percentage and message
- `progress.Dots` - Dot-based progress indicator

### Custom Progress Renderers

Implement the `ProgressRenderer` interface for custom styling or use `RenderFunc` for inline custom renderers.

### Spinner Methods

- `spinner.New(message string, options ...SpinnerOption) *Spinner` - Create and initialize a new animated spinner
- `spinner.Start()` - Start the spinner animation in a background goroutine
- `spinner.Stop()` - Stop the spinner animation and show successful completion with green checkmark
- `spinner.Complete(message string)` - Complete the spinner with optional success message (TaskComponent interface method)
- `spinner.Fail(message string)` - Stop the spinner animation and show failure with red crossmark and optional error message
- `spinner.UpdateMessage(message string)` - Update the spinner message while running
- `spinner.IsRunning() bool` - Check if the spinner is currently animating
- `spinner.Message() string` - Get the current spinner message
- `spinner.Color() ansi.Color` - Get the spinner's configured color
- `spinner.CurrentColor(frame int) ansi.Color` - Get the color for a specific animation frame (handles rotation)
- `spinner.Elapsed() time.Duration` - Get elapsed time since spinner started
- `spinner.ShowElapsed() bool` - Check if elapsed time will be shown on completion
- `spinner.State() SpinnerState` - Get the current completion state (SpinnerCompleted or SpinnerFailed)

### Spinner Options

- `spinner.WithColor(color ansi.Color)` - Set fixed color (overrides automatic rotation)
- `spinner.WithRenderer(renderer SpinnerRenderer)` - Set custom animation renderer
- `spinner.WithInterval(interval time.Duration)` - Set animation frame interval (default: 100ms)
- `spinner.WithOutput(w io.Writer)` - Set custom output writer
- `spinner.WithShowElapsed(show bool)` - Control whether elapsed time is displayed on completion (default: true)

### Built-in Spinner Renderers

- `spinner.Dots` - 8-frame braille spinner animation (⠋⠙⠹⠸⠼⠴⠦⠧)
- `spinner.Clock` - 4-frame subset for slower animation
- `spinner.Arrow` - Directional arrows rotating (→↓←↑)

### Custom Spinner Renderers

Implement the `SpinnerRenderer` interface for custom animations or use `RenderFunc` for inline custom renderers.

### TaskComponent Interface

The TaskComponent interface enables polymorphic task management, allowing SpinGroup to work with both Spinner and Progress components:

```go
type TaskComponent interface {
    Start()                    // Begin showing the component
    Complete(message string)   // Mark successful completion
    Fail(message string)       // Mark failure with error message
    SetOutput(io.Writer)       // Redirect output for frame integration
}
```

Both `*Spinner` and `*Progress` implement this interface, enabling mixed usage in SpinGroup.

### SpinGroup Methods

- `spinner.NewSpinGroup(title string, options ...SpinGroupOption) *SpinGroup` - Create a new spin group for sequential task execution using TaskComponent instances (Spinner or Progress)
- `spinGroup.AddTask(name string, component TaskComponent, taskFunc func(TaskComponent) error)` - Add a task with its associated component and function that receives the component for dynamic updates
- `spinGroup.Run() error` - Execute all tasks sequentially, returning first error encountered
- `spinGroup.RunInFrame() error` - Execute all tasks within a frame for organized display
- `spinGroup.TaskCount() int` - Get the number of tasks in the group
- `spinGroup.Title() string` - Get the spin group title

### SpinGroup Options

- `spinner.WithSpinGroupOutput(w io.Writer)` - Set custom output writer for the spin group


## Examples

Run the examples to see all features in action:

```bash
# Frame component examples
cd examples/frame
go run .

# Progress component examples
cd examples/progress
go run .

# Spinner component examples
cd examples/spinner
go run .

# SpinGroup component examples
cd examples/spingroup
go run .
```

The frame examples demonstrate:
- Basic frame usage
- Nested frames with color inheritance
- Different frame styles
- Dividers and formatting
- ANSI template processing
- Complex nested layouts
- Icon usage and status indicators
- Terminal control sequences

The progress examples demonstrate:
- Basic progress bar creation and updates
- Built-in progress renderers (Bar, Minimal, Dots)
- Custom renderers with ProgressRenderer interface and RenderFunc
- Custom colors and widths
- Increment and completion methods
- Real-time progress updates with messages  
- Seamless integration with frames using single-line updates
- Flexible three-section layout using proportional column system
- Advanced styling with emoji and time-based custom renderers

The spinner examples demonstrate:
- Basic spinner animations with automatic color rotation (Red→Blue→Cyan→Magenta)
- Built-in spinner renderers (Dots, Clock, Arrow)
- Custom renderers with SpinnerRenderer interface and RenderFunc
- Fixed color override with WithColor option
- Real-time message updates while spinning
- **Dual completion modes**: successful completion with green checkmark (✓) and failure with red crossmark (✗)
- **State management**: explicit tracking of completion status with SpinnerState
- Seamless integration with frames using single-line updates
- Custom animation intervals and timing control
- Thread-safe operations and proper resource cleanup

The SpinGroup examples demonstrate:
- Sequential task execution using mixed TaskComponent instances (Spinner and Progress)
- **Mixed Component Usage**: Combine indefinite tasks (Spinners) with definite tasks (Progress bars) in the same workflow
- **Definite vs Indefinite Tasks**: Use Progress bars for tasks with known steps/completion and Spinners for unpredictable operations
- Custom component configurations for each task (colors, renderers, intervals, widths)
- Error handling with automatic failure detection, visual feedback (red ✗), and early termination
- Frame integration for organized display with nested frames
- **Success and Failure States**: Visual indicators show green checkmarks (✓) for success and red crossmarks (✗) for failures
- Polymorphic API with `Run()` and `RunInFrame()` methods that work with any TaskComponent
- Thread-safe task addition and execution

## Architecture

### Core Packages

- **`ansi`** - ANSI color codes, styles, template formatting, icons, and terminal control sequences
- **`frame`** - Frame component for bordered content areas with nested frame support
- **`progress`** - Progress component for interactive progress bars with extensible renderers
- **`spinner`** - Spinner component for animated loading indicators and sequential task management

### Design Principles

- **io.Writer Interface**: All components implement standard Go interfaces
- **Functional Options**: Flexible configuration using the options pattern
- **Renderer Pattern**: Extensible styling through interface-based renderers
- **Template Processing**: Rich text formatting with `{{style+color:text}}` syntax
- **Responsive Design**: Automatic adaptation to terminal width
- **Color Inheritance**: Nested components inherit parent colors appropriately
- **Single-Line Updates**: Progress components update in place within frames for clean real-time feedback
- **Flexible Layout System**: Multi-column layouts with proportional width allocation and minimum constraints

## Development

### Prerequisites

- Go 1.21 or later (requires `min()` and `max()` built-in functions)
- [Task](https://taskfile.dev/) for build automation

### Commands

```bash
# Run tests
task test

# Run linter
task lint

# Build project
go build

# Run specific test
go test ./frame -run TestFrameBasic

# Tidy dependencies
go mod tidy
```

### Testing

Tests use buffer-based output verification to capture and validate terminal output.

## Contributing

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Follow existing patterns**:
   - Use the `ansi` package for all formatting
   - Implement `io.Writer` for output components
   - Use functional options for configuration
   - Add comprehensive tests with buffer verification
4. **Run tests and linting**: `task test && task lint`
5. **Commit changes**: `git commit -m 'Add amazing feature'`
6. **Push to branch**: `git push origin feature/amazing-feature`
7. **Open a Pull Request**

### Code Style

- Follow standard Go conventions
- Use dot imports for tests: `import . "github.com/pseudomuto/gooey/ansi"`
- Test both visual output and behavior
- Document public APIs with godoc comments
- Maintain consistency with existing component patterns

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Inspiration

This library is inspired by [Shopify's cli-ui](https://github.com/Shopify/cli-ui) Ruby gem, adapted for Go with idiomatic patterns and enhanced terminal capabilities.

