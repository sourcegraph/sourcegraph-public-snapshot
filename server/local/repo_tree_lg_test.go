// +build exectest

package local_test

import (
	"testing"

	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/server/testserver"
	"src.sourcegraph.com/sourcegraph/util/testutil"
)

func TestRepoTree_Search_lg(t *testing.T) {
	t.Skip("flaky") // see https://circleci.com/gh/sourcegraph/sourcegraph/5670

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

	repoRev := sourcegraph.RepoRevSpec{RepoSpec: sourcegraph.RepoSpec{URI: "myrepo"}, Rev: "master"}
	_, err = a.Client.RepoTree.Search(ctx, &sourcegraph.RepoTreeSearchOp{
		Rev: repoRev,
		Opt: &sourcegraph.RepoTreeSearchOptions{
			SearchOptions: vcs.SearchOptions{Query: "hello", QueryType: vcs.FixedQuery},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}
