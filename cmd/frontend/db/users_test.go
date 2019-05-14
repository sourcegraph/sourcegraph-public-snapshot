package db

import (
	"reflect"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/pkg/db/globalstatedb"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
)

// usernamesForTests is a list of test cases containing valid and invalid usernames and org names.
var usernamesForTests = []struct {
	name      string
	wantValid bool
}{
	{"nick", true},
	{"n1ck", true},
	{"Nick2", true},
	{"N-S", true},
	{"nick-s", true},
	{"renfred-xh", true},
	{"renfred-x-h", true},
	{"deadmau5", true},
	{"deadmau-5", true},
	{"3blindmice", true},
	{"777", true},
	{"7-7", true},
	{"long-butnotquitelongenoughtoreachlimit", true},

	{"nick-", false},
	{"nick--s", false},
	{"nick--sny", false},
	{"nick.com", false},
	{"nick_s", false},
	{"_", false},
	{"_nick", false},
	{"nick_", false},
	{"ke$ha", false},
	{"ni%k", false},
	{"#nick", false},
	{"@nick", false},
	{"", false},
	{"nick s", false},
	{" ", false},
	{"-", false},
	{"--", false},
	{"-s", false},
	{"レンフレッド", false},
	{"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", false},
}

func TestUsers_ValidUsernames(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)

	for _, test := range usernamesForTests {
		t.Run(test.name, func(t *testing.T) {
			valid := true
			if _, err := Users.Create(ctx, NewUser{Username: test.name}); err != nil {
				e, ok := err.(errCannotCreateUser)
				if ok && (e.Code() == "users_username_max_length" || e.Code() == "users_username_valid_chars") {
					valid = false
				} else {
					t.Fatal(err)
				}
			}
			if valid != test.wantValid {
				t.Errorf("%q: got valid %v, want %v", test.name, valid, test.wantValid)
			}
		})
	}
}

