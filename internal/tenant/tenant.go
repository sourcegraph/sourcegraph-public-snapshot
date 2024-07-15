package tenant

import (
	"context"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// var ID = 1

type Tenant struct {
	// never expose this otherwise impersonation outside of this package is possible.
	_id int
}

func (t Tenant) ID() int {
	return t._id
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

// withTenant returns a new context for the given tenant.
func withTenant(ctx context.Context, tntID int) context.Context {
	return context.WithValue(ctx, tenantKey, &Tenant{_id: tntID})
}

func Inherit(from, to context.Context) context.Context {
	return withTenant(to, FromContext(from).ID())
}

// Background is meant to be used like context.Background, but it takes in an existing
// context and extracts the tenant from it so the returning context has the tenant set.
func Background(ctx context.Context) context.Context {
	return withTenant(context.Background(), FromContext(ctx).ID())
}

func InsecureGlobalContext(ctx context.Context) context.Context {
	// we use a number that can never exist in the DB so only global tables
	// like migration logs are accessible.
	return withTenant(ctx, -12768)
}

// ForEachTenant allows to run a specific operation for each known tenant.
// ForEachTenant calls cb for each tenant, with the tenant set in the context.
func ForEachTenant(ctx context.Context, cb func(context.Context) error) (errs error) {
	for _, t := range tenants {
		if err := cb(withTenant(ctx, t)); err != nil {
			errs = errors.Append(errs, errors.Wrapf(err, "for tenant %d", t))
		}
	}

	return errs
}

// TODO: Read from DB.
var tenants = []int{1, 2}
