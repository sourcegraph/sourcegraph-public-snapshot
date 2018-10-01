package userpasswd

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
	"github.com/sourcegraph/sourcegraph/schema"
)

const providerType = "builtin"

type provider struct {
	c *schema.BuiltinAuthProvider
}

// ConfigID implements auth.Provider.
func (provider) ConfigID() auth.ProviderConfigID { return auth.ProviderConfigID{Type: providerType} }

// Config implements auth.Provider.
func (p provider) Config() schema.AuthProviders { return schema.AuthProviders{Builtin: p.c} }

// Refresh implements auth.Provider.
func (p provider) Refresh(context.Context) error { return nil }

// CachedInfo implements auth.Provider.
func (p provider) CachedInfo() *auth.ProviderInfo {
	return &auth.ProviderInfo{
		DisplayName: "Builtin username-password authentication",
	}
}
