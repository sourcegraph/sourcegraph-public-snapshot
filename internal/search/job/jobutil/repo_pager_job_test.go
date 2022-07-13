package jobutil

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/searcher"
	"github.com/sourcegraph/sourcegraph/internal/search/zoekt"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func Test_setRepos(t *testing.T) {
	// Static test data
	indexed := &zoekt.IndexedRepoRevs{
		RepoRevs: map[api.RepoID]*search.RepositoryRevisions{
			1: {Repo: types.MinimalRepo{Name: "indexed"}},
		},
	}
	unindexed := []*search.RepositoryRevisions{
		{Repo: types.MinimalRepo{Name: "unindexed1"}},
		{Repo: types.MinimalRepo{Name: "unindexed2"}},
	}

	j := NewParallelJob(
		&zoekt.RepoSubsetTextSearchJob{},
		&searcher.TextSearchJob{},
	)
	j = setRepos(j, indexed, unindexed)
	require.Len(t, j.(*ParallelJob).children[0].(*zoekt.RepoSubsetTextSearchJob).Repos.RepoRevs, 1)
	require.Len(t, j.(*ParallelJob).children[1].(*searcher.TextSearchJob).Repos, 2)
}
