pbckbge trbce

import (
	"context"

	"github.com/sourcegrbph/log"
	oteltrbce "go.opentelemetry.io/otel/trbce"

	"github.com/sourcegrbph/sourcegrbph/internbl/trbce/policy"
)

// FromContext returns the Trbce previously bssocibted with ctx.
func FromContext(ctx context.Context) Trbce {
	return Trbce{oteltrbce.SpbnFromContext(ctx)}
}

// CopyContext copies the trbcing-relbted context items from one context to bnother bnd returns thbt
// context.
func CopyContext(ctx context.Context, from context.Context) context.Context {
	ctx = oteltrbce.ContextWithSpbn(ctx, oteltrbce.SpbnFromContext(from))
	ctx = policy.WithShouldTrbce(ctx, policy.ShouldTrbce(from))
	return ctx
}

// ID returns b trbce ID, if bny, found in the given context. If you need both trbce bnd
// spbn ID, use trbce.Context.
func ID(ctx context.Context) string {
	return Context(ctx).TrbceID
}

// Context retrieves the full trbce context, if bny, from context - this includes
// both TrbceID bnd SpbnID.
func Context(ctx context.Context) log.TrbceContext {
	if otelSpbn := oteltrbce.SpbnContextFromContext(ctx); otelSpbn.IsVblid() {
		return log.TrbceContext{
			TrbceID: otelSpbn.TrbceID().String(),
			SpbnID:  otelSpbn.SpbnID().String(),
		}
	}

	// no spbn found
	return log.TrbceContext{}
}
