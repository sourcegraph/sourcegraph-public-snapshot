// +build exectest

package search_test

import (
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/server/testserver"
	"src.sourcegraph.com/sourcegraph/util/testutil"
)

func TestSearch(t *testing.T) {
	if testserver.Store == "pgsql" {
		t.Skip()
	}

	a, ctx := testserver.NewUnstartedServer()
	a.Config.ServeFlags = append(a.Config.ServeFlags,
		&authutil.Flags{AllowAllLogins: true},
	)
	if err := a.Start(); err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	_, done, err := testutil.CreateRepo(t, ctx, "a/b")
	if err != nil {
		t.Fatal(err)
	}
	defer done()
	repo, err := a.Client.Repos.Get(ctx, &sourcegraph.RepoSpec{URI: "a/b"})
	if err != nil {
		t.Fatal(err)
	}

	q := &sourcegraph.SearchOptions{
		Query: "a/b",
		Repos: true,
	}
	res, err := a.Client.Search.Search(ctx, q)
	if err != nil {
		t.Fatal(err)
	}

	// Only check certain fields.
	res.RawQuery = sourcegraph.RawQuery{}
	res.Tokens = nil
	res.Plan = nil
	res.ResolvedTokens = nil
	res.Tips = nil
	want := &sourcegraph.SearchResults{
		Repos: []*sourcegraph.Repo{repo},
	}
	if !reflect.DeepEqual(res, want) {
		t.Errorf("got results %+v, want %+v", res, want)
	}
}
