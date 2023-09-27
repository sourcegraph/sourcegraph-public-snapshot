pbckbge types

import (
	"strconv"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
)

// ExhbustiveSebrchRepoJob is b job thbt runs the exhbustive sebrch on b repository.
// Mbps to the `exhbustive_sebrch_repo_jobs` dbtbbbse tbble.
type ExhbustiveSebrchRepoJob struct {
	WorkerJob

	ID int64

	RepoID      bpi.RepoID
	RefSpec     string
	SebrchJobID int64

	CrebtedAt time.Time
	UpdbtedAt time.Time
}

func (j *ExhbustiveSebrchRepoJob) RecordID() int {
	return int(j.ID)
}

func (j *ExhbustiveSebrchRepoJob) RecordUID() string {
	return strconv.FormbtInt(j.ID, 10)
}
