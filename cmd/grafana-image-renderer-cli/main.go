package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/MacroPower/grafana-image-renderer-sdk-go/pkg/client"
	"github.com/MacroPower/grafana-image-renderer-sdk-go/pkg/sequencer"
)

func addDefaultFlags(
	f *flag.FlagSet,
	apiURL *string,
	apiKeyOrBasicAuth *string,
	dashboardID *string,
	panelID *int,
	renderWidth *int,
	renderHeight *int,
	timeout *time.Duration,
	startTimeMs *int64,
) {
	f.StringVar(apiURL, "api-url", "", "Grafana API URL")
	f.StringVar(apiKeyOrBasicAuth, "api-key-or-basic-auth", "", "Grafana authorization, either an API key or basic auth")
	f.StringVar(dashboardID, "dashboard", "", "ID of the dashboard")
	f.IntVar(panelID, "panel", 0, "ID of the panel, 0 = Entire dashboard")
	f.IntVar(renderWidth, "width", 1920, "The width of the image")
	f.IntVar(renderHeight, "height", 1080, "The height of the image")
	f.DurationVar(timeout, "timeout", 1*time.Minute, "Timeout of the render request")
	f.Int64Var(startTimeMs, "start-time", 0, "The starting timestamp (Unix MS) of the render")
}

func fromUnixMs(ms int64) time.Time {
	return time.Unix(ms/int64(1000), (ms%int64(1000))*int64(1000000))
}

func main() {
	var (
		apiURL            string
		apiKeyOrBasicAuth string
		dashboardID       string
		panelID           int
		renderWidth       int
		renderHeight      int
		timeout           time.Duration
		startTimeMs       int64

		sequenceCommand = flag.NewFlagSet("sequence", flag.ExitOnError)
		startFrame      = sequenceCommand.Int("start-frame", 1, "The first frame to render")
		endFrame        = sequenceCommand.Int("end-frame", 2, "The last frame to render")
		frameInterval   = sequenceCommand.Duration("frame-interval", 5*time.Minute, "Time progression between frames, positive = forward, negative = backward")
		startPadding    = sequenceCommand.Duration("start-padding", 0, "Duration to add to the start of the frame")
		endPadding      = sequenceCommand.Duration("end-padding", 0, "Duration to add to the end of the frame")
		maxConcurrency  = sequenceCommand.Int("max-concurrency", 5, "Maximum number of concurrent render requests")
		workerDelay     = sequenceCommand.Duration("worker-delay", 2*time.Second, "Delay worker startup")
		outDirectory    = sequenceCommand.String("out-directory", "frames", "Directory to write rendered frames to")

		imageCommand = flag.NewFlagSet("image", flag.ExitOnError)
		endTime      = imageCommand.Int64("end-time", 0, "The ending timestamp (Unix MS) of the render")
		outFile      = imageCommand.String("out-file", "img.png", "The file to write")
	)

	if len(os.Args) < 2 {
		fmt.Println("image or sequence subcommand is required")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "image":
		addDefaultFlags(
			imageCommand,
			&apiURL, &apiKeyOrBasicAuth, &dashboardID, &panelID,
			&renderWidth, &renderHeight, &timeout, &startTimeMs,
		)
		imageCommand.Parse(os.Args[2:])
	case "sequence":
		addDefaultFlags(
			sequenceCommand,
			&apiURL, &apiKeyOrBasicAuth, &dashboardID, &panelID,
			&renderWidth, &renderHeight, &timeout, &startTimeMs,
		)
		sequenceCommand.Parse(os.Args[2:])
	default:
		fmt.Println("image or sequence subcommand is required")
		os.Exit(1)
	}

	argErr := false
	if apiURL == "" {
		fmt.Println("api-url is required")
		argErr = true
	}
	if dashboardID == "" {
		fmt.Println("dashboard is required")
		argErr = true
	}
	if startTimeMs == 0 {
		fmt.Println("start-time is required")
		argErr = true
	}
	if imageCommand.Parsed() {
		if *endTime == 0 {
			fmt.Println("end-time is required")
			argErr = true
		}
	}
	if sequenceCommand.Parsed() {
		if *maxConcurrency < 1 {
			fmt.Println("max-concurrency must be at least 1")
			argErr = true
		}
		if *startFrame < 1 {
			fmt.Println("start-frame must be 1 or higher")
			argErr = true
		}
		if *startFrame > *endFrame {
			fmt.Println("end-frame must be after start-frame")
			argErr = true
		}
	}
	if argErr {
		fmt.Println("Use --help for more information")
		os.Exit(1)
	}

	client := client.NewClient(apiURL, apiKeyOrBasicAuth, http.DefaultClient)

	rf := func(start time.Time, end time.Time) ([]byte, int, error) {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		b, code, err := client.Render(ctx, dashboardID, start, end, panelID, renderWidth, renderHeight)

		return b, code, err
	}

	started := time.Now()
	startTime := fromUnixMs(startTimeMs)

	if imageCommand.Parsed() {
		b, _, err := rf(startTime, fromUnixMs(*endTime))
		if err != nil {
			panic(err)
		}

		err = ioutil.WriteFile(*outFile, b, 0644)
		if err != nil {
			panic(err)
		}
	}

	if sequenceCommand.Parsed() {
		seq := sequencer.FrameSequencer{
			Renderer:       rf,
			Start:          startTime,
			Interval:       *frameInterval,
			StartPadding:   *startPadding,
			EndPadding:     *endPadding,
			WorkerDelay:    *workerDelay,
			MaxConcurrency: *maxConcurrency,
			SaveCallback: func(b []byte, n int) error {
				filename := filepath.Join(*outDirectory, fmt.Sprintf("%06d.png", n))
				return ioutil.WriteFile(filename, b, 0644)
			},
		}
		seq.Sequence(*startFrame, *endFrame)
	}

	fmt.Printf("Completed all work in %f seconds", time.Since(started).Seconds())
}
