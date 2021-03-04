package resolvers

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/batches"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

type batchChangesCodeHostResolver struct {
	codeHost   *batches.CodeHost
	credential *database.UserCredential
}

var _ graphqlbackend.BatchChangesCodeHostResolver = &batchChangesCodeHostResolver{}

func (c *batchChangesCodeHostResolver) ExternalServiceKind() string {
	return extsvc.TypeToKind(c.codeHost.ExternalServiceType)
}

func (c *batchChangesCodeHostResolver) ExternalServiceURL() string {
	return c.codeHost.ExternalServiceID
}

func (c *batchChangesCodeHostResolver) Credential() graphqlbackend.BatchChangesCredentialResolver {
	if c.credential != nil {
		return &batchChangesCredentialResolver{credential: c.credential}
	}
	return nil
}

func (c *batchChangesCodeHostResolver) RequiresSSH() bool {
	return c.codeHost.RequiresSSH
}

// TODO(campaigns-deprecation): Remove this wrapper type. It just exists to fulfil the interface
// of graphqlbackend.CampaignsCodeHostConnectionResolver.
type campaignsCodeHostResolver struct {
	graphqlbackend.BatchChangesCodeHostResolver
}

var _ graphqlbackend.CampaignsCodeHostResolver = &campaignsCodeHostResolver{}

func (c *campaignsCodeHostResolver) ExternalServiceKind() string {
	return c.BatchChangesCodeHostResolver.ExternalServiceKind()
}

func (c *campaignsCodeHostResolver) ExternalServiceURL() string {
	return c.BatchChangesCodeHostResolver.ExternalServiceURL()
}

func (c *campaignsCodeHostResolver) Credential() graphqlbackend.CampaignsCredentialResolver {
	return c.BatchChangesCodeHostResolver.Credential()
}

func (c *campaignsCodeHostResolver) RequiresSSH() bool {
	return c.BatchChangesCodeHostResolver.RequiresSSH()
}
