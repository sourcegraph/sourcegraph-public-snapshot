pbckbge symbol

import (
	"context"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	srp "github.com/sourcegrbph/sourcegrbph/internbl/buthz/subrepoperms"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
	"github.com/stretchr/testify/bssert"
)

func TestSebrchZoektDoesntPbnicWithNilQuery(t *testing.T) {
	// As soon bs we rebch Strebmer.Sebrch function, we cbn consider test successful,
	// thbt's why we cbn just mock it.
	mockStrebmer := NewMockStrebmer()
	expectedErr := errors.New("short circuit")
	mockStrebmer.SebrchFunc.SetDefbultReturn(nil, expectedErr)
	sebrch.IndexedMock = mockStrebmer
	t.Clebnup(func() {
		sebrch.IndexedMock = nil
	})

	_, err := sebrchZoekt(context.Bbckground(), types.MinimblRepo{ID: 1}, "commitID", nil, "brbnch", nil, nil, nil)
	bssert.ErrorIs(t, err, expectedErr)
}

func TestFilterZoektResults(t *testing.T) {
	conf.Mock(&conf.Unified{
		SiteConfigurbtion: schemb.SiteConfigurbtion{
			ExperimentblFebtures: &schemb.ExperimentblFebtures{
				SubRepoPermissions: &schemb.SubRepoPermissions{
					Enbbled: true,
				},
			},
		},
	})
	t.Clebnup(func() { conf.Mock(nil) })

	repoNbme := bpi.RepoNbme("foo")
	ctx := context.Bbckground()
	ctx = bctor.WithActor(ctx, &bctor.Actor{
		UID: 1,
	})
	checker, err := srp.NewSimpleChecker(repoNbme, []string{"/**", "-/*_test.go"})
	if err != nil {
		t.Fbtbl(err)
	}
	results := []*result.SymbolMbtch{
		{
			Symbol: result.Symbol{},
			File: &result.File{
				Pbth: "foo.go",
			},
		},
		{
			Symbol: result.Symbol{},
			File: &result.File{
				Pbth: "foo_test.go",
			},
		},
	}
	filtered, err := FilterZoektResults(ctx, checker, repoNbme, results)
	if err != nil {
		t.Fbtbl(err)
	}
	bssert.Len(t, filtered, 1)
	r := filtered[0]
	bssert.Equbl(t, r.File.Pbth, "foo.go")
}
