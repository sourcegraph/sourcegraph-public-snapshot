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

	if authutil.ActiveFlags.MirrorsWaitlist != "none" {
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
		}
	}

	actor.MirrorsNext = true

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
func VerifyRepoPerms(ctx context.Context, actor auth.Actor, method, repoURI string) error {
	// Confirm that the repo is private.
	repoStore := store.ReposFromContextOrNil(ctx)
	if repoStore == nil {
		return grpc.Errorf(codes.Unimplemented, "no repo store in context", method)
	}
	if r, err := repoStore.Get(ctx, repoURI); err != nil {
		return err
	} else if !r.Private {
		return nil
	}

	err := grpc.Errorf(codes.PermissionDenied, "repo not available (%s): user does not have access", method)
	if actor.UID == 0 {
		// If the user is unauthenticated, check if the scope has special access.
		if VerifyScopeHasAccess(ctx, actor.Scope, method) {
			return nil
		} else {
			return grpc.Errorf(codes.PermissionDenied, "repo not available (%s): scope does not have access (%#v)", method, actor.Scope)
		}
	}

	if !actor.MirrorsNext || actor.RepoPerms == nil {
		return err
	}

	perms, ok := actor.RepoPerms.(*repoPerms)
	if !ok || perms.visibleRepos == nil {
		return err
	}

	if val, ok := perms.visibleRepos[repoURI]; !ok || !val {
		return err
	}

	return nil
}

// GetActorPrivateRepos returns the list of private repos visible to the current actor.
func GetActorPrivateRepos(ctx context.Context, actor auth.Actor, method string) []string {
	privateRepos := make([]string, 0)

	if !authutil.ActiveFlags.MirrorsNext || actor.UID == 0 || !actor.MirrorsNext || actor.RepoPerms == nil {
		return privateRepos
	}

	perms, ok := actor.RepoPerms.(*repoPerms)
	if !ok || perms.visibleRepos == nil {
		return privateRepos
	}

	for r, v := range perms.visibleRepos {
		if !v {
			continue
		}
		privateRepos = append(privateRepos, r)
	}

	return privateRepos
}
