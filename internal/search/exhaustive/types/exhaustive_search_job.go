package types

import (
	"strconv"
	"time"
)

// ExhaustiveSearchJob is a job that runs the exhaustive search.
// Maps to the `exhaustive_search_jobs` database table.
type ExhaustiveSearchJob struct {
	WorkerJob

	ID int64

	// InitiatorID is the user ID of the user who initiated the resolution job.
	// Currently, this is always the person who created the search.
	InitiatorID int32

	Query string

	CreatedAt time.Time
	UpdatedAt time.Time

	// The aggregate state of the job. This is only set when the job is returned
	// from ListSearchJobs. This state is different from WorkerJob.State, because it
	// reflects the combined state of all jobs created as part of the search job.
	AggState JobState
}

func (j *ExhaustiveSearchJob) RecordID() int {
	return int(j.ID)
}

func (j *ExhaustiveSearchJob) RecordUID() string {
	return strconv.FormatInt(j.ID, 10)
}
