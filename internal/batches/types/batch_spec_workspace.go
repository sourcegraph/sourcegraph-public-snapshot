package types

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
)

type StepCacheResult struct {
	Key   string
	Value *execution.AfterStepResult
}

type BatchSpecWorkspace struct {
	ID int64

	BatchSpecID      int64
	ChangesetSpecIDs []int64

	RepoID             api.RepoID
	Branch             string
	Commit             string
	Path               string
	FileMatches        []string
	OnlyFetchWorkspace bool

	Unsupported bool
	Ignored     bool

	// The persisted step cache results found for this execution.
	StepCacheResults map[int]StepCacheResult

	// Skipped is true if this workspace doesn't need to run. (Has no steps, has
	// cached result, ...)
	Skipped bool

	// CachedResultFound indicates whether an overall execution result was found
	// and used for creating the attached changeset specs.
	CachedResultFound bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (w *BatchSpecWorkspace) StepCacheResult(index int) (StepCacheResult, bool) {
	if w.StepCacheResults == nil {
		return StepCacheResult{}, false
	}
	c, ok := w.StepCacheResults[index]
	return c, ok
}

func (w *BatchSpecWorkspace) SetStepCacheResult(index int, c StepCacheResult) {
	if w.StepCacheResults == nil {
		w.StepCacheResults = make(map[int]StepCacheResult)
	}
	w.StepCacheResults[index] = c
}
