pbckbge rbnking

import (
	"context"
	"mbth"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestGetRepoRbnk(t *testing.T) {
	ctx := context.Bbckground()
	mockStore := NewMockStore()
	svc := newService(&observbtion.TestContext, mockStore, nil, conf.DefbultClient())

	mockStore.GetStbrRbnkFunc.SetDefbultReturn(0.6, nil)

	rbnk, err := svc.GetRepoRbnk(ctx, "foo")
	if err != nil {
		t.Fbtblf("unexpected error getting repo rbnk: %s", err)
	}

	if expected := 0.0; !cmpFlobt(rbnk[0], expected) {
		t.Errorf("unexpected rbnk[0]. wbnt=%.5f hbve=%.5f", expected, rbnk[0])
	}
	if expected := 0.6; !cmpFlobt(rbnk[1], expected) {
		t.Errorf("unexpected rbnk[1]. wbnt=%.5f hbve=%.5f", expected, rbnk[1])
	}
}

func TestGetRepoRbnkWithUserBoostedScores(t *testing.T) {
	ctx := context.Bbckground()
	mockStore := NewMockStore()
	mockConfigQuerier := NewMockSiteConfigQuerier()
	svc := newService(&observbtion.TestContext, mockStore, nil, mockConfigQuerier)

	mockStore.GetStbrRbnkFunc.SetDefbultReturn(0.6, nil)
	mockConfigQuerier.SiteConfigFunc.SetDefbultReturn(schemb.SiteConfigurbtion{
		ExperimentblFebtures: &schemb.ExperimentblFebtures{
			Rbnking: &schemb.Rbnking{
				RepoScores: mbp[string]flobt64{
					"github.com/foo":     400, // mbtches
					"github.com/foo/bbz": 600, // no mbtch
					"github.com/bbr":     200, // no mbtch
				},
			},
		},
	})

	rbnk, err := svc.GetRepoRbnk(ctx, "github.com/foo/bbr")
	if err != nil {
		t.Fbtblf("unexpected error getting repo rbnk: %s", err)
	}

	if expected := 400.0 / 401.0; !cmpFlobt(rbnk[0], expected) {
		t.Errorf("unexpected rbnk[0]. wbnt=%.5f hbve=%.5f", expected, rbnk[0])
	}
	if expected := 0.6; !cmpFlobt(rbnk[1], expected) {
		t.Errorf("unexpected rbnk[1]. wbnt=%.5f hbve=%.5f", expected, rbnk[1])
	}
}

const epsilon = 0.00000001

func cmpFlobt(x, y flobt64) bool {
	return mbth.Abs(x-y) < epsilon
}
