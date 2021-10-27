package unindexed

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	zoektutil "github.com/sourcegraph/sourcegraph/internal/search/zoekt"
)

// RepoFetcher is an object that exposes an interface to retrieve repos to
// search from Zoekt. The interface exposes a Get(ctx) method that allows
// parameterizing repo fetches by context.
type RepoFetcher struct {
	args              *search.TextParameters
	mode              search.GlobalSearchMode
	onMissingRepoRevs zoektutil.OnMissingRepoRevs
}

// repoData represents an object of repository revisions to search.
type repoData interface {
	AsList() []*search.RepositoryRevisions
	IsIndexed() bool
}

type IndexedMap map[api.RepoID]*search.RepositoryRevisions

func (m IndexedMap) AsList() []*search.RepositoryRevisions {
	reposList := make([]*search.RepositoryRevisions, 0, len(m))
	for _, repo := range m {
		reposList = append(reposList, repo)
	}
	return reposList
}

func (IndexedMap) IsIndexed() bool {
	return true
}

type UnindexedList []*search.RepositoryRevisions

func (ul UnindexedList) AsList() []*search.RepositoryRevisions {
	return ul
}

func (UnindexedList) IsIndexed() bool {
	return false
}

func NewRepoFetcher(stream streaming.Sender, args *search.TextParameters) RepoFetcher {
	return RepoFetcher{
		mode:              args.Mode,
		args:              args,
		onMissingRepoRevs: zoektutil.MissingRepoRevStatus(stream),
	}
}

// GetPartition returns the repository data to run structural search on. Importantly, it
// allows parameterizing the request to specify a context, for when multiple
// Get() calls are required with different limits or timeouts.
func (r *RepoFetcher) GetPartition(ctx context.Context) ([]repoData, error) {
	globalSearch := r.args.Mode == search.ZoektGlobalSearch
	request, err := zoektutil.NewIndexedSearchRequest(ctx, r.args, globalSearch, search.TextRequest, r.onMissingRepoRevs)
	if err != nil {
		return nil, err
	}
	repoSets := []repoData{UnindexedList(request.UnindexedRepos())} // unindexed included by default
	if r.mode != search.SearcherOnly {
		repoSets = append(repoSets, IndexedMap(request.IndexedRepos()))
	}
	return repoSets, nil
}
