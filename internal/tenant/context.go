package tenant

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type contextKey int

const tenantKey contextKey = iota

// FromContext returns the tenant from a given context, or ok=false when no tenant
// is set in the passed context.
func FromContext(ctx context.Context) (tnt *Tenant, ok bool) {
	tnt, ok = ctx.Value(tenantKey).(*Tenant)
	return tnt, ok
}

// withTenant returns a new context for the given tenant.
// 🚨 SECURITY: This method is intentionally not exported, so that there we control
// inside this package only when impersonation happens.
func withTenant(ctx context.Context, tntID int) context.Context {
	return context.WithValue(ctx, tenantKey, &Tenant{_id: tntID})
}

// Inherit returns a new context with the tenant from the given context. This can be
// useful to carry forward other context properties like trace spans etc.
//
// It errors if the given context is already for another tenant.
func Inherit(from, to context.Context) (context.Context, error) {
	fromTenant, fromTenantOk := FromContext(from)
	toTenant, toTenantOk := FromContext(to)

	if toTenantOk && fromTenantOk && fromTenant.ID() != toTenant.ID() {
		return nil, errors.New("cannot inherit into a context with a tenant")
	}

	if !fromTenantOk {
		return to, nil
	}

	return withTenant(to, fromTenant.ID()), nil
}

// TestContext can be used in tests to create a context with a tenant without having
// to worry about wiring up a lot of resources. Future datastores that depend on tenants
// should use this function to create a context with a tenant and auto-provision it.
//
// Note: You must not call this function outside of tests, it will panic.
func TestContext() context.Context {
	if !testutil.IsTest {
		panic("only call this function in tests")
	}
	return withTenant(context.Background(), 1)
}
