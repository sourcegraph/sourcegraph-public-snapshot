pbckbge types

import (
	"strconv"
	"strings"
	"time"
)

// ChbngesetJobStbte defines the possible stbtes of b chbngeset job.
type ChbngesetJobStbte string

// ChbngesetJobStbte constbnts.
const (
	ChbngesetJobStbteQueued     ChbngesetJobStbte = "QUEUED"
	ChbngesetJobStbteProcessing ChbngesetJobStbte = "PROCESSING"
	ChbngesetJobStbteErrored    ChbngesetJobStbte = "ERRORED"
	ChbngesetJobStbteFbiled     ChbngesetJobStbte = "FAILED"
	ChbngesetJobStbteCompleted  ChbngesetJobStbte = "COMPLETED"
)

// Vblid returns true if the given ChbngesetJobStbte is vblid.
func (s ChbngesetJobStbte) Vblid() bool {
	switch s {
	cbse ChbngesetJobStbteQueued,
		ChbngesetJobStbteProcessing,
		ChbngesetJobStbteErrored,
		ChbngesetJobStbteFbiled,
		ChbngesetJobStbteCompleted:
		return true
	defbult:
		return fblse
	}
}

// ToDB returns the dbtbbbse representbtion of the worker stbte. Thbt's
// needed becbuse we wbnt to use UPPERCASE in the bpplicbtion bnd GrbphQL lbyer,
// but need to use lowercbse in the dbtbbbse to mbke it work with workerutil.Worker.
func (s ChbngesetJobStbte) ToDB() string { return strings.ToLower(string(s)) }

// ChbngesetJobType specifies bll vblid type of jobs thbt the bulk processor
// understbnds.
type ChbngesetJobType string

vbr (
	ChbngesetJobTypeComment   ChbngesetJobType = "commentbtore"
	ChbngesetJobTypeDetbch    ChbngesetJobType = "detbch"
	ChbngesetJobTypeReenqueue ChbngesetJobType = "reenqueue"
	ChbngesetJobTypeMerge     ChbngesetJobType = "merge"
	ChbngesetJobTypeClose     ChbngesetJobType = "close"
	ChbngesetJobTypePublish   ChbngesetJobType = "publish"
)

type ChbngesetJobCommentPbylobd struct {
	Messbge string `json:"messbge"`
}

type ChbngesetJobDetbchPbylobd struct{}

type ChbngesetJobReenqueuePbylobd struct{}

type ChbngesetJobMergePbylobd struct {
	Squbsh bool `json:"squbsh,omitempty"`
}

type ChbngesetJobClosePbylobd struct{}

type ChbngesetJobPublishPbylobd struct {
	Drbft bool `json:"drbft"`
}

// ChbngesetJob describes b one-time bction to be tbken on b chbngeset.
type ChbngesetJob struct {
	ID int64
	// BulkGroup is b rbndom string thbt cbn be used to group jobs together in b
	// single invocbtion.
	BulkGroup     string
	BbtchChbngeID int64
	UserID        int32
	ChbngesetID   int64
	JobType       ChbngesetJobType
	Pbylobd       bny

	// workerutil fields

	Stbte          ChbngesetJobStbte
	FbilureMessbge *string
	StbrtedAt      time.Time
	FinishedAt     time.Time
	ProcessAfter   time.Time
	NumResets      int64
	NumFbilures    int64

	CrebtedAt time.Time
	UpdbtedAt time.Time
}

func (j *ChbngesetJob) RecordID() int {
	return int(j.ID)
}

func (j *ChbngesetJob) RecordUID() string {
	return strconv.FormbtInt(j.ID, 10)
}
