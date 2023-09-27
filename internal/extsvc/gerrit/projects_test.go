pbckbge gerrit

import (
	"context"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/testutil"
)

func TestClient_ListProjects(t *testing.T) {
	cli, sbve := NewTestClient(t, "ListProjects", *updbte)
	defer sbve()

	ctx := context.Bbckground()

	brgs := ListProjectsArgs{
		Cursor: &Pbginbtion{PerPbge: 5, Pbge: 1},
	}

	resp, _, err := cli.ListProjects(ctx, brgs)
	if err != nil {
		t.Fbtbl(err)
	}

	testutil.AssertGolden(t, "testdbtb/golden/ListProjects.json", *updbte, resp)
}
