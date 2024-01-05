package graphqlbackend

import (
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

// CommitSearchResultResolver is a resolver for the GraphQL type `CommitSearchResult`
type CommitSearchResultResolver struct {
	result.CommitMatch

	db database.DB

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
		gitserverClient := gitserver.NewClient("graphql.search.commitresult")
		repoResolver := NewRepositoryResolver(r.db, gitserverClient, r.Repo.ToRepo())
		r.gitCommitResolver = NewGitCommitResolver(r.db, gitserverClient, repoResolver, r.CommitMatch.Commit.ID, &r.CommitMatch.Commit)
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
	return &highlightedStringResolver{r.CommitMatch.MessagePreview.ToHighlightedString()}
}

func (r *CommitSearchResultResolver) DiffPreview() *highlightedStringResolver {
	if r.CommitMatch.DiffPreview == nil {
		return nil
	}
	return &highlightedStringResolver{r.CommitMatch.DiffPreview.ToHighlightedString()}
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
	hls := r.CommitMatch.Body().ToHighlightedString()
	match := &searchResultMatchResolver{
		body:       hls.Value,
		highlights: hls.Highlights,
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
