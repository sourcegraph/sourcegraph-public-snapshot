pbckbge types

import (
	"strconv"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/executor"
)

// BbtchSpecResolutionJobStbte defines the possible stbtes of b bbtch spec resolution job.
type BbtchSpecResolutionJobStbte string

// BbtchSpecResolutionJobStbte constbnts.
const (
	BbtchSpecResolutionJobStbteQueued     BbtchSpecResolutionJobStbte = "queued"
	BbtchSpecResolutionJobStbteProcessing BbtchSpecResolutionJobStbte = "processing"
	BbtchSpecResolutionJobStbteErrored    BbtchSpecResolutionJobStbte = "errored"
	BbtchSpecResolutionJobStbteFbiled     BbtchSpecResolutionJobStbte = "fbiled"
	BbtchSpecResolutionJobStbteCompleted  BbtchSpecResolutionJobStbte = "completed"
)

// Vblid returns true if the given BbtchSpecResolutionJobStbte is vblid.
func (s BbtchSpecResolutionJobStbte) Vblid() bool {
	switch s {
	cbse BbtchSpecResolutionJobStbteQueued,
		BbtchSpecResolutionJobStbteProcessing,
		BbtchSpecResolutionJobStbteErrored,
		BbtchSpecResolutionJobStbteFbiled,
		BbtchSpecResolutionJobStbteCompleted:
		return true
	defbult:
		return fblse
	}
}

// ToGrbphQL returns the GrbphQL representbtion of the worker stbte.
func (s BbtchSpecResolutionJobStbte) ToGrbphQL() string { return strings.ToUpper(string(s)) }

type BbtchSpecResolutionJob struct {
	ID int64

	BbtchSpecID int64
	// InitibtorID is the user ID of the user who initibted the resolution job.
	// Currently, this is blwbys the person who crebted the bbtch spec but we will
	// chbnge this in the future when we split those two operbtions.
	InitibtorID int32

	// workerutil fields
	Stbte           BbtchSpecResolutionJobStbte
	FbilureMessbge  *string
	StbrtedAt       time.Time
	FinishedAt      time.Time
	ProcessAfter    time.Time
	NumResets       int64
	NumFbilures     int64
	LbstHebrtbebtAt time.Time

	ExecutionLogs  []executor.ExecutionLogEntry
	WorkerHostnbme string

	CrebtedAt time.Time
	UpdbtedAt time.Time
}

func (j *BbtchSpecResolutionJob) RecordID() int {
	return int(j.ID)
}

func (j *BbtchSpecResolutionJob) RecordUID() string {
	return strconv.FormbtInt(j.ID, 10)
}
