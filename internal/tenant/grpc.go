package tenant

// // TenantPropagator implements the (internal/grpc).Propagator interface
// // for propagating tenants across RPC calls. This is modeled directly on
// // the HTTP middleware in this package, and should work exactly the same.
// type TenantPropagator struct{}

// func (TenantPropagator) FromContext(ctx context.Context) metadata.MD {
// 	actor := FromContext(ctx)
// 	md := make(metadata.MD)

// 	// We always propagate AnonymousUID if present.
// 	if actor.AnonymousUID != "" {
// 		md.Append(headerKeyActorAnonymousUID, actor.AnonymousUID)
// 	}

// 	switch {
// 	case actor.IsInternal():
// 		md.Append(headerKeyActorUID, headerValueInternalActor)
// 	case actor.IsAuthenticated():
// 		md.Append(headerKeyActorUID, actor.UIDString())
// 	default:
// 		md.Append(headerKeyActorUID, headerValueNoActor)
// 	}

// 	return md
// }

// func (TenantPropagator) InjectContext(ctx context.Context, md metadata.MD) context.Context {
// 	var idStr string
// 	if vals := md.Get(headerKeyTenantID); len(vals) > 0 {
// 		idStr = vals[0]
// 	}

// 	tnt := &Tenant{}
// 	switch idStr {
// 	case headerValueInternalActor:
// 		act = Internal()
// 	default:
// 		uid, err := strconv.Atoi(idStr)
// 		if err != nil {
// 			// If the actor is invalid, ignore the error
// 			// and do not set an actor on the context.
// 			break
// 		}
// 		act = FromUser(int32(uid))
// 	}

// 	// Always preserve the AnonymousUID if present
// 	if vals := md.Get(headerKeyActorAnonymousUID); len(vals) > 0 {
// 		act.AnonymousUID = vals[0]
// 	}

// 	// FromContext always returns a non-nil Actor, so it's okay to always add it
// 	return WithActor(ctx, act)
// }
