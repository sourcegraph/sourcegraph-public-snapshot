package threads

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func newThreadConnectionFiltersFromDB(ctx context.Context, opt dbThreadsListOptions) (graphqlbackend.ThreadConnectionFilters, error) {
	opt.LimitOffset = nil
	// TODO!(sqs) security respect repo perms
	allThreads, err := dbThreads{}.List(ctx, opt)
	if err != nil {
		return nil, err
	}
	return &threadConnectionFilters{ToThreadOrThreadPreviews(toThreads(allThreads), nil)}, nil
}

func newThreadConnectionFiltersFromConst(allThreads []graphqlbackend.ToThreadOrThreadPreview) graphqlbackend.ThreadConnectionFilters {
	return &threadConnectionFilters{allThreads: allThreads}
}

type threadConnectionFilters struct {
	allThreads []graphqlbackend.ToThreadOrThreadPreview
}

func (f *threadConnectionFilters) Repository(ctx context.Context) ([]graphqlbackend.RepositoryFilter, error) {
	// TODO!(sqs) security respect repo perms
	repos := map[api.RepoID]int32{}
	for _, t := range f.allThreads {
		repos[t.Common().Internal_RepositoryID()]++
	}

	filters := make([]graphqlbackend.RepositoryFilter, 0, len(repos))
	for repoID, count := range repos {
		repo, err := graphqlbackend.RepositoryByDBID(ctx, repoID)
		if err != nil {
			return nil, err
		}
		filters = append(filters, graphqlbackend.RepositoryFilter{Repository_: repo, Count_: count})
	}
	return filters, nil
}
