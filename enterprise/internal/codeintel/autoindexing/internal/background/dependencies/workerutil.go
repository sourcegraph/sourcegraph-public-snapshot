package dependencies

import (
	"strconv"
	"time"
)

// dependencyIndexingJob is a subset of the lsif_dependency_indexing_jobs table and acts as the
// queue and execution record for indexing the dependencies of a particular completed upload.
type dependencyIndexingJob struct {
	ID                  int        `json:"id"`
	State               string     `json:"state"`
	FailureMessage      *string    `json:"failureMessage"`
	StartedAt           *time.Time `json:"startedAt"`
	FinishedAt          *time.Time `json:"finishedAt"`
	ProcessAfter        *time.Time `json:"processAfter"`
	NumResets           int        `json:"numResets"`
	NumFailures         int        `json:"numFailures"`
	UploadID            int        `json:"uploadId"`
	ExternalServiceKind string     `json:"externalServiceKind"`
	ExternalServiceSync time.Time  `json:"externalServiceSync"`
}

func (u dependencyIndexingJob) RecordID() int {
	return u.ID
}

func (u dependencyIndexingJob) RecordUID() string {
	return strconv.Itoa(u.ID)
}

// dependencySyncingJob is a subset of the lsif_dependency_syncing_jobs table and acts as the
// queue and execution record for indexing the dependencies of a particular completed upload.
type dependencySyncingJob struct {
	ID             int        `json:"id"`
	State          string     `json:"state"`
	FailureMessage *string    `json:"failureMessage"`
	StartedAt      *time.Time `json:"startedAt"`
	FinishedAt     *time.Time `json:"finishedAt"`
	ProcessAfter   *time.Time `json:"processAfter"`
	NumResets      int        `json:"numResets"`
	NumFailures    int        `json:"numFailures"`
	UploadID       int        `json:"uploadId"`
}

func (u dependencySyncingJob) RecordID() int {
	return u.ID
}

func (u dependencySyncingJob) RecordUID() string {
	return strconv.Itoa(u.ID)
}
