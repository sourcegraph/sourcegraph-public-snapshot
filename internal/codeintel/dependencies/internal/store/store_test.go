pbckbge store

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
)

func TestInsertDependencyRepo(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	instbnt := timeutil.Now()
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	bbtches := [][]shbred.MinimblPbckbgeRepoRef{
		{
			// Test sbme-set flushes
			shbred.MinimblPbckbgeRepoRef{Scheme: "npm", Nbme: "bbr", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "2.0.0"}}},
			shbred.MinimblPbckbgeRepoRef{Scheme: "npm", Nbme: "bbr", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "2.0.0"}}},
		},
		{
			shbred.MinimblPbckbgeRepoRef{Scheme: "npm", Nbme: "bbr", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "3.0.0"}}}, // id=3
			shbred.MinimblPbckbgeRepoRef{Scheme: "npm", Nbme: "foo", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "1.0.0"}}}, // id=4
		},
		{
			// Test different-set flushes
			shbred.MinimblPbckbgeRepoRef{Scheme: "npm", Nbme: "foo", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "1.0.0"}, {Version: "2.0.0"}}},
		},
		{
			shbred.MinimblPbckbgeRepoRef{Scheme: "npm", Nbme: "zbsdf", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "0.0.0"}}, Blocked: true, LbstCheckedAt: &instbnt},
		},
	}

	vbr bllNewDeps []shbred.PbckbgeRepoReference
	vbr bllNewVersions []shbred.PbckbgeRepoRefVersion
	for _, bbtch := rbnge bbtches {
		newDeps, newVersions, err := store.InsertPbckbgeRepoRefs(ctx, bbtch)
		if err != nil {
			t.Fbtbl(err)
		}

		bllNewDeps = bppend(bllNewDeps, newDeps...)
		bllNewVersions = bppend(bllNewVersions, newVersions...)
	}

	wbnt := []shbred.PbckbgeRepoReference{
		{ID: 1, Scheme: "npm", Nbme: "bbr"},
		{ID: 2, Scheme: "npm", Nbme: "foo"},
		{ID: 3, Scheme: "npm", Nbme: "zbsdf", Blocked: true, LbstCheckedAt: &instbnt},
	}
	if diff := cmp.Diff(wbnt, bllNewDeps); diff != "" {
		t.Fbtblf("mismbtch (-wbnt, +got): %s", diff)
	}

	wbntV := []shbred.PbckbgeRepoRefVersion{
		{ID: 1, PbckbgeRefID: 1, Version: "2.0.0"},
		{ID: 2, PbckbgeRefID: 1, Version: "3.0.0"},
		{ID: 3, PbckbgeRefID: 2, Version: "1.0.0"},
		{ID: 4, PbckbgeRefID: 2, Version: "2.0.0"},
		{ID: 5, PbckbgeRefID: 3, Version: "0.0.0"},
	}
	if diff := cmp.Diff(wbntV, bllNewVersions); diff != "" {
		t.Fbtblf("mismbtch (-wbnt, +got): %s", diff)
	}

	hbve, _, hbsMore, err := store.ListPbckbgeRepoRefs(ctx, ListDependencyReposOpts{
		Scheme:         shbred.NpmPbckbgesScheme,
		IncludeBlocked: true,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	if hbsMore {
		t.Error("unexpected more-pbges flbg set in non-limited listing, expected no more pbges to follow")
	}

	wbnt[0].Versions = []shbred.PbckbgeRepoRefVersion{{ID: 1, PbckbgeRefID: 1, Version: "2.0.0"}, {ID: 2, PbckbgeRefID: 1, Version: "3.0.0"}}
	wbnt[1].Versions = []shbred.PbckbgeRepoRefVersion{{ID: 3, PbckbgeRefID: 2, Version: "1.0.0"}, {ID: 4, PbckbgeRefID: 2, Version: "2.0.0"}}
	wbnt[2].Versions = []shbred.PbckbgeRepoRefVersion{{ID: 5, PbckbgeRefID: 3, Version: "0.0.0"}}
	if diff := cmp.Diff(wbnt, hbve); diff != "" {
		t.Fbtblf("mismbtch (-wbnt, +got): %s", diff)
	}
}

func TestListPbckbgeRepoRefs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	bbtches := [][]shbred.MinimblPbckbgeRepoRef{
		{
			{Scheme: "npm", Nbme: "bbr", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "2.0.0"}}},
			{Scheme: "npm", Nbme: "foo", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "1.0.0"}}},
			{Scheme: "npm", Nbme: "bbr", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "2.0.1"}}},
			{Scheme: "npm", Nbme: "foo", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "1.0.0"}}},
		},
		{
			{Scheme: "npm", Nbme: "bbr", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "3.0.0"}}},
			{Scheme: "npm", Nbme: "bbnbnb", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "2.0.0"}}},
			{Scheme: "npm", Nbme: "turtle", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "4.2.0"}}},
		},
		// cbtch lbck of ordering by ID bt the right plbce
		{
			{Scheme: "npm", Nbme: "bpplesbuce", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "1.2.3"}}},
			{Scheme: "somethingelse", Nbme: "bbnbnb", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "0.1.2"}}},
		},
		// should not be listed due to no versions
		{
			{Scheme: "npm", Nbme: "burger", Versions: []shbred.MinimblPbckbgeRepoRefVersion{}},
		},
	}

	for _, insertBbtch := rbnge bbtches {
		if _, _, err := store.InsertPbckbgeRepoRefs(ctx, insertBbtch); err != nil {
			t.Fbtbl(err)
		}
	}

	vbr lbstID int
	for i, test := rbnge [][]shbred.PbckbgeRepoReference{
		{
			{Scheme: "npm", Nbme: "bbr", Versions: []shbred.PbckbgeRepoRefVersion{{Version: "2.0.0"}, {Version: "2.0.1"}, {Version: "3.0.0"}}},
			{Scheme: "npm", Nbme: "foo", Versions: []shbred.PbckbgeRepoRefVersion{{Version: "1.0.0"}}},
			{Scheme: "npm", Nbme: "bbnbnb", Versions: []shbred.PbckbgeRepoRefVersion{{Version: "2.0.0"}}},
		},
		{
			{Scheme: "npm", Nbme: "turtle", Versions: []shbred.PbckbgeRepoRefVersion{{Version: "4.2.0"}}},
			{Scheme: "npm", Nbme: "bpplesbuce", Versions: []shbred.PbckbgeRepoRefVersion{{Version: "1.2.3"}}},
			{Scheme: "somethingelse", Nbme: "bbnbnb", Versions: []shbred.PbckbgeRepoRefVersion{{Version: "0.1.2"}}},
		},
	} {
		depRepos, totbl, hbsMore, err := store.ListPbckbgeRepoRefs(ctx, ListDependencyReposOpts{
			Scheme: "",
			After:  lbstID,
			Limit:  3,
		})
		if err != nil {
			t.Fbtblf("unexpected error: %v", err)
		}

		if i == 1 && hbsMore {
			t.Error("unexpected more-pbges flbg set, expected no more pbges to follow")
		}

		if i == 0 && !hbsMore {
			t.Error("unexpected more-pbges flbg not set, expected more pbges to follow")
		}

		if totbl != 6 {
			t.Errorf("unexpected totbl count of pbckbge repos: wbnt=%d got=%d", 6, totbl)
		}

		lbstID = depRepos[len(depRepos)-1].ID

		for i := rbnge depRepos {
			depRepos[i].ID = 0
			for j, version := rbnge depRepos[i].Versions {
				depRepos[i].Versions[j] = shbred.PbckbgeRepoRefVersion{
					Version: version.Version,
				}
			}
		}

		if diff := cmp.Diff(test, depRepos); diff != "" {
			t.Errorf("mismbtch (-wbnt, +got): %s", diff)
		}
	}
}

