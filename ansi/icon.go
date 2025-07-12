package ansi

import "fmt"

const (
	// Status icons
	CheckMark    Icon = "✓"
	CrossMark    Icon = "✗"
	Circle       Icon = "○"
	FilledCircle Icon = "●"
	Warning      Icon = "⚠"
	Info         Icon = "ℹ"
	Question     Icon = "?"
	Exclamation  Icon = "!"

	// Checkbox icons
	CheckboxEmpty    Icon = "☐"
	CheckboxChecked  Icon = "☑"
	CheckboxCrossed  Icon = "☒"
	CheckboxProgress Icon = "⧖"

	// Arrow icons
	ArrowUp    Icon = "↑"
	ArrowDown  Icon = "↓"
	ArrowLeft  Icon = "←"
	ArrowRight Icon = "→"

	// Progress indicators
	Spinner1 Icon = "⠋"
	Spinner2 Icon = "⠙"
	Spinner3 Icon = "⠹"
	Spinner4 Icon = "⠸"
	Spinner5 Icon = "⠼"
	Spinner6 Icon = "⠴"
	Spinner7 Icon = "⠦"
	Spinner8 Icon = "⠧"

	// Geometric shapes
	Star      Icon = "★"
	StarEmpty Icon = "☆"
	Diamond   Icon = "◆"
	Square    Icon = "■"
	Triangle  Icon = "▲"
	Heart     Icon = "♥"

	// Common symbols
	Play   Icon = "▶"
	Pause  Icon = "⏸"
	Stop   Icon = "⏹"
	Record Icon = "●"
	Fast   Icon = "⚡"
	Slow   Icon = "🐌"
	Fire   Icon = "🔥"

	// Success/Error indicators
	Success Icon = "🎉"
	Error   Icon = "💥"
	Bug     Icon = "🐛"
	Fix     Icon = "🔧"

	// File/Folder icons
	File       Icon = "📄"
	Folder     Icon = "📁"
	FolderOpen Icon = "📂"

	// Network/Communication
	Download Icon = "⬇"
	Upload   Icon = "⬆"
	Link     Icon = "🔗"

	// Time/Clock
	Clock     Icon = "🕐"
	Hourglass Icon = "⏳"

	// Misc
	Gear    Icon = "⚙"
	Lock    Icon = "🔒"
	Key     Icon = "🔑"
	Shield  Icon = "🛡"
	Target  Icon = "🎯"
	Rocket  Icon = "🚀"
	Unicorn Icon = "🦄"

	// Separators
	Dash       Icon = "─"
	DoubleDash Icon = "═"
	Dot        Icon = "•"
	Bullet     Icon = "◦"
)

var (
	// TaskIcons for task status indicators
	TaskIcons = &IconSet{
		Name: "Task Status",
		Icons: map[string]Icon{
			"pending":     Circle,
			"in_progress": FilledCircle,
			"completed":   CheckMark,
			"failed":      CrossMark,
		},
	}

	// ChecklistIcons for checklist items
	ChecklistIcons = &IconSet{
		Name: "Checklist",
		Icons: map[string]Icon{
			"unchecked":   CheckboxEmpty,
			"in_progress": CheckboxProgress,
			"checked":     CheckboxChecked,
			"failed":      CheckboxCrossed,
		},
	}

	// StatusIcons for general status indicators
	StatusIcons = &IconSet{
		Name: "Status",
		Icons: map[string]Icon{
			"success": CheckMark,
			"error":   CrossMark,
			"warning": Warning,
			"info":    Info,
		},
	}

	// SpinnerIcons for progress animations
	SpinnerIcons = &IconSet{
		Name: "Spinner",
		Icons: map[string]Icon{
			"1": Spinner1,
			"2": Spinner2,
			"3": Spinner3,
			"4": Spinner4,
			"5": Spinner5,
			"6": Spinner6,
			"7": Spinner7,
			"8": Spinner8,
		},
	}

	// Global icon registry instance
	DefaultIconRegistry = NewIconRegistry()
)

type (
	// Icon represents a Unicode symbol/glyph
	Icon string

	// IconSet provides collections of related icons
	IconSet struct {
		Name  string
		Icons map[string]Icon
	}

	// IconRegistry manages global icon sets
	IconRegistry struct {
		sets map[string]*IconSet
	}
)

// NewIconRegistry creates a new icon registry
func NewIconRegistry() *IconRegistry {
	registry := &IconRegistry{
		sets: make(map[string]*IconSet),
	}

	// Register default icon sets
	registry.RegisterSet(TaskIcons)
	registry.RegisterSet(ChecklistIcons)
	registry.RegisterSet(StatusIcons)
	registry.RegisterSet(SpinnerIcons)

	return registry
}

// GetTaskIcon returns a task icon by status
func GetTaskIcon(status string) Icon {
	if icon, exists := TaskIcons.GetIcon(status); exists {
		return icon
	}
	return Question
}

// GetChecklistIcon returns a checklist icon by status
func GetChecklistIcon(status string) Icon {
	if icon, exists := ChecklistIcons.GetIcon(status); exists {
		return icon
	}
	return Question
}

// GetStatusIcon returns a status icon by status
func GetStatusIcon(status string) Icon {
	if icon, exists := StatusIcons.GetIcon(status); exists {
		return icon
	}
	return Question
}

// GetSpinnerIcon returns a spinner icon by frame
func GetSpinnerIcon(frame string) Icon {
	if icon, exists := SpinnerIcons.GetIcon(frame); exists {
		return icon
	}
	return Spinner1
}

// FormatIcon formats an icon with color and optional text
func FormatIcon(icon Icon, color Color, text string) string {
	if text == "" {
		return icon.Colorize(color)
	}
	return fmt.Sprintf("%s %s", icon.Colorize(color), text)
}

// String returns the icon as a string
func (i Icon) String() string {
	return string(i)
}

// Colorize applies a color to the icon
func (i Icon) Colorize(color Color) string {
	return color.Colorize(i.String())
}

// Sprint returns the icon with optional color formatting
func (i Icon) Sprint(color Color) string {
	return color.Sprint(i.String())
}

// GetIcon returns an icon from the set by name
func (is *IconSet) GetIcon(name string) (Icon, bool) {
	icon, exists := is.Icons[name]
	return icon, exists
}

// ListIcons returns all icon names in the set
func (is *IconSet) ListIcons() []string {
	names := make([]string, 0, len(is.Icons))
	for name := range is.Icons {
		names = append(names, name)
	}
	return names
}

// RegisterSet adds an icon set to the registry
func (ir *IconRegistry) RegisterSet(set *IconSet) {
	ir.sets[set.Name] = set
}

// GetIcon retrieves an icon by set name and icon name
func (ir *IconRegistry) GetIcon(setName, iconName string) (Icon, bool) {
	set, exists := ir.sets[setName]
	if !exists {
		return "", false
	}
	return set.GetIcon(iconName)
}

// GetSet retrieves an icon set by name
func (ir *IconRegistry) GetSet(name string) (*IconSet, bool) {
	set, exists := ir.sets[name]
	return set, exists
}

// ListSets returns all registered icon set names
func (ir *IconRegistry) ListSets() []string {
	names := make([]string, 0, len(ir.sets))
	for name := range ir.sets {
		names = append(names, name)
	}
	return names
}
