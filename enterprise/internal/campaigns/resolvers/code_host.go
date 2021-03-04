package resolvers

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

type campaignsCodeHostResolver struct {
	codeHost   *campaigns.CodeHost
	credential *database.UserCredential
}

var _ graphqlbackend.CampaignsCodeHostResolver = &campaignsCodeHostResolver{}

func (c *campaignsCodeHostResolver) ExternalServiceKind() string {
	return extsvc.TypeToKind(c.codeHost.ExternalServiceType)
}

func (c *campaignsCodeHostResolver) ExternalServiceURL() string {
	return c.codeHost.ExternalServiceID
}

func (c *campaignsCodeHostResolver) Credential() graphqlbackend.CampaignsCredentialResolver {
	if c.credential != nil {
		return &campaignsCredentialResolver{credential: c.credential}
	}
	return nil
}

func (c *campaignsCodeHostResolver) RequiresSSH() bool {
	return c.codeHost.RequiresSSH
}
