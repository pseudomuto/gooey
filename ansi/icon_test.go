package ansi_test

import (
	"testing"

	. "github.com/pseudomuto/gooey/ansi"
	"github.com/stretchr/testify/require"
)

func TestIcon(t *testing.T) {
	// Test Icon String() method
	require.Equal(t, "✓", CheckMark.String())
	require.Equal(t, "✗", CrossMark.String())
	require.Equal(t, "○", Circle.String())
	require.Equal(t, "●", FilledCircle.String())
	require.Equal(t, "☐", CheckboxEmpty.String())
	require.Equal(t, "☑", CheckboxChecked.String())
	require.Equal(t, "☒", CheckboxCrossed.String())
	require.Equal(t, "⧖", CheckboxProgress.String())
}

func TestIconColorize(t *testing.T) {
	// Test icon colorization
	greenCheck := CheckMark.Colorize(Green)
	require.Contains(t, greenCheck, "✓")
	require.Contains(t, greenCheck, Green.String())

	redCross := CrossMark.Colorize(Red)
	require.Contains(t, redCross, "✗")
	require.Contains(t, redCross, Red.String())
}

func TestIconSprint(t *testing.T) {
	// Test icon Sprint method
	yellowCircle := FilledCircle.Sprint(Yellow)
	require.Contains(t, yellowCircle, "●")
	require.Contains(t, yellowCircle, Yellow.String())
}

func TestIconSets(t *testing.T) {
	// Test TaskIcons set
	require.NotNil(t, TaskIcons)
	require.Equal(t, "Task Status", TaskIcons.Name)

	icon, exists := TaskIcons.GetIcon("pending")
	require.True(t, exists)
	require.Equal(t, Circle, icon)

	icon, exists = TaskIcons.GetIcon("completed")
	require.True(t, exists)
	require.Equal(t, CheckMark, icon)

	// Test ChecklistIcons set
	require.NotNil(t, ChecklistIcons)
	require.Equal(t, "Checklist", ChecklistIcons.Name)

	icon, exists = ChecklistIcons.GetIcon("unchecked")
	require.True(t, exists)
	require.Equal(t, CheckboxEmpty, icon)

	icon, exists = ChecklistIcons.GetIcon("checked")
	require.True(t, exists)
	require.Equal(t, CheckboxChecked, icon)
}

func TestIconRegistry(t *testing.T) {
	registry := NewIconRegistry()
	require.NotNil(t, registry)

	// Test getting icons from registry
	icon, exists := registry.GetIcon("Task Status", "pending")
	require.True(t, exists)
	require.Equal(t, Circle, icon)

	icon, exists = registry.GetIcon("Checklist", "checked")
	require.True(t, exists)
	require.Equal(t, CheckboxChecked, icon)

	// Test non-existent icon
	_, exists = registry.GetIcon("NonExistent", "test")
	require.False(t, exists)
}

func TestGetIconFunctions(t *testing.T) {
	// Test convenience functions
	require.Equal(t, Circle, GetTaskIcon("pending"))
	require.Equal(t, CheckMark, GetTaskIcon("completed"))
	require.Equal(t, Question, GetTaskIcon("invalid"))

	require.Equal(t, CheckboxEmpty, GetChecklistIcon("unchecked"))
	require.Equal(t, CheckboxChecked, GetChecklistIcon("checked"))
	require.Equal(t, Question, GetChecklistIcon("invalid"))

	require.Equal(t, CheckMark, GetStatusIcon("success"))
	require.Equal(t, CrossMark, GetStatusIcon("error"))
	require.Equal(t, Question, GetStatusIcon("invalid"))

	require.Equal(t, Spinner1, GetSpinnerIcon("1"))
	require.Equal(t, Spinner8, GetSpinnerIcon("8"))
	require.Equal(t, Spinner1, GetSpinnerIcon("invalid"))
}

func TestFormatIcon(t *testing.T) {
	// Test formatting with color only
	result := FormatIcon(CheckMark, Green, "")
	require.Contains(t, result, "✓")
	require.Contains(t, result, Green.String())

	// Test formatting with color and text
	result = FormatIcon(CheckMark, Green, "Success")
	require.Contains(t, result, "✓")
	require.Contains(t, result, Green.String())
	require.Contains(t, result, "Success")
}

func TestIconSetListIcons(t *testing.T) {
	names := TaskIcons.ListIcons()
	require.Contains(t, names, "pending")
	require.Contains(t, names, "completed")
	require.Contains(t, names, "in_progress")
	require.Contains(t, names, "failed")
}

func TestIconRegistryListSets(t *testing.T) {
	registry := NewIconRegistry()
	sets := registry.ListSets()
	require.Contains(t, sets, "Task Status")
	require.Contains(t, sets, "Checklist")
	require.Contains(t, sets, "Status")
	require.Contains(t, sets, "Spinner")
}
