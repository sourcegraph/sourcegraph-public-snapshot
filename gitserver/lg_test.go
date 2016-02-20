// +build exectest

package gitserver_test

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net/url"
	"testing"

	"golang.org/x/crypto/ssh"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/fed"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/server/testserver"
	"src.sourcegraph.com/sourcegraph/util/testutil"
)

func TestGitServerWithAnonymousReaders(t *testing.T) {
	var tests = []interface{}{
		// Clone test.
		gitCloneTest{false, false, "http", []string{}},
		gitCloneTest{false, true, "http", []string{}},
		// TODO: Maybe unauthed cloning over SSH should be
		// enabled with the `AllowAnonymousReaders` flag set.
		gitCloneTest{true, false, "ssh", []string{}},
		gitCloneTest{false, true, "ssh", []string{}},

		// Shallow clone tests.
		gitCloneTest{false, true, "http", []string{"--depth", "1"}},
		gitCloneTest{false, true, "ssh", []string{"--depth", "1"}},

		// Push tests.
		gitPushTest{true, false, "http", true, false},
		gitPushTest{false, true, "http", true, false},
		gitPushTest{true, false, "ssh", true, false},
		gitPushTest{false, true, "ssh", true, false},

		// Vary permissions.
		gitPushTest{true, true, "ssh", false, false},
		gitPushTest{false, true, "ssh", false, true},
		gitPushTest{false, true, "ssh", true, true},
		gitPushTest{true, true, "http", false, false},
		gitPushTest{false, true, "http", false, true},
		gitPushTest{false, true, "http", true, true},
	}
	testGitServer(t, &authutil.Flags{Source: "local", AllowAnonymousReaders: true}, tests)
}

func TestGitServerWithAuth(t *testing.T) {
	var tests = []interface{}{
		// Clone test.
		gitCloneTest{true, false, "http", []string{}},
		gitCloneTest{false, true, "http", []string{}},
		gitCloneTest{true, false, "ssh", []string{}},
		gitCloneTest{false, true, "ssh", []string{}},

		// Shallow clone tests.
		gitCloneTest{false, true, "http", []string{"--depth", "1"}},
		gitCloneTest{false, true, "ssh", []string{"--depth", "1"}},

		// Push tests.
		gitPushTest{true, false, "http", true, false},
		gitPushTest{false, true, "http", true, false},
		gitPushTest{true, false, "ssh", true, false},
		gitPushTest{false, true, "ssh", true, false},

		// Vary permissions.
		gitPushTest{true, true, "ssh", false, false},
		gitPushTest{false, true, "ssh", false, true},
		gitPushTest{false, true, "ssh", true, true},
		gitPushTest{true, true, "http", false, false},
		gitPushTest{false, true, "http", false, true},
		gitPushTest{false, true, "http", true, true},
	}
	testGitServer(t, &authutil.Flags{Source: "local", AllowAllLogins: true}, tests)
}

type gitCloneTest struct {
	expectError   bool
	authenticated bool
	protocol      string // http or ssh
	args          []string
}

func (t gitCloneTest) String() string {
	return fmt.Sprintf("Clone over %s, expect error %t: authenticated: %t, args: %v", t.protocol, t.expectError, t.authenticated, t.args)
}

type gitPushTest struct {
	expectError   bool
	authenticated bool
	protocol      string // http or ssh
	canWrite      bool   // user should have write access.
	isAdmin       bool   // user should be admin.
}

func (t gitPushTest) String() string {
	return fmt.Sprintf("Push over %s, expect error %t: authenticated: %t, can write: %t, is admin: %t",
		t.protocol, t.expectError, t.authenticated, t.canWrite, t.isAdmin)
}

func testGitServer(t *testing.T, authFlags *authutil.Flags, tests []interface{}) {
	t.Parallel()

	// Start test server.
	server, ctx := testserver.NewUnstartedServer()
	server.Config.Serve.NoWorker = true
	server.Config.ServeFlags = append(server.Config.ServeFlags, &fed.Flags{IsRoot: true}, authFlags)
	if err := server.Start(); err != nil {
		t.Fatal("Unable to start the test server:", err)
	}
	defer server.Close()

	// Create a test user.
	userSpec, err := server.Client.Accounts.Create(ctx, &sourcegraph.NewAccount{Login: "u", Email: "u@example.com", Password: "p"})
	if err != nil {
		t.Fatal(err)
	}

	// Link a SSH key to the user.
	linkedKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	linkedKeyPublic, err := ssh.NewPublicKey(&linkedKey.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	_, err = server.Client.UserKeys.AddKey(ctx,
		&sourcegraph.SSHPublicKey{Name: "testkey", Key: ssh.MarshalAuthorizedKey(linkedKeyPublic)})
	if err != nil {
		t.Fatal(err)
	}

	// Create SSH key not linked with the user.
	unlinkedKey, err := rsa.GenerateKey(rand.Reader, 1024)
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
	authedURL.User = url.UserPassword("u", "p")

	remoteURL := func(protocol string, authenticated bool) string {
		if protocol == "ssh" {
			return repo.SSHCloneURL
		} else if authenticated {
			return authedURL.String()
		}
		return repo.HTTPCloneURL
	}
	sshKey := func(protocol string, authenticated bool) *rsa.PrivateKey {
		if protocol != "ssh" {
			return nil
		}
		if authenticated {
			return linkedKey
		}
		return unlinkedKey
	}

	// Run the tests.
	for _, it := range tests {
		switch test := it.(type) {
		case gitCloneTest:
			err := testutil.CloneRepo(remoteURL(test.protocol, test.authenticated), "",
				sshKey(test.protocol, test.authenticated), test.args)
			if (test.expectError && err == nil) || (!test.expectError && err != nil) {
				t.Errorf("FAILED: %s : %v", test.String(), err)
			}
		case gitPushTest:
			user := &sourcegraph.User{UID: userSpec.UID, Login: userSpec.Login, Write: test.canWrite, Admin: test.isAdmin}
			if _, err = server.Client.Accounts.Update(ctx, user); err != nil {
				t.Errorf("Error while updating user permissions: %s", err)
				continue
			}
			err := testutil.PushRepo(ctx, remoteURL(test.protocol, test.authenticated), repo.HTTPCloneURL, sshKey(test.protocol, test.authenticated), map[string]string{"unique.txt": test.String()})
			if (test.expectError && err == nil) || (!test.expectError && err != nil) {
				t.Errorf("FAILED: %s : %v", test.String(), err)
			}
		default:
			t.Errorf("Invalid test type: %T", it)
		}
	}
}
