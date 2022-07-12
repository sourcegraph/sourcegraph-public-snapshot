package background

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"time"

	"go.uber.org/atomic"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/analytics"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

type key int

var jobsKey key

type backgroundJobs struct {
	wg    sync.WaitGroup
	count atomic.Int32

	results chan string
}

// Context creates a context that can have background jobs added to it with background.Run
func Context(ctx context.Context) context.Context {
	return context.WithValue(ctx, jobsKey, &backgroundJobs{
		results: make(chan string, 10), // reasonable default
	})
}

func loadFromContext(ctx context.Context) *backgroundJobs {
	return ctx.Value(jobsKey).(*backgroundJobs)
}

// Run starts the given job and registers it so that background.Wait does not exit until
// this job is complete.
//
// Jobs get a context timeout of 30 seconds.
func Run(ctx context.Context, job func(ctx context.Context, out *std.Output), verbose bool) {
	jobs := loadFromContext(ctx)
	jobs.wg.Add(1)
	jobs.count.Add(1)

	b := new(bytes.Buffer)
	out := std.NewOutput(b, verbose)
	go func() {
		jobCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		job(jobCtx, out)
		jobs.results <- strings.TrimSpace(b.String())
	}()
}

// Wait blocks until jobs registered in context are complete, rendering their results as
// they complete.
func Wait(ctx context.Context, out *std.Output) {
	jobs := loadFromContext(ctx)
	count := jobs.count.Load()
	if count == 0 {
		return // no jobs registered
	}
	start := time.Now() // start clock for additional time waited

	out.Write("") // separator for background output
	out.VerboseLine(output.Styledf(output.StylePending, "Waiting for remaining background jobs to complete (%d total)...", count))
	go func() {
		for r := range jobs.results {
			if r != "" {
				out.Write(r)
			}
			jobs.wg.Done()
		}
	}()
	jobs.wg.Wait()

	// Done!
	close(jobs.results)
	out.VerboseLine(output.Line(output.EmojiSuccess, output.StyleSuccess, "Background jobs done!"))
	analytics.LogEvent(ctx, "background_wait", nil, start)
}
