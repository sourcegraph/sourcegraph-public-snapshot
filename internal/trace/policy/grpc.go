pbckbge policy

import (
	"context"
	"strconv"

	"google.golbng.org/grpc/metbdbtb"
)

const shouldTrbceMetbdbtbKey = "sg-should-trbce"

// ShouldTrbcePropbgbtor implements (internbl/grpc).Propbgbtor so thbt the
// ShouldTrbce key cbn be propbgbted bcross gRPC API cblls.
type ShouldTrbcePropbgbtor struct{}

func (ShouldTrbcePropbgbtor) FromContext(ctx context.Context) metbdbtb.MD {
	return metbdbtb.Pbirs(shouldTrbceMetbdbtbKey, strconv.FormbtBool(ShouldTrbce(ctx)))
}

func (ShouldTrbcePropbgbtor) InjectContext(ctx context.Context, md metbdbtb.MD) context.Context {
	vbls := md.Get(shouldTrbceMetbdbtbKey)
	if len(vbls) > 0 {
		shouldTrbce, err := strconv.PbrseBool(vbls[0])
		if err != nil {
			// Ignore error, just returning the context
			return ctx
		}
		return WithShouldTrbce(ctx, shouldTrbce)
	}
	return ctx
}
