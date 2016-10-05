package backend

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"gopkg.in/inconshreveable/log15.v2"

	"context"

	"google.golang.org/grpc"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/rcache"
	localcli "sourcegraph.com/sourcegraph/sourcegraph/services/backend/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
)

var Async = &async{}

type async struct{}
type asyncWorker struct{}

// StartAsyncWorkers will start async workers to consume jobs from the queue.
func StartAsyncWorkers(ctx context.Context) {
	w := &asyncWorker{}
	for i := 0; i < localcli.Flags.NumAsyncWorkers; i++ {
		go func() {
			for {
				didWork := w.try(ctx)
				if !didWork {
					// didn't do anything, sleep
					time.Sleep(5 * time.Second)
				}
			}
		}()
	}
}

func (s *async) RefreshIndexes(ctx context.Context, op *sourcegraph.AsyncRefreshIndexesOp) (err error) {
	if Mocks.Async.RefreshIndexes != nil {
		return Mocks.Async.RefreshIndexes(ctx, op)
	}

	ctx, done := trace(ctx, "Async", "RefreshIndexes", op, &err)
	defer done()

	// We filter out repos we can't index at this stage, so that our
	// metrics on async success vs failure aren't polluted by repos we do
	// not support.
	ok, err := s.shouldRefreshIndex(ctx, op)
	if err != nil {
		return err
	} else if !ok {
		asyncRefreshIndexesUnsupported.Inc()
		return nil
	}

	// Keep track of who triggered a refresh
	if actor := authpkg.ActorFromContext(ctx); actor.IsAuthenticated() {
		op.Source = fmt.Sprintf("%s (UID %s %s)", op.Source, actor.UID, actor.Login)
	}

	args, err := json.Marshal(op)
	if err != nil {
		return err
	}
	err = localstore.Queue.Enqueue(ctx, &localstore.Job{
		Type: "RefreshIndexes",
		Args: args,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *async) shouldRefreshIndex(ctx context.Context, op *sourcegraph.AsyncRefreshIndexesOp) (bool, error) {
	rev, err := Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: op.Repo})
	if err != nil {
		return false, err
	}
	inv, err := Repos.GetInventory(ctx, &sourcegraph.RepoRevSpec{
		Repo:     op.Repo,
		CommitID: rev.CommitID,
	})
	if err != nil {
		return false, err
	}
	for _, l := range inv.Languages {
		// We currently only support Go in universe
		if l.Name == "Go" {
			return true, nil
		}
	}
	return false, nil
}

// try attempts to lock a job and do it. Returns true if work was done
func (s *asyncWorker) try(ctx context.Context) bool {
	j, err := localstore.Queue.LockJob(ctx)
	if err != nil {
		log15.Debug("Queue.LockJob failed", "err", err)
		return false
	}
	if j == nil {
		return false
	}

	err = s.doSafe(ctx, j.Job)
	if err != nil {
		err = j.MarkError(err.Error())
		if err != nil {
			log15.Debug("Queue Job.Error failed", "err", err)
		}
		return true
	}
	err = j.MarkSuccess()
	if err != nil {
		log15.Debug("Queue Job.Delete failed", "err", err)
	}
	return true
}

// doSafe is a wrapper which recovers from panics
func (s *asyncWorker) doSafe(ctx context.Context, job *localstore.Job) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic when running job %+v: %v", job, r)
		}
	}()
	err = s.do(ctx, job)
	return
}

func (s *asyncWorker) do(ctx context.Context, job *localstore.Job) error {
	switch job.Type {
	case "RefreshIndexes":
		op := &sourcegraph.AsyncRefreshIndexesOp{}
		err := json.Unmarshal(job.Args, op)
		if err != nil {
			return err
		}
		return s.refreshIndexes(ctx, op)
	case "NOOP":
		return nil
	default:
		return fmt.Errorf("unknown async job type %s", job.Type)
	}
}

func (s *asyncWorker) refreshIndexes(ctx context.Context, op *sourcegraph.AsyncRefreshIndexesOp) error {
	ctx, release, ok := rcache.TryAcquireMutex(ctx, fmt.Sprintf("async/refreshindex/%d", op.Repo))
	if !ok {
		// We already have a running job for this repo, but it may be
		// indexing an older commit. So lets just try again later.
		op.Source = op.Source + " (mutex)"
		args, err := json.Marshal(op)
		if err != nil {
			return err
		}
		return localstore.Queue.Enqueue(ctx, &localstore.Job{
			Type:  "RefreshIndexes",
			Args:  args,
			Delay: 10 * time.Minute,
		})
	}
	defer release()

	err := Defs.RefreshIndex(ctx, &sourcegraph.DefsRefreshIndexOp{
		Repo:                op.Repo,
		RefreshRefLocations: true,
		Force:               op.Force,
	})
	if err != nil {
		return grpc.Errorf(grpc.Code(err), "Def.RefreshIndex failed on repo %d from source %s: %s", op.Repo, op.Source, err)
	}

	return nil
}

var asyncRefreshIndexesUnsupported = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "src",
	Subsystem: "async",
	Name:      "refresh_indexes_unsupported",
	Help:      "Number of repos we skip indexing.",
})

func init() {
	prometheus.MustRegister(asyncRefreshIndexesUnsupported)
}

type MockAsync struct {
	RefreshIndexes func(v0 context.Context, v1 *sourcegraph.AsyncRefreshIndexesOp) error
}

func (s *MockAsync) MockRefreshIndexes(t *testing.T, want *sourcegraph.AsyncRefreshIndexesOp) (called *bool) {
	called = new(bool)
	s.RefreshIndexes = func(ctx context.Context, got *sourcegraph.AsyncRefreshIndexesOp) error {
		*called = true
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got AsyncRefeshIndexesOp %+v, want %+v", got, want)
		}
		return nil
	}
	return
}
