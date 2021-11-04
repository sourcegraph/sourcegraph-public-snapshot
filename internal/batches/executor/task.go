package executor

import (
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/lib/batches/template"

	"github.com/sourcegraph/src-cli/internal/batches/graphql"
	"github.com/sourcegraph/src-cli/internal/batches/repozip"
)

type Task struct {
	Repository *graphql.Repository

	// Path is the folder relative to the repository's root in which the steps
	// should be executed. "" means root.
	Path string
	// OnlyFetchWorkspace determines whether the repository archive contains
	// the complete repository or just the files in Path (and additional files,
	// see RepoFetcher).
	// If Path is "" then this setting has no effect.
	OnlyFetchWorkspace bool

	Steps []batcheslib.Step

	// TODO(mrnugget): this should just be a single BatchSpec field instead, if
	// we can make it work with caching
	BatchChangeAttributes *template.BatchChangeAttributes `json:"-"`
	Template              *batcheslib.ChangesetTemplate   `json:"-"`
	TransformChanges      *batcheslib.TransformChanges    `json:"-"`

	Archive repozip.Archive `json:"-"`

	CachedResultFound bool                      `json:"-"`
	CachedResult      execution.AfterStepResult `json:"-"`
}

func (t *Task) ArchivePathToFetch() string {
	if t.OnlyFetchWorkspace {
		return t.Path
	}
	return ""
}

func (t *Task) cacheKey() TaskCacheKey {
	return TaskCacheKey{t}
}
