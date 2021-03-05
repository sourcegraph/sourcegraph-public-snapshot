package backend

import (
	"context"
	"fmt"

	"github.com/google/zoekt"
	zoektquery "github.com/google/zoekt/query"
)

// FakeSearcher is a zoekt.Searcher that returns a predefined search Result.
type FakeSearcher struct {
	Result *zoekt.SearchResult

	Repos []*zoekt.RepoListEntry

	// Default all unimplemented zoekt.Searcher methods to panic.
	zoekt.Searcher
}

func (ss *FakeSearcher) Search(ctx context.Context, q zoektquery.Q, opts *zoekt.SearchOptions) (*zoekt.SearchResult, error) {
	if ss.Result == nil {
		return &zoekt.SearchResult{}, nil
	}
	return ss.Result, nil
}

func (ss *FakeSearcher) StreamSearch(ctx context.Context, q zoektquery.Q, opts *zoekt.SearchOptions, z zoekt.Sender) error {
	return (&StreamSearchAdapter{ss}).StreamSearch(ctx, q, opts, z)
}

func (ss *FakeSearcher) List(ctx context.Context, q zoektquery.Q) (*zoekt.RepoList, error) {
	return &zoekt.RepoList{Repos: ss.Repos}, nil
}

func (ss *FakeSearcher) String() string {
	return fmt.Sprintf("FakeSearcher(Result = %v, Repos = %v)", ss.Result, ss.Repos)
}
