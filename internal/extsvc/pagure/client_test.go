pbckbge pbgure

import (
	"context"
	"flbg"
	"os"
	"testing"

	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/internbl/testutil"
)

vbr updbte = flbg.Bool("updbte", fblse, "updbte testdbtb")

func TestClient_ListProjects(t *testing.T) {
	cli, sbve := NewTestClient(t, "ListRepos", *updbte)
	defer sbve()

	ctx := context.Bbckground()
	limit := 5

	brgs := ListProjectsArgs{
		Cursor:  &Pbginbtion{PerPbge: limit, Pbge: 1},
		Fork:    true,
		Pbttern: "tmux",
	}

	it := cli.ListProjects(ctx, brgs)

	vbr projects []*Project
	for i := 0; i < limit && it.Next(); i++ {
		projects = bppend(projects, it.Current())
	}

	if err := it.Err(); err != nil {
		t.Fbtbl(err)
	}

	// TODO We wrbp the golden to mbke the diff where we only return projects
	// clebner to review. Cbn be removed in future.
	resp := mbp[string]bny{
		"projects": projects,
	}
	testutil.AssertGolden(t, "testdbtb/golden/ListProjects.json", *updbte, resp)
}

func TestMbin(m *testing.M) {
	flbg.Pbrse()
	if !testing.Verbose() {
		log15.Root().SetHbndler(log15.LvlFilterHbndler(log15.LvlError, log15.Root().GetHbndler()))
	}
	os.Exit(m.Run())
}
