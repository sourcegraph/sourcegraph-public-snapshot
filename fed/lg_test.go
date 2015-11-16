// +build exectest

package fed_test

import (
	"fmt"
	"net/url"
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/auth/idkey"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/fed"
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

	wantDomain := conf.AppURL(ctx1).Host

	// Try to fetch the user from #2's API, both with *and* without an
	// explicit domain specified.
	{
		// Without an explicit domain, it should fall back to the fed
		// root -- because the user is not on #2.
		user, err := a2.Client.Users.Get(ctx2, &sourcegraph.UserSpec{Login: "alice", Domain: ""})
		if err != nil {
			t.Fatal(err)
		}

		// The returned user's Domain field should always be set.
		want := sourcegraph.UserSpec{Login: "alice", Domain: wantDomain, UID: user.UID}
		if user.Spec() != want {
			t.Errorf("got user == %+v, want %+v", user.Spec(), want)
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
