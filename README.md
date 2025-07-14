# Gooey

A Go CLI UI library inspired by [Shopify's cli-ui](https://github.com/Shopify/cli-ui), providing beautiful terminal interfaces for command-line applications.

## Features

- **Frame Components**: Create bordered content areas with nested frame support
- **Progress Components**: Interactive progress bars with extensible renderers, adaptive width calculation, real-time updates, and seamless frame integration
- **Spinner Components**: Animated loading indicators with automatic color rotation (Red→Blue→Cyan→Magenta), multiple animation styles, and real-time message updates
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

### Basic Frame Usage

```go
package main

import (
    "github.com/pseudomuto/gooey/ansi"
    "github.com/pseudomuto/gooey/components/frame"
)

func main() {
    // Create a basic frame
    f := frame.Open("My Application")
    f.Println("Welcome to my CLI tool!")
    f.Println("This content is automatically formatted")
    f.Close()
}
```

### Nested Frames with Colors

```go
// Outer frame
outer := frame.Open("Deployment", frame.WithColor(ansi.Blue))
outer.Println("Starting deployment process...")

// Inner frame for specific task
inner := frame.Open("Database Migration", frame.WithColor(ansi.Green))
inner.Println("Migrating database schema...")
inner.Println("Migration completed successfully")
inner.Close()

outer.Println("Deployment completed!")
outer.Close()
```

### Frame Styles and Dividers

```go
// Box style (default)
boxFrame := frame.Open("Box Style", frame.WithStyle(frame.Box))
boxFrame.Println("This uses full box borders")
boxFrame.Divider("Section Break")
boxFrame.Println("Content after divider")
boxFrame.Close()

// Bracket style
bracketFrame := frame.Open("Bracket Style", frame.WithStyle(frame.Bracket))
bracketFrame.Println("This uses simple bracket markers")
bracketFrame.Close()
```

### Progress Bars

```go
package main

import (
    "fmt"
    "io"
    "time"
    
    "github.com/pseudomuto/gooey/ansi"
    "github.com/pseudomuto/gooey/components/progress"
    "github.com/pseudomuto/gooey/components/frame"
)

func main() {
    // Basic progress bar
    p := progress.New("Downloading files", 100)
    for i := 0; i <= 100; i += 10 {
        p.Update(i, fmt.Sprintf("Downloaded %d files", i))
        time.Sleep(100 * time.Millisecond)
    }
    p.Complete("All files downloaded!")

    // Different built-in styles
    p1 := progress.New("Processing", 50, 
        progress.WithRenderer(progress.Bar), 
        progress.WithColor(ansi.Green))
        
    p2 := progress.New("Uploading", 25, 
        progress.WithRenderer(progress.Minimal), 
        progress.WithColor(ansi.Blue))
        
    p3 := progress.New("Installing", 20, 
        progress.WithRenderer(progress.Dots), 
        progress.WithColor(ansi.Yellow),
        progress.WithWidth(30))

    // Custom renderer with function
    p4 := progress.New("Processing", 10,
        progress.WithRenderer(progress.RenderFunc(func(p *progress.Progress, w io.Writer) {
            percentage := p.Percentage()
            if percentage < 50 {
                fmt.Fprintf(w, "%s: 🔴 %.1f%%", p.Title(), percentage)
            } else {
                fmt.Fprintf(w, "%s: 🟢 %.1f%%", p.Title(), percentage)
            }
        })))

    // Progress bars work seamlessly within frames
    f := frame.Open("Deployment", frame.WithColor(ansi.Cyan))
    p5 := progress.New("Building", 8, 
        progress.WithColor(ansi.Green),
        progress.WithOutput(f))
    
    for i := 0; i <= 8; i++ {
        p5.Update(i, fmt.Sprintf("Step %d", i))
        time.Sleep(200 * time.Millisecond)
    }
    p5.Complete("Build completed!")
    f.Close()
}
```

### Spinner Animations

```go
package main

import (
    "fmt"
    "io"
    "time"
    
    "github.com/pseudomuto/gooey/ansi"
    "github.com/pseudomuto/gooey/components/frame"
    "github.com/pseudomuto/gooey/components/spinner"
)

func main() {
    // Basic spinner with automatic color rotation (Red→Blue→Cyan→Magenta)
    s := spinner.New("Loading data...")
    s.Start()
    time.Sleep(3 * time.Second)
    s.Stop()

    // Custom spinner with fixed color and no elapsed time display
    s2 := spinner.New("Processing files...",
        spinner.WithColor(ansi.Green),
        spinner.WithRenderer(spinner.Clock),
        spinner.WithInterval(200*time.Millisecond),
        spinner.WithShowElapsed(false)) // Disable elapsed time display
    
    s2.Start()
    time.Sleep(2 * time.Second)
    
    s2.UpdateMessage("Almost done...")
    time.Sleep(1 * time.Second)
    
    s2.Stop()

    // Different built-in animation styles
    s3 := spinner.New("Dots animation", 
        spinner.WithRenderer(spinner.Dots))    // 8-frame braille spinner
        
    s4 := spinner.New("Clock animation", 
        spinner.WithRenderer(spinner.Clock))   // 4-frame subset
        
    s5 := spinner.New("Arrow animation", 
        spinner.WithRenderer(spinner.Arrow))   // Directional arrows

    // Custom renderer with function
    s6 := spinner.New("Custom animation",
        spinner.WithRenderer(spinner.RenderFunc(func(s *spinner.Spinner, frame int, w io.Writer) {
            icons := []string{"◐", "◓", "◑", "◒"}
            color := s.CurrentColor(frame) // Access automatic color rotation
            icon := icons[frame%len(icons)]
            fmt.Fprintf(w, "%s %s", ansi.Icon(icon).Colorize(color), s.Message())
        })))

    // Spinners work seamlessly within frames
    f := frame.Open("Deployment", frame.WithColor(ansi.Cyan))
    s7 := spinner.New("Building application...", 
        spinner.WithOutput(f))
    
    s7.Start()
    time.Sleep(2 * time.Second)
    s7.UpdateMessage("Finalizing...")
    time.Sleep(1 * time.Second)
    s7.Stop()
    
    f.Println("Build completed!")
    f.Close()
}
```

### ANSI Template Formatting

```go
// Using ansi.Formatter for template processing
fmtr := ansi.NewFormatter(os.Stdout)
f := frame.Open("{{rocket:}} Deployment", frame.WithColor(ansi.Blue), frame.WithOutput(fmtr))
f.Println("Processing {{bold+cyan:important}} data...")
f.Println("Status: {{green:SUCCESS}}")
f.Println("Result: {{check:}} Task completed")
f.Close()

// One-off formatting without creating a formatter
formatted := ansi.Format("{{bold+red:Error}}: Operation failed")
fmt.Println(formatted)

// Quick colorization with sprintf-style formatting
text := ansi.Colorize("Progress: {{green:%d%%}} complete", 75)
fmt.Println(text)
```

### Icons and Status Indicators

```go
// Using predefined icons
fmt.Println(ansi.CheckMark.Colorize(ansi.Green), "Task completed")
fmt.Println(ansi.Warning.Colorize(ansi.Yellow), "Warning message")

// Using icon sets
taskIcon := ansi.GetTaskIcon("completed")
fmt.Printf("%s Task finished\n", taskIcon.Colorize(ansi.Green))

// Template-based icon usage
fmt.Println(ansi.Format("{{check:}} Success"))
fmt.Println(ansi.Format("{{warning:}} Warning"))
fmt.Println(ansi.Format("{{error:}} Failed"))
fmt.Println(ansi.Format("{{rocket:}} Launching..."))

// Terminal control
fmt.Print(ansi.ClearScreen)
fmt.Print(ansi.MoveCursor(10, 5))
fmt.Println("Text at specific position")
```

### Flexible Layout System

The `term.SectionLayout` provides a powerful system for creating multi-column layouts with proportional widths:

```go
import "github.com/pseudomuto/gooey/term"

// Create a 3-column layout with 20%, 60%, 20% proportions
layout := term.NewSectionLayout(100, 1, 3, 1)
widths := layout.SectionWidths() // Returns [20, 60, 20]

// With minimum width constraints
layout = term.NewSectionLayout(100, 1, 3, 1).WithMinWidths(10, 20, 8)
widths = layout.SectionWidths() // Respects minimums

// 2-column layout
layout = term.NewSectionLayout(100, 2, 3) // 40:60 split

// 4-column equal layout
layout = term.NewSectionLayout(100, 1, 1, 1, 1) // 25% each

// Float weights work too
layout = term.NewSectionLayout(100, 0.5, 1.5, 0.5) // 20:60:20

// Text formatting utilities
formatted := term.TruncateAndPad("Hello", 10) // "Hello     "
formatted = term.TruncateAndPad("Hello World", 8) // "Hello..."
```

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
- `progress.Update(current int, message string)` - Update progress value and optional message
- `progress.Increment(message string)` - Increment progress by 1 with optional message
- `progress.Complete(message string)` - Mark progress as 100% complete with final message

#### Progress Getters

- `progress.Current() int` - Get current progress value
- `progress.Total() int` - Get total progress value
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

```go
// Implement the ProgressRenderer interface
type customRenderer struct{}

func (r *customRenderer) Render(p *progress.Progress, w io.Writer) {
    fmt.Fprintf(w, "Custom: %s %.1f%%", p.Title(), p.Percentage())
}

// Use with WithRenderer option
p := progress.New("Task", 100, progress.WithRenderer(&customRenderer{}))

// Or use RenderFunc for inline custom renderers
p := progress.New("Task", 100, 
    progress.WithRenderer(progress.RenderFunc(func(p *progress.Progress, w io.Writer) {
        // Custom rendering logic here
    })))
```

### Spinner Methods

- `spinner.New(message string, options ...SpinnerOption) *Spinner` - Create and initialize a new animated spinner
- `spinner.Start()` - Start the spinner animation in a background goroutine
- `spinner.Stop()` - Stop the spinner animation and show completion with elapsed time
- `spinner.UpdateMessage(message string)` - Update the spinner message while running
- `spinner.IsRunning() bool` - Check if the spinner is currently animating
- `spinner.Message() string` - Get the current spinner message
- `spinner.Color() ansi.Color` - Get the spinner's configured color
- `spinner.CurrentColor(frame int) ansi.Color` - Get the color for a specific animation frame (handles rotation)
- `spinner.Elapsed() time.Duration` - Get elapsed time since spinner started
- `spinner.ShowElapsed() bool` - Check if elapsed time will be shown on completion

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

```go
// Implement the SpinnerRenderer interface
type customRenderer struct{}

func (r *customRenderer) Render(s *spinner.Spinner, frame int, w io.Writer) {
    color := s.CurrentColor(frame) // Access automatic color rotation
    fmt.Fprintf(w, "Custom: %s %s", color.Colorize("●"), s.Message())
}

// Use with WithRenderer option
s := spinner.New("Task", spinner.WithRenderer(&customRenderer{}))

// Or use RenderFunc for inline custom renderers
s := spinner.New("Task", 
    spinner.WithRenderer(spinner.RenderFunc(func(s *spinner.Spinner, frame int, w io.Writer) {
        // Custom rendering logic here
        // s.CurrentColor(frame) provides automatic color rotation
    })))
```

### Layout System API

#### SectionLayout

- `term.NewSectionLayout(totalWidth int, weights ...float64) SectionLayout` - Create layout with proportional column weights
- `layout.WithMinWidths(minWidths ...int) SectionLayout` - Set minimum width constraints per column
- `layout.SectionWidths() []int` - Calculate actual column widths with smart scaling

#### Text Utilities

- `term.TruncateAndPad(text string, maxWidth int) string` - Truncate text with "..." and pad to exact width
- `term.PrintableWidth(text string) int` - Get display width excluding ANSI escape sequences
- `term.TruncateString(text string, maxWidth int) string` - Truncate string while preserving ANSI formatting
- `term.StripCodes(text string) string` - Remove ANSI escape sequences

#### Mathematical Utilities

- `term.Max(a, b int) int` - Return the larger of two integers
- `term.Min(a, b int) int` - Return the smaller of two integers

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
- Seamless integration with frames using single-line updates
- Custom animation intervals and timing control
- Thread-safe operations and proper resource cleanup

## Architecture

### Core Packages

- **`ansi`** - ANSI color codes, styles, template formatting, icons, and terminal control sequences
- **`components`** - Reusable UI components (Frame, Progress, Spinner)
- **`term`** - Terminal utilities, width detection, flexible layout system, and mathematical functions

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

- Go 1.24.4 or later
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
go test ./components/frame -run TestFrameBasic

# Tidy dependencies
go mod tidy
```

### Testing

Tests use buffer-based output verification:

```go
func TestFrame(t *testing.T) {
    var buf bytes.Buffer
    frame := frame.Open("Test", frame.WithOutput(&buf))
    frame.Println("test content")
    frame.Close()

    output := buf.String()
    require.Contains(t, output, "test content")
}
```

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

