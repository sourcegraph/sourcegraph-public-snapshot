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
	withBars bool
)

func init() {
	flag.DurationVar(&duration, "progress", 5*time.Second, "time to take in the progress bar and pending samples")
	flag.BoolVar(&verbose, "verbose", false, "enable verbose mode")
	flag.BoolVar(&withBars, "with-bars", false, "show status bars on top of progress bar")
}

func main() {
	flag.Parse()

	out := output.NewOutput(flag.CommandLine.Output(), output.OutputOpts{
		Verbose: verbose,
	})

	if withBars {
		demoProgressWithBars(out, duration)
	} else {
		demo(out, duration)
	}
}

func demo(out *output.Output, duration time.Duration) {
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
		for range ticker.C {
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
		for range ticker.C {
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
		for range ticker.C {
			now := time.Now()
			if now.After(until) {
				return
			}

			pending.Updatef("Waiting for another %s", time.Until(until))
		}
	}()

	out.Write("")
	block := out.Block(output.Line(output.EmojiSuccess, output.StyleSuccess, "Done!"))
	block.Write("Here is some additional information.\nIt even line wraps.")
	block.Close()
}

func demoProgressWithBars(out *output.Output, duration time.Duration) {
	var wg sync.WaitGroup
	progress := out.ProgressWithStatusBars([]output.ProgressBar{
		{Label: "Running steps", Max: 1.0},
	}, []*output.StatusBar{
		output.NewStatusBarWithLabel("github.com/sourcegraph/src-cli"),
		output.NewStatusBarWithLabel("github.com/sourcegraph/sourcegraph"),
	}, nil)

	wg.Add(1)
	go func() {
		ticker := time.NewTicker(duration / 10)
		defer ticker.Stop()
		defer wg.Done()

		i := 0
		for range ticker.C {
			i += 1
			if i > 10 {
				return
			}

			progress.Verbosef("%slog line %d", output.StyleWarning, i)
		}
	}()

	wg.Add(1)
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		defer wg.Done()

		start := time.Now()
		until := start.Add(duration)
		for range ticker.C {
			now := time.Now()
			if now.After(until) {
				return
			}

			elapsed := time.Since(start)

			if elapsed < 5*time.Second {
				if elapsed < 1*time.Second {
					progress.StatusBarUpdatef(0, "Downloading archive...")
					progress.StatusBarUpdatef(1, "Downloading archive...")

				} else if elapsed > 1*time.Second && elapsed < 2*time.Second {
					progress.StatusBarUpdatef(0, `comby -in-place 'fmt.Sprintf("%%d", :[v])' 'strconv.Itoa(:[v])' main.go`)
					progress.StatusBarUpdatef(1, `comby -in-place 'fmt.Sprintf("%%d", :[v])' 'strconv.Itoa(:[v])' pkg/main.go pkg/utils.go`)

				} else if elapsed > 2*time.Second && elapsed < 4*time.Second {
					progress.StatusBarUpdatef(0, `goimports -w main.go`)
					if elapsed > (2*time.Second + 500*time.Millisecond) {
						progress.StatusBarUpdatef(1, `goimports -w pkg/main.go pkg/utils.go`)
					}

				} else if elapsed > 4*time.Second && elapsed < 5*time.Second {
					progress.StatusBarCompletef(1, `Done!`)
					if elapsed > (4*time.Second + 500*time.Millisecond) {
						progress.StatusBarCompletef(0, `Done!`)
					}
				}
			}

			if elapsed > 5*time.Second && elapsed < 6*time.Second {
				progress.StatusBarResetf(0, "github.com/sourcegraph/code-intel", `Downloading archive...`)
				if elapsed > (5*time.Second + 200*time.Millisecond) {
					progress.StatusBarResetf(1, "github.com/sourcegraph/srcx86", `Downloading archive...`)
				}
			} else if elapsed > 6*time.Second && elapsed < 7*time.Second {
				progress.StatusBarUpdatef(1, `comby -in-place 'fmt.Sprintf("%%d", :[v])' 'strconv.Itoa(:[v])' main.go (%s)`)
				if elapsed > (6*time.Second + 100*time.Millisecond) {
					progress.StatusBarUpdatef(0, `comby -in-place 'fmt.Sprintf("%%d", :[v])' 'strconv.Itoa(:[v])' main.go`)
				}
			} else if elapsed > 7*time.Second && elapsed < 8*time.Second {
				progress.StatusBarCompletef(0, "Done!")
				if elapsed > (7*time.Second + 320*time.Millisecond) {
					progress.StatusBarCompletef(1, "Done!")
				}
			}

			progress.SetValue(0, float64(now.Sub(start))/float64(duration))
		}
	}()

	wg.Wait()

	progress.Complete()
}
