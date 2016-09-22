package accesscontrol

import (
	"context"
	"reflect"
	"testing"

	opentracing "github.com/opentracing/opentracing-go"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github"
	githubmock "sourcegraph.com/sourcegraph/sourcegraph/services/ext/github/mocks"
)

// testContext with mock stubs for GitHubRepoGetter
func testContext() (context.Context, *githubmock.GitHubRepoGetter) {
	var m githubmock.GitHubRepoGetter
	ctx := context.Background()
	ctx = github.WithRepos(ctx, &m)
	ctx = authpkg.WithActor(ctx, authpkg.Actor{UID: 1, Login: "test"})
	ctx = github.WithMockHasAuthedUser(ctx, true)
	_, ctx = opentracing.StartSpanFromContext(ctx, "dummy")
	return ctx, &m
}

func TestUserHasReadAccessAll(t *testing.T) {
	ctx, mock := testContext()

	type testcase struct {
		title                     string
		inputRepos                []*sourcegraph.Repo
		shouldCallGitHub          bool
		mockGitHubAccessibleRepos []*sourcegraph.Repo
		expRepos                  []*sourcegraph.Repo
	}
	testRepos_ := map[string]*sourcegraph.Repo{
		"a": {URI: "a"},
		"b": {URI: "b", Private: true},
		"c": {URI: "c", Private: true},
		"d": {URI: "d", Private: true},
		"e": {URI: "e", Private: true},
	}
	testRepos := func(uris ...string) (r []*sourcegraph.Repo) {
		for _, uri := range uris {
			r = append(r, testRepos_[uri])
		}
		return
	}

	testcases := []testcase{{
		title:                     "allow public repo access",
		inputRepos:                testRepos("a"),
		shouldCallGitHub:          false,
		mockGitHubAccessibleRepos: nil,
		expRepos:                  testRepos("a"),
	}, {
		title:                     "allow private repo access",
		inputRepos:                testRepos("b"),
		shouldCallGitHub:          true,
		mockGitHubAccessibleRepos: testRepos("b"),
		expRepos:                  testRepos("b"),
	}, {
		title:                     "private repo denied",
		inputRepos:                testRepos("b"),
		shouldCallGitHub:          true,
		mockGitHubAccessibleRepos: nil,
		expRepos:                  nil,
	}, {
		title:                     "public repo access, selected private repo access, inaccessible private repo denied",
		inputRepos:                testRepos("a", "b", "c"),
		shouldCallGitHub:          true,
		mockGitHubAccessibleRepos: testRepos("b"),
		expRepos:                  testRepos("a", "b"),
	}, {
		title:                     "edge case: no input repos",
		inputRepos:                nil,
		shouldCallGitHub:          false,
		mockGitHubAccessibleRepos: nil,
		expRepos:                  nil,
	}, {
		title:                     "private not in list of accessible",
		inputRepos:                testRepos("b"),
		shouldCallGitHub:          true,
		mockGitHubAccessibleRepos: testRepos("c"),
		expRepos:                  nil,
	}, {
		title:                     "public not in list of accessible (still allowed)",
		inputRepos:                testRepos("a"),
		shouldCallGitHub:          false,
		mockGitHubAccessibleRepos: testRepos("c"),
		expRepos:                  testRepos("a"),
	}, {
		title:                     "public not in list of accessible (still allowed) and private not either",
		inputRepos:                testRepos("a", "b"),
		shouldCallGitHub:          true,
		mockGitHubAccessibleRepos: testRepos("c"),
		expRepos:                  testRepos("a"),
	}, {
		title:                     "public and private repos accessible, one private denied",
		inputRepos:                testRepos("a", "b", "c", "d"),
		shouldCallGitHub:          true,
		mockGitHubAccessibleRepos: testRepos("c", "b"),
		expRepos:                  testRepos("a", "b", "c"),
	}}

	for _, test := range testcases {
		calledListAccessible := mock.MockListAccessible(ctx, test.mockGitHubAccessibleRepos)

		gotRepos, err := VerifyUserHasReadAccessAll(ctx, "Repos.List", test.inputRepos)
		if err != nil {
			t.Fatal(err)
		}
		if *calledListAccessible != test.shouldCallGitHub {
			if test.shouldCallGitHub {
				t.Errorf("expected GitHub API to be called for permissions check, but it wasn't")
			} else {
				t.Errorf("did not expect GitHub API to be called for permissions check, but it was")
			}
		}
		if !reflect.DeepEqual(gotRepos, test.expRepos) {
			t.Errorf("in test case %s, expected %+v, got %+v", test.title, test.expRepos, gotRepos)
		}
	}
}

