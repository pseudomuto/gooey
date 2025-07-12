# Gooey

A Go CLI UI library inspired by [Shopify's cli-ui](https://github.com/Shopify/cli-ui), providing beautiful terminal interfaces for command-line applications.

## Features

- **Frame Components**: Create bordered content areas with nested frame support
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

## API Reference

### Frame Methods

- `frame.Open(title string, options ...FrameOption) *Frame` - Create and open a new frame
- `frame.Close()` - Close the current frame with timing information
- `frame.Print(format string, args ...any)` - Print formatted content without newline
- `frame.Println(format string, args ...any)` - Print formatted content with newline
- `frame.Divider(text string)` - Add a divider line with optional text

### Frame Options

- `frame.WithColor(color ansi.Color)` - Set frame border color
- `frame.WithStyle(style FrameStyle)` - Set frame style (Box or Bracket)
- `frame.WithOutput(w io.Writer)` - Set custom output writer

### Frame Styles

- `frame.Box` - Full box borders with complete enclosure
- `frame.Bracket` - Simple bracket-style markers

## Examples

Run the example to see all features in action:

```bash
cd examples/frame
go run .
```

This will demonstrate:

- Basic frame usage
- Nested frames with color inheritance
- Different frame styles
- Dividers and formatting
- ANSI template processing
- Complex nested layouts
- Icon usage and status indicators
- Terminal control sequences

## Architecture

### Core Packages

- **`ansi`** - ANSI color codes, styles, template formatting, icons, and terminal control sequences
- **`components`** - Reusable UI components (Frame, etc.)
- **`term`** - Terminal utilities and width detection

### Design Principles

- **io.Writer Interface**: All components implement standard Go interfaces
- **Functional Options**: Flexible configuration using the options pattern
- **Template Processing**: Rich text formatting with `{{style+color:text}}` syntax
- **Responsive Design**: Automatic adaptation to terminal width
- **Color Inheritance**: Nested components inherit parent colors appropriately

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

