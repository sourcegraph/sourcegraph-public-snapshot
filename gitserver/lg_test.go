// +build exectest

package gitserver_test

import (
	"net/url"
	"testing"

	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/fed"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/server/testserver"
	"src.sourcegraph.com/sourcegraph/util/testutil"
)

func TestGitClone_noAuth(t *testing.T) {
	testGitClone_noAuth_withCloneArgs(t, nil)
}

func TestGitClone_noAuth_shallowClone(t *testing.T) {
	testGitClone_noAuth_withCloneArgs(t, []string{"--depth=1"})
}

func testGitClone_noAuth_withCloneArgs(t *testing.T, cloneArgs []string) {
	if testserver.Store == "pgsql" {
		t.Skip()
	}

	t.Parallel()

	a, ctx := testserver.NewUnstartedServer()
	a.Config.ServeFlags = append(a.Config.ServeFlags,
		&authutil.Flags{Source: "none", AllowAnonymousReaders: true},
	)
	if err := a.Start(); err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	_, done, err := testutil.CreateAndPushRepo(t, ctx, "myrepo")
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	repo, err := a.Client.Repos.Get(ctx, &sourcegraph.RepoSpec{URI: "myrepo"})
	if err != nil {
		t.Fatal(err)
	}

	if err := testutil.CloneRepo(t, repo.HTTPCloneURL, "", cloneArgs); err != nil {
		t.Fatalf("git clone %v %s: %s", cloneArgs, repo.HTTPCloneURL, err)
	}
}

func TestGitClone_authRequired(t *testing.T) {
	if testserver.Store == "pgsql" {
		t.Skip()
	}

	t.Parallel()

	a, ctx := testserver.NewUnstartedServer()
	a.Config.ServeFlags = append(a.Config.ServeFlags,
		&authutil.Flags{Source: "local", AllowAllLogins: true},
		&fed.Flags{IsRoot: true},
	)
	if err := a.Start(); err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	user, err := a.Client.Accounts.Create(ctx, &sourcegraph.NewAccount{Login: "u", Email: "u@example.com", Password: "p"})
	if err != nil {
		t.Fatal(err)
	}

	authedCtx := a.AsUIDWithScope(ctx, int(user.UID), []string{"user:write"})

	_, done, err := testutil.CreateRepo(t, authedCtx, "myrepo")
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	repo, err := a.Client.Repos.Get(authedCtx, &sourcegraph.RepoSpec{URI: "myrepo"})
	if err != nil {
		t.Fatal(err)
	}

	authedCloneURL, err := url.Parse(repo.HTTPCloneURL)
	if err != nil {
		t.Fatal(err)
	}
	authedCloneURL.User = url.UserPassword("u", "p")
	unauthedCloneURL := repo.HTTPCloneURL
	repo.HTTPCloneURL = authedCloneURL.String()

	if _, err := testutil.PushRepo(t, authedCtx, repo, nil); err != nil {
		t.Fatal(err)
	}

	// Can't clone if unauthed.
	if err := testutil.CloneRepo(t, unauthedCloneURL, "", nil); err == nil {
		t.Fatalf("git clone %s: err == nil, wanted auth denied", unauthedCloneURL)
	}

	// Can clone if authed.
	if err := testutil.CloneRepo(t, authedCloneURL.String(), "", nil); err != nil {
		t.Fatalf("git clone %s: %s", authedCloneURL, err)
	}
}
