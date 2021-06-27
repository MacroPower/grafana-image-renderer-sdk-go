package sequencer

import (
	"fmt"
	"time"
)

// RenderFunc is used to define render behavior. You should define the
// call that should be made to the renderer. In its sumplest form, it
// should simply call client.Render with the time parameters passed.
type RenderFunc func(time.Time, time.Time) ([]byte, int, error)

// SaveFunc can be used to handle render results. The caller will pass
// the rendered results as []byte, and include an int that identifies
// the render in the sequence.
type SaveFunc func([]byte, int) error

// Sequencer defines and manages a render sequence.
type Sequencer interface {
	Sequence(int, int)
}

type FrameSequencer struct {
	// See RenderFunc for more information.
	Renderer RenderFunc

	// The start time of the frame sequence. If using a positive
	// interval, this should be the start of the range to capture. If
	// using a negative interval, it is the end of the range.
	Start time.Time

	// Interval is the time progression between frames.
	Interval time.Duration

	// Padding can be added or subtracted from the frame.
	StartPadding, EndPadding time.Duration

	// Maximum number of concurrent render requests.
	MaxConcurrency int

	// See SaveFunc for more information.
	SaveCallback SaveFunc
}

type frame struct {
	num        int
	start, end time.Time
}

func (s *FrameSequencer) Sequence(start, end int) {
	if start < 1 || start > end {
		panic("malformed sequence")
	}

	numFrames := 1 + end - start
	in := make(chan frame, numFrames)
	out := make(chan error, numFrames)

	maxConcurrency := s.MaxConcurrency
	if maxConcurrency > numFrames {
		maxConcurrency = numFrames
	}
	for i := 0; i < maxConcurrency; i++ {
		go s.renderWorker(i, in, out)
	}

	for i := start; i <= end; i++ {
		frameStart := s.Start.Add(s.Interval * time.Duration(i-1))
		frameEnd := frameStart.Add(s.Interval)

		if frameStart.After(frameEnd) {
			// This allows users to inverse the frame order
			// by passing a negative interval.
			oldFrameEnd := frameEnd
			frameEnd = frameStart
			frameStart = oldFrameEnd
		}

		frameStart = frameStart.Add(s.StartPadding)
		frameEnd = frameEnd.Add(s.EndPadding)

		in <- frame{i, frameStart, frameEnd}
	}
	close(in)

	for i := 1; i <= numFrames; i++ {
		err := <-out
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (s *FrameSequencer) renderWorker(in <-chan frame, out chan<- error) {
	for f := range in {
		startTime := time.Now()
		fmt.Printf("Rendering frame %d\n", f.num)

		b, code, err := s.Renderer(f.start, f.end)
		if err != nil || code != 200 {
			out <- fmt.Errorf("worker %d error: code %d: %v", n, code, err)
			continue
		}

		fmt.Printf("Frame %d rendered in %f seconds\n", f.num, time.Since(startTime).Seconds())

		out <- s.SaveCallback(b, f.num)
	}
}
