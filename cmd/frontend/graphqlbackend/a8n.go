package graphqlbackend

import (
	"context"
	"database/sql"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/externallink"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
)

// NewA8NResolver will be set by enterprise
var NewA8NResolver func(*sql.DB) A8NResolver

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

type UpdateCampaignArgs struct {
	Input struct {
		ID          graphql.ID
		Name        *string
		Description *string
	}
}

type DeleteCampaignArgs struct {
	Campaign graphql.ID
}

type CreateChangesetsArgs struct {
	Input []struct {
		Repository graphql.ID
		ExternalID string
	}
}

type A8NResolver interface {
	CreateCampaign(ctx context.Context, args *CreateCampaignArgs) (CampaignResolver, error)
	UpdateCampaign(ctx context.Context, args *UpdateCampaignArgs) (CampaignResolver, error)
	CampaignByID(ctx context.Context, id graphql.ID) (CampaignResolver, error)
	Campaigns(ctx context.Context, args *graphqlutil.ConnectionArgs) (CampaignsConnectionResolver, error)
	DeleteCampaign(ctx context.Context, args *DeleteCampaignArgs) (*EmptyResponse, error)

	CreateChangesets(ctx context.Context, args *CreateChangesetsArgs) ([]ChangesetResolver, error)
	ChangesetByID(ctx context.Context, id graphql.ID) (ChangesetResolver, error)
	Changesets(ctx context.Context, args *graphqlutil.ConnectionArgs) (ChangesetsConnectionResolver, error)

	AddChangesetsToCampaign(ctx context.Context, args *AddChangesetsToCampaignArgs) (CampaignResolver, error)
}

var onlyInEnterprise = errors.New("campaigns and changesets are only available in enterprise")

func (r *schemaResolver) AddChangesetsToCampaign(ctx context.Context, args *AddChangesetsToCampaignArgs) (CampaignResolver, error) {
	if r.a8nResolver == nil {
		return nil, onlyInEnterprise
	}
	return r.a8nResolver.AddChangesetsToCampaign(ctx, args)
}

func (r *schemaResolver) CreateCampaign(ctx context.Context, args *CreateCampaignArgs) (CampaignResolver, error) {
	if r.a8nResolver == nil {
		return nil, onlyInEnterprise
	}
	return r.a8nResolver.CreateCampaign(ctx, args)
}

func (r *schemaResolver) UpdateCampaign(ctx context.Context, args *UpdateCampaignArgs) (CampaignResolver, error) {
	if r.a8nResolver == nil {
		return nil, onlyInEnterprise
	}
	return r.a8nResolver.UpdateCampaign(ctx, args)
}

func (r *schemaResolver) DeleteCampaign(ctx context.Context, args *DeleteCampaignArgs) (*EmptyResponse, error) {
	if r.a8nResolver == nil {
		return nil, onlyInEnterprise
	}
	return r.a8nResolver.DeleteCampaign(ctx, args)
}

func (r *schemaResolver) Campaigns(ctx context.Context, args *graphqlutil.ConnectionArgs) (CampaignsConnectionResolver, error) {
	if r.a8nResolver == nil {
		return nil, onlyInEnterprise
	}
	return r.a8nResolver.Campaigns(ctx, args)
}

func (r *schemaResolver) CreateChangesets(ctx context.Context, args *CreateChangesetsArgs) ([]ChangesetResolver, error) {
	if r.a8nResolver == nil {
		return nil, onlyInEnterprise
	}
	return r.a8nResolver.CreateChangesets(ctx, args)
}

func (r *schemaResolver) Changesets(ctx context.Context, args *graphqlutil.ConnectionArgs) (ChangesetsConnectionResolver, error) {
	if r.a8nResolver == nil {
		return nil, onlyInEnterprise
	}
	return r.a8nResolver.Changesets(ctx, args)
}

type ChangesetCountsArgs struct {
	From *DateTime
	To   *DateTime
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
	ChangesetCountsOverTime(ctx context.Context, args *ChangesetCountsArgs) ([]ChangesetCountsResolver, error)
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
	Events(ctx context.Context, args *struct{ graphqlutil.ConnectionArgs }) (ChangesetEventsConnectionResolver, error)
}

type ChangesetEventsConnectionResolver interface {
	Nodes(ctx context.Context) ([]ChangesetEventResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type ChangesetEventResolver interface {
	ID() graphql.ID
	Changeset(ctx context.Context) (ChangesetResolver, error)
	CreatedAt() DateTime
}

type ChangesetCountsResolver interface {
	Date() DateTime
	Total() int32
	Merged() int32
	Closed() int32
	Open() int32
	OpenApproved() int32
	OpenChangesRequested() int32
	OpenPending() int32
}
