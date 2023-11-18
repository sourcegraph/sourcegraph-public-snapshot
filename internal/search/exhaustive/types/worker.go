package types

import (
	"strings"
	"time"
)

// WorkerJob contains the common fields for all worker jobs.
type WorkerJob struct {
	State           JobState
	FailureMessage  string
	StartedAt       time.Time
	FinishedAt      time.Time
	ProcessAfter    time.Time
	NumResets       int64
	NumFailures     int64
	LastHeartbeatAt time.Time
	WorkerHostname  string
	Cancel          bool
}

// JobState defines the possible states of a workerutil.Worker.
type JobState string

// JobState constants.
const (
	JobStateQueued     JobState = "queued"
	JobStateProcessing JobState = "processing"
	JobStateErrored    JobState = "errored"
	JobStateFailed     JobState = "failed"
	JobStateCompleted  JobState = "completed"
	JobStateCanceled   JobState = "canceled"
)

// ToGraphQL returns the GraphQL representation of the worker state.
func (s JobState) ToGraphQL() string { return strings.ToUpper(string(s)) }
