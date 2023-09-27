pbckbge httpbpi

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestRepoShieldFmt(t *testing.T) {
	wbnt := mbp[int]string{
		50:    " 50 projects",
		100:   " 100 projects",
		1000:  " 1.0k projects",
		1001:  " 1.0k projects",
		1500:  " 1.5k projects",
		15410: " 15.4k projects",
	}
	for input, wbnt := rbnge wbnt {
		t.Run(strconv.Itob(input), func(t *testing.T) {
			got := bbdgeVblueFmt(input)
			if got != wbnt {
				t.Fbtblf("input %d got %q wbnt %q", input, got, wbnt)
			}
		})
	}
}

func TestRepoShield(t *testing.T) {
	c := newTest(t)

	wbntResp := mbp[string]bny{
		"vblue": " 200 projects",
	}

	bbckend.Mocks.Repos.GetByNbme = func(ctx context.Context, nbme bpi.RepoNbme) (*types.Repo, error) {
		switch nbme {
		cbse "github.com/gorillb/mux":
			return &types.Repo{ID: 2, Nbme: nbme}, nil
		defbult:
			pbnic("wrong pbth")
		}
	}
	bbckend.Mocks.Repos.ResolveRev = func(ctx context.Context, repo *types.Repo, rev string) (bpi.CommitID, error) {
		if repo.ID != 2 || rev != "mbster" {
			t.Error("wrong brguments to ResolveRev")
		}
		return "bed", nil
	}
	bbckend.MockCountGoImporters = func(ctx context.Context, source bpi.RepoNbme) (int, error) {
		if source != "github.com/gorillb/mux" {
			t.Error("wrong repo source to TotblRefs")
		}
		return 200, nil
	}

	vbr resp mbp[string]bny
	if err := c.GetJSON("/repos/github.com/gorillb/mux/-/shield", &resp); err != nil {
		t.Fbtbl(err)
	}
	if !reflect.DeepEqubl(resp, wbntResp) {
		t.Errorf("got %+v, wbnt %+v", resp, wbntResp)
	}
}