func TestUsers_Create_SiteAdmin(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)

	if _, err := globalstatedb.Get(ctx); err != nil {
		t.Fatal(err)
	}

	// Create site admin.
	user, err := Users.Create(ctx, NewUser{
		Email:                 "a@a.com",
		Username:              "u",
		Password:              "p",
		EmailVerificationCode: "c",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !user.SiteAdmin {
		t.Fatal("!user.SiteAdmin")
	}

	// Creating a non-site-admin now that the site has already been initialized.
	u2, err := Users.Create(ctx, NewUser{
		Email:                 "a2@a2.com",
		Username:              "u2",
		Password:              "p2",
		EmailVerificationCode: "c2",
	})
	if err != nil {
		t.Fatal(err)
	}
	if u2.SiteAdmin {
		t.Fatal("want u2 not site admin because site is already initialized")
	}
	// Similar to the above, but expect an error because we pass FailIfNotInitialUser: true.
	_, err = Users.Create(ctx, NewUser{
		Email:                 "a3@a3.com",
		Username:              "u3",
		Password:              "p3",
		EmailVerificationCode: "c3",
		FailIfNotInitialUser:  true,
	})
	if want := (errCannotCreateUser{"site_already_initialized"}); err != want {
		t.Fatalf("got error %v, want %v", err, want)
	}

	// Delete the site admin.
	if err := Users.Delete(ctx, user.ID); err != nil {
		t.Fatal(err)
	}

	// Disallow creating a site admin when a user already exists (even if the site is not yet initialized).
	if _, err := dbconn.Global.ExecContext(ctx, "UPDATE site_config SET initialized=false"); err != nil {
		t.Fatal(err)
	}
	u4, err := Users.Create(ctx, NewUser{
		Email:                 "a4@a4.com",
		Username:              "u4",
		Password:              "p4",
		EmailVerificationCode: "c4",
	})
	if err != nil {
		t.Fatal(err)
	}
	if u4.SiteAdmin {
		t.Fatal("want u4 not site admin because site is already initialized")
	}
	// Similar to the above, but expect an error because we pass FailIfNotInitialUser: true.
	if _, err := dbconn.Global.ExecContext(ctx, "UPDATE site_config SET initialized=false"); err != nil {
		t.Fatal(err)
	}
	_, err = Users.Create(ctx, NewUser{
		Email:                 "a5@a5.com",
		Username:              "u5",
		Password:              "p5",
		EmailVerificationCode: "c5",
		FailIfNotInitialUser:  true,
	})
	if want := (errCannotCreateUser{"initial_site_admin_must_be_first_user"}); err != want {
		t.Fatalf("got error %v, want %v", err, want)
	}
}

func TestUsers_CheckAndDecrementInviteQuota(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)

	user, err := Users.Create(ctx, NewUser{
		Email:                 "a@a.com",
		Username:              "u",
		Password:              "p",
		EmailVerificationCode: "c",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Check default invite quota.
	var inviteQuota int
	row := dbconn.Global.QueryRowContext(ctx, "SELECT invite_quota FROM users WHERE id=$1", user.ID)
	if err := row.Scan(&inviteQuota); err != nil {
		t.Fatal(err)
	}
	// Check that it's within some reasonable bounds. The upper bound number here can increased
	// if we increase the default.
	if lo, hi := 0, 15; inviteQuota <= lo || inviteQuota > hi {
		t.Fatalf("got default user invite quota %d, want in [%d,%d)", inviteQuota, lo, hi)
	}

	// Decrementing should succeed while we have remaining quota. Keep going until we exhaust it.
	// Since the quota is fairly low, this isn't too slow.
	for inviteQuota > 0 {
		if ok, err := Users.CheckAndDecrementInviteQuota(ctx, user.ID); !ok || err != nil {
			t.Fatal("initial CheckAndDecrementInviteQuota failed:", err)
		}
		inviteQuota--
	}

	// Now our quota is exhausted, and CheckAndDecrementInviteQuota should fail.
	if ok, err := Users.CheckAndDecrementInviteQuota(ctx, user.ID); ok || err != nil {
		t.Fatalf("over-limit CheckAndDecrementInviteQuota #1: got error %v", err)
	}

	// Check again that we're still over quota, just in case.
	if ok, err := Users.CheckAndDecrementInviteQuota(ctx, user.ID); ok || err != nil {
		t.Fatalf("over-limit CheckAndDecrementInviteQuota #2: got error %v", err)
	}
}

func TestUsers_ListCount(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)

	user, err := Users.Create(ctx, NewUser{
		Email:                 "a@a.com",
		Username:              "u",
		Password:              "p",
		EmailVerificationCode: "c",
	})
	if err != nil {
		t.Fatal(err)
	}
	user.Tags = []string{}

	if count, err := Users.Count(ctx, &UsersListOptions{}); err != nil {
		t.Fatal(err)
	} else if want := 1; count != want {
		t.Errorf("got %d, want %d", count, want)
	}
	if users, err := Users.List(ctx, &UsersListOptions{}); err != nil {
		t.Fatal(err)
	} else if users, want := normalizeUsers(users), normalizeUsers([]*types.User{user}); !reflect.DeepEqual(users, want) {
		t.Errorf("got %+v, want %+v", users, want)
	}

	if count, err := Users.Count(ctx, &UsersListOptions{UserIDs: []int32{}}); err != nil {
		t.Fatal(err)
	} else if want := 0; count != want {
		t.Errorf("got %d, want %d", count, want)
	}
	if users, err := Users.List(ctx, &UsersListOptions{UserIDs: []int32{}}); err != nil {
		t.Fatal(err)
	} else if len(users) > 0 {
		t.Errorf("got %d, want empty", len(users))
	}

	if users, err := Users.List(ctx, &UsersListOptions{}); err != nil {
		t.Fatal(err)
	} else if users, want := normalizeUsers(users), normalizeUsers([]*types.User{user}); !reflect.DeepEqual(users, want) {
		t.Errorf("got %+v, want %+v", users, want)
	}

	if err := Users.Delete(ctx, user.ID); err != nil {
		t.Fatal(err)
	}

	if count, err := Users.Count(ctx, &UsersListOptions{}); err != nil {
		t.Fatal(err)
	} else if want := 0; count != want {
		t.Errorf("got %d, want %d", count, want)
	}
}

func TestUsers_Update(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)

	user, err := Users.Create(ctx, NewUser{
		Email:                 "a@a.com",
		Username:              "u",
		Password:              "p",
		EmailVerificationCode: "c",
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := Users.Update(ctx, user.ID, UserUpdate{
		Username:    "u1",
		DisplayName: strptr("d1"),
		AvatarURL:   strptr("a1"),
	}); err != nil {
		t.Fatal(err)
	}
	user, err = Users.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if want := "u1"; user.Username != want {
		t.Errorf("got username %q, want %q", user.Username, want)
	}
	if want := "d1"; user.DisplayName != want {
		t.Errorf("got display name %q, want %q", user.DisplayName, want)
	}
	if want := "a1"; user.AvatarURL != want {
		t.Errorf("got avatar URL %q, want %q", user.AvatarURL, want)
	}

	if err := Users.Update(ctx, user.ID, UserUpdate{
		DisplayName: strptr(""),
	}); err != nil {
		t.Fatal(err)
	}
	user, err = Users.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if want := "u1"; user.Username != want {
		t.Errorf("got username %q, want %q", user.Username, want)
	}
	if user.DisplayName != "" {
		t.Errorf("got display name %q, want nil", user.DisplayName)
	}
	if want := "a1"; user.AvatarURL != want {
		t.Errorf("got avatar URL %q, want %q", user.AvatarURL, want)
	}

	// Can't update to duplicate username.
	user2, err := Users.Create(ctx, NewUser{
		Email:                 "a2@a.com",
		Username:              "u2",
		Password:              "p2",
		EmailVerificationCode: "c2",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := Users.Update(ctx, user2.ID, UserUpdate{Username: "u1"}); err == nil {
		t.Fatal("want error when updating user to existing username")
	}

	// Can't update nonexistent user.
	if err := Users.Update(ctx, 12345, UserUpdate{Username: "u12345"}); err == nil {
		t.Fatal("want error when updating nonexistent user")
	}
}

func TestUsers_GetByVerifiedEmail(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)

	user, err := Users.Create(ctx, NewUser{
		Email:                 "a@a.com",
		Username:              "u",
		Password:              "p",
		EmailVerificationCode: "c",
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Users.GetByVerifiedEmail(ctx, "a@a.com"); !errcode.IsNotFound(err) {
		t.Errorf("for unverified email, got error %v, want IsNotFound", err)
	}

	if err := UserEmails.SetVerified(ctx, user.ID, "a@a.com", true); err != nil {
		t.Fatal(err)
	}

	gotUser, err := Users.GetByVerifiedEmail(ctx, "a@a.com")
	if err != nil {
		t.Fatal(err)
	}
	if gotUser.ID != user.ID {
		t.Errorf("got user %d, want %d", gotUser.ID, user.ID)
	}
}

func TestUsers_Delete(t *testing.T) {
	for name, hard := range map[string]bool{"": false, "_Hard": true} {
		t.Run("TestUsers_Delete"+name, func(t *testing.T) {
			if testing.Short() {
				t.Skip()
			}
			ctx := dbtesting.TestContext(t)

			otherUser, err := Users.Create(ctx, NewUser{Username: "other"})
			if err != nil {
				t.Fatal(err)
			}

			user, err := Users.Create(ctx, NewUser{
				Email:                 "a@a.com",
				Username:              "u",
				Password:              "p",
				EmailVerificationCode: "c",
			})
			if err != nil {
				t.Fatal(err)
			}

			// Create settings for the user, and for another user authored by this user.
			if _, err := Settings.CreateIfUpToDate(ctx, api.SettingsSubject{User: &user.ID}, nil, &user.ID, "{}"); err != nil {
				t.Fatal(err)
			}
			if _, err := Settings.CreateIfUpToDate(ctx, api.SettingsSubject{User: &otherUser.ID}, nil, &user.ID, "{}"); err != nil {
				t.Fatal(err)
			}

			// Create a repository to comply with the postgres repo constraint.
			if err := Repos.Upsert(ctx, api.InsertRepoOp{Name: "myrepo", Description: "", Fork: false, Enabled: true}); err != nil {
				t.Fatal(err)
			}
			repo, err := Repos.GetByName(ctx, "myrepo")
			if err != nil {
				t.Fatal(err)
			}

			// Create a discussion thread to confirm that deletion properly removes
			// threads and their associated comments.
			newThread, err := DiscussionThreads.Create(ctx, &types.DiscussionThread{
				AuthorUserID: user.ID,
				Title:        "Hello world",
				Target: &types.DiscussionThreadTargetRepo{
					RepoID:   repo.ID,
					Path:     strPtr("foo/bar/mux.go"),
					Branch:   strPtr("master"),
					Revision: strPtr("0c1a96370c1a96370c1a96370c1a96370c1a9637"),
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			newComment, err := DiscussionComments.Create(ctx, &types.DiscussionComment{
				ThreadID:     newThread.ID,
				AuthorUserID: user.ID,
				Contents:     "Thread contents",
			})
			if err != nil {
				t.Fatal(err)
			}

			if hard {
				// Hard delete user.
				if err := Users.HardDelete(ctx, user.ID); err != nil {
					t.Fatal(err)
				}
			} else {
				// Delete user.
				if err := Users.Delete(ctx, user.ID); err != nil {
					t.Fatal(err)
				}
			}

			// User no longer exists.
			_, err = Users.GetByID(ctx, user.ID)
			if !errcode.IsNotFound(err) {
				t.Errorf("got error %v, want ErrUserNotFound", err)
			}
			users, err := Users.List(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(users) > 1 {
				// The otherUser should still exist, which is why we check for 1 not 0.
				t.Errorf("got %d users, want 1", len(users))
			}

			// User's settings no longer exist.
			if settings, err := Settings.GetLatest(ctx, api.SettingsSubject{User: &user.ID}); err != nil {
				t.Error(err)
			} else if settings != nil {
				t.Errorf("got settings %+v, want nil", settings)
			}
			// Settings authored by user still exist but have nil author.
			if settings, err := Settings.GetLatest(ctx, api.SettingsSubject{User: &otherUser.ID}); err != nil {
				t.Fatal(err)
			} else if settings.AuthorUserID != nil {
				t.Errorf("got author %v, want nil", *settings.AuthorUserID)
			}

			// Can't delete already-deleted user.
			err = Users.Delete(ctx, user.ID)
			if !errcode.IsNotFound(err) {
				t.Errorf("got error %v, want ErrUserNotFound", err)
			}

			// Confirm discussion thread/comment no longer exists.
			_, err = DiscussionThreads.Get(ctx, newThread.ID)
			if _, ok := err.(*ErrThreadNotFound); !ok {
				t.Fatal("expected ErrThreadNotFound")
			}
			_, err = DiscussionComments.Get(ctx, newComment.ID)
			if _, ok := err.(*ErrCommentNotFound); !ok {
				t.Fatal("expected ErrCommentNotFound")
			}
		})
	}
}

func normalizeUsers(users []*types.User) []*types.User {
	for _, u := range users {
		u.CreatedAt = u.CreatedAt.Local().Round(time.Second)
		u.UpdatedAt = u.UpdatedAt.Local().Round(time.Second)
	}
	return users
}
