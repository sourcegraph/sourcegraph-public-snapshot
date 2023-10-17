package usagestats

import (
	"context"
	"fmt"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
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

	gitDirSize, err := db.GitserverRepos().GetGitserverGitDirSize(ctx)
	if err != nil {
		return nil, err
	}
	total.GitDirBytes = uint64(gitDirSize)

	repos, err := search.ListAllIndexed(ctx, search.Indexed())
	if err != nil {
		return nil, err
	}

	total.NewLinesCount = repos.Stats.NewLinesCount
	total.DefaultBranchNewLinesCount = repos.Stats.DefaultBranchNewLinesCount
	total.OtherBranchesNewLinesCount = repos.Stats.OtherBranchesNewLinesCount

	return &total, nil
}

func GetRepositorySizeHistorgram(ctx context.Context, db database.DB) ([]RepoSizeBucket, error) {
	kb := int64(1000)
	mb := kb * kb
	gb := kb * mb

	var sizes []int64
	sizes = append(sizes, 0)
	sizes = append(sizes, kb)
	sizes = append(sizes, mb)
	sizes = append(sizes, gb)
	sizes = append(sizes, 5*gb)
	sizes = append(sizes, 15*gb)
	sizes = append(sizes, 25*gb)
	sizes = append(sizes, 50*gb)
	sizes = append(sizes, 100*gb)

	var results []RepoSizeBucket

	baseStore := basestore.NewWithHandle(db.Handle())

	getCount := func(start int64, end *int64) (int64, bool, error) {
		baseQuery := "select coalesce(count(repo_size_bytes), 0) from gitserver_repos where clone_status = 'cloned' "
		upperBound := sqlf.Sprintf("and true")
		if end != nil {
			upperBound = sqlf.Sprintf("and repo_size_bytes < %s", *end)
		}
		return basestore.ScanFirstInt64(baseStore.Query(ctx, sqlf.Sprintf("%s and repo_size_bytes >= %s %s", sqlf.Sprintf(baseQuery), start, upperBound)))
	}

	for i := 1; i < len(sizes); i++ {
		start := sizes[i-1]
		end := sizes[i]
		count, got, err := getCount(start, &end)
		if err != nil {
			return nil, err
		} else if !got {
			continue
		}
		results = append(results, RepoSizeBucket{
			Lt:    &end,
			Gte:   start,
			Count: count,
		})
	}

	// get the infinite value (everything greater than the last bucket)
	last := sizes[len(sizes)-1]
	inf, got, err := getCount(last, nil)
	if err != nil {
		return nil, err
	}
	if got {
		results = append(results, RepoSizeBucket{
			Gte:   last,
			Count: inf,
		})
	}
	return results, nil
}

type RepoSizeBucket struct {
	Lt    *int64 `json:"lt,omitempty"`
	Gte   int64  `json:"gte,omitempty"`
	Count int64  `json:"count"`
}

func (r RepoSizeBucket) String() string {
	if r.Lt != nil {
		return fmt.Sprintf("Gte: %d, Lt: %d, Count: %d", r.Gte, *r.Lt, r.Count)
	}
	return fmt.Sprintf("Gte: %d, Count: %d", r.Gte, r.Count)
}
