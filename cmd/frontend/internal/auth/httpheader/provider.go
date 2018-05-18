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

// ID implements auth.Provider.
func (provider) ID() auth.ProviderID { return auth.ProviderID{Type: providerType} }

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
