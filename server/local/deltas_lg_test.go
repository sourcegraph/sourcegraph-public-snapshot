// +build exectest

package local_test

import (
	"testing"

	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/server/testserver"
	"src.sourcegraph.com/sourcegraph/util/testutil"
)

func TestDeltas_lg(t *testing.T) {
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
