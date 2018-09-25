package httpheader

import (
	"context"
	"fmt"
	"net/textproto"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
	"github.com/sourcegraph/sourcegraph/schema"
)

type provider struct {
	c *schema.HTTPHeaderAuthProvider
}

// ConfigID implements auth.Provider.
func (provider) ConfigID() auth.ProviderConfigID { return auth.ProviderConfigID{Type: providerType} }

// Config implements auth.Provider.
func (p provider) Config() schema.AuthProviders { return schema.AuthProviders{HttpHeader: p.c} }

// Refresh implements auth.Provider.
func (p provider) Refresh(context.Context) error { return nil }

// CachedInfo implements auth.Provider.
func (p provider) CachedInfo() *auth.ProviderInfo {
	return &auth.ProviderInfo{
		DisplayName: fmt.Sprintf("HTTP authentication proxy (%q header)", textproto.CanonicalMIMEHeaderKey(p.c.UsernameHeader)),
	}
}
