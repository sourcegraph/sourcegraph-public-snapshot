package resolvers

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/db"
)

type campaignsCodeHostResolver struct {
	externalServiceKind string
	externalServiceURL  string
	credential          *db.UserCredential
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
		return &campaignsCredentialResolver{credential: c.credential}
	}
	return nil
}
