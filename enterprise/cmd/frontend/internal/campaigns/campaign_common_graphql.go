package campaigns

import (
	"context"
	"sort"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

type threadsGetter interface {
	// getThreads returns a list of threads (and thread previews) in the campaign.
	getThreads(ctx context.Context) ([]graphqlbackend.ToThreadOrThreadPreview, error)
}

// gqlCampaignThreadDerived implements the fields/methods of the Campaign and CampaignPrevie GraphQL
// types that are derived from the campaign's set of threads.
type gqlCampaignThreadDerived struct {
	getThreads func(context.Context) ([]*graphqlbackend.Thread, error)
}

func campaignRepositories(ctx context.Context, campaign threadsGetter) ([]*graphqlbackend.RepositoryResolver, error) {
	threads, err := campaign.getThreads(ctx)
	if err != nil {
		return nil, err
	}

	byRepositoryDBID := map[api.RepoID]*graphqlbackend.RepositoryResolver{}
	for _, thread := range threads {
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

func campaignCommits(ctx context.Context, campaign threadsGetter) ([]*graphqlbackend.GitCommitResolver, error) {
	rcs, err := campaignRepositoryComparisons(ctx, campaign)
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

func campaignRepositoryComparisons(ctx context.Context, campaign threadsGetter) ([]graphqlbackend.RepositoryComparison, error) {
	threads, err := campaign.getThreads(ctx)
	if err != nil {
		return nil, err
	}

	rcs := make([]graphqlbackend.RepositoryComparison, 0, len(threads))
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
