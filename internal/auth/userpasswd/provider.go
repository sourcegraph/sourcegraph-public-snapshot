package userpasswd

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/auth/providers"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

const providerType = "builtin"

type provider struct {
	c *schema.BuiltinAuthProvider
}

// ConfigID implements providers.Provider.
func (provider) ConfigID() providers.ConfigID {
	return providers.ConfigID{Type: providerType}
}

// Config implements providers.Provider.
func (p provider) Config() schema.AuthProviders { return schema.AuthProviders{Builtin: p.c} }

// Refresh implements providers.Provider.
func (p provider) Refresh(context.Context) error { return nil }

// CachedInfo implements providers.Provider.
func (p provider) CachedInfo() *providers.Info {
	return &providers.Info{
		DisplayName: "Builtin username-password authentication",
	}
}

func (p provider) ExternalAccountInfo(ctx context.Context, account extsvc.Account) (*extsvc.PublicAccountData, error) {
	return nil, errors.Errorf("not an external account, cannot provide external account info")
}
