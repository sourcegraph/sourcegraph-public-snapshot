package campaigns

import (
	"context"
	"sort"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// gqlCampaignThreadDerived implements the fields/methods of the Campaign and CampaignPrevie GraphQL
// types that are derived from the campaign's set of threads.
type gqlCampaignThreadDerived struct {
	getThreads func(context.Context) ([]*graphqlbackend.Thread, error)
}

func campaignRepositories(ctx context.Context, threads []graphqlbackend.Thread) ([]*graphqlbackend.RepositoryResolver, error) {
	byRepositoryDBID := map[api.RepoID]*graphqlbackend.RepositoryResolver{}
	for _, thread := range threadNodes {
		repo, err := thread.Repository(ctx)
		if err != nil {
			return nil, err
		}
		key := repo.DBID()
		if _, seen := byRepositoryDBID[key]; !seen {
			byRepositoryDBID[key] = repo
		}
	}

	repos := make([]*graphqlbackend.RepositoryResolver, 0, len(byRepositoryDBID))
	for _, repo := range byRepositoryDBID {
		repos = append(repos, repo)
	}
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].DBID() < repos[j].DBID()
	})
	return repos, nil
}

func campaignCommits(ctx context.Context, threads []graphqlbackend.Thread) ([]*graphqlbackend.GitCommitResolver, error) {
	rcs, err := campaignRepositoryComparisons(ctx, threads)
	if err != nil {
		return nil, err
	}

	var allCommits []*graphqlbackend.GitCommitResolver
	for _, rc := range rcs {
		cc := rc.Commits(&graphqlutil.ConnectionArgs{})
		commits, err := cc.Nodes(ctx)
		if err != nil {
			return nil, err
		}
		allCommits = append(allCommits, commits...)
	}
	return allCommits, nil
}

func campaignRepositoryComparisons(ctx context.Context, threads []graphqlbackend.Thread) ([]*graphqlbackend.RepositoryComparisonResolver, error) {
	rcs := make([]*graphqlbackend.RepositoryComparisonResolver, 0, len(threads))
	for _, thread := range threads {
		rc, err := thread.RepositoryComparison(ctx)
		if err != nil {
			return nil, err
		}
		if rc != nil {
			rcs = append(rcs, rc)
		}
	}
	return rcs, nil
}
