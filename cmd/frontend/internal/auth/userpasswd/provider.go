package userpasswd

import (
	"context"

	authprovider "github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/schema"
)

const providerType = "builtin"

type provider struct {
	c *schema.BuiltinAuthProvider
}

// ConfigID implements providers.Provider.
func (provider) ConfigID() authprovider.ProviderConfigID {
	return authprovider.ProviderConfigID{Type: providerType}
}

// Config implements providers.Provider.
func (p provider) Config() schema.AuthProviders { return schema.AuthProviders{Builtin: p.c} }

// Refresh implements providers.Provider.
func (p provider) Refresh(context.Context) error { return nil }

// CachedInfo implements providers.Provider.
func (p provider) CachedInfo() *authprovider.ProviderInfo {
	return &authprovider.ProviderInfo{
		DisplayName: "Builtin username-password authentication",
	}
}
