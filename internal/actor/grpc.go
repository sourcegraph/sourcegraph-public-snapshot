package actor

import (
	"context"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ActorPropagator implements the (internal/grpc).Propagator interface
// for propagating actors across RPC calls. This is modeled directly on
// the HTTP middleware in this package, and should work exactly the same.
type ActorPropagator struct{}

func (ActorPropagator) FromContext(ctx context.Context) metadata.MD {
	actor := FromContext(ctx)
	md := make(metadata.MD)

	// We always propagate AnonymousUID if present.
	if actor.AnonymousUID != "" {
		md.Append(headerKeyActorAnonymousUID, actor.AnonymousUID)
	}

	switch {
	case actor.IsInternal():
		md.Append(headerKeyActorUID, headerValueInternalActor)
	case actor.IsAuthenticated():
		md.Append(headerKeyActorUID, actor.UIDString())
	default:
		md.Append(headerKeyActorUID, headerValueNoActor)
	}

	return md
}

func (ActorPropagator) InjectContext(ctx context.Context, md metadata.MD) (context.Context, error) {
	var uidStr string
	if vals := md.Get(headerKeyActorUID); len(vals) > 0 {
		uidStr = vals[0]
	}

	act := &Actor{}
	switch uidStr {
	case "", headerValueNoActor:
	case headerValueInternalActor:
		act = Internal()
	default:
		uid, err := strconv.Atoi(uidStr)
		if err != nil {
			// The actor is invalid.
			return ctx, status.New(codes.InvalidArgument, errors.Wrap(err, "bad actor value in metadata").Error()).Err()
		}
		act = FromUser(int32(uid))
	}

	// Always preserve the AnonymousUID if present
	if vals := md.Get(headerKeyActorAnonymousUID); len(vals) > 0 {
		act.AnonymousUID = vals[0]
	}

	// FromContext always returns a non-nil Actor, so it's okay to always add it
	return WithActor(ctx, act), nil
}