func TestVerifyAccess(t *testing.T) {
	asUID := func(uid int) context.Context {
		var actor auth.Actor
		switch uid {
		case 1:
			actor = auth.Actor{
				UID:   1,
				Write: true,
				Admin: true,
			}
		case 2:
			actor = auth.Actor{
				UID:   2,
				Write: true,
			}
		default:
			actor = auth.Actor{
				UID: uid,
			}
		}
		return auth.WithActor(context.Background(), actor)
	}

	var uid int
	var ctx context.Context

	// Test that UID 1 has all access
	uid = 1
	ctx = asUID(uid)

	if err := VerifyUserHasWriteAccess(ctx, "Repos.Create", nil); err != nil {
		t.Fatalf("user %v should have write access; got: %v\n", uid, err)
	}
	if err := VerifyUserHasAdminAccess(ctx, "Repos.Create"); err != nil {
		t.Fatalf("user %v should have admin access; got: %v\n", uid, err)
	}

	// Test that UID 2 has only write access
	uid = 2
	ctx = asUID(uid)

	if err := VerifyUserHasWriteAccess(ctx, "Repos.Create", nil); err != nil {
		t.Fatalf("user %v should have write access; got: %v\n", uid, err)
	}
	if err := VerifyUserHasAdminAccess(ctx, "Repos.Create"); err == nil {
		t.Fatalf("user %v should not have admin access; got access\n", uid)
	}

	// Test that UID 3 has no write/admin access, excluding to whitelisted methods
	uid = 3
	ctx = asUID(uid)

	if err := VerifyUserHasWriteAccess(ctx, "Repos.Create", nil); err == nil {
		t.Fatalf("user %v should not have write access; got access\n", uid)
	}
	if err := VerifyUserHasAdminAccess(ctx, "Repos.Create"); err == nil {
		t.Fatalf("user %v should not have admin access; got access\n", uid)
	}
	if err := VerifyUserHasWriteAccess(ctx, "MirrorRepos.cloneRepo", nil); err != nil {
		t.Fatalf("user %v should have MirrorRepos.cloneRepo access; got %v\n", uid, err)
	}

	// Test that unauthed context has no write/admin access
	uid = 0
	ctx = asUID(uid)

	if err := VerifyUserHasWriteAccess(ctx, "Repos.Create", nil); err == nil {
		t.Fatalf("user %v should not have write access; got access\n", uid)
	}
	if err := VerifyUserHasAdminAccess(ctx, "Repos.Create"); err == nil {
		t.Fatalf("user %v should not have admin access; got access\n", uid)
	}
	if err := VerifyUserHasWriteAccess(ctx, "MirrorRepos.cloneRepo", nil); err == nil {
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

	if err := VerifyUserHasWriteAccess(ctx, "Repos.Create", nil); err == nil {
		t.Fatalf("user %v should not have write access; got access\n", uid)
	}

	uid = 1234
	ctx = asUID(uid)

	if err := VerifyUserHasWriteAccess(ctx, "Repos.Create", nil); err == nil {
		t.Fatalf("user %v should not have write access; got access\n", uid)
	}
}

func asUID(uid int) context.Context {
	return auth.WithActor(context.Background(), auth.Actor{
		UID: uid,
	})
}
