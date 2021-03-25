package result

import (
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type RepoMatch types.RepoName

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
