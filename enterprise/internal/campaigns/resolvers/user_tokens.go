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

func (r *Resolver) CreateCampaignsCredential(ctx context.Context, args *graphqlbackend.CreateCampaignsCredentialArgs) (graphqlbackend.CampaignsCredentialResolver, error) {
	return &campaignsCredentialResolver{}, errors.New("not implemented")
}

func (r *Resolver) DeleteCampaignsCredential(ctx context.Context, args *graphqlbackend.DeleteCampaignsCredentialArgs) (*graphqlbackend.EmptyResponse, error) {
	return &graphqlbackend.EmptyResponse{}, errors.New("not implemented")
}

func (r *Resolver) CampaignsCredentialByID(ctx context.Context, id graphql.ID) (graphqlbackend.CampaignsCredentialResolver, error) {
	return &campaignsCredentialResolver{}, nil
}

func (r *Resolver) CampaignsCodeHosts(ctx context.Context, args *graphqlbackend.ListCampaignsCodeHostsArgs) (graphqlbackend.CampaignsCodeHostConnectionResolver, error) {
	return &campaignsCodeHostConnectionResolver{}, nil
}

type campaignsCredentialResolver struct{}

var _ graphqlbackend.CampaignsCredentialResolver = &campaignsCredentialResolver{}

func (c *campaignsCredentialResolver) ID() graphql.ID {
	return graphql.ID("stub")
}

func (c *campaignsCredentialResolver) ExternalServiceKind() string {
	return extsvc.KindGitHub
}

func (c *campaignsCredentialResolver) ExternalServiceURL() string {
	return "https://github.com/"
}

func (c *campaignsCredentialResolver) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: time.Now()}
}

type campaignsCodeHostResolver struct{}

var _ graphqlbackend.CampaignsCodeHostResolver = &campaignsCodeHostResolver{}

func (c *campaignsCodeHostResolver) ExternalServiceKind() string {
	return extsvc.KindGitHub
}

func (c *campaignsCodeHostResolver) ExternalServiceURL() string {
	return "https://github.com/"
}

func (c *campaignsCodeHostResolver) Credential() graphqlbackend.CampaignsCredentialResolver {
	return &campaignsCredentialResolver{}
}

type campaignsCodeHostConnectionResolver struct{}

var _ graphqlbackend.CampaignsCodeHostConnectionResolver = &campaignsCodeHostConnectionResolver{}

func (c *campaignsCodeHostConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return 1, nil
}

func (c *campaignsCodeHostConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
}

func (c *campaignsCodeHostConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.CampaignsCodeHostResolver, error) {
	return []graphqlbackend.CampaignsCodeHostResolver{&campaignsCodeHostResolver{}}, nil
}
