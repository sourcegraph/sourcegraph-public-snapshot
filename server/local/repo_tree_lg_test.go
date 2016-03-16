// +build exectest

package local_test

import (
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/auth/authutil"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/server/testserver"
	"sourcegraph.com/sourcegraph/sourcegraph/util/testutil"
)

func TestRepoTree_Search_lg(t *testing.T) {
	t.Parallel()

	a, ctx := testserver.NewUnstartedServer()
	a.Config.ServeFlags = append(a.Config.ServeFlags,
		&authutil.Flags{Source: "none", AllowAnonymousReaders: true},
	)
	if err := a.Start(); err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	_, commitID, done, err := testutil.CreateAndPushRepo(t, ctx, "myrepo")
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	repoRev := sourcegraph.RepoRevSpec{RepoSpec: sourcegraph.RepoSpec{URI: "myrepo"}, Rev: "master", CommitID: commitID}
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
