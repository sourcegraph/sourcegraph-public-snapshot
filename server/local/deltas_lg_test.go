// +build exectest

package local_test

import (
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/auth/authutil"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/server/testserver"
	"sourcegraph.com/sourcegraph/sourcegraph/util/testutil"
)

func TestDeltas_lg(t *testing.T) {
	t.Parallel()

	a, ctx := testserver.NewUnstartedServer()
	a.Config.ServeFlags = append(a.Config.ServeFlags,
		&authutil.Flags{Source: "none", AllowAnonymousReaders: true},
	)
	if err := a.Start(); err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	_, _, done, err := testutil.CreateAndPushRepo(t, ctx, "myrepo")
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	deltaSpec := &sourcegraph.DeltaSpec{
		Base: sourcegraph.RepoRevSpec{RepoSpec: sourcegraph.RepoSpec{URI: "myrepo"}, Rev: "master"},
		Head: sourcegraph.RepoRevSpec{RepoSpec: sourcegraph.RepoSpec{URI: "myrepo"}, Rev: "master"},
	}
	delta, err := a.Client.Deltas.Get(ctx, deltaSpec)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := a.Client.Deltas.ListFiles(ctx, &sourcegraph.DeltasListFilesOp{Ds: delta.DeltaSpec()}); err != nil {
		t.Fatal(err)
	}
}
