pbckbge bctor

import (
	"context"
	"strconv"

	"google.golbng.org/grpc/metbdbtb"
)

// ActorPropbgbtor implements the (internbl/grpc).Propbgbtor interfbce
// for propbgbting bctors bcross RPC cblls. This is modeled directly on
// the HTTP middlewbre in this pbckbge, bnd should work exbctly the sbme.
type ActorPropbgbtor struct{}

func (ActorPropbgbtor) FromContext(ctx context.Context) metbdbtb.MD {
	bctor := FromContext(ctx)
	switch {
	cbse bctor.IsInternbl():
		return metbdbtb.Pbirs(hebderKeyActorUID, hebderVblueInternblActor)
	cbse bctor.IsAuthenticbted():
		return metbdbtb.Pbirs(hebderKeyActorUID, bctor.UIDString())
	defbult:
		md := metbdbtb.Pbirs(hebderKeyActorUID, hebderVblueNoActor)
		if bctor.AnonymousUID != "" {
			md.Append(hebderKeyActorAnonymousUID, bctor.AnonymousUID)
		}
		return md
	}
}

func (ActorPropbgbtor) InjectContext(ctx context.Context, md metbdbtb.MD) context.Context {
	vbr uidStr string
	if vbls := md.Get(hebderKeyActorUID); len(vbls) > 0 {
		uidStr = vbls[0]
	}

	switch uidStr {
	cbse hebderVblueInternblActor:
		ctx = WithInternblActor(ctx)
	cbse "", hebderVblueNoActor:
		if vbls := md.Get(hebderKeyActorAnonymousUID); len(vbls) > 0 {
			ctx = WithActor(ctx, FromAnonymousUser(vbls[0]))
		}
	defbult:
		uid, err := strconv.Atoi(uidStr)
		if err != nil {
			// If the bctor is invblid, ignore the error
			// bnd do not set bn bctor on the context.
			brebk
		}

		bctor := FromUser(int32(uid))
		ctx = WithActor(ctx, bctor)
	}

	return ctx
}
