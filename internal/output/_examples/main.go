package main

import (
	"flag"
	"strings"
	"sync"
	"time"

	"github.com/sourcegraph/src-cli/internal/output"
)

var (
	duration time.Duration
	verbose  bool
)

func init() {
	flag.DurationVar(&duration, "progress", 5*time.Second, "time to take in the progress bar and pending samples")
	flag.BoolVar(&verbose, "verbose", false, "enable verbose mode")
}

func main() {
	flag.Parse()

	out := output.NewOutput(flag.CommandLine.Output(), output.OutputOpts{
		Verbose: verbose,
	})

	var wg sync.WaitGroup
	progress := out.Progress([]output.ProgressBar{
		{Label: "A", Max: 1.0},
		{Label: "BB", Max: 1.0, Value: 0.5},
		{Label: strings.Repeat("X", 200), Max: 1.0},
	}, nil)

	wg.Add(1)
	go func() {
		ticker := time.NewTicker(duration / 20)
		defer ticker.Stop()
		defer wg.Done()

		i := 0
		for _ = range ticker.C {
			i += 1
			if i > 20 {
				return
			}

			progress.Verbosef("%slog line %d", output.StyleWarning, i)
		}
	}()

	wg.Add(1)
	go func() {
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		defer wg.Done()

		start := time.Now()
		until := start.Add(duration)
		for _ = range ticker.C {
			now := time.Now()
			if now.After(until) {
				return
			}

			progress.SetValue(0, float64(now.Sub(start))/float64(duration))
			progress.SetValue(1, 0.5+float64(now.Sub(start))/float64(duration)/2)
			progress.SetValue(2, 2*float64(now.Sub(start))/float64(duration))
		}
	}()

	wg.Wait()
	progress.Complete()

	func() {
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

		pending := out.Pending(output.Linef("", output.StylePending, "Starting pending ticker"))
		defer pending.Complete(output.Line(output.EmojiSuccess, output.StyleSuccess, "Ticker done!"))

		until := time.Now().Add(duration)
		for _ = range ticker.C {
			now := time.Now()
			if now.After(until) {
				return
			}

			pending.Updatef("Waiting for another %s", until.Sub(time.Now()))
		}
	}()

	out.Write("")
	block := out.Block(output.Line(output.EmojiSuccess, output.StyleSuccess, "Done!"))
	block.Write("Here is some additional information.\nIt even line wraps.")
}
