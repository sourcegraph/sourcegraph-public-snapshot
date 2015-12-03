package accesscontrol

import (
	"testing"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func TestVerifyAccess(t *testing.T) {
	t.Skip("TODO(pararth): fix failing test after you merge your stuff with my removal of OAuth2")

	getUserPermissionsFromRoot = func(ctx context.Context, actor auth.Actor) (*sourcegraph.UserPermissions, error) {
		switch actor.UID {
		case 1:
			return &sourcegraph.UserPermissions{
				UID:      int32(actor.UID),
				ClientID: actor.ClientID,
				Read:     true,
				Write:    true,
				Admin:    true,
			}, nil
		case 2:
			return &sourcegraph.UserPermissions{
				UID:      int32(actor.UID),
				ClientID: actor.ClientID,
				Read:     true,
				Write:    true,
			}, nil
		case 3:
			return &sourcegraph.UserPermissions{
				UID:      int32(actor.UID),
				ClientID: actor.ClientID,
				Read:     true,
			}, nil
		default:
			return &sourcegraph.UserPermissions{
				UID:      int32(actor.UID),
				ClientID: actor.ClientID,
			}, nil
		}
	}

	var uid int
	var ctx context.Context

	// Test that UID 1 has all access
	uid = 1
	ctx = asUID(uid)

	if err := VerifyUserHasWriteAccess(ctx, "Repos.Create"); err != nil {
		t.Fatalf("user %v should have write access; got: %v\n", uid, err)
	}
	if err := VerifyUserHasAdminAccess(ctx, "Repos.Create"); err != nil {
		t.Fatalf("user %v should have admin access; got: %v\n", uid, err)
	}

	// Test that UID 2 has only write access
	uid = 2
	ctx = asUID(uid)

	if err := VerifyUserHasWriteAccess(ctx, "Repos.Create"); err != nil {
		t.Fatalf("user %v should have write access; got: %v\n", uid, err)
	}
	if err := VerifyUserHasAdminAccess(ctx, "Repos.Create"); err == nil {
		t.Fatalf("user %v should not have admin access; got access\n", uid)
	}

	// Test that UID 3 has no write/admin access
	uid = 3
	ctx = asUID(uid)

	if err := VerifyUserHasWriteAccess(ctx, "Repos.Create"); err == nil {
		t.Fatalf("user %v should not have write access; got access\n", uid)
	}
	if err := VerifyUserHasAdminAccess(ctx, "Repos.Create"); err == nil {
		t.Fatalf("user %v should not have admin access; got access\n", uid)
	}

	// Test that unauthed context has no write/admin access
	uid = 0
	ctx = asUID(uid)

	if err := VerifyUserHasWriteAccess(ctx, "Repos.Create"); err == nil {
		t.Fatalf("user %v should not have write access; got access\n", uid)
	}
	if err := VerifyUserHasAdminAccess(ctx, "Repos.Create"); err == nil {
		t.Fatalf("user %v should not have admin access; got access\n", uid)
	}

	// Test that UID 2 loses write access when it is restricted to admins
	authutil.ActiveFlags.RestrictWriteAccess = true

	uid = 2
	ctx = asUID(uid)

	if err := VerifyUserHasWriteAccess(ctx, "Repos.Create"); err == nil {
		t.Fatalf("user %v should not have write access; got access\n", uid)
	}

	// Test that for local auth, all authenticated users have write access,
	// but unauthenticated users don't.
	authutil.ActiveFlags = authutil.Flags{
		Source: "local",
	}

	uid = 0
	ctx = asUID(uid)

	if err := VerifyUserHasWriteAccess(ctx, "Repos.Create"); err == nil {
		t.Fatalf("user %v should not have write access; got access\n", uid)
	}

	uid = 1234
	ctx = asUID(uid)

	if err := VerifyUserHasWriteAccess(ctx, "Repos.Create"); err != nil {
		t.Fatalf("user %v should have write access; got: %v\n", uid, err)
	}
}

func asUID(uid int) context.Context {
	return auth.WithActor(context.Background(), auth.Actor{
		UID:      uid,
		ClientID: "xyz",
	})
}
