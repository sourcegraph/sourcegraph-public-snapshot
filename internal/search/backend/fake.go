package backend

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sourcegraph/zoekt"
	zoektquery "github.com/sourcegraph/zoekt/query"
)

// FakeStreamer is a zoekt.Streamer that returns predefined search results
type FakeStreamer struct {
	Results     []*zoekt.SearchResult
	SearchError error

	Repos     []*zoekt.RepoListEntry
	ListError error

	// Default all unimplemented zoekt.Searcher methods to panic.
	zoekt.Searcher
}

// Search returns a single search result. If there is more than one predefined result, it concatenates
// their file lists together.
func (ss *FakeStreamer) Search(ctx context.Context, q zoektquery.Q, opts *zoekt.SearchOptions) (*zoekt.SearchResult, error) {
	if ss.SearchError != nil {
		return nil, ss.SearchError
	}

	res := &zoekt.SearchResult{}
	for _, result := range ss.Results {
		res.Files = append(res.Files, result.Files...)
		res.Stats.Add(result.Stats)
	}

	return res, nil
}

func (ss *FakeStreamer) StreamSearch(ctx context.Context, q zoektquery.Q, opts *zoekt.SearchOptions, z zoekt.Sender) error {
	if ss.SearchError != nil {
		return ss.SearchError
	}

	// Send out a stats-only event, to mimic a common approach in Zoekt
	z.Send(&zoekt.SearchResult{
		Stats: zoekt.Stats{
			Crashes: 0,
			Wait:    2 * time.Millisecond,
		},
		Progress: zoekt.Progress{
			MaxPendingPriority: 0,
		},
	})

	for _, r := range ss.Results {
		// Make sure to copy results before sending
		res := &zoekt.SearchResult{}
		res.Files = append(res.Files, r.Files...)
		res.Stats.Add(r.Stats)

		z.Send(res)
	}
	return nil
}

func (ss *FakeStreamer) List(ctx context.Context, q zoektquery.Q, opt *zoekt.ListOptions) (*zoekt.RepoList, error) {
	if ss.ListError != nil {
		return nil, ss.ListError
	}

	if opt == nil {
		opt = &zoekt.ListOptions{}
	}

	list := &zoekt.RepoList{}
	if opt.Field == zoekt.RepoListFieldReposMap {
		list.ReposMap = make(zoekt.ReposMap)
		for _, r := range ss.Repos {
			list.ReposMap[r.Repository.ID] = zoekt.MinimalRepoListEntry{
				HasSymbols: r.Repository.HasSymbols,
				Branches:   r.Repository.Branches,
			}
		}
	} else {
		list.Repos = ss.Repos
	}

	for _, r := range ss.Repos {
		list.Stats.Add(&r.Stats)
	}
	list.Stats.Repos = len(ss.Repos)

	return list, nil
}

func (ss *FakeStreamer) Close() {}

func (ss *FakeStreamer) Streamer() string {
	var parts []string
	if ss.Results != nil {
		parts = append(parts, fmt.Sprintf("Results = %v", ss.Results))
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
	return fmt.Sprintf("FakeStreamer(%s)", strings.Join(parts, ", "))
}
