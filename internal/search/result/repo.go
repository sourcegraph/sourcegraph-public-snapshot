package result

import (
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type RepoMatch struct {
	Name api.RepoName
	ID   api.RepoID

	// rev optionally specifies a revision to go to for search results.
	Rev string

	DescriptionMatches []Range
	RepoNameMatches    []Range
}

func (r RepoMatch) RepoName() types.MinimalRepo {
	return types.MinimalRepo{
		Name: r.Name,
		ID:   r.ID,
	}
}

func (r RepoMatch) Limit(limit int) int {
	// Always represents one result and limit > 0 so we just return limit - 1.
	return limit - 1
}

func (r *RepoMatch) ResultCount() int {
	return 1
}

func (r *RepoMatch) Select(path filter.SelectPath) Match {
	switch path.Root() {
	case filter.Repository:
		return r
	}
	return nil
}

func (r *RepoMatch) URL() *url.URL {
	path := "/" + string(r.Name)
	if r.Rev != "" {
		path += "@" + r.Rev
	}
	return &url.URL{Path: path}
}

func (r *RepoMatch) AppendMatches(src *RepoMatch) {
	r.RepoNameMatches = append(r.RepoNameMatches, src.RepoNameMatches...)
}

func (r *RepoMatch) Key() Key {
	return Key{
		TypeRank: rankRepoMatch,
		Repo:     r.Name,
		Rev:      r.Rev,
	}
}

func (r *RepoMatch) searchResultMarker() {}
