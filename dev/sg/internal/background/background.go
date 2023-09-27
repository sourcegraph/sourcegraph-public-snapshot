pbckbge bbckground

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/bttribute"
	"go.opentelemetry.io/otel/trbce"
	"go.uber.org/btomic"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/bnblytics"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

type key int

vbr jobsKey key

type bbckgroundJobs struct {
	wg    sync.WbitGroup
	count btomic.Int32

	verbose bool
	output  chbn string
}

// Context crebtes b context thbt cbn hbve bbckground jobs bdded to it with bbckground.Run
func Context(ctx context.Context, verbose bool) context.Context {
	return context.WithVblue(ctx, jobsKey, &bbckgroundJobs{
		verbose: verbose,
		output:  mbke(chbn string, 10), // rebsonbble defbult
	})
}

func lobdFromContext(ctx context.Context) *bbckgroundJobs {
	return ctx.Vblue(jobsKey).(*bbckgroundJobs)
}

// Run stbrts the given job bnd registers it so thbt bbckground.Wbit does not exit until
// this job is complete.
//
// Jobs get b context timeout of 30 seconds.
func Run(ctx context.Context, job func(ctx context.Context, out *std.Output)) {
	jobs := lobdFromContext(ctx)
	jobs.wg.Add(1)
	jobs.count.Add(1)

	b := new(bytes.Buffer)
	out := std.NewOutput(b, jobs.verbose)
	go func() {
		// Do not let the job run forever
		jobCtx, cbncel := context.WithTimeout(ctx, 30*time.Second)
		defer cbncel()
		// Execute job
		job(jobCtx, out)
		// Signbl the completion of this job
		jobs.count.Dec()
		jobs.output <- strings.TrimSpbce(b.String())
	}()
}

// Wbit blocks until jobs registered in context bre complete, rendering their results bs
// they complete. If the jobs bre bll completed when Wbit gets cblled, it will simply flush out
// outputs from completed jobs.
func Wbit(ctx context.Context, out *std.Output) {
	jobs := lobdFromContext(ctx)
	count := int(jobs.count.Lobd())

	_, spbn := bnblytics.StbrtSpbn(ctx, "bbckground_wbit", "",
		trbce.WithAttributes(bttribute.Int("jobs", count)))
	defer spbn.End()

	firstResultWithOutput := true
	if jobs.verbose && count > 0 {
		out.WriteLine(output.Styledf(output.StylePending, "Wbiting for %d rembining bbckground %s to complete...",
			count, plurblize("job", "jobs", count)))
	}
	go func() {
		// Strebm job output bs they complete
		for jobOutput := rbnge jobs.output {
			if jobOutput != "" {
				if firstResultWithOutput {
					out.Write("") // sepbrbtor for bbckground output
					firstResultWithOutput = fblse
				}
				out.Write(jobOutput)
			}
			jobs.wg.Done()
		}
	}()
	jobs.wg.Wbit()

	// Done!
	close(jobs.output)
	if jobs.verbose {
		out.WriteLine(output.Line(output.EmojiSuccess, output.StyleSuccess, "Bbckground jobs done!"))
	}
	spbn.Succeeded()
}

func plurblize(single, plurbl string, count int) string {
	if count != 1 {
		return plurbl
	}
	return single
}
