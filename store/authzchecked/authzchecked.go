package authzchecked

import (
	"errors"
	"os"

	"src.sourcegraph.com/sourcegraph/auth"

	"golang.org/x/net/context"
)

// TODO(public-release): Add wrappers for:
//
// * user settings, org settings, etc., that require the
//   authorized user to be that user or an org member.
//
// * build data fs (will need to use a repo opener or something)
//
// * users - a user can only list their own emails, only sg internal can create users
//
// * graph data - units, defs, etc.
//
// * repo SSH keys
//
// * repo settings
//
// * orgs
//
// * VCS - must have read perms on repo to use VCS
//
// * ext auth tokens - only a user can see these for themselves
//
// etc

// checkActorUID returns ErrForbidden if the context's actor is not an
// authenticated user or if the UID doesn't match the specified UID.
func checkActorUID(ctx context.Context, uid int) error {
	if actor := auth.ActorFromContext(ctx); actor.UID == uid {
		return nil
	}
	return os.ErrPermission
}

// checkSiteAdmin returns ErrSiteAdminOnly if the context's actor is
// not a site admin.
func checkSiteAdmin(ctx context.Context) error {
	if actor := auth.ActorFromContext(ctx); actor.HasAdminAccess() {
		return nil
	}
	return ErrSiteAdminOnly
}

var ErrSiteAdminOnly = errors.New("only site admins may perform this operation")
