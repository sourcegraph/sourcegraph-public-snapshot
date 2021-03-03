package resolvers

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
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

type campaignsCredentialResolver struct {
	credential *database.UserCredential
}

var _ graphqlbackend.CampaignsCredentialResolver = &campaignsCredentialResolver{}

func (c *campaignsCredentialResolver) ID() graphql.ID {
	return marshalCampaignsCredentialID(c.credential.ID)
}

func (c *campaignsCredentialResolver) ExternalServiceKind() string {
	return extsvc.TypeToKind(c.credential.ExternalServiceType)
}

func (c *campaignsCredentialResolver) ExternalServiceURL() string {
	// This is usually the code host URL.
	return c.credential.ExternalServiceID
}

func (c *campaignsCredentialResolver) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: c.credential.CreatedAt}
}
