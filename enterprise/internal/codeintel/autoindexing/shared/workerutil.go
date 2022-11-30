package shared

import "time"

// DependencyIndexingJob is a subset of the lsif_dependency_indexing_jobs table and acts as the
// queue and execution record for indexing the dependencies of a particular completed upload.
type DependencyIndexingJob struct {
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

func (u DependencyIndexingJob) RecordID() int {
	return u.ID
}

// DependencySyncingJob is a subset of the lsif_dependency_syncing_jobs table and acts as the
// queue and execution record for indexing the dependencies of a particular completed upload.
type DependencySyncingJob struct {
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

func (u DependencySyncingJob) RecordID() int {
	return u.ID
}
