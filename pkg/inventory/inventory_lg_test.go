// +build exectest

package inventory_test

import (
	"reflect"
	"testing"
	"time"

	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/pkg/inventory"
	"src.sourcegraph.com/sourcegraph/server/testserver"
	"src.sourcegraph.com/sourcegraph/util/testutil"
)

func TestBuildRepo_serverside_hosted_lg(t *testing.T) {
	if testserver.Store == "pgsql" {
		t.Skip()
	}

	t.Parallel()

	a, ctx := testserver.NewUnstartedServer()
	a.Config.Serve.NoWorker = true
	a.Config.ServeFlags = append(a.Config.ServeFlags,
		&authutil.Flags{Source: "none", AllowAnonymousReaders: true},
	)
	if err := a.Start(); err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	repo, done, err := testutil.CreateRepo(t, ctx, "myrepo")
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	// Sample Go files for language detection.
	files := map[string]string{"a.go": "a"}
	if _, err := testutil.PushRepo(t, ctx, repo, files); err != nil {
		t.Fatal(err)
	}

	// Check inventory.
	cl, _ := sourcegraph.NewClientFromContext(ctx)
	inv, err := cl.Repos.GetInventory(ctx, &sourcegraph.RepoRevSpec{RepoSpec: repo.RepoSpec()})
	if err != nil {
		t.Fatal(err)
	}
	want := &inventory.Inventory{Languages: []*inventory.Lang{{Name: "Go", TotalBytes: 1, Type: "programming"}}}
	if !reflect.DeepEqual(inv, want) {
		t.Errorf("got inventory %+v, want %+v", inv, want)
	}

	// Check that repo.Language was automatically set.
	time.Sleep(1 * time.Second)
	repo, err = cl.Repos.Get(ctx, &sourcegraph.RepoSpec{URI: repo.URI})
	if err != nil {
		t.Fatal(err)
	}
	if want := "Go"; repo.Language != want {
		t.Errorf("got language %q, want %q", repo.Language, want)
	}
}
