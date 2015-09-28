// +build exectest

package fed_test

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/auth/idkey"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/fed"
	"src.sourcegraph.com/sourcegraph/fed/discover"
	"src.sourcegraph.com/sourcegraph/server/testserver"
	"src.sourcegraph.com/sourcegraph/util/testutil"
)

// TestFederation starts 2 servers (#1 and #2) and tests that API
// calls to #2 for repos that live on #1 are transparently served by
// #2 communicating with #1 on the client's behalf.
func TestFederation(t *testing.T) {
	if testserver.Store == "pgsql" {
		// TODO(sqs): Test when the origin server is DB-based.
		t.Skip()
	}

	a1, ctx1 := testserver.NewUnstartedServer()
	a1.Config.ServeFlags = append(a1.Config.ServeFlags,
		&fed.Flags{IsRoot: true},
		&authutil.Flags{Source: "local", OAuth2AuthServer: true},
	)
	if err := a1.Start(); err != nil {
		t.Fatal(err)
	}
	defer a1.Close()

	_, a1HTTPPort, err := net.SplitHostPort(conf.AppURL(ctx1).Host)
	if err != nil {
		t.Fatal(err)
	}

	// Set the HTTP_DISCOVERY_PORT env var to a1's HTTP port so that
	// when we exec a2, it inherits that value. This means that
	// discovering localhost/foo/bar will hit a1.
	origEnv := os.Getenv("HTTP_DISCOVERY_PORT")
	if err := os.Setenv("HTTP_DISCOVERY_PORT", a1HTTPPort); err != nil {
		t.Fatal(err)
	}
	origVar := discover.TestingHTTPPort
	discover.TestingHTTPPort, _ = strconv.Atoi(a1HTTPPort)
	if err := os.Setenv("HTTP_DISCOVERY_INSECURE", "t"); err != nil {
		t.Fatal(err)
	}
	discover.InsecureHTTP = true
	defer func() {
		// Revert back to original values.
		if err := os.Setenv("HTTP_DISCOVERY_PORT", origEnv); err != nil {
			t.Fatal(err)
		}
		discover.TestingHTTPPort = origVar
		discover.InsecureHTTP = false
		if err := os.Setenv("HTTP_DISCOVERY_INSECURE", ""); err != nil {
			t.Fatal(err)
		}
	}()

	// Start the server (#2) that our client will contact.
	a2, ctx2 := testserver.NewUnstartedServer()
	a2.Config.ServeFlags = append(a2.Config.ServeFlags,
		&fed.Flags{RootURLStr: conf.AppURL(ctx1).String()},
	)
	if err := a2.Start(); err != nil {
		t.Fatal(err)
	}
	defer a2.Close()

	ctx2 = a1.AsUID(ctx2, 1)

	{
		// Register #2 as a client of #1.
		k2 := idkey.FromContext(ctx2)
		jwks, err := k2.MarshalJWKSPublicKey()
		if err != nil {
			t.Fatal(err)
		}
		_, err = a1.Client.RegisteredClients.Create(ctx1, &sourcegraph.RegisteredClient{
			ID:         k2.ID,
			ClientName: "server2",
			JWKS:       string(jwks),
			Type:       sourcegraph.RegisteredClientType_SourcegraphServer,
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	testRepoFederation(t, a1, ctx1, a2, ctx2)
	testUserFederation(t, a1, ctx1, a2, ctx2)
}

// testRepoFederation tests that #2 serves #1's repos to the client by
// transparently communicating with #1.
func testRepoFederation(t *testing.T, a1 *testserver.Server, ctx1 context.Context, a2 *testserver.Server, ctx2 context.Context) {
	// Create the repo that #1 owns.
	_, done, err := testutil.CreateRepo(t, ctx1, "a/b")
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	{
		// Check that discovery finds the repo on server #1.
		info, err := discover.Repo(context.Background(), "localhost/a/b")
		if err != nil {
			t.Fatal(err)
		}

		// Determine whether discovery was successful by seeing the gRPC
		// and HTTP endpoints that `info` holds.
		ctx, err := info.NewContext(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if v, want := sourcegraph.GRPCEndpoint(ctx), sourcegraph.GRPCEndpoint(ctx1); *v != *want {
			t.Errorf("discovery: got GRPC endpoint == %q, want %q", v, want)
		}
	}

	// Try to access the repo from #2's API.
	repo, err := a2.Client.Repos.Get(ctx2, &sourcegraph.RepoSpec{URI: "localhost/a/b"})
	if err != nil {
		if grpc.Code(err) == codes.NotFound {
			// To help debug, list all repos that DO exist on #1.
			listAllRepos(t, ctx1, "server #1")
			listAllRepos(t, ctx2, "server #2")
		}
		t.Fatal(err)
	}
	if want := "a/b"; repo.URI != want {
		t.Errorf("got repo.URI == %q, want %q", repo.URI, want)
	}

	{
		// Check that server #2 doesn't have the discovery meta
		// tags. Only the server that owns the repo should advertise
		// the meta tags.

		orig := discover.TestingHTTPPort
		_, a2HTTPPort, err := net.SplitHostPort(conf.AppURL(ctx2).Host)
		if err != nil {
			t.Fatal(err)
		}
		discover.TestingHTTPPort, _ = strconv.Atoi(a2HTTPPort)

		// Try to run discovery against server #2. It should succeed
		// and return server #2's information.
		//
		// TODO(sqs): Make server #2 return the info of the servers
		// that actually host the repos, instead of its own info, to
		// cut down on intermediate proxying.
		repos := []string{"localhost/a/b", "localhost/localhost/a/b"}
		for _, repo := range repos {
			if _, err := discover.Repo(context.Background(), repo); err != nil {
				t.Fatalf("Discover %q on server #2: got err == %v", repo, err)
			}
		}

		// Revert to previous value.
		discover.TestingHTTPPort = orig
	}
}

// testUserFederation tests that #2 serves #1's users to the client by
// transparently communicating with #1.
func testUserFederation(t *testing.T, a1 *testserver.Server, ctx1 context.Context, a2 *testserver.Server, ctx2 context.Context) {
	// Create the user that #1 owns.
	user1, err := testutil.CreateAccount(t, ctx1, "alice")
	if err != nil {
		t.Fatal(err)
	}

	{
		// Ensure that since the user was created directly on #1, it
		// doesn't have a domain set (the Domain field is only set on
		// users when they are created/fetched via federation).
		if user1.Domain != "" {
			t.Errorf("got user1.Domain == %q, want empty", user1.Domain)
		}

		// Check that this is still true, even when we re-fetch the user.
		user1Obj, err := a1.Client.Users.Get(ctx1, &sourcegraph.UserSpec{Login: "alice"})
		if err != nil {
			t.Fatal(err)
		}
		if user1Obj.Domain != "" {
			t.Errorf("got user1Obj.Domain == %q, want empty", user1Obj.Domain)
		}
	}

	wantDomain := conf.AppURL(ctx1).Host

	// Try to fetch the user from #2's API, both with *and* without an
	// explicit domain specified.
	{
		// Without an explicit domain, it should fail -- because the user
		// is not on #2.
		//
		// TODO(sqs): I temporarily made this succeed by using the
		// federation root domain. Is this desirable?
		_, err := a2.Client.Users.Get(ctx2, &sourcegraph.UserSpec{Login: "alice", Domain: ""})
		// if grpc.Code(err) != codes.NotFound {
		// 	t.Fatalf("get alice from #2: got err %v, want codes.NotFound", err)
		// }
		if err != nil {
			t.Fatal(err)
		}
	}
	{
		// With an explicit domain of "localhost", it should succeed,
		// because that's where the user's account lives.
		user, err := a2.Client.Users.Get(ctx2, &sourcegraph.UserSpec{Login: "alice", Domain: wantDomain})
		if err != nil {
			if grpc.Code(err) == codes.NotFound {
				// To help debug, list all users that DO exist on #1.
				listAllUsers(t, ctx1, fmt.Sprintf("server #1 (with domain %q)", wantDomain))
			}
			t.Fatal(err)
		}

		// The returned user's Domain field should always be set.
		want := sourcegraph.UserSpec{Login: "alice", Domain: wantDomain, UID: user.UID}
		if user.Spec() != want {
			t.Errorf("got user == %+v, want %+v", user.Spec(), want)
		}
	}
}

func listAllRepos(t *testing.T, ctx context.Context, label string) {
	repos, err := sourcegraph.NewClientFromContext(ctx).Repos.List(ctx, &sourcegraph.RepoListOptions{})
	if err == nil {
		t.Logf("%s has %d repos", label, len(repos.Repos))
		for _, repo := range repos.Repos {
			t.Logf(" - %s", repo.URI)
		}
	} else {
		t.Errorf("warning: listing repos on %s failed: %s", label, err)
	}

}

func listAllUsers(t *testing.T, ctx context.Context, label string) {
	users, err := sourcegraph.NewClientFromContext(ctx).Users.List(ctx, &sourcegraph.UsersListOptions{})
	if err == nil {
		t.Logf("%s has %d users", label, len(users.Users))
		for _, user := range users.Users {
			t.Logf(" - %s (domain %q, UID %d)", user.Login, user.Domain, user.UID)
		}
	} else {
		t.Errorf("warning: listing users on %s failed: %s", label, err)
	}
}
