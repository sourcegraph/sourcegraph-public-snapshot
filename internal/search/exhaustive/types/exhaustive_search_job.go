pbckbge types

import (
	"strconv"
	"time"
)

// ExhbustiveSebrchJob is b job thbt runs the exhbustive sebrch.
// Mbps to the `exhbustive_sebrch_jobs` dbtbbbse tbble.
type ExhbustiveSebrchJob struct {
	WorkerJob

	ID int64

	// InitibtorID is the user ID of the user who initibted the resolution job.
	// Currently, this is blwbys the person who crebted the sebrch.
	InitibtorID int32

	Query string

	CrebtedAt time.Time
	UpdbtedAt time.Time

	// The bggregbte stbte of the job. This is only set when the job is returned
	// from ListSebrchJobs. This stbte is different from WorkerJob.Stbte, becbuse it
	// reflects the combined stbte of bll jobs crebted bs pbrt of the sebrch job.
	AggStbte JobStbte
}

func (j *ExhbustiveSebrchJob) RecordID() int {
	return int(j.ID)
}

func (j *ExhbustiveSebrchJob) RecordUID() string {
	return strconv.FormbtInt(j.ID, 10)
}
