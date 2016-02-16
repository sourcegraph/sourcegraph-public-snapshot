package accesscontrol

import (
	"strings"

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

// SetMirrorRepoPerms checks if the PrivateMirrors feature is enabled
// and the actor corresponds to a logged-in user, and sets the
// appropriate waitlist state and repo permissions for the actor.
func SetMirrorRepoPerms(ctx context.Context, actor *auth.Actor) {
	if !authutil.ActiveFlags.PrivateMirrors || actor == nil || actor.UID == 0 {
		return
	}

	if authutil.ActiveFlags.MirrorsWaitlist != "none" {
		waitlistedUser, err := store.WaitlistFromContext(ctx).GetUser(elevatedActor(ctx), int32(actor.UID))
		if err != nil {
			if _, ok := err.(*store.WaitlistedUserNotFoundError); !ok {
				log15.Debug("Error fetching waitlisted user", "uid", actor.UID, "error", err)
			}
			return
		}

		if waitlistedUser.GrantedAt == nil {
			// User is on the waitlist.
			actor.MirrorsWaitlist = true
			return
		}
	}

	actor.PrivateMirrors = true

	// User has access to private mirrors. Save the visible repos
	// in the context actor.
	userRepos, err := store.RepoPermsFromContext(ctx).ListUserRepos(elevatedActor(ctx), int32(actor.UID))
	if err != nil {
		log15.Debug("Error listing visible repos for user", "uid", actor.UID, "error", err)
		return
	}

	visibleRepos := make(map[string]bool)
	for _, repo := range userRepos {
		visibleRepos[repo] = true
	}
	actor.RepoPerms = &repoPerms{visibleRepos: visibleRepos}
}

// VerifyRepoPerms checks if a repoURI is visible to the actor in the given context.
func VerifyRepoPerms(ctx context.Context, actor auth.Actor, method, repoURI string) error {
	for scope := range actor.Scope {
		if strings.HasPrefix(scope, "repo:") {
			repo := strings.TrimPrefix(scope, "repo:")
			if repo == repoURI {
				return nil
			}
		}
	}

	// Short-circuit: If the user is unauthenticated and the scope has special access,
	// grant access directly to avoid infinite cycles back to this function from reposStore.Get.
	if actor.UID == 0 && VerifyScopeHasAccess(ctx, actor.Scope, method) {
		return nil
	}

	// Confirm that the repo is private.
	if r, err := store.ReposFromContext(ctx).Get(elevatedActor(ctx), repoURI); err != nil {
		return err
	} else if !r.Private {
		return nil
	}

	err := grpc.Errorf(codes.Unauthenticated, "repo not available (%s): user does not have access", method)
	if actor.UID == 0 || !actor.PrivateMirrors || actor.RepoPerms == nil {
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
// If PrivateMirrors is not enabled, or if the actor has access to all private repos (eg. internal command)
// then the returned slice will be nil.
// If the slice is non-nil but empty, the actor has access to no private repos on this server.
func GetActorPrivateRepos(ctx context.Context, actor auth.Actor, method string) []string {
	if !authutil.ActiveFlags.PrivateMirrors {
		return nil
	}

	if VerifyScopeHasAccess(ctx, actor.Scope, method) {
		return nil
	}

	privateRepos := make([]string, 0)

	if actor.UID == 0 || !actor.PrivateMirrors || actor.RepoPerms == nil {
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

// GetActorPrivateRepoMap returns the list of private repos visible to the context actor.
// If PrivateMirrors is not enabled, or if the actor has access to all private repos (eg. internal command)
// then the returned map will be nil.
// If the map is non-nil but empty, the actor has access to no private repos on this server.
func GetActorPrivateRepoMap(ctx context.Context, method string) map[string]bool {
	var visiblePrivateRepos map[string]bool
	actorPrivateRepos := GetActorPrivateRepos(ctx, auth.ActorFromContext(ctx), method)

	// Note: if actorPrivateRepos is nil, either the mirrors next feature
	// is not enabled, or the actor has access to all private repos on this
	// server.
	if actorPrivateRepos != nil {
		visiblePrivateRepos = make(map[string]bool)
		for _, repo := range actorPrivateRepos {
			visiblePrivateRepos[repo] = true
		}
	}

	return visiblePrivateRepos
}

// elevatedActor returns an actor with admin access to the stores.
//
// CAUTION: use this function only in cases where it is required
// to complete an operation with elevated access, for example when
// creating an account when a user signs up. DO NOT USE this actor
// to complete requests that will return store data in the response.
func elevatedActor(ctx context.Context) context.Context {
	return auth.WithActor(ctx, auth.Actor{Scope: map[string]bool{"internal:elevated": true}})
}
