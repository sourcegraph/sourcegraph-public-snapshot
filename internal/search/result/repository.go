package result

import "github.com/sourcegraph/sourcegraph/internal/api"

type RepoMatch struct {
	Name api.RepoName
	ID   api.RepoID

	// rev optionally specifies a revision to go to for search results.
	Rev string
}
