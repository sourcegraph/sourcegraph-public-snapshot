package resolvers

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
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

func (c *campaignsCredentialResolver) HasSSHKey() bool {
	has := false
	switch a := c.credential.Credential.(type) {
	case *auth.OAuthBearerToken:
		has = a.SSHKey != ""
	case *auth.BasicAuth:
		has = a.SSHKey != ""
	}
	return has
}
