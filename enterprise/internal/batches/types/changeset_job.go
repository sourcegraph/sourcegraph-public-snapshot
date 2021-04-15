package types

import "time"

// ChangesetJobType specifies all valid type of jobs that the bulk processor
// understands.
type ChangesetJobType string

var (
	ChangesetJobTypeComment ChangesetJobType = "commentatore"
)

type ChangesetJobCommentPayload struct {
	Message string `json:"message"`
}

// ChangesetJob describes a one-time action to be taken on a changeset.
type ChangesetJob struct {
	ID int64
	// BulkGroup is a random string that can be used to group jobs together in a
	// single invocation.
	BulkGroup     string
	BatchChangeID int64
	UserID        int32
	ChangesetID   int64
	JobType       ChangesetJobType
	Payload       interface{}

	// workerutil fields

	State          string
	FailureMessage *string
	StartedAt      time.Time
	FinishedAt     time.Time
	ProcessAfter   time.Time
	NumResets      int64
	NumFailures    int64

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (j *ChangesetJob) RecordID() int {
	return int(j.ID)
}
