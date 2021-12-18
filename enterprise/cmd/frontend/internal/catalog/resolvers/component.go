package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/catalog"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

type componentResolver struct {
	component catalog.Component
	db        database.DB
}

func (r *componentResolver) ID() graphql.ID {
	return relay.MarshalID("Component", r.component.Name) // TODO(sqs)
}

func (r *componentResolver) Name() string {
	return r.component.Name
}

func (r *componentResolver) Description() *string {
	if r.component.Description == "" {
		return nil
	}
	return &r.component.Description
}

func (r *componentResolver) Lifecycle() gql.ComponentLifecycle {
	return gql.ComponentLifecycle(r.component.Lifecycle)
}

func (r *componentResolver) URL() string {
	return "/catalog/components/" + string(r.Name())
}

func (r *componentResolver) Kind() gql.ComponentKind {
	return gql.ComponentKind(r.component.Kind)
}

func (r *componentResolver) sourceRepoResolver(ctx context.Context) (*gql.RepositoryResolver, error) {
	// 🚨 SECURITY: database.Repos.Get uses the authzFilter under the hood and
	// filters out repositories that the user doesn't have access to.
	repo, err := r.db.Repos().GetByName(ctx, r.component.SourceRepo)
	if err != nil {
		return nil, err
	}

	return gql.NewRepositoryResolver(r.db, repo), nil
}

func (r *componentResolver) sourceCommitResolver(ctx context.Context) (*gql.GitCommitResolver, error) {
	repoResolver, err := r.sourceRepoResolver(ctx)
	if err != nil {
		return nil, err
	}
	return gql.NewGitCommitResolver(r.db, repoResolver, api.CommitID(r.component.SourceCommit), nil), nil
}

func (r *componentResolver) SourceLocations(ctx context.Context) ([]*gql.GitTreeEntryResolver, error) {
	commitResolver, err := r.sourceCommitResolver(ctx)
	if err != nil {
		return nil, err
	}
	var locs []*gql.GitTreeEntryResolver
	for _, sourcePath := range r.component.SourcePaths {
		locs = append(locs, gql.NewGitTreeEntryResolver(r.db, commitResolver, gql.CreateFileInfo(sourcePath, false)))
	}
	return locs, nil
}
