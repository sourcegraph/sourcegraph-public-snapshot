package accesscontrol

import (
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/fed"
	"src.sourcegraph.com/sourcegraph/svc"
	"src.sourcegraph.com/sourcegraph/svc/middleware/remote"
)

// VerifyUserHasWriteAccess checks if the user in the current context
// is authorized to make write requests to this server.
// This method always returns nil when the user has write access,
// and returns a non-nil error when access cannot be granted.
// If the cmdline flag auth.restrict-write-access is set, this method
// will check if the authenticated user has admin privileges.
func VerifyUserHasWriteAccess(ctx context.Context, method string) error {
	if !authutil.ActiveFlags.HasUserAccounts() {
		// no user accounts on the server, so everyone has write access.
		return nil
	}

	actor := auth.ActorFromContext(ctx)
	if !actor.IsAuthenticated() {
		// Check if the actor is authorized with an access token
		// having a scope. This token is set in package sgx on server
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
		for _, scope := range actor.Scope {
			// internal server commands have default write access.
			if scope == "internal:cli" {
				return nil
			}

			// workers have write access on the Builds server.
			if scope == "worker:build" && strings.HasPrefix(method, "Builds.") {
				return nil
			}
		}
		return grpc.Errorf(codes.Unauthenticated, "write operation (%s) denied: no authenticated user in current context", method)
	}

	if authutil.ActiveFlags.RestrictWriteAccess {
		return VerifyUserHasAdminAccess(ctx, method)
	}

	// Get UserPermissions info for this user from the root server.
	// Making such a call to root server for every write operation will be quite slow, so
	// cache the user permissions on the client (i.e. local instance).
	if perms, err := getUserPermissionsFromRoot(ctx); err != nil {
		return err
	} else if !(perms.Write || perms.Admin) {
		return grpc.Errorf(codes.PermissionDenied, "write operation (%s) denied: user does not have write access", method)
	} else {
		return nil
	}
}

func VerifyUserHasAdminAccess(ctx context.Context, method string) error {
	if !authutil.ActiveFlags.HasUserAccounts() {
		// no user accounts on the server, so everyone has admin access.
		return nil
	}

	var isAdmin bool
	actor := auth.ActorFromContext(ctx)

	if authutil.ActiveFlags.IsLocal() {
		// check local auth server for user's admin privileges.
		user, err := svc.Users(ctx).Get(ctx, &sourcegraph.UserSpec{UID: int32(actor.UID)})
		if err != nil {
			return grpc.Errorf(codes.Internal, "admin operation (%s) denied: could not complete permissions check for user %v: %v", method, actor.UID, err)
		}
		isAdmin = user.Admin
	} else {
		// get UserPermissions info for this user from the root server.
		if perms, err := getUserPermissionsFromRoot(ctx); err != nil {
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

func getUserPermissionsFromRoot(ctx context.Context) (*sourcegraph.UserPermissions, error) {
	// TODO: cache UserPermissions to avoid making a call to root server for every
	// write/admin operation.
	rootGRPCURL, err := fed.Config.RootGRPCEndpoint()
	if err != nil {
		return nil, err
	}
	actor := auth.ActorFromContext(ctx)
	ctx = sourcegraph.WithGRPCEndpoint(ctx, rootGRPCURL)
	ctx = svc.WithServices(ctx, remote.Services)
	rootCl := sourcegraph.NewClientFromContext(ctx)
	userPermissions, err := rootCl.RegisteredClients.GetUserPermissions(ctx, &sourcegraph.UserPermissionsOptions{
		UID:        int32(actor.UID),
		ClientSpec: &sourcegraph.RegisteredClientSpec{ID: actor.ClientID},
	})
	if err != nil {
		return nil, err
	}
	return userPermissions, nil
}
