package graphqlbackend

import (
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

// CommitSearchResultResolver is a resolver for the GraphQL type `CommitSearchResult`
type CommitSearchResultResolver struct {
	result.CommitMatch

	db dbutil.DB

	// gitCommitResolver should not be used directly since it may be uninitialized.
	// Use Commit() instead.
	gitCommitResolver *GitCommitResolver
	gitCommitOnce     sync.Once
}

func (r *CommitSearchResultResolver) Commit() *GitCommitResolver {
	r.gitCommitOnce.Do(func() {
		if r.gitCommitResolver != nil {
			return
		}
		repoResolver := NewRepositoryResolver(r.db, r.RepoName.ToRepo())
		r.gitCommitResolver = toGitCommitResolver(repoResolver, r.db, r.CommitMatch.Commit.ID, &r.CommitMatch.Commit)
	})
	return r.gitCommitResolver
}

func (r *CommitSearchResultResolver) Refs() []*GitRefResolver {
	out := make([]*GitRefResolver, 0, len(r.CommitMatch.Refs))
	for _, ref := range r.CommitMatch.Refs {
		out = append(out, &GitRefResolver{
			repo: r.Commit().Repository(),
			name: ref,
		})
	}
	return out
}

func (r *CommitSearchResultResolver) SourceRefs() []*GitRefResolver {
	out := make([]*GitRefResolver, 0, len(r.CommitMatch.SourceRefs))
	for _, ref := range r.CommitMatch.SourceRefs {
		out = append(out, &GitRefResolver{
			repo: r.Commit().Repository(),
			name: ref,
		})
	}
	return out
}

func (r *CommitSearchResultResolver) MessagePreview() *highlightedStringResolver {
	if r.CommitMatch.MessagePreview == nil {
		return nil
	}
	return &highlightedStringResolver{*r.CommitMatch.MessagePreview}
}

func (r *CommitSearchResultResolver) DiffPreview() *highlightedStringResolver {
	if r.CommitMatch.DiffPreview == nil {
		return nil
	}
	return &highlightedStringResolver{*r.CommitMatch.DiffPreview}
}

func (r *CommitSearchResultResolver) Icon() string {
	return "data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0iVVRGLTgiPz48IURPQ1RZUEUgc3ZnIFBVQkxJQyAiLS8vVzNDLy9EVEQgU1ZHIDEuMS8vRU4iICJodHRwOi8vd3d3LnczLm9yZy9HcmFwaGljcy9TVkcvMS4xL0RURC9zdmcxMS5kdGQiPjxzdmcgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIiB4bWxuczp4bGluaz0iaHR0cDovL3d3dy53My5vcmcvMTk5OS94bGluayIgdmVyc2lvbj0iMS4xIiB3aWR0aD0iMjQiIGhlaWdodD0iMjQiIHZpZXdCb3g9IjAgMCAyNCAyNCI+PHBhdGggZD0iTTE3LDEyQzE3LDE0LjQyIDE1LjI4LDE2LjQ0IDEzLDE2LjlWMjFIMTFWMTYuOUM4LjcyLDE2LjQ0IDcsMTQuNDIgNywxMkM3LDkuNTggOC43Miw3LjU2IDExLDcuMVYzSDEzVjcuMUMxNS4yOCw3LjU2IDE3LDkuNTggMTcsMTJNMTIsOUEzLDMgMCAwLDAgOSwxMkEzLDMgMCAwLDAgMTIsMTVBMywzIDAgMCwwIDE1LDEyQTMsMyAwIDAsMCAxMiw5WiIgLz48L3N2Zz4="
}

func (r *CommitSearchResultResolver) Label() Markdown {
	return Markdown(r.CommitMatch.Label())
}

func (r *CommitSearchResultResolver) URL() string {
	return r.CommitMatch.URL().String()
}

func (r *CommitSearchResultResolver) Detail() Markdown {
	return Markdown(r.CommitMatch.Detail())
}

func (r *CommitSearchResultResolver) Matches() []*searchResultMatchResolver {
	match := &searchResultMatchResolver{
		body:       r.CommitMatch.Body.Value,
		highlights: r.CommitMatch.Body.Highlights,
		url:        r.Commit().URL(),
	}
	matches := []*searchResultMatchResolver{match}
	return matches
}

func (r *CommitSearchResultResolver) ToRepository() (*RepositoryResolver, bool) { return nil, false }
func (r *CommitSearchResultResolver) ToFileMatch() (*FileMatchResolver, bool)   { return nil, false }
func (r *CommitSearchResultResolver) ToCommitSearchResult() (*CommitSearchResultResolver, bool) {
	return r, true
}

func (r *CommitSearchResultResolver) ResultCount() int32 {
	return 1
}
