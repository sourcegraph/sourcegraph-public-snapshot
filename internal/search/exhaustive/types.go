package exhaustive

import (
	"strconv"
	"strings"
	"time"
)

// JobState defines the possible states of a workerutil.Worker.
type JobState string

// JobState constants.
const (
	JobStateQueued     JobState = "queued"
	JobStateProcessing JobState = "processing"
	JobStateErrored    JobState = "errored"
	JobStateFailed     JobState = "failed"
	JobStateCompleted  JobState = "completed"
)

// ToGraphQL returns the GraphQL representation of the worker state.
func (s JobState) ToGraphQL() string { return strings.ToUpper(string(s)) }

type Job struct {
	ID int64

	// InitiatorID is the user ID of the user who initiated the resolution job.
	// Currently, this is always the person who created the search.
	InitiatorID int32

	Query string

	// workerutil fields
	State           JobState
	FailureMessage  *string
	StartedAt       time.Time
	FinishedAt      time.Time
	ProcessAfter    time.Time
	NumResets       int64
	NumFailures     int64
	LastHeartbeatAt time.Time
	WorkerHostname  string
	Cancel          bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (j *Job) RecordID() int {
	return int(j.ID)
}

func (j *Job) RecordUID() string {
	return strconv.FormatInt(j.ID, 10)
}
