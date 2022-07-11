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

func Context(ctx context.Context) context.Context {
	return context.WithValue(ctx, jobsKey, &backgroundJobs{
		results: make(chan string, 10),
	})
}

func loadFromContext(ctx context.Context) *backgroundJobs {
	return ctx.Value(jobsKey).(*backgroundJobs)
}

func Run(ctx context.Context, job func(out *std.Output), verbose bool) {
	jobs := loadFromContext(ctx)
	jobs.wg.Add(1)
	jobs.count.Add(1)
	b := new(bytes.Buffer)
	out := std.NewOutput(b, verbose)
	go func() {
		job(out)
		jobs.results <- strings.TrimSpace(b.String())
	}()
}

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
