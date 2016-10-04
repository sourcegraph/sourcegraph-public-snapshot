package localstore

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"context"

	"github.com/keegancsmith/que-go"
	"gopkg.in/gorp.v1"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil"
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

// Job contains the fields necessary to do a Job
type Job struct {
	// Type determines what to do
	Type string

	// Args is passed to the worker
	Args []byte

	// Delay will ensure at least Delay time passes before popping the Job
	// off the queue.
	Delay time.Duration
}

// LockedJob is a job returned from the queue. You must call MarkSuccess or
// MarkError when done.
type LockedJob struct {
	*Job
	success func() error
	error   func(string) error
}

// NewLockedJob constructs a new LockedJob
func NewLockedJob(j *Job, success func() error, error func(string) error) *LockedJob {
	return &LockedJob{
		Job:     j,
		success: success,
		error:   error,
	}
}

// MarkSuccess marks the Job as successful and deletes it from the queue
func (j *LockedJob) MarkSuccess() error { return j.success() }

// MarkError marks the job as failed with reason. It will put it back on the
// queue for later processing.
func (j *LockedJob) MarkError(reason string) error { return j.error(reason) }

// QueueStats captures statistics of what is in the queue for a Job Type
type QueueStats struct {
	NumJobs          int
	NumJobsWithError int
}

type queue struct{}

// Enqueue puts j onto the queue
func (q *queue) Enqueue(ctx context.Context, j *Job) error {
	if Mocks.Queue.Enqueue != nil {
		return Mocks.Queue.Enqueue(ctx, j)
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
func (q *queue) LockJob(ctx context.Context) (*LockedJob, error) {
	if Mocks.Queue.LockJob != nil {
		return Mocks.Queue.LockJob(ctx)
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
			log15.Debug("Job.MarkError ignoring", "type", j.Type, "lastReason", j.LastError.String, "reason", reason)
			return j.Delete()
		}
	}

	return NewLockedJob(q.fromQue(j), j.Delete, errFunc), nil
}

// Stats returns statistics about the queue per Job Type
func (q *queue) Stats(ctx context.Context) (map[string]QueueStats, error) {
	if Mocks.Queue.Stats != nil {
		return Mocks.Queue.Stats(ctx)
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
	qs := map[string]QueueStats{}
	for _, row := range stats {
		s := row.(*stat)
		qs[s.Type] = QueueStats{
			NumJobs:          s.NumJobs,
			NumJobsWithError: s.NumJobsWithError,
		}
	}
	return qs, nil
}

func (q *queue) toQue(j *Job) *que.Job {
	return &que.Job{
		Type:  j.Type,
		Args:  j.Args,
		RunAt: time.Now().Add(j.Delay),
	}
}

func (q *queue) fromQue(j *que.Job) *Job {
	return &Job{
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

type MockQueue struct {
	Enqueue func(ctx context.Context, j *Job) error
	LockJob func(ctx context.Context) (*LockedJob, error)
	Stats   func(ctx context.Context) (map[string]QueueStats, error)
}

func (s *MockQueue) MockEnqueue(t *testing.T, wantJob *Job) (called *bool) {
	called = new(bool)
	s.Enqueue = func(ctx context.Context, job *Job) error {
		*called = true
		if !reflect.DeepEqual(job, wantJob) {
			t.Errorf("got job {Type:%s Args:%s}, want {Type:%s Args:%s}", job.Type, string(job.Args), wantJob.Type, string(wantJob.Args))
		}
		return nil
	}
	return
}

func (s *MockQueue) MockLockJob_Return(t *testing.T, job *Job) (called, calledSuccess, calledError *bool) {
	called = new(bool)
	calledSuccess = new(bool)
	calledError = new(bool)
	j := NewLockedJob(
		job,
		func() error {
			*calledSuccess = true
			return nil
		},
		func(_ string) error {
			*calledError = true
			return nil
		},
	)
	if job == nil {
		j = nil
	}
	s.LockJob = func(ctx context.Context) (*LockedJob, error) {
		*called = true
		return j, nil
	}
	return
}
