package types

import (
	"strconv"
	"time"
)

// ExhaustiveSearchRepoRevisionJob is a job that runs the exhaustive search on a revision of a repository.
// Maps to the `exhaustive_search_repo_revision_jobs` database table.
type ExhaustiveSearchRepoRevisionJob struct {
	WorkerJob

	ID int64

	SearchRepoJobID int64
	Revision        string

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (j *ExhaustiveSearchRepoRevisionJob) RecordID() int {
	return int(j.ID)
}

func (j *ExhaustiveSearchRepoRevisionJob) RecordUID() string {
	return strconv.FormatInt(j.ID, 10)
}
