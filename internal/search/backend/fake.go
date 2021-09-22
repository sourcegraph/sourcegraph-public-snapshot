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
	sr, _ := ss.Search(ctx, q, opts)
	z.Send(sr)
	return nil
}

func (ss *FakeSearcher) List(ctx context.Context, q zoektquery.Q, opt *zoekt.ListOptions) (*zoekt.RepoList, error) {
	list := &zoekt.RepoList{}
	if opt != nil && opt.Minimal {
		list.Minimal = make(map[uint32]*zoekt.MinimalRepoListEntry, len(ss.Repos))
		for _, r := range ss.Repos {
			list.Minimal[r.Repository.ID] = &zoekt.MinimalRepoListEntry{
				HasSymbols: r.Repository.HasSymbols,
				Branches:   r.Repository.Branches,
			}
		}
	} else {
		list.Repos = ss.Repos
	}

	return list, nil
}

func (ss *FakeSearcher) String() string {
	return fmt.Sprintf("FakeSearcher(Result = %v, Repos = %v)", ss.Result, ss.Repos)
}
