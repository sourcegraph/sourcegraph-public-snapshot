package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// ChangesetCampaigns is the implementation of the GraphQL type ChangesetCampaignsMutation. If it is
// not set at runtime, a "not implemented" error is returned to API clients who invoke it.
//
// This is contributed by enterprise.
var ChangesetCampaigns ChangesetCampaignsResolver

// ChangesetCampaignByID is called to look up a ChangesetCampaign given its GraphQL ID.
func ChangesetCampaignByID(ctx context.Context, id graphql.ID) (ChangesetCampaign, error) {
	if ChangesetCampaigns == nil {
		return nil, errors.New("changesetCampaigns is not implemented")
	}
	return ChangesetCampaigns.ChangesetCampaignByID(ctx, id)
}

// ChangesetCampaignsDefinedIn returns an instance of the GraphQL ChangesetCampaignConnection type
// with the list of changeset campaigns defined in a project.
func ChangesetCampaignsDefinedIn(ctx context.Context, project graphql.ID, arg *graphqlutil.ConnectionArgs) (ChangesetCampaignConnection, error) {
	if ChangesetCampaigns == nil {
		return nil, errors.New("changesetCampaigns is not implemented")
	}
	return ChangesetCampaigns.ChangesetCampaignsDefinedIn(ctx, project, arg)
}

func (schemaResolver) ChangesetCampaigns() (ChangesetCampaignsResolver, error) {
	if ChangesetCampaigns == nil {
		return nil, errors.New("changesetCampaigns is not implemented")
	}
	return ChangesetCampaigns, nil
}

// ChangesetCampaignsResolver is the interface for the GraphQL type ChangesetCampaignsMutation.
type ChangesetCampaignsResolver interface {
	CreateChangesetCampaign(context.Context, *CreateChangesetCampaignArgs) (ChangesetCampaign, error)
	UpdateChangesetCampaign(context.Context, *UpdateChangesetCampaignArgs) (ChangesetCampaign, error)
	DeleteChangesetCampaign(context.Context, *DeleteChangesetCampaignArgs) (*EmptyResponse, error)

	// ChangesetCampaignByID is called by the ChangesetCampaignByID func but is not in the GraphQL
	// API.
	ChangesetCampaignByID(context.Context, graphql.ID) (ChangesetCampaign, error)

	// ChangesetCampaignsDefinedIn is called by the ChangesetCampaignsDefinedIn func but is not in
	// the GraphQL API.
	ChangesetCampaignsDefinedIn(ctx context.Context, project graphql.ID, arg *graphqlutil.ConnectionArgs) (ChangesetCampaignConnection, error)
}

type CreateChangesetCampaignArgs struct {
	Input struct {
		Project     graphql.ID
		Name        string
		Description *string
	}
}

type UpdateChangesetCampaignArgs struct {
	Input struct {
		ID          graphql.ID
		Name        *string
		Description *string
	}
}

type DeleteChangesetCampaignArgs struct {
	ChangesetCampaign graphql.ID
}

// ChangesetCampaign is the interface for the GraphQL type ChangesetCampaign.
type ChangesetCampaign interface {
	ID() graphql.ID
	Project(context.Context) (Project, error)
	Name() string
	Description() *string
	URL() string
}

// ChangesetCampaignConnection is the interface for the GraphQL type ChangesetCampaignConnection.
type ChangesetCampaignConnection interface {
	Nodes(context.Context) ([]ChangesetCampaign, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}
