package frame

import (
	"sync"

	"github.com/pseudomuto/gooey/ansi"
)

type frameStack struct {
	frames []*Frame
	mutex  sync.RWMutex
}

func (fs *frameStack) push(frame *Frame) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	fs.frames = append(fs.frames, frame)
}

func (fs *frameStack) pop() *Frame {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	if len(fs.frames) == 0 {
		return nil
	}

	frame := fs.frames[len(fs.frames)-1]
	fs.frames = fs.frames[:len(fs.frames)-1]
	return frame
}

func (fs *frameStack) current() *Frame {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()
	if len(fs.frames) == 0 {
		return nil
	}

	return fs.frames[len(fs.frames)-1]
}

func (fs *frameStack) depth() int {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	return len(fs.frames)
}

// frameDepth returns the 1-based depth position of a specific frame in the stack
// Returns 0 if the frame is not found
func (fs *frameStack) frameDepth(frame *Frame) int {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()
	for i, f := range fs.frames {
		if f == frame {
			return i + 1 // 1-based depth
		}
	}

	return 0
}

// frameColors returns the colors of frames in the stack up to the specified depth
func (fs *frameStack) frameColors(maxDepth int) []ansi.Color {
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
