package result

import "github.com/sourcegraph/sourcegraph/internal/api"

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
