package userpasswd

import (
	"context"

	"sourcegraph.com/cmd/frontend/auth/providers"
	"sourcegraph.com/schema"
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
