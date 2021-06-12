package sequencer

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"
)

type RenderFunc func(time.Time, time.Time) ([]byte, int, error)

type Sequencer interface {
	Sequence(int, int)
}

type FrameSequencer struct {
	Renderer          RenderFunc
	Start             time.Time
	Interval, Padding time.Duration
	MaxConcurrency    int
	OutDirectory      string
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

	for i := 0; i < s.MaxConcurrency; i++ {
		go s.renderWorker(in, out)
	}

	for i := start; i <= end; i++ {
		frameStart := s.Start.Add(s.Interval * time.Duration(i-1))
		frameEnd := frameStart.Add(s.Interval - s.Padding)

		in <- frame{i, frameStart, frameEnd}
	}
	close(in)

	for i := 1; i <= numFrames; i++ {
		err := <-out
		if err != nil {
			fmt.Print(err)
		}
	}
}

func (s *FrameSequencer) renderWorker(in <-chan frame, out chan<- error) {
	for f := range in {
		startTime := time.Now()
		fmt.Printf("Rendering frame %d\n", f.num)

		b, code, err := s.Renderer(f.start, f.end)
		if err != nil {
			out <- fmt.Errorf("%d error: %v", code, err)
			continue
		}

		fmt.Printf("Frame %d rendered in %f seconds\n", f.num, time.Since(startTime).Seconds())

		filename := filepath.Join(s.OutDirectory, fmt.Sprintf("%06d.png", f.num))
		out <- ioutil.WriteFile(filename, b, 0644)
	}
}
