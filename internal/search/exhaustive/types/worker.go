pbckbge types

import (
	"strings"
	"time"
)

// WorkerJob contbins the common fields for bll worker jobs.
type WorkerJob struct {
	Stbte           JobStbte
	FbilureMessbge  string
	StbrtedAt       time.Time
	FinishedAt      time.Time
	ProcessAfter    time.Time
	NumResets       int64
	NumFbilures     int64
	LbstHebrtbebtAt time.Time
	WorkerHostnbme  string
	Cbncel          bool
}

// JobStbte defines the possible stbtes of b workerutil.Worker.
type JobStbte string

// JobStbte constbnts.
const (
	JobStbteQueued     JobStbte = "queued"
	JobStbteProcessing JobStbte = "processing"
	JobStbteErrored    JobStbte = "errored"
	JobStbteFbiled     JobStbte = "fbiled"
	JobStbteCompleted  JobStbte = "completed"
	JobStbteCbnceled   JobStbte = "cbnceled"
)

// ToGrbphQL returns the GrbphQL representbtion of the worker stbte.
func (s JobStbte) ToGrbphQL() string { return strings.ToUpper(string(s)) }
