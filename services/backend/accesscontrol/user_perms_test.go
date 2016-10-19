package accesscontrol

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	opentracing "github.com/opentracing/opentracing-go"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
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
	ctx = authpkg.WithActor(ctx, &authpkg.Actor{UID: "1", Login: "test"})
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
	}, {
		title:                     "preserve input order",
		inputRepos:                testRepos("b", "a"),
		shouldCallGitHub:          true,
		mockGitHubAccessibleRepos: testRepos("b"),
		expRepos:                  testRepos("b", "a"),
	}, {
		title:                     "preserve input order with some denied",
		inputRepos:                testRepos("c", "b", "d", "a"),
		shouldCallGitHub:          true,
		mockGitHubAccessibleRepos: testRepos("c", "d"),
		expRepos:                  testRepos("c", "d", "a"),
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

type MockRepos struct {
	_Get                     func(ctx context.Context, repo int32) (*sourcegraph.Repo, error)
	_GetByURI                func(ctx context.Context, repo string) (*sourcegraph.Repo, error)
	_UnsafeDangerousGetByURI func(ctx context.Context, repo string) (*sourcegraph.Repo, error)
}

func (m *MockRepos) Get(ctx context.Context, repo int32) (*sourcegraph.Repo, error) {
	return m._Get(ctx, repo)
}

func (m *MockRepos) GetByURI(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
	return m._GetByURI(ctx, repo)
}

func (m *MockRepos) UnsafeDangerousGetByURI(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
	return m._UnsafeDangerousGetByURI(ctx, repo)
}

func TestVerifyUserHasReadAccessToDefRepoRefs(t *testing.T) {
	ctx, mock := testContext()

	type testcase struct {
		title                     string
		input                     []*sourcegraph.DeprecatedDefRepoRef
		shouldCallGitHub          bool
		mockGitHubAccessibleRepos []*sourcegraph.Repo
		expected                  []*sourcegraph.DeprecatedDefRepoRef
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
	testDefs := func(uris ...string) (d []*sourcegraph.DeprecatedDefRepoRef) {
		for _, uri := range uris {
			d = append(d, &sourcegraph.DeprecatedDefRepoRef{Repo: uri})
		}
		return
	}

	testcases := []testcase{{
		title:                     "allow public repo access",
		input:                     testDefs("a"),
		shouldCallGitHub:          false,
		mockGitHubAccessibleRepos: testRepos("a"),
		expected:                  testDefs("a"),
	}, {
		title:                     "multiple repo defs",
		input:                     testDefs("a", "b", "c", "d"),
		shouldCallGitHub:          true,
		mockGitHubAccessibleRepos: testRepos("a", "c", "d"),
		expected:                  testDefs("a", "c", "d"),
	}, {
		title:                     "multiple defs same repo",
		input:                     testDefs("a", "c", "a", "d", "a", "b"),
		shouldCallGitHub:          true,
		mockGitHubAccessibleRepos: testRepos("a"),
		expected:                  testDefs("a", "a", "a"),
	}, {
		title:                     "allow private repo access",
		input:                     testDefs("b"),
		shouldCallGitHub:          true,
		mockGitHubAccessibleRepos: testRepos("b"),
		expected:                  testDefs("b"),
	}, {
		title:                     "private repo denied",
		input:                     testDefs("b"),
		shouldCallGitHub:          true,
		mockGitHubAccessibleRepos: nil,
		expected:                  []*sourcegraph.DeprecatedDefRepoRef{},
	}, {
		title:                     "public repo access, selected private repo access, inaccessible private repo denied",
		input:                     testDefs("a", "b", "c"),
		shouldCallGitHub:          true,
		mockGitHubAccessibleRepos: testRepos("b"),
		expected:                  testDefs("a", "b"),
	}, {
		title:                     "edge case: no input repos",
		input:                     nil,
		shouldCallGitHub:          false,
		mockGitHubAccessibleRepos: nil,
		expected:                  []*sourcegraph.DeprecatedDefRepoRef{},
	}, {
		title:                     "private not in list of accessible",
		input:                     testDefs("b"),
		shouldCallGitHub:          true,
		mockGitHubAccessibleRepos: testRepos("c"),
		expected:                  []*sourcegraph.DeprecatedDefRepoRef{},
	}, {
		title:                     "public not in list of accessible (still allowed)",
		input:                     testDefs("a"),
		shouldCallGitHub:          false,
		mockGitHubAccessibleRepos: testRepos("c"),
		expected:                  testDefs("a"),
	}, {
		title:                     "public not in list of accessible (still allowed) and private not either",
		input:                     testDefs("a", "b"),
		shouldCallGitHub:          true,
		mockGitHubAccessibleRepos: testRepos("c"),
		expected:                  testDefs("a"),
	}, {
		title:                     "public and private repos accessible, one private denied",
		input:                     testDefs("a", "b", "c", "d"),
		shouldCallGitHub:          true,
		mockGitHubAccessibleRepos: testRepos("c", "b"),
		expected:                  testDefs("a", "b", "c"),
	}, {
		title:                     "preserve input order",
		input:                     testDefs("b", "a"),
		shouldCallGitHub:          true,
		mockGitHubAccessibleRepos: testRepos("b"),
		expected:                  testDefs("b", "a"),
	}, {
		title:                     "preserve input order with some denied",
		input:                     testDefs("c", "b", "d", "a"),
		shouldCallGitHub:          true,
		mockGitHubAccessibleRepos: testRepos("c", "d"),
		expected:                  testDefs("c", "d", "a"),
	}}

	for _, test := range testcases {
		t.Run(test.title, func(t *testing.T) {
			calledListAccessible := mock.MockListAccessible(ctx, test.mockGitHubAccessibleRepos)

			Repos = &MockRepos{
				_UnsafeDangerousGetByURI: func(ctx context.Context, repo string) (*sourcegraph.Repo, error) {
					if r, ok := testRepos_[repo]; ok {
						return r, nil
					}
					return nil, errors.New("repo not found")
				},
			}

			got, err := VerifyUserHasReadAccessToDefRepoRefs(ctx, "GlobalRefs.Get", test.input)
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
			if !reflect.DeepEqual(got, test.expected) {
				t.Errorf("in test case %s, expected %+v, got %+v", test.title, test.expected, got)
			}
		})
	}
}

func TestVerifyAccess(t *testing.T) {
	var uid string
	var ctx context.Context

	// Test that UID 3 has no write/admin access, excluding to whitelisted methods
	uid = "3"
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
	uid = ""
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
	// data
	uid = "1"
	uid2 := "3"

	ctx = asUID(uid2)
	if err := VerifyUserSelfOrAdmin(ctx, "Users.ListEmails", uid2); err != nil {
		t.Fatalf("user %v should have read access; got: %v\n", uid, err)
	}
	if err := VerifyUserSelfOrAdmin(ctx, "Users.ListEmails", uid); err == nil {
		t.Fatalf("user %v should not have read access; got access\n", uid2)
	}

	// Test that for local auth, all authenticated users have write access,
	// but unauthenticated users don't.
	uid = ""
	ctx = asUID(uid)

	if err := VerifyUserHasWriteAccess(ctx, "Repos.Create", nil); err == nil {
		t.Fatalf("user %v should not have write access; got access\n", uid)
	}

	uid = "1234"
	ctx = asUID(uid)

	if err := VerifyUserHasWriteAccess(ctx, "Repos.Create", nil); err == nil {
		t.Fatalf("user %v should not have write access; got access\n", uid)
	}
}

func TestVerifyActorHasRepoURIAccess(t *testing.T) {
	ctx, mock := testContext()

	tests := []struct {
		title                string
		repoURI              string
		authorizedForPrivate bool // True here simulates that the actor has access to private repos when asking GitHub API. False simulates that they don't.
		want                 bool
	}{
		{
			title:                `private repo URI begins with "github.com/", actor unauthorized for private repo access`,
			repoURI:              "github.com/user/privaterepo",
			authorizedForPrivate: false,
			want:                 false,
		},
		{
			title:                `private repo URI begins with "GitHub.com/", actor unauthorized for private repo access`,
			repoURI:              "GitHub.com/User/PrivateRepo",
			authorizedForPrivate: false,
			want:                 false,
		},
		{
			title:                `private repo URI begins with "github.com/", actor authorized for private repo access`,
			repoURI:              "github.com/user/privaterepo",
			authorizedForPrivate: true,
			want:                 true,
		},
		{
			title:                `private repo URI begins with "GitHub.com/", actor authorized for private repo access`,
			repoURI:              "GitHub.com/User/PrivateRepo",
			authorizedForPrivate: true,
			want:                 true,
		},
		{
			title:   `public repo URI begins with "github.com/"`,
			repoURI: "github.com/user/publicrepo",
			want:    true,
		},
		{
			title:   `public repo URI begins with "GitHub.com/"`,
			repoURI: "GitHub.com/User/PublicRepo",
			want:    true,
		},
		{
			title:   `repo URI begins with "bitbucket.org/"; not supported at this time, so permission denied`,
			repoURI: "bitbucket.org/foo/bar",
			want:    false,
		},
		{
			title:   `repo URI that is local (neither GitHub nor a remote URI)`,
			repoURI: "sourcegraph/sourcegraph",
			want:    true,
		},
	}
	for _, test := range tests {
		// Mocked GitHub API responses differ depending on value of test.authorizedForPrivate.
		// If true, then "github.com/user/privaterepo" repo exists, otherwise it's not found.
		mock.Get_ = func(_ context.Context, uri string) (*sourcegraph.Repo, error) {
			switch uri := strings.ToLower(uri); {
			case uri == "github.com/user/privaterepo" && test.authorizedForPrivate:
				return &sourcegraph.Repo{URI: "github.com/User/PrivateRepo"}, nil
			case uri == "github.com/user/publicrepo":
				return &sourcegraph.Repo{URI: "github.com/User/PublicRepo"}, nil
			default:
				return nil, legacyerr.Errorf(legacyerr.NotFound, "repo not found")
			}
		}

		actor := &auth.Actor{UID: "1"}
		const repoID = 1
		got := VerifyActorHasRepoURIAccess(ctx, actor, "Repos.GetByURI", repoID, test.repoURI)
		if want := test.want; got != want {
			t.Errorf("%s: got %v, want %v", test.title, got, want)
		}
	}
}

func TestLocalRepoURI(t *testing.T) {
	tests := []struct {
		repoURI string
		want    bool
	}{
		{repoURI: "github.com/user/repo", want: false},
		{repoURI: "example.com", want: false},
		{repoURI: "user/repo", want: true},
		{repoURI: "a/b/c", want: true},
		{repoURI: "a/b", want: true},
		{repoURI: "a", want: true},
		{repoURI: "", want: true},
	}
	for _, test := range tests {
		if got, want := localRepoURI(test.repoURI), test.want; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	}
}

func asUID(uid string) context.Context {
	return auth.WithActor(context.Background(), &auth.Actor{
		UID: uid,
	})
}
