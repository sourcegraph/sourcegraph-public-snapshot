package resolvers

import (
	"context"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

func (r *Resolver) CreateCampaignsUserToken(ctx context.Context, args *graphqlbackend.CreateCampaignsUserTokenArgs) (graphqlbackend.CampaignsUserTokenResolver, error) {
	return &campaignsUserTokenResolver{}, errors.New("not implemented")
}

func (r *Resolver) DeleteCampaignsUserToken(ctx context.Context, args *graphqlbackend.DeleteCampaignsUserTokenArgs) (*graphqlbackend.EmptyResponse, error) {
	return &graphqlbackend.EmptyResponse{}, errors.New("not implemented")
}

func (r *Resolver) CampaignsUserTokenByID(ctx context.Context, id graphql.ID) (graphqlbackend.CampaignsUserTokenResolver, error) {
	return &campaignsUserTokenResolver{}, nil
}

func (r *Resolver) ConfiguredExternalServices(ctx context.Context, userID graphql.ID) (graphqlbackend.ConfiguredExternalServicesConnectionResolver, error) {
	return &configuredExternalServicesConnectionResolver{}, nil
}

type campaignsUserTokenResolver struct{}

var _ graphqlbackend.CampaignsUserTokenResolver = &campaignsUserTokenResolver{}

func (c *campaignsUserTokenResolver) ID() graphql.ID {
	return graphql.ID("stub")
}

func (c *campaignsUserTokenResolver) ExternalServiceKind() string {
	return extsvc.KindGitHub
}

func (c *campaignsUserTokenResolver) ExternalServiceURL() string {
	return "https://github.com/"
}

func (c *campaignsUserTokenResolver) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: time.Now()}
}

type configuredExternalServiceResolver struct{}

var _ graphqlbackend.ConfiguredExternalServiceResolver = &configuredExternalServiceResolver{}

func (c *configuredExternalServiceResolver) ExternalServiceKind() string {
	return extsvc.KindGitHub
}

func (c *configuredExternalServiceResolver) ExternalServiceURL() string {
	return "https://github.com/"
}

func (c *configuredExternalServiceResolver) ConfiguredToken() graphqlbackend.CampaignsUserTokenResolver {
	return &campaignsUserTokenResolver{}
}

type configuredExternalServicesConnectionResolver struct{}

var _ graphqlbackend.ConfiguredExternalServicesConnectionResolver = &configuredExternalServicesConnectionResolver{}

func (c *configuredExternalServicesConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return 1, nil
}

func (c *configuredExternalServicesConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
}

func (c *configuredExternalServicesConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.ConfiguredExternalServiceResolver, error) {
	return []graphqlbackend.ConfiguredExternalServiceResolver{&configuredExternalServiceResolver{}}, nil
}
