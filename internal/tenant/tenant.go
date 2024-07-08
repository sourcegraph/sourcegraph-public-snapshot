package tenant

import "context"

var ID = 1

type Tenant struct {
	ID int
}

type contextKey int

const tenantKey contextKey = iota

// FromContext returns the tenant from a given context.
func FromContext(ctx context.Context) *Tenant {
	tnt, ok := ctx.Value(tenantKey).(*Tenant)
	if !ok || tnt == nil {
		return &Tenant{}
	}
	return tnt
}

// WithTenant returns a new context for the given tenant.
func WithTenant(ctx context.Context, tntID int) context.Context {
	return context.WithValue(ctx, tenantKey, &Tenant{ID: tntID})
}
