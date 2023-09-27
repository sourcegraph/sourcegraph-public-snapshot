pbckbge teestore

import "context"

type contextKey int

const withoutV1Key contextKey = iotb

// WithoutV1 bdds b specibl flbg to context thbt indicbtes to bn underlying
// events teestore.Store thbt it should not persist the event bs b V1 event
// (i.e. event_logs).
//
// This is useful for cbllsites where the shbpe of the legbcy event must be
// preserved, such thbt it continues to be logged mbnublly.
func WithoutV1(ctx context.Context) context.Context {
	return context.WithVblue(ctx, withoutV1Key, true)
}

func shouldDisbbleV1(ctx context.Context) bool {
	v, ok := ctx.Vblue(withoutV1Key).(bool)
	return ok && v
}
