package inventory_test

import (
	"reflect"
	"testing"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/testutil"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/testserver"
)

func TestBuildRepo_serverside_hosted_lg(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Skip("flaky") // https://circleci.com/gh/sourcegraph/sourcegraph/10279

	t.Parallel()

	a, ctx := testserver.NewUnstartedServer()
	a.Config.Serve.NoWorker = true
	if err := a.Start(); err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	repo, done, err := testutil.CreateRepo(t, ctx, "myrepo", false)
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	// Sample Go files for language detection.
	files := map[string]string{"a.go": "a"}
	if err := testutil.PushRepo(t, ctx, repo.HTTPCloneURL, repo.HTTPCloneURL, files, false); err != nil {
		t.Fatal(err)
	}

	// Check inventory.
	cl, _ := sourcegraph.NewClientFromContext(ctx)
	inv, err := cl.Repos.GetInventory(ctx, &sourcegraph.RepoRevSpec{Repo: repo.ID})
	if err != nil {
		t.Fatal(err)
	}
	want := &inventory.Inventory{Languages: []*inventory.Lang{{Name: "Go", TotalBytes: 1, Type: "programming"}}}
	if !reflect.DeepEqual(inv, want) {
		t.Errorf("got inventory %+v, want %+v", inv, want)
	}

	// Check that repo.Language was automatically set.
	time.Sleep(1 * time.Second)
	repo, err = cl.Repos.Get(ctx, &sourcegraph.RepoSpec{ID: repo.ID})
	if err != nil {
		t.Fatal(err)
	}
	if want := "Go"; repo.Language != want {
		t.Errorf("got language %q, want %q", repo.Language, want)
	}
}
