package tenant

import (
	"context"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// TenantPropagator implements the (internal/grpc).Propagator interface
// for propagating tenants across RPC calls. This is modeled directly on
// the HTTP middleware in this package, and should work exactly the same.
type TenantPropagator struct{}

func (TenantPropagator) FromContext(ctx context.Context) metadata.MD {
	tenant, err := FromContext(ctx)
	md := make(metadata.MD)

	if err != nil {
		md.Append(headerKeyTenantID, headerValueNoTenant)
	} else {
		md.Append(headerKeyTenantID, strconv.Itoa(tenant.ID()))
	}

	return md
}

func (TenantPropagator) InjectContext(ctx context.Context, md metadata.MD) (context.Context, error) {
	var idStr string
	if vals := md.Get(headerKeyTenantID); len(vals) > 0 {
		idStr = vals[0]
	}

	switch idStr {
	case "", headerValueNoTenant:
		// Nothing to do, empty tenant.
		return ctx, nil
	default:
		uid, err := strconv.Atoi(idStr)
		if err != nil {
			// The tenant value is invalid.
			return ctx, status.New(codes.InvalidArgument, errors.Wrap(err, "bad tenant value in metadata").Error()).Err()
		}
		return withTenant(ctx, uid), nil
	}
}
