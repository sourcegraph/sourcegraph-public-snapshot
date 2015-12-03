// +build exectest

package fed_test

import (
	"net/url"
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/auth/idkey"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/fed"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
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
		&authutil.Flags{Source: "local", OAuth2AuthServer: true, AllowAllLogins: true},
	)
	if err := a1.Start(); err != nil {
		t.Fatal(err)
	}
	defer a1.Close()

	// Start the server (#2) that our client will contact.
	a2, ctx2 := testserver.NewUnstartedServer()
	a2.Config.ServeFlags = append(a2.Config.ServeFlags,
		&fed.Flags{RootURLStr: conf.AppURL(ctx1).String()},
	)
	if err := a2.Start(); err != nil {
		t.Fatal(err)
	}
	defer a2.Close()

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

	testUserFederation(t, a1, ctx1, a2, ctx2)
}

func urlsEqualIgnoreRootSlash(a, b *url.URL) bool {
	a2 := *a
	b2 := *b
	if a2.Path == "/" {
		a2.Path = ""
	}
	if b2.Path == "/" {
		b2.Path = ""
	}
	return a2 == b2
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

	{
		// Without an explicit domain, it should fall back to the
		// local server, which has no such user.
		_, err := a2.Client.Users.Get(ctx2, &sourcegraph.UserSpec{Login: "alice", Domain: ""})
		if grpc.Code(err) != codes.NotFound {
			t.Fatalf("got err == %v, want NotFound", err)
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
