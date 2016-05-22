// +build exectest

package backend_test

import (
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/testserver"
	"sourcegraph.com/sourcegraph/sourcegraph/util/testutil"
)

func TestDeltas_lg(t *testing.T) {
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

	repo := sourcegraph.RepoSpec{URI: "myrepo"}
	deltaSpec := &sourcegraph.DeltaSpec{
		Base: sourcegraph.RepoRevSpec{RepoSpec: repo, CommitID: commitID},
		Head: sourcegraph.RepoRevSpec{RepoSpec: repo, CommitID: commitID},
	}
	delta, err := a.Client.Deltas.Get(ctx, deltaSpec)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := a.Client.Deltas.ListFiles(ctx, &sourcegraph.DeltasListFilesOp{Ds: delta.DeltaSpec()}); err != nil {
		t.Fatal(err)
	}
}
