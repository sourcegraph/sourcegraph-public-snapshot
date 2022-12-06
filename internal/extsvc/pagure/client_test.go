package pagure

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/testutil"
)

var update = flag.Bool("update", false, "update testdata")

func TestClient_ListProjects(t *testing.T) {
	cli, save := NewTestClient(t, "ListRepos", *update)
	defer save()

	ctx := context.Background()
	limit := 5

	args := ListProjectsArgs{
		Cursor:  &Pagination{PerPage: limit, Page: 1},
		Fork:    true,
		Pattern: "tmux",
	}

	it := cli.ListProjects(ctx, args)

	var projects []*Project
	for i := 0; i < limit && it.Next(); i++ {
		projects = append(projects, it.Current())
	}

	if err := it.Err(); err != nil {
		t.Fatal(err)
	}

	// TODO We wrap the golden to make the diff where we only return projects
	// cleaner to review. Can be removed in future.
	resp := map[string]any{
		"projects": projects,
	}
	testutil.AssertGolden(t, "testdata/golden/ListProjects.json", *update, resp)
}

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log15.Root().SetHandler(log15.LvlFilterHandler(log15.LvlError, log15.Root().GetHandler()))
	}
	os.Exit(m.Run())
}
