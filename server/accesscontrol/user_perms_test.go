package accesscontrol

import (
	"testing"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func TestVerifyAccess(t *testing.T) {
	authutil.ActiveFlags.Source = "local"
	asUID := func(uid int) context.Context {
		var user sourcegraph.User
		switch uid {
		case 1:
			user = sourcegraph.User{
				UID:   1,
				Write: true,
				Admin: true,
			}
		case 2:
			user = sourcegraph.User{
				UID:   2,
				Write: true,
			}
		default:
			user = sourcegraph.User{
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
}
