package result

import (
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type RepoMatch struct {
	Name api.RepoName
	ID   api.RepoID

	// rev optionally specifies a revision to go to for search results.
	Rev string
}

func (r RepoMatch) RepoName() types.RepoName {
	return types.RepoName{
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
	switch path.Type {
	case filter.Repository:
		return r
	}
	return nil
}

func (r *RepoMatch) URL() *url.URL {
	rawPath := "/" + escapePathForURL(string(r.Name))
	if r.Rev != "" {
		rawPath += "@" + escapePathForURL(r.Rev)
	}
	return &url.URL{RawPath: rawPath}
}

func (r *RepoMatch) searchResultMarker() {}

// TODO(camdencheek): put this in a proper shared place outside of graphqlbackend
// escapePathForURL escapes path (e.g. repository name, revspec) for use in a Sourcegraph URL.
// For niceness/readability, we do NOT escape slashes but we do escape other characters like '#'
// that are necessary for correctness.
func escapePathForURL(path string) string {
	return strings.ReplaceAll(url.PathEscape(path), "%2F", "/")
}
