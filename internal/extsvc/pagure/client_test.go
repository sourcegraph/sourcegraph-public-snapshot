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

	args := ListProjectsArgs{
		Cursor:  &Pagination{PerPage: 5, Page: 1},
		Fork:    true,
		Pattern: "tmux",
	}

	resp, err := cli.ListProjects(ctx, args)
	if err != nil {
		t.Fatal(err)
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
