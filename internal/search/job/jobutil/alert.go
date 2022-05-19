package jobutil

import (
	"context"
	"math"
	"time"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/search"
	searchalert "github.com/sourcegraph/sourcegraph/internal/search/alert"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewAlertJob creates a job that translates errors from child jobs
// into alerts when necessary.
func NewAlertJob(inputs *run.SearchInputs, child job.Job) job.Job {
	if _, ok := child.(*NoopJob); ok {
		return child
	}
	return &alertJob{
		inputs: inputs,
		child:  child,
	}
}

type alertJob struct {
	inputs *run.SearchInputs
	child  job.Job
}

func (j *alertJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	tr, ctx, stream, finish := job.StartSpan(ctx, stream, j)
	defer func() { finish(alert, err) }()
	tr.TagFields(trace.LazyFields(j.Tags))

	start := time.Now()
	countingStream := streaming.NewResultCountingStream(stream)
	statsObserver := streaming.NewStatsObservingStream(countingStream)
	jobAlert, err := j.child.Run(ctx, clients, statsObserver)

	ao := searchalert.Observer{
		Db:           clients.DB,
		SearchInputs: j.inputs,
		HasResults:   countingStream.Count() > 0,
	}
	if err != nil {
		ao.Error(ctx, err)
	}
	observerAlert, err := ao.Done()

	// We have an alert for context timeouts and we have a progress
	// notification for timeouts. We don't want to show both, so we only show
	// it if no repos are marked as timedout. This somewhat couples us to how
	// progress notifications work, but this is the third attempt at trying to
	// fix this behaviour so we are accepting that.
	if errors.Is(err, context.DeadlineExceeded) {
		if !statsObserver.Status.Any(search.RepoStatusTimedout) {
			usedTime := time.Since(start)
			suggestTime := longer(2, usedTime)
			return search.AlertForTimeout(usedTime, suggestTime, j.inputs.OriginalQuery, j.inputs.PatternType), nil
		} else {
			err = nil
		}
	}

	return search.MaxPriorityAlert(jobAlert, observerAlert), err
}

func (j *alertJob) Name() string {
	return "AlertJob"
}

func (j *alertJob) Tags() []log.Field {
	return []log.Field{
		trace.Stringer("query", j.inputs.Query),
		log.String("originalQuery", j.inputs.OriginalQuery),
		trace.Stringer("patternType", j.inputs.PatternType),
		log.Bool("onSourcegraphDotCom", j.inputs.OnSourcegraphDotCom),
		trace.Stringer("features", j.inputs.Features),
		trace.Stringer("protocol", j.inputs.Protocol),
	}
}

// longer returns a suggested longer time to wait if the given duration wasn't long enough.
func longer(n int, dt time.Duration) time.Duration {
	dt2 := func() time.Duration {
		Ndt := time.Duration(n) * dt
		dceil := func(x float64) time.Duration {
			return time.Duration(math.Ceil(x))
		}
		switch {
		case math.Floor(Ndt.Hours()) > 0:
			return dceil(Ndt.Hours()) * time.Hour
		case math.Floor(Ndt.Minutes()) > 0:
			return dceil(Ndt.Minutes()) * time.Minute
		case math.Floor(Ndt.Seconds()) > 0:
			return dceil(Ndt.Seconds()) * time.Second
		default:
			return 0
		}
	}()
	lowest := 2 * time.Second
	if dt2 < lowest {
		return lowest
	}
	return dt2
}
