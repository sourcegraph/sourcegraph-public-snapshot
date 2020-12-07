package util

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/efritz/pentimento"
)

// MaxDisplayLines is the number of lines that will be displayed before truncation.
const MaxDisplayLines = 50

// MaxDisplayWidth is the number of columns that can be used to draw a progress bar.
const MaxDisplayWidth = 80

// ParallelFn groups an error-returning function with a description that can be displayed
// by runParallel.
type ParallelFn struct {
	Fn          func(ctx context.Context) error
	Description func() string
	Total       func() int
	Finished    func() int
}

// braille is an animated spinner based off of the characters used by yarn.
var braille = pentimento.NewAnimatedString([]string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}, pentimento.DefaultInterval)

// RunParallel runs each function in parallel. Returns the first error to occur. The
// number of invocations is limited by concurrency.
func RunParallel(ctx context.Context, concurrency int, fns []ParallelFn) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	queue := make(chan int, len(fns))     // worker input
	errs := make(chan errPair, len(fns))  // worker output
	var wg sync.WaitGroup                 // denotes active writers of errs channel
	pendingMap := newPendingMap(len(fns)) // state tracker

	// queue all work up front
	for i := range fns {
		queue <- i
	}
	close(queue)

	// launch workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			runFunctions(ctx, fns, pendingMap, queue, errs)
		}()
	}

	// block until completion or error
	err := monitor(ctx, fns, pendingMap, errs, concurrency)
	if err != nil {
		cancel()    // stop workers
		wg.Wait()   // wait for workers to drain
		close(errs) // close output channel
	}

	return err
}

// errPair bundles an error value with the function index from which it was produced.
type errPair struct {
	i   int
	err error
}

// runFunctions is the worker body. It will pull an index off of the work queue,
// mark that index as pending, then send the index and the value resulting from
// the invocation of the function at that index onto the errors channel.
func runFunctions(ctx context.Context, fns []ParallelFn, pendingMap *pendingMap, queue <-chan int, errs chan<- errPair) {
	for {
		select {
		case i, ok := <-queue:
			if !ok {
				return
			}

			pendingMap.set(i)
			errs <- errPair{i, fns[i].Fn(ctx)}

		case <-ctx.Done():
			return
		}
	}
}

// monitor waits for all functions to complete, an error, or the context to be
// canceled. The first error encountered is returned. The current state of the
// pending map is periodically written to the screen. All content written to the
// screen is removed at exit of this function.
func monitor(ctx context.Context, fns []ParallelFn, pendingMap *pendingMap, errs <-chan errPair, concurrency int) error {
	return pentimento.PrintProgress(func(p *pentimento.Printer) error {
		defer func() {
			// Clear last progress update on exit
			_ = p.Reset()
		}()

		for pendingMap.size() != 0 {
			select {
			case pair := <-errs:
				if pair.err != nil && pair.err != context.Canceled {
					return pair.err
				}

				// Nil-valued error, remove it from the pending map
				pendingMap.remove(pair.i)

			case <-time.After(time.Millisecond * 250):
				// Update screen

			case <-ctx.Done():
				return ctx.Err()
			}

			_ = p.WriteContent(formatUpdate(fns, pendingMap, concurrency))
		}

		return nil
	})
}

// formatUpdate constructs a content object with a number of lines indicating the in progress
// and head-of-queue tasks, as well as a progress bar.
func formatUpdate(fns []ParallelFn, pendingMap *pendingMap, concurrency int) *pentimento.Content {
	keys := pendingMap.keys()
	content := pentimento.NewContent()

	for _, i := range keys[:numLines(concurrency, len(keys))] {
		if pendingMap.get(i) {
			content.AddLine(fmt.Sprintf("%s %s", braille, fns[i].Description()))
		} else {
			content.AddLine(fmt.Sprintf("%s %s", " ", fns[i].Description()))
		}
	}

	total := 0
	finished := 0
	for _, fn := range fns {
		total += fn.Total()
		finished += fn.Finished()
	}

	content.AddLine("")
	content.AddLine(formatProgressBar(total, finished))
	return content
}

// numLines determines how many lines to display in formatUpdate.
func numLines(concurrency, numTasks int) int {
	return int(math.Min(float64(concurrency*2), math.Min(float64(numTasks), float64(MaxDisplayLines))))
}

// formatProgressBar constructs a progress bar string based on the relationship between the
// total and finished parameters.
func formatProgressBar(total, finished int) string {
	maxWidth := MaxDisplayWidth - 4 - digits(total) - digits(finished)
	width := int(float64(maxWidth) * float64(finished) / float64(total))

	var arrow string
	if width < maxWidth {
		arrow = ">"
	}

	return fmt.Sprintf(
		"[%s%s%s] %d/%d",
		strings.Repeat("=", width),
		arrow,
		strings.Repeat(" ", maxWidth-width-len(arrow)),
		finished,
		total,
	)
}

// digits returns the number of digits of n.
func digits(n int) int {
	if n >= 10 {
		return 1 + digits(n/10)
	}
	return 1
}
