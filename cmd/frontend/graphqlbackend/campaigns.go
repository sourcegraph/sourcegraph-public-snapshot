package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// Campaigns is the implementation of the GraphQL type CampaignsMutation. If it is not set at
// runtime, a "not implemented" error is returned to API clients who invoke it.
//
// This is contributed by enterprise.
var Campaigns CampaignsResolver

// CampaignByID is called to look up a Campaign given its GraphQL ID.
func CampaignByID(ctx context.Context, id graphql.ID) (Campaign, error) {
	if Campaigns == nil {
		return nil, errors.New("campaigns is not implemented")
	}
	return Campaigns.CampaignByID(ctx, id)
}

// CampaignsInNamespace returns an instance of the GraphQL CampaignConnection type with the list of
// campaigns defined in a namespace.
func CampaignsInNamespace(ctx context.Context, namespace graphql.ID, arg *graphqlutil.ConnectionArgs) (CampaignConnection, error) {
	if Campaigns == nil {
		return nil, errors.New("campaigns is not implemented")
	}
	return Campaigns.CampaignsInNamespace(ctx, namespace, arg)
}

func (schemaResolver) Campaigns() (CampaignsResolver, error) {
	if Campaigns == nil {
		return nil, errors.New("campaigns is not implemented")
	}
	return Campaigns, nil
}

// CampaignsResolver is the interface for the GraphQL type CampaignsMutation.
type CampaignsResolver interface {
	CreateCampaign(context.Context, *CreateCampaignArgs) (Campaign, error)
	UpdateCampaign(context.Context, *UpdateCampaignArgs) (Campaign, error)
	DeleteCampaign(context.Context, *DeleteCampaignArgs) (*EmptyResponse, error)
	AddThreadsToCampaign(context.Context, *AddRemoveThreadsToFromCampaignArgs) (*EmptyResponse, error)
	RemoveThreadsFromCampaign(context.Context, *AddRemoveThreadsToFromCampaignArgs) (*EmptyResponse, error)

	// CampaignByID is called by the CampaignByID func but is not in the GraphQL API.
	CampaignByID(context.Context, graphql.ID) (Campaign, error)

	// CampaignsInNamespace is called by the CampaignsInNamespace func but is not in the GraphQL
	// API.
	CampaignsInNamespace(ctx context.Context, namespace graphql.ID, arg *graphqlutil.ConnectionArgs) (CampaignConnection, error)
}

type CreateCampaignArgs struct {
	Input struct {
		Namespace   graphql.ID
		Name        string
		Description *string
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

type AddRemoveThreadsToFromCampaignArgs struct {
	Campaign graphql.ID
	Threads  []graphql.ID
}

// Campaign is the interface for the GraphQL type Campaign.
type Campaign interface {
	ID() graphql.ID
	Namespace(context.Context) (*NamespaceResolver, error)
	Name() string
	Description() *string
	URL(context.Context) (string, error)
	Threads(context.Context, *graphqlutil.ConnectionArgs) (DiscussionThreadConnection, error)
	Repositories(context.Context) ([]*RepositoryResolver, error)
	Commits(context.Context) ([]*GitCommitResolver, error)
	RepositoryComparisons(context.Context) ([]*RepositoryComparisonResolver, error)
}

// CampaignConnection is the interface for the GraphQL type CampaignConnection.
type CampaignConnection interface {
	Nodes(context.Context) ([]Campaign, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}
