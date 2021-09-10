package types

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
)

type BatchSpecWorkspace struct {
	ID int64

	BatchSpecID      int64
	ChangesetSpecIDs []int64

	RepoID             api.RepoID
	Branch             string
	Commit             string
	Path               string
	Steps              []batcheslib.Step
	FileMatches        []string
	OnlyFetchWorkspace bool

	CreatedAt time.Time
	UpdatedAt time.Time
}
