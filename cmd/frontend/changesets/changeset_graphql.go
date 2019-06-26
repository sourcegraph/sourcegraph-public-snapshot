package changesets

import (
	"context"
	"encoding/json"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

// 🚨 SECURITY: TODO!(sqs): There are virtually no security checks here and they MUST be added.

// gqlChangeset implements the GraphQL type Changeset.
type gqlChangeset struct{ db *types.DiscussionThread }

func (GraphQLResolver) ChangesetFor(t *types.DiscussionThread) (graphqlbackend.Changeset, error) {
	return &gqlChangeset{t}, nil
}

type threadSettings struct {
	Deltas []struct {
		Repository graphql.ID
		Base, Head string
	}
}

func (v *gqlChangeset) getSettings() (*threadSettings, error) {
	if v.db.Settings == nil {
		return &threadSettings{}, nil
	}
	var settings threadSettings
	if err := json.Unmarshal([]byte(*v.db.Settings), &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

func (v *gqlChangeset) Repositories(ctx context.Context) ([]*graphqlbackend.RepositoryResolver, error) {
	settings, err := v.getSettings()
	if err != nil {
		return nil, err
	}

	rs := make([]*graphqlbackend.RepositoryResolver, len(settings.Deltas))
	for i, delta := range settings.Deltas {
		var err error
		rs[i], err = graphqlbackend.RepositoryByID(ctx, delta.Repository)
		if err != nil {
			return nil, err
		}
	}
	return rs, nil
}

func (v *gqlChangeset) Commits(ctx context.Context) ([]*graphqlbackend.GitCommitResolver, error) {
	rcs, err := v.RepositoryComparisons(ctx)
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

func (v *gqlChangeset) RepositoryComparisons(ctx context.Context) ([]*graphqlbackend.RepositoryComparisonResolver, error) {
	settings, err := v.getSettings()
	if err != nil {
		return nil, err
	}

	rcs := make([]*graphqlbackend.RepositoryComparisonResolver, len(settings.Deltas))
	for i, delta := range settings.Deltas {
		repo, err := graphqlbackend.RepositoryByID(ctx, delta.Repository)
		if err != nil {
			return nil, err
		}
		rcs[i], err = graphqlbackend.NewRepositoryComparison(ctx, repo, &graphqlbackend.RepositoryComparisonInput{
			Base: &delta.Base,
			Head: &delta.Head,
		})
		if err != nil {
			return nil, err
		}
	}
	return rcs, nil
}
