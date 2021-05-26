package executor

import (
	"github.com/sourcegraph/src-cli/internal/batches"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
)

type Task struct {
	Repository *graphql.Repository

	// Path is the folder relative to the repository's root in which the steps
	// should be executed.
	Path string
	// OnlyFetchWorkspace determines whether the repository archive contains
	// the complete repository or just the files in Path (and additional files,
	// see RepoFetcher).
	// If Path is "" then this setting has no effect.
	OnlyFetchWorkspace bool

	Steps []batches.Step

	// TODO(mrnugget): this should just be a single BatchSpec field instead, if
	// we can make it work with caching
	BatchChangeAttributes *BatchChangeAttributes     `json:"-"`
	Template              *batches.ChangesetTemplate `json:"-"`
	TransformChanges      *batches.TransformChanges  `json:"-"`

	Archive batches.RepoZip `json:"-"`

	CachedResultFound bool                `json:"-"`
	CachedResult      stepExecutionResult `json:"-"`
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
