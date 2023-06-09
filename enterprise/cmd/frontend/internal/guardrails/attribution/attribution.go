package attribution

import (
	"context"
	"fmt"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Service is an attribution service which searches for matches on snippets of
// code.
type Service struct {
	// SearchClient is used to find attribution on the local instance.
	SearchClient client.SearchClient
}

// SnippetAttributions is holds the collection of attributions for a snippet.
type SnippetAttributions struct {
	// RepositoryNames is the list of repository names. We intend on mixing
	// names from both the local instance as well as from sourcegraph.com. So
	// we intentionally use a string since the name may not represent a
	// repository available on this instance.
	//
	// Note: for now this is a simple slice, we likely will expand what is
	// represented here and it will change into a struct capturing more
	// information.
	RepositoryNames []string

	// TotalCount is the total number of repository attributions we found
	// before stopping the search.
	//
	// Note: if we didn't finish searching the full corpus then LimitHit will
	// be true. For filtering use case this means if LimitHit is true you need
	// to be conservative with TotalCount and assume it could be higher.
	TotalCount int

	// LimitHit is true if we stopped searching before looking into the full
	// corpus. If LimitHit is true then it is possible there are more than
	// TotalCount attributions.
	LimitHit bool
}

// SnippetAttribution will search the instances indexed code for code matching
// snippet and return the attribution results.
func (c *Service) SnippetAttribution(ctx context.Context, snippet string, limit int) (*SnippetAttributions, error) {
	const (
		version    = "V3"
		searchMode = search.Precise
		protocol   = search.Batch
	)

	patternType := "literal"
	searchQuery := fmt.Sprintf("type:file select:repo index:only count:%d content:%q", limit, snippet)

	inputs, err := c.SearchClient.Plan(
		ctx,
		version,
		&patternType,
		searchQuery,
		searchMode,
		protocol,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create search plan")
	}

	// TODO(keegancsmith) Reading the SearchClient code it seems to miss out
	// on some of the observability that we instead add in at a later stage.
	// For example the search dataset in honeycomb will be missing. Will have
	// to follow-up with observability and maybe solve it for all users.
	//
	// Note: In our current API we could just store repo names in seen. But it
	// is safer to rely on searches ranking for result stability than doing
	// something like sorting by name from the map.
	var (
		mu        sync.Mutex
		seen      map[api.RepoID]struct{}
		repoNames []string
		limitHit  bool
	)
	_, err = c.SearchClient.Execute(ctx, streaming.StreamFunc(func(ev streaming.SearchEvent) {
		mu.Lock()
		defer mu.Unlock()

		limitHit = limitHit || ev.Stats.IsLimitHit

		for _, m := range ev.Results {
			repo := m.RepoName()
			if _, ok := seen[repo.ID]; ok {
				continue
			}
			seen[repo.ID] = struct{}{}
			repoNames = append(repoNames, string(repo.Name))
		}
	}), inputs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute search")
	}

	// Note: Our search API is missing total count internally, but Zoekt does
	// expose this. For now we just count what we found.
	totalCount := len(repoNames)
	if len(repoNames) > limit {
		repoNames = repoNames[:limit]
	}

	return &SnippetAttributions{
		RepositoryNames: repoNames,
		TotalCount:      totalCount,
		LimitHit:        limitHit,
	}, nil
}
