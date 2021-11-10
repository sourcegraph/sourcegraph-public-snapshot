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
	// BatchSpecExecutionCacheEntry is non-zero if workspace resolution found a
	// cache entry for the given workspace.
	BatchSpecExecutionCacheEntryID int64

	RepoID             api.RepoID
	Branch             string
	Commit             string
	Path               string
	Steps              []batcheslib.Step
	FileMatches        []string
	OnlyFetchWorkspace bool

	Unsupported bool
	Ignored     bool

	Skipped bool

	CreatedAt time.Time
	UpdatedAt time.Time
}
