package accesscontrol

import (
	"gopkg.in/inconshreveable/log15.v2"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/store"
)

type repoPerms struct {
	visibleRepos map[string]bool
}

// SetMirrorRepoPerms checks if the MirrorsNext feature is enabled
// and the actor corresponds to a logged-in user, and sets the
// appropriate waitlist state and repo permissions for the actor.
func SetMirrorRepoPerms(ctx context.Context, actor *auth.Actor) *auth.Actor {
	if !authutil.ActiveFlags.MirrorsNext || actor == nil || actor.UID == 0 {
		return actor
	}

	waitlistStore := store.WaitlistFromContextOrNil(ctx)
	if waitlistStore == nil {
		log15.Debug("Waitlist store unavailable")
		return actor
	}

	waitlistedUser, err := waitlistStore.GetUser(ctx, int32(actor.UID))
	if err != nil {
		if err, ok := err.(*store.WaitlistedUserNotFoundError); !ok {
			log15.Debug("Error fetching waitlisted user", "uid", actor.UID, "error", err)
		}
		return actor
	}

	if waitlistedUser.GrantedAt == nil {
		// User is on the waitlist.
		actor.MirrorsWaitlist = true
		return actor
	} else {
		actor.MirrorsNext = true
	}

	// User has access to private mirrors. Save the visible repos
	// in the context actor.
	repoPermsStore := store.RepoPermsFromContextOrNil(ctx)
	if repoPermsStore == nil {
		log15.Debug("Repo perms store unavailable")
		return actor
	}

	userRepos, err := repoPermsStore.ListUserRepos(ctx, int32(actor.UID))
	if err != nil {
		log15.Debug("Error listing visible repos for user", "uid", actor.UID, "error", err)
		return actor
	}

	visibleRepos := make(map[string]bool)
	for _, repo := range userRepos {
		visibleRepos[repo] = true
	}
	actor.RepoPerms = &repoPerms{visibleRepos: visibleRepos}
	return actor
}

// VerifyRepoPerms checks if a repoURI is visible to the actor in the given context.
func VerifyRepoPerms(actor auth.Actor, method, repoURI string) error {
	if actor.UID == 0 {
		return grpc.Errorf(codes.Unauthenticated, "repo not available (%s): no authenticated user in current context", method)
	}

	err := grpc.Errorf(codes.PermissionDenied, "repo not available (%s): user does not have access", method)
	if !actor.MirrorsNext || actor.RepoPerms == nil {
		return err
	}

	perms, ok := actor.RepoPerms.(*repoPerms)
	if !ok || perms.visibleRepos == nil {
		return err
	}

	if val, ok := perms.visibleRepos[repoURI]; !ok || val == false {
		return err
	}

	return nil
}
