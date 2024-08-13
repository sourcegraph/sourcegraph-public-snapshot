package tenant

import (
	"context"
	"runtime/pprof"

	"go.uber.org/atomic"

	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type contextKey int

const tenantKey contextKey = iota

var ErrNoTenantInContext = errors.New("no tenant in context")

// FromContext returns the tenant from a given context. It returns
// ErrNoTenantInContext if the context does not have a tenant.
func FromContext(ctx context.Context) (*Tenant, error) {
	tnt, ok := ctx.Value(tenantKey).(*Tenant)
	if !ok {
		if pprofMissingTenant != nil {
			// We want to track every stack trace, so need a unique value for the event
			eventValue := pprofUniqID.Add(1)

			// skip stack for Add and this function (2).
			pprofMissingTenant.Add(eventValue, 2)
		}

		return nil, ErrNoTenantInContext
	}
	return tnt, nil
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
	fromTenant, fromTenantErr := FromContext(from)
	toTenant, toTenantErr := FromContext(to)

	if toTenantErr == nil && fromTenantErr == nil && fromTenant.ID() != toTenant.ID() {
		return nil, errors.New("cannot inherit into a context with a tenant")
	}

	if fromTenantErr != nil {
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

var tenantCounter atomic.Int64

// NewTestContext is like TestContext, but it will return a context with a new
// tenant every time.
func NewTestContext() context.Context {
	if !testutil.IsTest {
		panic("only call this function in tests")
	}
	return withTenant(context.Background(), int(tenantCounter.Inc()))
}

var pprofUniqID atomic.Int64
var pprofMissingTenant = func() *pprof.Profile {
	if !shouldLogNoTenant() {
		return nil
	}
	return pprof.NewProfile("missing_tenant")
}()
