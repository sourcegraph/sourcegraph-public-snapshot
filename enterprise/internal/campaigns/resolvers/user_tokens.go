package resolvers

import (
	"context"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

const campaignsCredentialIDKind = "CampaignsCredential"

func marshalCampaignsCredentialID(id int64) graphql.ID {
	return relay.MarshalID(campaignsCredentialIDKind, id)
}

func unmarshalCampaignsCredentialID(id graphql.ID) (cid int64, err error) {
	err = relay.UnmarshalSpec(id, &cid)
	return
}

func (r *Resolver) CreateCampaignsCredential(ctx context.Context, args *graphqlbackend.CreateCampaignsCredentialArgs) (graphqlbackend.CampaignsCredentialResolver, error) {
	return &campaignsCredentialResolver{}, errors.New("not implemented")
}

func (r *Resolver) DeleteCampaignsCredential(ctx context.Context, args *graphqlbackend.DeleteCampaignsCredentialArgs) (*graphqlbackend.EmptyResponse, error) {
	dbID, err := unmarshalCampaignsCredentialID(args.CampaignsCredential)
	if err != nil {
		return nil, err
	}
	// Validate the credential with the ID exists so we don't swallow an error here.
	_, err = r.store.GetCampaignCredential(ctx, ee.GetCampaignCredentialOpts{ID: dbID})
	if err != nil {
		return nil, err
	}
	if err := r.store.DeleteCampaignCredential(ctx, ee.DeleteCampaignCredentialOpts{ID: dbID}); err != nil {
		return nil, err
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *Resolver) CampaignsCredentialByID(ctx context.Context, id graphql.ID) (graphqlbackend.CampaignsCredentialResolver, error) {
	dbID, err := unmarshalCampaignsCredentialID(id)
	if err != nil {
		return nil, err
	}
	credential, err := r.store.GetCampaignCredential(ctx, ee.GetCampaignCredentialOpts{ID: dbID})
	if err != nil && err != ee.ErrNoResults {
		return nil, err
	}
	if credential == nil {
		return nil, nil
	}
	return &campaignsCredentialResolver{credential: credential}, nil
}

func (r *Resolver) CampaignsCodeHosts(ctx context.Context, args *graphqlbackend.ListCampaignsCodeHostsArgs) (graphqlbackend.CampaignsCodeHostConnectionResolver, error) {
	return &campaignsCodeHostConnectionResolver{}, nil
}

type campaignsCredentialResolver struct {
	id                  int64
	externalServiceKind string
	externalServiceURL  string
	createdAt           time.Time
}

var _ graphqlbackend.CampaignsCredentialResolver = &campaignsCredentialResolver{}

func (c *campaignsCredentialResolver) ID() graphql.ID {
	return marshalCampaignsCredentialID(c.id)
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

func (c *campaignsCodeHostConnectionResolver) compute(ctx context.Context) {
	q := "SELECT external_service_type, external_service_id from repo where external_service_type IN ('github','gitlab','bitbucketServer') group by external_service_type, external_service_id"
}

type credential struct {
	createdAt time.Time
}
