package gerrit

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/testutil"
)

func TestClient_ListProjects(t *testing.T) {
	cli, save := NewTestClient(t, "ListProjects", *update)
	defer save()

	ctx := context.Background()

	args := ListProjectsArgs{
		Cursor: &Pagination{PerPage: 5, Page: 1},
	}

	resp, _, err := cli.ListProjects(ctx, args)
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertGolden(t, "testdata/golden/ListProjects.json", *update, resp)
}
