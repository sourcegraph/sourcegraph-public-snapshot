package backend

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/zoekt"
	zoektquery "github.com/sourcegraph/zoekt/query"
)

// FakeSearcher is a zoekt.Searcher that returns a predefined search Result.
type FakeSearcher struct {
	Result      *zoekt.SearchResult
	SearchError error

	Repos     []*zoekt.RepoListEntry
	ListError error

	// Default all unimplemented zoekt.Searcher methods to panic.
	zoekt.Searcher
}

func (ss *FakeSearcher) Search(ctx context.Context, q zoektquery.Q, opts *zoekt.SearchOptions) (*zoekt.SearchResult, error) {
	if ss.SearchError != nil {
		return nil, ss.SearchError
	}
	res := ss.Result
	if res == nil {
		res = &zoekt.SearchResult{}
	}

	// Copy since downstream could mutate
	sr := *res
	sr.Files = append([]zoekt.FileMatch{}, sr.Files...)
	res = &sr

	return res, nil
}

func (ss *FakeSearcher) StreamSearch(ctx context.Context, q zoektquery.Q, opts *zoekt.SearchOptions, z zoekt.Sender) error {
	sr, err := ss.Search(ctx, q, opts)
	if err != nil {
		return err
	}
	z.Send(sr)
	return nil
}

func (ss *FakeSearcher) List(ctx context.Context, q zoektquery.Q, opt *zoekt.ListOptions) (*zoekt.RepoList, error) {
	if ss.ListError != nil {
		return nil, ss.ListError
	}

	list := &zoekt.RepoList{}
	if opt != nil && opt.Minimal { //nolint:staticcheck // See https://github.com/sourcegraph/sourcegraph/issues/45814
		list.Minimal = make(map[uint32]*zoekt.MinimalRepoListEntry, len(ss.Repos)) //nolint:staticcheck // See https://github.com/sourcegraph/sourcegraph/issues/45814
		for _, r := range ss.Repos {
			list.Minimal[r.Repository.ID] = &zoekt.MinimalRepoListEntry{ //nolint:staticcheck // See https://github.com/sourcegraph/sourcegraph/issues/45814
				HasSymbols: r.Repository.HasSymbols,
				Branches:   r.Repository.Branches,
			}
		}
	} else {
		list.Repos = ss.Repos
	}

	return list, nil
}

func (ss *FakeSearcher) Close() {}

func (ss *FakeSearcher) String() string {
	var parts []string
	if ss.Result != nil {
		parts = append(parts, fmt.Sprintf("Result = %v", ss.Result))
	}
	if ss.Repos != nil {
		parts = append(parts, fmt.Sprintf("Repos = %v", ss.Repos))
	}
	if ss.SearchError != nil {
		parts = append(parts, fmt.Sprintf("SearchError = %v", ss.SearchError))
	}
	if ss.ListError != nil {
		parts = append(parts, fmt.Sprintf("ListError = %v", ss.ListError))
	}
	return fmt.Sprintf("FakeSearcher(%s)", strings.Join(parts, ", "))
}
