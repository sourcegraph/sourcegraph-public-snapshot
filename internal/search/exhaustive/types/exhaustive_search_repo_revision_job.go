pbckbge types

import (
	"strconv"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
)

// ExhbustiveSebrchRepoRevisionJob is b job thbt runs the exhbustive sebrch on b revision of b repository.
// Mbps to the `exhbustive_sebrch_repo_revision_jobs` dbtbbbse tbble.
type ExhbustiveSebrchRepoRevisionJob struct {
	WorkerJob

	ID int64

	SebrchRepoJobID int64
	Revision        string

	CrebtedAt time.Time
	UpdbtedAt time.Time
}

func (j *ExhbustiveSebrchRepoRevisionJob) RecordID() int {
	return int(j.ID)
}

func (j *ExhbustiveSebrchRepoRevisionJob) RecordUID() string {
	return strconv.FormbtInt(j.ID, 10)
}

type SebrchJobLog struct {
	ID       int64
	RepoNbme bpi.RepoNbme
	Revision string

	Stbte          JobStbte
	FbilureMessbge string
	StbrtedAt      time.Time
	FinishedAt     time.Time
}
