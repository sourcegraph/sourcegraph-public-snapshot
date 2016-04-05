package accesscontrol

import (
	"testing"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
)

func TestVerifyAccess(t *testing.T) {
	asUID := func(uid int) context.Context {
		var user *sourcegraph.User
		switch uid {
		case 1:
			user = &sourcegraph.User{
				UID:   1,
				Write: true,
				Admin: true,
			}
		case 2:
			user = &sourcegraph.User{
				UID:   2,
				Write: true,
			}
		default:
			user = &sourcegraph.User{
				UID: int32(uid),
			}
		}
		return auth.WithActor(context.Background(), auth.GetActorFromUser(user))
	}

	var uid int
	var ctx context.Context

	// Test that UID 1 has all access
	uid = 1
	ctx = asUID(uid)

	if err := VerifyUserHasWriteAccess(ctx, "Repos.Create", ""); err != nil {
		t.Fatalf("user %v should have write access; got: %v\n", uid, err)
	}
	if err := VerifyUserHasAdminAccess(ctx, "Repos.Create"); err != nil {
		t.Fatalf("user %v should have admin access; got: %v\n", uid, err)
	}

	// Test that UID 2 has only write access
	uid = 2
	ctx = asUID(uid)

	if err := VerifyUserHasWriteAccess(ctx, "Repos.Create", ""); err != nil {
		t.Fatalf("user %v should have write access; got: %v\n", uid, err)
	}
	if err := VerifyUserHasAdminAccess(ctx, "Repos.Create"); err == nil {
		t.Fatalf("user %v should not have admin access; got access\n", uid)
	}

	// Test that UID 3 has no write/admin access, excluding to whitelisted methods
	uid = 3
	ctx = asUID(uid)

	if err := VerifyUserHasWriteAccess(ctx, "Repos.Create", ""); err == nil {
		t.Fatalf("user %v should not have write access; got access\n", uid)
	}
	if err := VerifyUserHasAdminAccess(ctx, "Repos.Create"); err == nil {
		t.Fatalf("user %v should not have admin access; got access\n", uid)
	}
	if err := VerifyUserHasWriteAccess(ctx, "MirrorRepos.cloneRepo", ""); err != nil {
		t.Fatalf("user %v should have MirrorRepos.cloneRepo access; got %v\n", uid, err)
	}

	// Test that unauthed context has no write/admin access
	uid = 0
	ctx = asUID(uid)

	if err := VerifyUserHasWriteAccess(ctx, "Repos.Create", ""); err == nil {
		t.Fatalf("user %v should not have write access; got access\n", uid)
	}
	if err := VerifyUserHasAdminAccess(ctx, "Repos.Create"); err == nil {
		t.Fatalf("user %v should not have admin access; got access\n", uid)
	}
	if err := VerifyUserHasWriteAccess(ctx, "MirrorRepos.cloneRepo", ""); err == nil {
		t.Fatalf("user %v should not have MirrorRepos.cloneRepo access; got access\n", uid)
	}

	// Test that user has read access for their own data, but not other users'
	// data, unless the user is admin.
	uid = 1
	var uid2 int = 2
	ctx = asUID(uid)

	if err := VerifyUserSelfOrAdmin(ctx, "Users.ListEmails", int32(uid)); err != nil {
		t.Fatalf("user %v should have read access; got: %v\n", uid, err)
	}
	// uid = 1 is admin, so they should have access.
	if err := VerifyUserSelfOrAdmin(ctx, "Users.ListEmails", int32(uid2)); err != nil {
		t.Fatalf("user %v should have read access; got: %v\n", uid, err)
	}
	ctx = asUID(uid2)
	if err := VerifyUserSelfOrAdmin(ctx, "Users.ListEmails", int32(uid)); err == nil {
		t.Fatalf("user %v should not have read access; got access\n", uid2)
	}

	// Test that for local auth, all authenticated users have write access,
	// but unauthenticated users don't.
	uid = 0
	ctx = asUID(uid)

	if err := VerifyUserHasWriteAccess(ctx, "Repos.Create", ""); err == nil {
		t.Fatalf("user %v should not have write access; got access\n", uid)
	}

	uid = 1234
	ctx = asUID(uid)

	if err := VerifyUserHasWriteAccess(ctx, "Repos.Create", ""); err == nil {
		t.Fatalf("user %v should not have write access; got access\n", uid)
	}
}

func asUID(uid int) context.Context {
	return auth.WithActor(context.Background(), auth.Actor{
		UID:      uid,
		ClientID: "xyz",
	})
}
