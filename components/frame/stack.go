package frame

import (
	"strings"
	"sync"

	"github.com/pseudomuto/gooey/ansi"
)

type FrameStack struct {
	frames []*Frame
	mutex  sync.RWMutex
}

func (fs *FrameStack) Push(frame *Frame) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	fs.frames = append(fs.frames, frame)
}

func (fs *FrameStack) Pop() *Frame {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	if len(fs.frames) == 0 {
		return nil
	}

	frame := fs.frames[len(fs.frames)-1]
	fs.frames = fs.frames[:len(fs.frames)-1]
	return frame
}

func (fs *FrameStack) Current() *Frame {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()
	if len(fs.frames) == 0 {
		return nil
	}
	return fs.frames[len(fs.frames)-1]
}

func (fs *FrameStack) Depth() int {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()
	return len(fs.frames)
}

// FrameDepth returns the 1-based depth position of a specific frame in the stack
// Returns 0 if the frame is not found
func (fs *FrameStack) FrameDepth(frame *Frame) int {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()
	for i, f := range fs.frames {
		if f == frame {
			return i + 1 // 1-based depth
		}
	}
	return 0
}

// GetFrameColors returns the colors of frames in the stack up to the specified depth
func (fs *FrameStack) GetFrameColors(maxDepth int) []ansi.Color {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	colors := make([]ansi.Color, 0, min(maxDepth, len(fs.frames)))
	for i := 0; i < min(maxDepth, len(fs.frames)); i++ {
		frameColorMutex.RLock()
		color := fs.frames[i].color
		if frameColorOverride != nil {
			color = *frameColorOverride
		}
		frameColorMutex.RUnlock()
		colors = append(colors, color)
	}
	return colors
}

// Prefix returns the prefix string for the current frame depth
func (fs *FrameStack) Prefix() string {
	depth := fs.Depth()
	if depth == 0 {
		return ""
	}

	var builder strings.Builder
	for i := range depth {
		if i == depth-1 {
			builder.WriteString(frameBranch)
		} else {
			builder.WriteString(frameVerticalPrefix)
		}
	}
	return builder.String()
}

// ContentPrefix returns the prefix that should be used for content inside the current frame
func (fs *FrameStack) ContentPrefix() string {
	depth := fs.Depth()
	if depth == 0 {
		return ""
	}

	var builder strings.Builder
	for range depth {
		builder.WriteString(frameVerticalPrefix)
	}
	return builder.String()
}

// PrefixWidth returns the width of the prefix for the current frame depth
func (fs *FrameStack) PrefixWidth() int {
	return fs.Depth() * 3
}
