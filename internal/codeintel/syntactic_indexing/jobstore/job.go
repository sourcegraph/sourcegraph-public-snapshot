package jobstore

import (
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type recordState string

const (
	Queued     recordState = "queued"
	Errored    recordState = "errored"
	Processing recordState = "processing"
	Completed  recordState = "completed"
)

// Unless marked otherwise, the columns in this
// record have a special meaning assigned to them by
// the queries dbworker performs. You can read more
// about the different fields and what they do here:
// https://docs-legacy.sourcegraph.com/dev/background-information/workers#database-backed-stores
type SyntacticIndexingJob struct {
	ID             int         `json:"id"`
	State          recordState `json:"state"`
	QueuedAt       time.Time   `json:"queuedAt"`
	StartedAt      *time.Time  `json:"startedAt"`
	FinishedAt     *time.Time  `json:"finishedAt"`
	ProcessAfter   *time.Time  `json:"processAfter"`
	NumResets      int         `json:"numResets"`
	NumFailures    int         `json:"numFailures"`
	FailureMessage *string     `json:"failureMessage"`
	ShouldReindex  bool        `json:"shouldReindex"`

	// The fields below are not part of the standard dbworker fields

	// Which commit to index
	Commit api.CommitID `json:"commit"`
	// Which repository id to index
	RepositoryID api.RepoID `json:"repositoryId"`
	// Name of repository being indexed
	RepositoryName string `json:"repositoryName"`
	// Which user scheduled this job
	EnqueuerUserID int32 `json:"enqueuerUserID"`
}

var _ workerutil.Record = SyntacticIndexingJob{}

func (i SyntacticIndexingJob) RecordID() int {
	return i.ID
}

func (i SyntacticIndexingJob) RecordUID() string {
	return strconv.Itoa(i.ID)
}

func ScanSyntacticIndexRecord(s dbutil.Scanner) (*SyntacticIndexingJob, error) {
	var job SyntacticIndexingJob
	if err := scanSyntacticIndexRecord(&job, s); err != nil {
		return nil, err
	}
	return &job, nil
}

func scanSyntacticIndexRecord(job *SyntacticIndexingJob, s dbutil.Scanner) error {

	// Make sure this is in sync with columnExpressions below...
	if err := s.Scan(
		&job.ID,
		&job.Commit,
		&job.QueuedAt,
		&job.State,
		&job.FailureMessage,
		&job.StartedAt,
		&job.FinishedAt,
		&job.ProcessAfter,
		&job.NumResets,
		&job.NumFailures,
		&job.RepositoryID,
		&job.RepositoryName,
		&job.ShouldReindex,
		&job.EnqueuerUserID,
	); err != nil {
		return err
	}

	return nil
}
