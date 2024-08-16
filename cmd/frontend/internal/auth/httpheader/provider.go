package httpheader

import (
	"context"
	"fmt"
	"net/textproto"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/providers"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/schema"
)

type provider struct {
	c *schema.HTTPHeaderAuthProvider
}

// ConfigID implements providers.Provider.
func (provider) ConfigID() providers.ConfigID {
	return providers.ConfigID{Type: providerType}
}

// Config implements providers.Provider.
func (p provider) Config() schema.AuthProviders { return schema.AuthProviders{HttpHeader: p.c} }

// CachedInfo implements providers.Provider.
func (p provider) CachedInfo() *providers.Info {
	return &providers.Info{
		DisplayName: fmt.Sprintf("HTTP authentication proxy (%q header)", textproto.CanonicalMIMEHeaderKey(p.c.UsernameHeader)),
	}
}

func (p *provider) ExternalAccountInfo(ctx context.Context, account extsvc.Account) (*extsvc.PublicAccountData, error) {
	return &extsvc.PublicAccountData{
		DisplayName: account.AccountID,
	}, nil
}

func (p *provider) Type() providers.ProviderType {
	return providers.ProviderTypeHTTPHeader
}
