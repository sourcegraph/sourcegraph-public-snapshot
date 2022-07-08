package background

import (
	"bytes"
	"context"
	"sync"

	"go.uber.org/atomic"

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
	go func() {
		b := new(bytes.Buffer)
		out := std.NewOutput(b, verbose)
		job(out)
		jobs.results <- b.String()
	}()
}

func Wait(ctx context.Context, out *std.Output) {
	jobs := loadFromContext(ctx)
	count := jobs.count.Load()
	if count == 0 {
		return // no jobs registered
	}

	p := out.Pending(output.Styledf(output.StylePending, "waiting for remaining background jobs to complete (%d total)...", count))

	go func() {
		for c := range jobs.results {
			out.Write(c)
			jobs.wg.Done()
		}
	}()

	jobs.wg.Wait()
	p.Destroy()
	close(jobs.results)

	out.VerboseLine(output.Line(output.EmojiSuccess, output.StyleSuccess, "Background jobs done!"))
}
