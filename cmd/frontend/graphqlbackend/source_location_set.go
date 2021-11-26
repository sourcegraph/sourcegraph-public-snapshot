package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

type SourceLocationSet interface {
	ID() graphql.ID
	Readme(context.Context) (FileResolver, error)
	Commits(context.Context, *graphqlutil.ConnectionArgs) (GitCommitConnectionResolver, error)
	Branches(context.Context, *GitRefConnectionArgs) (GitRefConnectionResolver, error)
	CodeOwners(context.Context, *graphqlutil.ConnectionArgs) (CodeOwnerConnectionResolver, error)
	Contributors(context.Context, *graphqlutil.ConnectionArgs) (ContributorConnectionResolver, error)
	Usage(context.Context) (ComponentUsageResolver, error)
	WhoKnows(context.Context, *WhoKnowsArgs) ([]WhoKnowsEdgeResolver, error)
	Cyclonedx(context.Context) (*string, error)
}

func (r *GitTreeEntryResolver) Commits(ctx context.Context, args *graphqlutil.ConnectionArgs) (GitCommitConnectionResolver, error) {
	return EnterpriseResolvers.catalogRootResolver.GitTreeEntryCommits(ctx, r, args)
}

func (r *GitTreeEntryResolver) Branches(ctx context.Context, args *GitRefConnectionArgs) (GitRefConnectionResolver, error) {
	return EnterpriseResolvers.catalogRootResolver.GitTreeEntryBranches(ctx, r, args)
}

func (r *GitTreeEntryResolver) Readme(ctx context.Context) (FileResolver, error) {
	return EnterpriseResolvers.catalogRootResolver.GitTreeEntryReadme(ctx, r)
}

func (r *GitTreeEntryResolver) CodeOwners(ctx context.Context, args *graphqlutil.ConnectionArgs) (CodeOwnerConnectionResolver, error) {
	return EnterpriseResolvers.catalogRootResolver.GitTreeEntryCodeOwners(ctx, r, args)
}

func (r *GitTreeEntryResolver) Contributors(ctx context.Context, args *graphqlutil.ConnectionArgs) (ContributorConnectionResolver, error) {
	return EnterpriseResolvers.catalogRootResolver.GitTreeEntryContributors(ctx, r, args)
}

func (r *GitTreeEntryResolver) Usage(ctx context.Context) (ComponentUsageResolver, error) {
	return EnterpriseResolvers.catalogRootResolver.GitTreeEntryUsage(ctx, r)
}

func (r *GitTreeEntryResolver) WhoKnows(ctx context.Context, args *WhoKnowsArgs) ([]WhoKnowsEdgeResolver, error) {
	return EnterpriseResolvers.catalogRootResolver.GitTreeEntryWhoKnows(ctx, r, args)
}

func (r *GitTreeEntryResolver) Cyclonedx(ctx context.Context) (*string, error) {
	return EnterpriseResolvers.catalogRootResolver.GitTreeEntryCyclonedx(ctx, r)
}
