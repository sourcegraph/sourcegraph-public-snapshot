package gitserver_test

import (
	"fmt"
	"net/url"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/testutil"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/testserver"
)

func TestGitServer(t *testing.T) {
	t.Skip("Flaky https://app.asana.com/0/87040567695724/150658343890371")
	if testing.Short() {
		t.Skip()
	}

	var tests = []interface{}{
		// Clone test.
		gitCloneTest{false, false, []string{}, false},
		gitCloneTest{false, true, []string{}, false},

		// Shallow clone tests.
		gitCloneTest{false, true, []string{"--depth", "1"}, false},

		// Empty fetch tests.
		gitCloneTest{false, true, []string{}, true},

		// Push tests.
		gitPushTest{true, false, true, false},
		gitPushTest{false, true, true, false},

		// Vary permissions.
		gitPushTest{true, true, false, false},
		gitPushTest{false, true, false, true},
		gitPushTest{false, true, true, true},
	}
	testGitServer(t, tests)
}

type gitCloneTest struct {
	expectError   bool
	authenticated bool
	args          []string
	emptyFetch    bool
}

func (t gitCloneTest) String() string {
	return fmt.Sprintf("Clone over HTTP, expect error %t: authenticated: %t, test empty fetch: %t, args: %v", t.expectError, t.authenticated, t.emptyFetch, t.args)
}

type gitPushTest struct {
	expectError   bool
	authenticated bool
	canWrite      bool // user should have write access.
	isAdmin       bool // user should be admin.
}

func (t gitPushTest) String() string {
	return fmt.Sprintf("Push over HTTP, expect error %t: authenticated: %t, can write: %t, is admin: %t",
		t.expectError, t.authenticated, t.canWrite, t.isAdmin)
}

func testGitServer(t *testing.T, tests []interface{}) {
	t.Parallel()

	// Start test server.
	server, ctx := testserver.NewUnstartedServer()
	server.Config.Serve.NoWorker = true
	if err := server.Start(); err != nil {
		t.Fatal("Unable to start the test server:", err)
	}
	defer server.Close()

	// Create a test user.
	const login = "alice"
	acct, err := server.Client.Accounts.Create(ctx, &sourcegraph.NewAccount{Login: login, Email: "u@example.com", Password: "p"})
	if err != nil {
		t.Fatal(err)
	}

	// Create repo with files.
	repo, _, done, err := testutil.CreateAndPushRepo(t, ctx, "myrepo")
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	authedURL, err := url.Parse(repo.HTTPCloneURL)
	if err != nil {
		t.Fatal(err)
	}
	authedURL.User = url.UserPassword(login, "p")

	remoteURL := func(authenticated bool) string {
		if authenticated {
			return authedURL.String()
		}
		return repo.HTTPCloneURL
	}

	// Run the tests.
	for _, it := range tests {
		switch test := it.(type) {
		case gitCloneTest:
			err := testutil.CloneRepo(t, remoteURL(test.authenticated), "", test.args, test.emptyFetch)
			if (test.expectError && err == nil) || (!test.expectError && err != nil) {
				t.Errorf("FAIL: %s : %v", test.String(), err)
			}
		case gitPushTest:
			user := &sourcegraph.User{UID: acct.UID, Login: login, Write: test.canWrite, Admin: test.isAdmin}
			if _, err = server.Client.Accounts.Update(ctx, user); err != nil {
				t.Errorf("Error while updating user permissions: %s", err)
				continue
			}
			authedCloneURL, err := testutil.AddSystemAuthToURL(ctx, "internal:write", repo.HTTPCloneURL)
			if err != nil {
				t.Errorf("Error while creating authed clone URL: %s", err)
			}

			err = testutil.PushRepo(t, ctx, remoteURL(test.authenticated), authedCloneURL, map[string]string{"unique.txt": test.String()}, false)
			if (test.expectError && err == nil) || (!test.expectError && err != nil) {
				t.Errorf("FAIL: %s : %v", test.String(), err)
			}
		default:
			t.Errorf("Invalid test type: %T", it)
		}
	}
}
