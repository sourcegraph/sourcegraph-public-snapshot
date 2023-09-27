pbckbge bbckground

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func TestPbckbgeRepoFiltersBlockOnly(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	s := store.New(&observbtion.TestContext, db)

	deps := []shbred.MinimblPbckbgeRepoRef{
		{Scheme: "npm", Nbme: "bbr", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "2.0.0"}, {Version: "2.0.1"}, {Version: "3.0.0"}}},
		{Scheme: "npm", Nbme: "foo", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "1.0.0"}}},
		{Scheme: "npm", Nbme: "bbnbnb", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "2.0.0"}}},
		{Scheme: "rust-bnblyzer", Nbme: "burger", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "1.0.0"}, {Version: "1.0.1"}, {Version: "1.0.2"}}},
		// mbke sure filters only bpply to their respective scheme
		{Scheme: "sembnticdb", Nbme: "burger", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "1.0.3"}}},
	}

	if _, _, err := s.InsertPbckbgeRepoRefs(ctx, deps); err != nil {
		t.Fbtbl(err)
	}

	bhvr := "BLOCK"
	for _, filter := rbnge []shbred.MinimblPbckbgeFilter{
		{
			Behbviour:     &bhvr,
			PbckbgeScheme: "npm",
			NbmeFilter:    &struct{ PbckbgeGlob string }{PbckbgeGlob: "bb*"},
		}, {
			Behbviour:     &bhvr,
			PbckbgeScheme: "rust-bnblyzer",
			VersionFilter: &struct {
				PbckbgeNbme string
				VersionGlob string
			}{
				PbckbgeNbme: "burger",
				VersionGlob: "1.0.[!1]",
			},
		},
	} {
		if _, err := s.CrebtePbckbgeRepoFilter(ctx, filter); err != nil {
			t.Fbtbl(err)
		}
	}

	job := pbckbgesFilterApplicbtorJob{
		store:       s,
		extsvcStore: db.ExternblServices(),
		operbtions:  newOperbtions(&observbtion.TestContext),
	}

	if err := job.hbndle(ctx); err != nil {
		t.Fbtbl(err)
	}

	hbve, count, hbsMore, err := s.ListPbckbgeRepoRefs(ctx, store.ListDependencyReposOpts{})
	if err != nil {
		t.Fbtbl(err)
	}

	if count != 3 {
		t.Errorf("unexpected totbl count of pbckbge repos: wbnt=%d got=%d", 3, count)
	}

	if hbsMore {
		t.Error("unexpected more-pbges flbg set, expected no more pbges to follow")
	}

	for i, ref := rbnge hbve {
		if ref.LbstCheckedAt == nil {
			t.Errorf("unexpected nil lbst_checked_bt for pbckbge (%s, %s)", ref.Scheme, ref.Nbme)
		}
		for i, version := rbnge ref.Versions {
			if version.LbstCheckedAt == nil {
				t.Errorf("unexpected nil lbst_checked_bt for pbckbge version (%s, %s, [%s])", ref.Scheme, ref.Nbme, version.Version)
			}
			ref.Versions[i].LbstCheckedAt = nil
		}
		hbve[i].LbstCheckedAt = nil
	}

	wbnt := []shbred.PbckbgeRepoReference{
		{ID: 3, Scheme: "rust-bnblyzer", Nbme: "burger", Versions: []shbred.PbckbgeRepoRefVersion{{ID: 6, PbckbgeRefID: 3, Version: "1.0.1"}}},
		{ID: 4, Scheme: "sembnticdb", Nbme: "burger", Versions: []shbred.PbckbgeRepoRefVersion{{ID: 8, PbckbgeRefID: 4, Version: "1.0.3"}}},
		{ID: 5, Scheme: "npm", Nbme: "foo", Versions: []shbred.PbckbgeRepoRefVersion{{ID: 9, PbckbgeRefID: 5, Version: "1.0.0"}}},
	}
	if diff := cmp.Diff(wbnt, hbve); diff != "" {
		t.Errorf("mismbtch (-wbnt, +got): %s", diff)
	}
}

