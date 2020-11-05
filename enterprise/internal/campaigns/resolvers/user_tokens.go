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

type campaignsCredentialResolver struct {
	externalServiceKind string
	externalServiceURL  string
	createdAt           time.Time
}

var _ graphqlbackend.CampaignsCredentialResolver = &campaignsCredentialResolver{}

func (c *campaignsCredentialResolver) ID() graphql.ID {
	return graphql.ID("stub")
}

func (c *campaignsCredentialResolver) ExternalServiceKind() string {
	return c.externalServiceKind
}

func (c *campaignsCredentialResolver) ExternalServiceURL() string {
	return c.externalServiceURL
}

func (c *campaignsCredentialResolver) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: c.createdAt}
}

type campaignsCodeHostResolver struct {
	externalServiceKind string
	externalServiceURL  string
	credential          *credential
}

var _ graphqlbackend.CampaignsCodeHostResolver = &campaignsCodeHostResolver{}

func (c *campaignsCodeHostResolver) ExternalServiceKind() string {
	return c.externalServiceKind
}

func (c *campaignsCodeHostResolver) ExternalServiceURL() string {
	return c.externalServiceURL
}

func (c *campaignsCodeHostResolver) Credential() graphqlbackend.CampaignsCredentialResolver {
	if c.credential != nil {
		return &campaignsCredentialResolver{externalServiceKind: c.externalServiceKind, externalServiceURL: c.externalServiceURL, createdAt: c.credential.createdAt}
	}
	return nil
}

type campaignsCodeHostConnectionResolver struct{}

var _ graphqlbackend.CampaignsCodeHostConnectionResolver = &campaignsCodeHostConnectionResolver{}

func (c *campaignsCodeHostConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return 4, nil
}

func (c *campaignsCodeHostConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
}

func (c *campaignsCodeHostConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.CampaignsCodeHostResolver, error) {
	return []graphqlbackend.CampaignsCodeHostResolver{&campaignsCodeHostResolver{
		externalServiceKind: extsvc.KindGitHub,
		externalServiceURL:  "https://github.com/",
		credential:          &credential{createdAt: time.Now()},
	}, &campaignsCodeHostResolver{
		externalServiceKind: extsvc.KindGitLab,
		externalServiceURL:  "https://gitlab.com/",
		credential:          &credential{createdAt: time.Now()},
	}, &campaignsCodeHostResolver{
		externalServiceKind: extsvc.KindBitbucketServer,
		externalServiceURL:  "https://bitbucket.sgdev.org/",
		credential:          &credential{createdAt: time.Now()},
	}, &campaignsCodeHostResolver{
		externalServiceKind: extsvc.KindGitHub,
		externalServiceURL:  "https://ghe.sgdev.org/",
	}}, nil
}

type credential struct {
	createdAt time.Time
}
