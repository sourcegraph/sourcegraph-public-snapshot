package job

import (
	"context"
	"math"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	searchalert "github.com/sourcegraph/sourcegraph/internal/search/alert"
	"github.com/sourcegraph/sourcegraph/internal/search/job/jobutil"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewAlertJob creates a job that translates errors from child jobs
// into alerts when necessary.
func NewAlertJob(inputs *run.SearchInputs, child Job) Job {
	if _, ok := child.(*noopJob); ok {
		return child
	}
	return &alertJob{
		inputs: inputs,
		child:  child,
	}
}

type alertJob struct {
	inputs *run.SearchInputs
	child  Job
}

func (j *alertJob) Run(ctx context.Context, db database.DB, stream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, stream, finish := jobutil.StartSpan(ctx, stream, j)
	defer func() { finish(alert, err) }()

	start := time.Now()
	countingStream := streaming.NewResultCountingStream(stream)
	statsObserver := streaming.NewStatsObservingStream(countingStream)
	jobAlert, err := j.child.Run(ctx, db, statsObserver)

	ao := searchalert.Observer{
		Db:           db,
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
