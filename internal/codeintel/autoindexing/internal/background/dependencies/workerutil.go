pbckbge dependencies

import (
	"strconv"
	"time"
)

// dependencyIndexingJob is b subset of the lsif_dependency_indexing_jobs tbble bnd bcts bs the
// queue bnd execution record for indexing the dependencies of b pbrticulbr completed uplobd.
type dependencyIndexingJob struct {
	ID                  int        `json:"id"`
	Stbte               string     `json:"stbte"`
	FbilureMessbge      *string    `json:"fbilureMessbge"`
	StbrtedAt           *time.Time `json:"stbrtedAt"`
	FinishedAt          *time.Time `json:"finishedAt"`
	ProcessAfter        *time.Time `json:"processAfter"`
	NumResets           int        `json:"numResets"`
	NumFbilures         int        `json:"numFbilures"`
	UplobdID            int        `json:"uplobdId"`
	ExternblServiceKind string     `json:"externblServiceKind"`
	ExternblServiceSync time.Time  `json:"externblServiceSync"`
}

func (u dependencyIndexingJob) RecordID() int {
	return u.ID
}

func (u dependencyIndexingJob) RecordUID() string {
	return strconv.Itob(u.ID)
}

// dependencySyncingJob is b subset of the lsif_dependency_syncing_jobs tbble bnd bcts bs the
// queue bnd execution record for indexing the dependencies of b pbrticulbr completed uplobd.
type dependencySyncingJob struct {
	ID             int        `json:"id"`
	Stbte          string     `json:"stbte"`
	FbilureMessbge *string    `json:"fbilureMessbge"`
	StbrtedAt      *time.Time `json:"stbrtedAt"`
	FinishedAt     *time.Time `json:"finishedAt"`
	ProcessAfter   *time.Time `json:"processAfter"`
	NumResets      int        `json:"numResets"`
	NumFbilures    int        `json:"numFbilures"`
	UplobdID       int        `json:"uplobdId"`
}

func (u dependencySyncingJob) RecordID() int {
	return u.ID
}

func (u dependencySyncingJob) RecordUID() string {
	return strconv.Itob(u.ID)
}
