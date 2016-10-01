package localstore

import (
	"fmt"
	"time"

	"context"

	"github.com/keegancsmith/que-go"
	"gopkg.in/gorp.v1"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store/mockstore"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
)

func init() {
	// From schema.sql in github.com/keegancsmith/que-go
	AppSchema.CreateSQL = append(AppSchema.CreateSQL, `
CREATE TABLE que_jobs
(
  priority    smallint    NOT NULL DEFAULT 100,
  run_at      timestamptz NOT NULL DEFAULT now(),
  job_id      bigserial   NOT NULL,
  job_class   text        NOT NULL,
  args        json        NOT NULL DEFAULT '[]'::json,
  error_count integer     NOT NULL DEFAULT 0,
  last_error  text,
  queue       text        NOT NULL DEFAULT '',

  CONSTRAINT que_jobs_pkey PRIMARY KEY (queue, priority, run_at, job_id)
);`)
	AppSchema.DropSQL = append(AppSchema.DropSQL, "DROP TABLE IF EXISTS que_jobs;")
}

const queueMaxAttempts = 2

var MockQueue *mockstore.Queue

type queue struct{}

// Enqueue puts j onto the queue
func (q *queue) Enqueue(ctx context.Context, j *store.Job) error {
	if MockQueue != nil {
		return MockQueue.Enqueue(ctx, j)
	}

	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Queue.Enqueue."+j.Type, nil); err != nil {
		return err
	}
	c, err := q.client(ctx)
	if err != nil {
		return err
	}
	return c.Enqueue(q.toQue(j))
}

// LockJob removes a job from queue, or returns nil if there is no
// jobs. You must call LockedJob.MarkSuccess or LockedJob.MarkError
// when done.
func (q *queue) LockJob(ctx context.Context) (*store.LockedJob, error) {
	if MockQueue != nil {
		return MockQueue.LockJob(ctx)
	}

	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Queue.LockJob", nil); err != nil {
		return nil, err
	}
	c, err := q.client(ctx)
	if err != nil {
		return nil, err
	}
	j, err := c.LockJob("")
	if err != nil || j == nil {
		return nil, err
	}

	// We don't want to retry jobs forever, we would rather log and drain
	// the queue.
	errFunc := j.Error
	if j.ErrorCount+1 >= queueMaxAttempts {
		errFunc = func(reason string) error {
			log15.Debug("store.Job.MarkError ignoring", "type", j.Type, "lastReason", j.LastError.String, "reason", reason)
			return j.Delete()
		}
	}

	return store.NewLockedJob(q.fromQue(j), j.Delete, errFunc), nil
}

// Stats returns statistics about the queue per Job Type
func (q *queue) Stats(ctx context.Context) (map[string]store.QueueStats, error) {
	if MockQueue != nil {
		return MockQueue.Stats(ctx)
	}

	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Queue.Stats", nil); err != nil {
		return nil, err
	}
	type stat struct {
		NumJobs          int    `db:"num_jobs"`
		NumJobsWithError int    `db:"num_jobs_with_error"`
		Type             string `db:"type"`
	}
	stats, err := appDBH(ctx).Select(&stat{}, `select job_class as type, count(1) as num_jobs, count(nullif(error_count, 0)) as num_jobs_with_error from que_jobs group by job_class;`)
	if err != nil {
		return nil, err
	}
	qs := map[string]store.QueueStats{}
	for _, row := range stats {
		s := row.(*stat)
		qs[s.Type] = store.QueueStats{
			NumJobs:          s.NumJobs,
			NumJobsWithError: s.NumJobsWithError,
		}
	}
	return qs, nil
}

func (q *queue) toQue(j *store.Job) *que.Job {
	return &que.Job{
		Type:  j.Type,
		Args:  j.Args,
		RunAt: time.Now().Add(j.Delay),
	}
}

func (q *queue) fromQue(j *que.Job) *store.Job {
	return &store.Job{
		Type: j.Type,
		Args: j.Args,
	}
}

func (q *queue) client(ctx context.Context) (*que.Client, error) {
	dbh := dbutil.GetUnderlyingSQLExecutor(appDBH(ctx))
	dbm, ok := dbh.(*gorp.DbMap)
	if !ok {
		return nil, fmt.Errorf("queue could not get underlying *sql.Db from appDBH")
	}
	return que.NewClient(dbm.Db), nil
}
