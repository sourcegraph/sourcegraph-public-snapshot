package usagestats

import (
	"context"

	"github.com/google/zoekt"
	"github.com/google/zoekt/query"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search"
)

type Repositories struct {
	// GitDirBytes is the amount of bytes stored in .git directories.
	GitDirBytes uint64

	// NewLinesCount is the number of newlines "\n" that appear in the zoekt
	// indexed documents. This is not exactly the same as line count, since it
	// will not include lines not terminated by "\n" (eg a file with no "\n",
	// or a final line without "\n").
	//
	// Note: Zoekt deduplicates documents across branches, so if a path has
	// the same contents on multiple branches, there is only one document for
	// it. As such that document's newlines is only counted once. See
	// DefaultBranchNewLinesCount and OtherBranchesNewLinesCount for counts
	// which do not deduplicate.
	NewLinesCount uint64

	// DefaultBranchNewLinesCount is the number of newlines "\n" in the default
	// branch.
	DefaultBranchNewLinesCount uint64

	// OtherBranchesNewLinesCount is the number of newlines "\n" in all
	// indexed branches except the default branch.
	OtherBranchesNewLinesCount uint64
}

func GetRepositories(ctx context.Context, db database.DB) (*Repositories, error) {
	var total Repositories

	stats, err := gitserver.NewClient(db).ReposStats(ctx)
	if err != nil {
		return nil, err
	}
	for _, stat := range stats {
		// In the rare case we haven't yet computed the stat (UpdatedAt ==
		// 0), we undercount the size.
		total.GitDirBytes += uint64(stat.GitDirBytes)
	}

	if !conf.SearchIndexEnabled() {
		return &total, nil
	}

	opts := &zoekt.ListOptions{Minimal: true}
	repos, err := search.Indexed().List(ctx, &query.Const{Value: true}, opts)
	if err != nil {
		return nil, err
	}

	total.NewLinesCount = repos.Stats.NewLinesCount
	total.DefaultBranchNewLinesCount = repos.Stats.DefaultBranchNewLinesCount
	total.OtherBranchesNewLinesCount = repos.Stats.OtherBranchesNewLinesCount

	return &total, nil
}