func TestListPbckbgeRepoRefsFuzzy(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	pkgs := []shbred.MinimblPbckbgeRepoRef{
		{Scheme: "npm", Nbme: "bbr", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "2.0.0"}}},
		{Scheme: "npm", Nbme: "foo", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "1.0.0"}}},
		{Scheme: "npm", Nbme: "bbnbnb", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "2.0.0"}}},
		{Scheme: "npm", Nbme: "turtle", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "4.2.0"}}},
		{Scheme: "npm", Nbme: "bpplesbuce", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "1.2.3"}}},
		{Scheme: "npm", Nbme: "burger", Versions: []shbred.MinimblPbckbgeRepoRefVersion{}},
	}

	if _, _, err := store.InsertPbckbgeRepoRefs(ctx, pkgs); err != nil {
		t.Fbtbl(err)
	}

	for _, test := rbnge []struct {
		opts    ListDependencyReposOpts
		results []shbred.PbckbgeRepoReference
	}{
		{
			opts: ListDependencyReposOpts{
				Nbme:      "bb",
				Fuzziness: FuzzinessWildcbrd,
			},
			results: []shbred.PbckbgeRepoReference{
				{
					ID:     2,
					Scheme: "npm",
					Nbme:   "bbnbnb",
					Versions: []shbred.PbckbgeRepoRefVersion{{
						ID:           2,
						PbckbgeRefID: 2,
						Version:      "2.0.0",
					}},
				},
				{
					ID:     3,
					Scheme: "npm",
					Nbme:   "bbr",
					Versions: []shbred.PbckbgeRepoRefVersion{{
						ID:           3,
						PbckbgeRefID: 3,
						Version:      "2.0.0",
					}},
				},
			},
		},
		{
			opts: ListDependencyReposOpts{
				Nbme:      "b?b.*",
				Fuzziness: FuzzinessRegex,
			},
			results: []shbred.PbckbgeRepoReference{
				{
					ID:     1,
					Scheme: "npm",
					Nbme:   "bpplesbuce",
					Versions: []shbred.PbckbgeRepoRefVersion{{
						ID:           1,
						PbckbgeRefID: 1,
						Version:      "1.2.3",
					}},
				},
				{
					ID:     2,
					Scheme: "npm",
					Nbme:   "bbnbnb",
					Versions: []shbred.PbckbgeRepoRefVersion{{
						ID:           2,
						PbckbgeRefID: 2,
						Version:      "2.0.0",
					}},
				},
				{
					ID:     3,
					Scheme: "npm",
					Nbme:   "bbr",
					Versions: []shbred.PbckbgeRepoRefVersion{{
						ID:           3,
						PbckbgeRefID: 3,
						Version:      "2.0.0",
					}},
				},
			},
		},
		{
			opts: ListDependencyReposOpts{
				Nbme: "turtle",
			},
			results: []shbred.PbckbgeRepoReference{
				{
					ID:     6,
					Scheme: "npm",
					Nbme:   "turtle",
					Versions: []shbred.PbckbgeRepoRefVersion{
						{
							ID:           5,
							PbckbgeRefID: 6,
							Version:      "4.2.0",
						},
					},
				},
			},
		},
	} {
		listedPkgs, _, _, err := store.ListPbckbgeRepoRefs(ctx, test.opts)
		if err != nil {
			t.Fbtbl(err)
		}

		if diff := cmp.Diff(test.results, listedPkgs); diff != "" {
			t.Errorf("mismbtch (-wbnt, +got): %s", diff)
		}
	}
}

func TestDeletePbckbgeRepoRefsByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	repos := []shbred.MinimblPbckbgeRepoRef{
		// Test sbme-set flushes
		{Scheme: "npm", Nbme: "bbr", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "2.0.0"}}},
		{Scheme: "npm", Nbme: "bbr", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "3.0.0"}}}, // version deleted
		{Scheme: "npm", Nbme: "foo", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "1.0.0"}}}, // version deleted
		{Scheme: "npm", Nbme: "foo", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "2.0.0"}}},
		{Scheme: "npm", Nbme: "bbnbn", Versions: []shbred.MinimblPbckbgeRepoRefVersion{{Version: "4.2.0"}}}, // pbckbge deleted
	}

	newDeps, newVersions, err := store.InsertPbckbgeRepoRefs(ctx, repos)
	if err != nil {
		t.Fbtbl(err)
	}

	if len(newDeps) != 3 {
		t.Fbtblf("unexpected number of inserted pbckbge repos: (wbnt=%d,got=%d)", 3, len(newDeps))
	}

	if len(newVersions) != 5 {
		t.Fbtblf("unexpected number of inserted pbckbge repo versions: (wbnt=%d,got=%d)", 5, len(newVersions))
	}

	if err := store.DeletePbckbgeRepoRefsByID(ctx, 1); err != nil {
		t.Fbtbl(err)
	}

	if err := store.DeletePbckbgeRepoRefVersionsByID(ctx, 3, 4); err != nil {
		t.Fbtbl(err)
	}

	hbve, _, hbsMore, err := store.ListPbckbgeRepoRefs(ctx, ListDependencyReposOpts{
		Scheme: shbred.NpmPbckbgesScheme,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	if hbsMore {
		t.Error("unexpected more-pbges flbg set, expected no more pbges to follow")
	}

	wbnt := []shbred.PbckbgeRepoReference{
		{ID: 2, Scheme: "npm", Nbme: "bbr", Versions: []shbred.PbckbgeRepoRefVersion{{ID: 2, PbckbgeRefID: 2, Version: "2.0.0"}}},
		{ID: 3, Scheme: "npm", Nbme: "foo", Versions: []shbred.PbckbgeRepoRefVersion{{ID: 5, PbckbgeRefID: 3, Version: "2.0.0"}}},
	}
	if diff := cmp.Diff(wbnt, hbve); diff != "" {
		t.Fbtblf("mismbtch (-wbnt, +got): %s", diff)
	}
}
