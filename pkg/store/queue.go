package store

import "time"

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
