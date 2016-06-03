package store

import "golang.org/x/net/context"

// Queue pushes and dequeues jobs. Note: we don't dequeue a job directly,
// instead we need to mark a job as finished. This allows us to pick up work
// when processing fails on it.
type Queue interface {
	// Enqueue puts j onto the queue
	Enqueue(ctx context.Context, j *Job) error

	// LockJob removes a job from queue, or returns nil if there is no
	// jobs. You must call LockedJob.MarkSuccess or LockedJob.MarkError
	// when done.
	LockJob(ctx context.Context) (*LockedJob, error)
}

// Job contains the fields necessary to do a Job
type Job struct {
	// Type determines what to do
	Type string

	// Args is passed to the worker
	Args []byte
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
