// +build exectest,buildtest

package sgx_test

import (
	"testing"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/fed"
	"src.sourcegraph.com/sourcegraph/server/testserver"
	"src.sourcegraph.com/sourcegraph/sgx"
	"src.sourcegraph.com/sourcegraph/util/testutil"
)

func TestBuildRepo_serverside_hosted_lg(t *testing.T) {
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

	commitID, close, err := testutil.CreateAndPushRepo(t, ctx, "r/r")
	if err != nil {
		t.Fatal(err)
	}
	defer close()

	// Build the repo.
	build, _, err := testutil.BuildRepoAndWait(t, ctx, "r/r", commitID)
	if err != nil {
		t.Fatal(err)
	}
	if !build.Success {
		t.Fatalf("build %s failed", build.Spec().IDString())
	}

	checkImport(t, ctx, a.Client, "r/r")
}

func TestBuildRepo_push_hosted_lg(t *testing.T) {
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

	_, close, err := testutil.CreateAndPushRepo(t, ctx, "r/rr")
	if err != nil {
		t.Fatal(err)
	}
	defer close()

	repo, err := a.Client.Repos.Get(ctx, &sourcegraph.RepoSpec{URI: "r/rr"})
	if err != nil {
		t.Fatal(err)
	}

	// Clone and build the repo locally.
	if err := cloneAndLocallyBuildRepo(t, a, repo, ""); err != nil {
		t.Fatal(err)
	}

	checkImport(t, ctx, a.Client, "r/rr")
}

func TestBuildRepo_serverside_hosted_authRequired_lg(t *testing.T) {
	if testserver.Store == "pgsql" {
		t.Skip()
	}

	t.Parallel()

	a, ctx := testserver.NewUnstartedServer()
	a.Config.ServeFlags = append(a.Config.ServeFlags,
		&authutil.Flags{Source: "local"},
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
	ctx = a.AsUID(ctx, int(user.UID))

	repo, done, err := testutil.CreateRepo(t, ctx, "r/r")
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	repo.HTTPCloneURL, _ = sgx.AddUsernamePasswordToCloneURL(repo.HTTPCloneURL, "u", "p")

	commitID, err := testutil.PushRepo(t, ctx, repo)
	if err != nil {
		t.Fatal(err)
	}

	// Build the repo.
	build, _, err := testutil.BuildRepoAndWait(t, ctx, "r/r", commitID)
	if err != nil {
		t.Fatal(err)
	}
	if !build.Success {
		t.Fatalf("build %s failed", build.Spec().IDString())
	}

	checkImport(t, ctx, a.Client, "r/r")
}

func TestBuildRepo_push_hosted_authRequired_lg(t *testing.T) {
	if testserver.Store == "pgsql" {
		t.Skip()
	}

	t.Parallel()

	a, ctx := testserver.NewUnstartedServer()
	a.Config.ServeFlags = append(a.Config.ServeFlags,
		&authutil.Flags{Source: "local"},
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
	ctx = a.AsUID(ctx, int(user.UID))

	repo, done, err := testutil.CreateRepo(t, ctx, "r/r")
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	repo.HTTPCloneURL, _ = sgx.AddUsernamePasswordToCloneURL(repo.HTTPCloneURL, "u", "p")

	if _, err := testutil.PushRepo(t, ctx, repo); err != nil {
		t.Fatal(err)
	}

	// Clone and build the repo locally.
	if err := cloneAndLocallyBuildRepo(t, a, repo, "u"); err != nil {
		t.Fatal(err)
	}

	checkImport(t, ctx, a.Client, "r/r")
}
