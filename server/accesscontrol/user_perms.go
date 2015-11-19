package accesscontrol

import (
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/fed"
	"src.sourcegraph.com/sourcegraph/svc"
)

// VerifyUserHasWriteAccess checks if the user in the current context
// is authorized to make write requests to this server.
//
// This method always returns nil when the user has write access,
// and returns a non-nil error when access cannot be granted.
// If the cmdline flag auth.restrict-write-access is set, this method
// will check if the authenticated user has admin privileges.
func VerifyUserHasWriteAccess(ctx context.Context, method string) error {
	return VerifyActorHasWriteAccess(ctx, auth.ActorFromContext(ctx), method)
}

// VerifyUserHasWriteAccess checks if the user in the current context
// is authorized to make admin requests to this server.
func VerifyUserHasAdminAccess(ctx context.Context, method string) error {
	return VerifyActorHasAdminAccess(ctx, auth.ActorFromContext(ctx), method)
}

// VerifyActorHasReadAccess checks if the given actor is authorized to make
// read requests to this server.
//
// Note that this function allows the caller to retrieve any user's access levels.
// This is meant for trusted server code living outside the scope of gRPC requests
// to verify user permissions, for example the SSH Git server. For all other cases,
// VerifyUserHasWriteAccess or VerifyUserHasAdminAccess should be used to authorize a user for gRPC operations.
func VerifyActorHasReadAccess(ctx context.Context, actor auth.Actor, method string) error {
	if !authutil.ActiveFlags.HasAccessControl() {
		// Access controls are disabled on the server, so everyone has read access.
		return nil
	}

	if authutil.ActiveFlags.AllowAnonymousReaders {
		return nil
	}

	if !actor.IsAuthenticated() {
		if len(actor.Scope) > 0 {
			return nil
		}
		return grpc.Errorf(codes.Unauthenticated, "read operation (%s) denied: no authenticated user in current context", method)
	}

	return nil
}

// VerifyActorHasWriteAccess checks if the given actor is authorized to make
// write requests to this server.
//
// Note that this function allows the caller to retrieve any user's access levels.
// This is meant for trusted server code living outside the scope of gRPC requests
// to verify user permissions, for example the SSH Git server. For all other cases,
// VerifyUserHasWriteAccess should be used to authorize a user for gRPC operations.
func VerifyActorHasWriteAccess(ctx context.Context, actor auth.Actor, method string) error {
	if authutil.ActiveFlags.RestrictWriteAccess {
		return VerifyActorHasAdminAccess(ctx, actor, method)
	}

	if !authutil.ActiveFlags.HasAccessControl() {
		// Access controls are disabled on the server, so everyone has write access.
		return nil
	}

	if !actor.IsAuthenticated() {
		if verifyScopeHasAccess(ctx, actor.Scope, method) {
			return nil
		}
		return grpc.Errorf(codes.Unauthenticated, "write operation (%s) denied: no authenticated user in current context", method)
	}

	// If auth source is local, we currently do not differentiate between
	// read and write access. The supported access levels for a user are:
	// unauthenticated, authenticated and admin. It is recommended for
	// instances with local auth to use the `auth.restrict-write-access`
	// flag to restrict write access to admin users.
	if authutil.ActiveFlags.IsLocal() {
		return nil
	}

	// If auth source is ldap, we currently do not differentiate between
	// read and write access. The supported access levels for a user are:
	// unauthenticated, authenticated and admin.
	if authutil.ActiveFlags.IsLDAP() {
		return nil
	}

	// Get UserPermissions info for this user from the root server.
	perms, err := getUserPermissionsFromRoot(ctx, actor)
	if err != nil {
		return err
	}

	if !perms.Write && !perms.Admin {
		return grpc.Errorf(codes.PermissionDenied, "write operation (%s) denied: user does not have write access", method)
	}

	return nil
}

// VerifyActorHasAdminAccess checks if the given actor is authorized to make
// admin requests to this server.
//
// Note that this function allows the caller to retrieve any user's access levels.
// This is meant for trusted server code living outside the scope of gRPC requests
// to verify user permissions, for example the SSH Git server. For all other cases,
// VerifyUserHasAdminAccess should be used to authorize a user for gRPC operations.
func VerifyActorHasAdminAccess(ctx context.Context, actor auth.Actor, method string) error {
	if !authutil.ActiveFlags.HasAccessControl() {
		// Access controls are disabled on the server, so everyone has admin access.
		return nil
	}

	if !actor.IsAuthenticated() {
		if verifyScopeHasAccess(ctx, actor.Scope, method) {
			return nil
		}
		return grpc.Errorf(codes.Unauthenticated, "admin operation (%s) denied: no authenticated user in current context", method)
	}

	var isAdmin bool
	if authutil.ActiveFlags.IsLocal() || authutil.ActiveFlags.IsLDAP() {
		// Check local auth server for user's admin privileges.
		user, err := svc.Users(ctx).Get(ctx, &sourcegraph.UserSpec{UID: int32(actor.UID)})
		if err != nil {
			return grpc.Errorf(codes.Internal, "admin operation (%s) denied: could not complete permissions check for user %v: %v", method, actor.UID, err)
		}
		isAdmin = user.Admin
	} else {
		// Get UserPermissions info for this user from the root server.
		if perms, err := getUserPermissionsFromRoot(ctx, actor); err != nil {
			return err
		} else {
			isAdmin = perms.Admin
		}
	}

	if !isAdmin {
		return grpc.Errorf(codes.PermissionDenied, "admin operation (%s) denied: user does not have admin status", method)
	}
	return nil
}

var getUserPermissionsFromRoot = func(ctx context.Context, actor auth.Actor) (*sourcegraph.UserPermissions, error) {
	// TODO: Cache UserPermissions to avoid making a call to root server for every
	// write/admin operation.
	rootCtx := fed.Config.NewRemoteContext(ctx)
	rootCl := sourcegraph.NewClientFromContext(rootCtx)
	userPermissions, err := rootCl.RegisteredClients.GetUserPermissions(rootCtx, &sourcegraph.UserPermissionsOptions{
		UID:        int32(actor.UID),
		ClientSpec: &sourcegraph.RegisteredClientSpec{ID: actor.ClientID},
	})
	if err != nil {
		return nil, err
	}
	return userPermissions, nil
}

// Check if the actor is authorized with an access token
// having a valid scope. This token is set in package sgx on server
// startup, and is only available to client commands spawned
// in the server process.
//
// !!!!!!!!!!!!!!!!!!!! DANGER(security) !!!!!!!!!!!!!!!!!!!!!!
// This does not check that the token is properly signed, since
// that is done in server/internal/oauth2util/grpc_middleware.go
// when parsing the request metadata and adding the actor to the
// context. To avoid additional latency from expensive public key
// operations, that check is not repeated here, but be careful
// about refactoring that check.
func verifyScopeHasAccess(ctx context.Context, scopes []string, method string) bool {
	for _, scope := range scopes {
		switch {
		case strings.HasPrefix(scope, "internal:"):
			// internal server commands have default write access.
			return true

		case scope == "worker:build":
			// workers have write access on the Builds server.
			if strings.HasPrefix(method, "Builds.") {
				return true
			}

		case strings.HasPrefix(scope, "app:"):
			// all apps have default write access.
			// TODO: configure app-specific permissions.
			return true
		}
	}
	return false
}
