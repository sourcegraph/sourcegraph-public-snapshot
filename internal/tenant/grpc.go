package tenant

import (
	"context"
	"strconv"

	"google.golang.org/grpc/metadata"
)

// TenantPropagator implements the (internal/grpc).Propagator interface
// for propagating tenants across RPC calls. This is modeled directly on
// the HTTP middleware in this package, and should work exactly the same.
type TenantPropagator struct{}

func (TenantPropagator) FromContext(ctx context.Context) metadata.MD {
	tenant := FromContext(ctx)
	md := make(metadata.MD)

	switch {
	case tenant.ID() == 0:
		md.Append(headerKeyTenantID, headerValueNoTenant)
	default:
		md.Append(headerKeyTenantID, strconv.Itoa(tenant.ID()))
	}

	return md
}

func (TenantPropagator) InjectContext(ctx context.Context, md metadata.MD) context.Context {
	var idStr string
	if vals := md.Get(headerKeyTenantID); len(vals) > 0 {
		idStr = vals[0]
	}

	switch idStr {
	case headerValueNoTenant:
		// Nothing to do, empty tenant.
		return ctx
	default:
		uid, err := strconv.Atoi(idStr)
		if err != nil {
			// If the actor is invalid, ignore the error
			// and do not set an actor on the context.
			return ctx
		}
		return withTenant(ctx, uid)
	}
}
