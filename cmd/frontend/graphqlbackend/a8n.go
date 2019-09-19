package graphqlbackend

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/externallink"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/pkg/a8n"
)

type AddChangesetsToCampaignArgs struct {
	Campaign   graphql.ID
	Changesets []graphql.ID
}

type CreateCampaignArgs struct {
	Input struct {
		Namespace   graphql.ID
		Name        string
		Description string
	}
}

type CreateChangesetsArgs struct {
	Input []struct {
		Repository graphql.ID
		ExternalID string
	}
}

type A8NResolver interface {
	CampaignByID(ctx context.Context, id graphql.ID) (CampaignResolver, error)
	ChangesetByID(ctx context.Context, id graphql.ID) (ChangesetResolver, error)

	AddChangesetsToCampaign(ctx context.Context, args *AddChangesetsToCampaignArgs) (CampaignResolver, error)
	CreateCampaign(ctx context.Context, args *CreateCampaignArgs) (CampaignResolver, error)
	Campaigns(ctx context.Context, args *graphqlutil.ConnectionArgs) (CampaignsConnectionResolver, error)
	CreateChangesets(ctx context.Context, args *CreateChangesetsArgs) ([]ChangesetResolver, error)
	Changesets(ctx context.Context, args *graphqlutil.ConnectionArgs) (ChangesetsConnectionResolver, error)
}

type CampaignResolver interface {
	ID() graphql.ID
	Name() string
	Description() string
	Author(ctx context.Context) (*UserResolver, error)
	URL(ctx context.Context) (string, error)
	Namespace(ctx context.Context) (n NamespaceResolver, err error)
	CreatedAt() DateTime
	UpdatedAt() DateTime
	Changesets(ctx context.Context, args struct{ graphqlutil.ConnectionArgs }) ChangesetsConnectionResolver
}

type CampaignsConnectionResolver interface {
	Nodes(ctx context.Context) ([]CampaignResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type ChangesetsConnectionResolver interface {
	Nodes(ctx context.Context) ([]ChangesetResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type ChangesetResolver interface {
	ID() graphql.ID
	CreatedAt() DateTime
	UpdatedAt() DateTime
	Title() (string, error)
	Body() (string, error)
	State() (a8n.ChangesetState, error)
	ExternalURL() (*externallink.Resolver, error)
	ReviewState() (a8n.ChangesetReviewState, error)
	Repository(ctx context.Context) (*RepositoryResolver, error)
	Campaigns(ctx context.Context, args *struct{ graphqlutil.ConnectionArgs }) (CampaignsConnectionResolver, error)
}
