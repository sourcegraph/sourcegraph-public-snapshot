package actor

import (
	"context"
	"strconv"

	"google.golang.org/grpc/metadata"
)

// ActorPropagator implements the (internal/grpc).Propagator interface
// for propagating actors across RPC calls. This is modeled directly on
// the HTTP middleware in this package, and should work exactly the same.
type ActorPropagator struct{}

func (ActorPropagator) FromContext(ctx context.Context) metadata.MD {
	actor := FromContext(ctx)
	switch {
	case actor.IsInternal():
		return metadata.Pairs(headerKeyActorUID, headerValueInternalActor)
	case actor.IsAuthenticated():
		return metadata.Pairs(headerKeyActorUID, actor.UIDString())
	default:
		md := metadata.Pairs(headerKeyActorUID, headerValueNoActor)
		if actor.AnonymousUID != "" {
			md.Append(headerKeyActorAnonymousUID, actor.AnonymousUID)
		}
		return md
	}
}

func (ActorPropagator) InjectContext(ctx context.Context, md metadata.MD) context.Context {
	var uidStr string
	if vals := md.Get(headerKeyActorUID); len(vals) > 0 {
		uidStr = vals[0]
	}

	switch uidStr {
	case headerValueInternalActor:
		ctx = WithInternalActor(ctx)
	case "", headerValueNoActor:
		if vals := md.Get(headerKeyActorAnonymousUID); len(vals) > 0 {
			ctx = WithActor(ctx, FromAnonymousUser(vals[0]))
		}
	default:
		uid, err := strconv.Atoi(uidStr)
		if err != nil {
			// If the actor is invalid, ignore the error
			// and do not set an actor on the context.
			break
		}

		actor := FromUser(int32(uid))
		ctx = WithActor(ctx, actor)
	}

	return ctx
}