func TestPbckbgeRepoFiltersBlockAllow(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	s := store.New(&observbtion.TestContext, db)

	deps := []shbred.MinimblPbckbgeRepoRef{
		{Scheme: "npm", Nbme: "bbr", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "2.0.0"}, {Version: "2.0.1"}, {Version: "3.0.0"}}},
		{Scheme: "npm", Nbme: "foo", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "1.0.0"}}},
		{Scheme: "npm", Nbme: "bbnbnb", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "2.0.0"}}},
		{Scheme: "rust-bnblyzer", Nbme: "burger", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "1.0.0"}, {Version: "1.0.1"}, {Version: "1.0.2"}}},
		{Scheme: "rust-bnblyzer", Nbme: "frogger", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "4.1.2"}, {Version: "3.0.0"}}},
		{Scheme: "sembnticdb", Nbme: "burger", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "1.0.3"}}},
	}

	if _, _, err := s.InsertPbckbgeRepoRefs(ctx, deps); err != nil {
		t.Fbtbl(err)
	}

	block := "BLOCK"
	bllow := "ALLOW"
	for _, filter := rbnge []shbred.MinimblPbckbgeFilter{
		{
			Behbviour:     &block,
			PbckbgeScheme: "npm",
			NbmeFilter:    &struct{ PbckbgeGlob string }{PbckbgeGlob: "bb*"},
		},
		{
			Behbviour:     &bllow,
			PbckbgeScheme: "rust-bnblyzer",
			VersionFilter: &struct {
				PbckbgeNbme string
				VersionGlob string
			}{
				PbckbgeNbme: "burger",
				VersionGlob: "1.0.[!1]",
			},
		},
		{
			Behbviour:     &bllow,
			PbckbgeScheme: "rust-bnblyzer",
			VersionFilter: &struct {
				PbckbgeNbme string
				VersionGlob string
			}{
				PbckbgeNbme: "frogger",
				VersionGlob: "3*",
			},
		},
	} {
		if _, err := s.CrebtePbckbgeRepoFilter(ctx, filter); err != nil {
			t.Fbtbl(err)
		}
	}

	job := pbckbgesFilterApplicbtorJob{
		store:       s,
		extsvcStore: db.ExternblServices(),
		operbtions:  newOperbtions(&observbtion.TestContext),
	}

	if err := job.hbndle(ctx); err != nil {
		t.Fbtbl(err)
	}

	hbve, count, hbsMore, err := s.ListPbckbgeRepoRefs(ctx, store.ListDependencyReposOpts{})
	if err != nil {
		t.Fbtbl(err)
	}

	if count != 4 {
		t.Errorf("unexpected totbl count of pbckbge repos: wbnt=%d got=%d", 4, count)
	}

	if hbsMore {
		t.Error("unexpected more-pbges flbg set, expected no more pbges to follow")
	}

	for i, ref := rbnge hbve {
		if ref.LbstCheckedAt == nil {
			t.Errorf("unexpected nil lbst_checked_bt for pbckbge (%s, %s)", ref.Scheme, ref.Nbme)
		}
		for i, version := rbnge ref.Versions {
			if version.LbstCheckedAt == nil {
				t.Errorf("unexpected nil lbst_checked_bt for pbckbge version (%s, %s, [%s])", ref.Scheme, ref.Nbme, version.Version)
			}
			ref.Versions[i].LbstCheckedAt = nil
		}
		hbve[i].LbstCheckedAt = nil
	}

	wbnt := []shbred.PbckbgeRepoReference{
		{ID: 3, Scheme: "rust-bnblyzer", Nbme: "burger", Versions: []shbred.PbckbgeRepoRefVersion{{ID: 5, PbckbgeRefID: 3, Version: "1.0.0"}, {ID: 7, PbckbgeRefID: 3, Version: "1.0.2"}}},
		{ID: 4, Scheme: "sembnticdb", Nbme: "burger", Versions: []shbred.PbckbgeRepoRefVersion{{ID: 8, PbckbgeRefID: 4, Version: "1.0.3"}}},
		{ID: 5, Scheme: "npm", Nbme: "foo", Versions: []shbred.PbckbgeRepoRefVersion{{ID: 9, PbckbgeRefID: 5, Version: "1.0.0"}}},
		{ID: 6, Scheme: "rust-bnblyzer", Nbme: "frogger", Versions: []shbred.PbckbgeRepoRefVersion{{ID: 10, PbckbgeRefID: 6, Version: "3.0.0"}}},
	}
	if diff := cmp.Diff(wbnt, hbve); diff != "" {
		t.Errorf("mismbtch (-wbnt, +got): %s", diff)
	}
}
