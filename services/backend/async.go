package backend

import (
	"encoding/json"
	"fmt"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	localcli "sourcegraph.com/sourcegraph/sourcegraph/services/backend/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/services/svc"
	"sourcegraph.com/sqs/pbtypes"
)

var Async sourcegraph.AsyncServer = &async{}

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

func (s *async) RefreshIndexes(ctx context.Context, op *sourcegraph.AsyncRefreshIndexesOp) (*pbtypes.Void, error) {
	// Keep track of who triggered a refresh
	if actor := authpkg.ActorFromContext(ctx); actor.IsAuthenticated() {
		op.Source = fmt.Sprintf("%s (UID %d %s)", op.Source, actor.UID, actor.Login)
	}

	args, err := json.Marshal(op)
	if err != nil {
		return nil, err
	}
	err = store.QueueFromContext(ctx).Enqueue(ctx, &store.Job{
		Type: "RefreshIndexes",
		Args: args,
	})
	if err != nil {
		return nil, err
	}
	return &pbtypes.Void{}, nil
}

// try attempts to lock a job and do it. Returns true if work was done
func (s *asyncWorker) try(ctx context.Context) bool {
	j, err := store.QueueFromContext(ctx).LockJob(ctx)
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
func (s *asyncWorker) doSafe(ctx context.Context, job *store.Job) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic when running job %+v: %v", job, r)
		}
	}()
	err = s.do(ctx, job)
	return
}

func (s *asyncWorker) do(ctx context.Context, job *store.Job) error {
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
	_, err := svc.Defs(ctx).RefreshIndex(ctx, &sourcegraph.DefsRefreshIndexOp{
		Repo:                op.Repo,
		RefreshRefLocations: true,
		Force:               op.Force,
	})
	if err != nil {
		return grpc.Errorf(grpc.Code(err), "Def.RefreshIndex failed on repo %d from source %s: %s", op.Repo, op.Source, err)
	}

	_, err = svc.Search(ctx).RefreshIndex(ctx, &sourcegraph.SearchRefreshIndexOp{
		Repos:         []int32{op.Repo},
		RefreshCounts: true,
		RefreshSearch: true,
	})
	if err != nil {
		return grpc.Errorf(grpc.Code(err), "Search.RefreshIndex failed on repo %d from source %s: %s", op.Repo, op.Source, err)
	}

	return nil
}
