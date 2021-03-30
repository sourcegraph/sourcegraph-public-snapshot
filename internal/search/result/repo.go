package result

import (
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
)

type RepoMatch struct {
	Name api.RepoName
	ID   api.RepoID

	// rev optionally specifies a revision to go to for search results.
	Rev string
}

func (r RepoMatch) Limit(limit int) int {
	// Always represents one result and limit > 0 so we just return limit - 1.
	return limit - 1
}

func (r *RepoMatch) ResultCount() int {
	return 1
}

func (r *RepoMatch) Select(path filter.SelectPath) Match {
	switch path.Type {
	case filter.Repository:
		return r
	}
	return nil
}

func (r *RepoMatch) searchResultMarker() {}
