package accesscontrol

import (
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
)

// VerifyUserHasReadAccess checks if the user in the current context
// is authorized to make write requests to this server.
//
// This method always returns nil when the user has write access,
// and returns a non-nil error when access cannot be granted.
// If the cmdline flag auth.restrict-write-access is set, this method
// will check if the authenticated user has admin privileges.
func VerifyUserHasReadAccess(ctx context.Context, method, repo string) error {
	return VerifyActorHasReadAccess(ctx, auth.ActorFromContext(ctx), method, repo)
}

// VerifyUserHasWriteAccess checks if the user in the current context
// is authorized to make write requests to this server.
//
// This method always returns nil when the user has write access,
// and returns a non-nil error when access cannot be granted.
// If the cmdline flag auth.restrict-write-access is set, this method
// will check if the authenticated user has admin privileges.
func VerifyUserHasWriteAccess(ctx context.Context, method, repo string) error {
	return VerifyActorHasWriteAccess(ctx, auth.ActorFromContext(ctx), method, repo)
}

// VerifyUserHasWriteAccess checks if the user in the current context
// is authorized to make admin requests to this server.
func VerifyUserHasAdminAccess(ctx context.Context, method string) error {
	return VerifyActorHasAdminAccess(ctx, auth.ActorFromContext(ctx), method)
}

// VerifyUserSelfOrAdmin checks if the user in the current context has
// the given uid, or if the actor has admin access on the server.
// This check should be used in cases where a request should succeed only
// if the request is for the user's own information, or if the ctx actor is an admin.
func VerifyUserSelfOrAdmin(ctx context.Context, method string, uid int32) error {
	if uid != 0 && auth.ActorFromContext(ctx).UID == int(uid) {
		return nil
	}

	return VerifyUserHasAdminAccess(ctx, method)
}

// VerifyClientSelfOrAdmin checks if the client in the current context has
// the given id, or if the actor has admin access on the server.
// This check should be used in cases where a request should succeed only
// if the request is for the client's own information, or if the ctx actor is an admin.
func VerifyClientSelfOrAdmin(ctx context.Context, method string, clientID string) error {
	if clientID != "" && auth.ActorFromContext(ctx).ClientID == clientID {
		return nil
	}

	return VerifyUserHasAdminAccess(ctx, method)
}

// VerifyActorHasReadAccess checks if the given actor is authorized to make
// read requests to this server.
//
// Note that this function allows the caller to retrieve any user's access levels.
// This is meant for trusted server code living outside the scope of gRPC requests
// to verify user permissions, for example the SSH Git server. For all other cases,
// VerifyUserHasWriteAccess or VerifyUserHasAdminAccess should be used to authorize a user for gRPC operations.
func VerifyActorHasReadAccess(ctx context.Context, actor auth.Actor, method, repo string) error {
	if !authutil.ActiveFlags.HasAccessControl() {
		// Access controls are disabled on the server, so everyone has read access.
		return nil
	}

	if authutil.ActiveFlags.PrivateMirrors && repo != "" {
		return VerifyRepoPerms(ctx, actor, method, repo)
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
func VerifyActorHasWriteAccess(ctx context.Context, actor auth.Actor, method, repo string) error {
	if !authutil.ActiveFlags.HasAccessControl() {
		// Access controls are disabled on the server, so everyone has write access.
		return nil
	}

	if authutil.ActiveFlags.PrivateMirrors {
		if method == "Repos.Create" && actor.PrivateMirrors {
			return nil
		} else if repo != "" {
			return VerifyRepoPerms(ctx, actor, method, repo)
		}
	}

	if !actor.IsAuthenticated() {
		if VerifyScopeHasAccess(ctx, actor.Scope, method) {
			return nil
		}
		return grpc.Errorf(codes.Unauthenticated, "write operation (%s) denied: no authenticated user in current context", method)
	}

	var hasWrite bool
	if inAuthenticatedWriteWhitelist(method) {
		hasWrite = true
	} else {
		hasWrite = actor.HasWriteAccess()
	}

	if !hasWrite {
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
		if VerifyScopeHasAccess(ctx, actor.Scope, method) {
			return nil
		}
		return grpc.Errorf(codes.Unauthenticated, "admin operation (%s) denied: no authenticated user in current context", method)
	}

	if !actor.HasAdminAccess() {
		return grpc.Errorf(codes.PermissionDenied, "admin operation (%s) denied: user does not have admin status", method)
	}
	return nil
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
func VerifyScopeHasAccess(ctx context.Context, scopes map[string]bool, method string) bool {
	if scopes == nil {
		return false
	}
	for scope := range scopes {
		switch {
		case strings.HasPrefix(scope, "internal:"):
			// internal server commands have default write access.
			return true

		case scope == "worker:build":
			return true

		case strings.HasPrefix(scope, "app:"):
			// all apps have default write access.
			// TODO: configure app-specific permissions.
			return true
		}
	}
	return false
}

// Check if we always allow write access to a method for an authenticated
// user.
func inAuthenticatedWriteWhitelist(method string) bool {
	switch method {
	case "MirrorRepos.CloneRepo":
		return true
	}
	return false
}
