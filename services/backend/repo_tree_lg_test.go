// +build exectest

package backend_test

import (
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/testserver"
	"sourcegraph.com/sourcegraph/sourcegraph/util/testutil"
)

func TestRepoTree_Search_lg(t *testing.T) {
	t.Parallel()

	a, ctx := testserver.NewUnstartedServer()
	if err := a.Start(); err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	_, commitID, done, err := testutil.CreateAndPushRepo(t, ctx, "myrepo")
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	repoRev := sourcegraph.RepoRevSpec{RepoSpec: sourcegraph.RepoSpec{URI: "myrepo"}, CommitID: commitID}
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
