// +build exectest

package gitserver_test

import (
	"crypto/rand"
	"crypto/rsa"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/fed"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/server/testserver"
	"src.sourcegraph.com/sourcegraph/util/testutil"
)

func TestGitServerWithAnonymousReaders(t *testing.T) {
	testGitServer(t, &authutil.Flags{Source: "none", AllowAnonymousReaders: true},
		gitServerTestExpectations{
			UnauthedCloneHTTP: false,
			AuthedCloneHTTP:   false,
			// TODO: Maybe unauthed cloning over SSH should be
			// enabled with the `AllowAnonymousReaders` flag set.
			UnauthedCloneSSH: true,
			AuthedCloneSSH:   false,
		})
}

func TestGitServerWithAuth(t *testing.T) {
	testGitServer(t, &authutil.Flags{Source: "local", AllowAllLogins: true},
		gitServerTestExpectations{
			UnauthedCloneHTTP: true,
			AuthedCloneHTTP:   false,
			UnauthedCloneSSH:  true,
			AuthedCloneSSH:    false,
		})
}

type gitServerTestExpectations struct {
	UnauthedCloneHTTP, AuthedCloneHTTP bool
	UnauthedCloneSSH, AuthedCloneSSH   bool
}

func testGitServer(t *testing.T, authFlags *authutil.Flags, expect gitServerTestExpectations) {
	if testserver.Store == "pgsql" {
		t.Skip()
	}

	t.Parallel()

	// Start test server.
	server, ctx := runTestServer(t, authFlags)
	defer server.Close()

	// Create a test user.
	_, err := server.Client.Accounts.Create(ctx, &sourcegraph.NewAccount{Login: "u", Email: "u@example.com", Password: "p"})
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
	authedCloneURL, err := url.Parse(repo.HTTPCloneURL)
	if err != nil {
		t.Fatal(err)
	}
	authedCloneURL.User = url.UserPassword("u", "p")

	// Run clone tests on the server.
	cloneArgs := [][]string{
		[]string{},
		[]string{"--depth", "1"},
	}
	for _, cloneArg := range cloneArgs {
		// Test unauthed clones.
		testGitClone(t, "Unauthed HTTP clone", expect.UnauthedCloneHTTP, repo.HTTPCloneURL, nil, cloneArg...)
		testGitClone(t, "Unauthed SSH clone", expect.UnauthedCloneSSH, repo.SSHCloneURL, unlinkedKey, cloneArg...)
		// Test authed clones.
		testGitClone(t, "Authed HTTP clone", expect.AuthedCloneHTTP, authedCloneURL.String(), nil, cloneArg...)
		testGitClone(t, "Authed SSH clone", expect.AuthedCloneSSH, repo.SSHCloneURL, linkedKey, cloneArg...)
	}
}

// TODO: test ssh/http push, with different scopes: "", "user:write", "user:admin"

func runTestServer(t *testing.T, authFlags *authutil.Flags) (*testserver.Server, context.Context) {
	server, ctx := testserver.NewUnstartedServer()
	server.Config.Serve.NoWorker = true
	server.Config.ServeFlags = append(server.Config.ServeFlags, &fed.Flags{IsRoot: true}, authFlags)
	if err := server.Start(); err != nil {
		t.Fatal("Unable to start the test server:", err)
	}
	return server, ctx
}

func testGitClone(t *testing.T, name string, errorExpected bool, cloneURL string, key *rsa.PrivateKey, cloneArgs ...string) {
	err := testutil.CloneRepo(t, cloneURL, "", key, cloneArgs)
	if (errorExpected && err == nil) || (!errorExpected && err != nil) {
		t.Errorf("%s: error expected: %t : git clone %s %s", name, errorExpected, strings.Join(cloneArgs, " "), cloneURL)
	}
}
