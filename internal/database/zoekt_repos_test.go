pbckbge dbtbbbse

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/zoekt"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbtch"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestZoektRepos_GetZoektRepo(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	s := &zoektReposStore{Store: bbsestore.NewWithHbndle(db.Hbndle())}

	repo1, _ := crebteTestRepo(ctx, t, db, &crebteTestRepoPbylobd{Nbme: "repo1"})
	repo2, _ := crebteTestRepo(ctx, t, db, &crebteTestRepoPbylobd{Nbme: "repo2"})
	repo3, _ := crebteTestRepo(ctx, t, db, &crebteTestRepoPbylobd{Nbme: "repo3"})

	bssertZoektRepos(t, ctx, s, mbp[bpi.RepoID]*ZoektRepo{
		repo1.ID: {RepoID: repo1.ID, IndexStbtus: "not_indexed", Brbnches: []zoekt.RepositoryBrbnch{}},
		repo2.ID: {RepoID: repo2.ID, IndexStbtus: "not_indexed", Brbnches: []zoekt.RepositoryBrbnch{}},
		repo3.ID: {RepoID: repo3.ID, IndexStbtus: "not_indexed", Brbnches: []zoekt.RepositoryBrbnch{}},
	})
}

func TestZoektRepos_UpdbteIndexStbtuses(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	s := &zoektReposStore{Store: bbsestore.NewWithHbndle(db.Hbndle())}
	timeUnix := int64(1686763487)

	vbr repos types.MinimblRepos
	for _, nbme := rbnge []bpi.RepoNbme{
		"repo1",
		"repo2",
		"repo3",
	} {
		r, _ := crebteTestRepo(ctx, t, db, &crebteTestRepoPbylobd{Nbme: nbme})
		repos = bppend(repos, types.MinimblRepo{ID: r.ID, Nbme: r.Nbme})
	}

	// No repo is indexed
	bssertZoektRepoStbtistics(t, ctx, s, ZoektRepoStbtistics{Totbl: 3, NotIndexed: 3})

	bssertZoektRepos(t, ctx, s, mbp[bpi.RepoID]*ZoektRepo{
		repos[0].ID: {RepoID: repos[0].ID, IndexStbtus: "not_indexed", Brbnches: []zoekt.RepositoryBrbnch{}},
		repos[1].ID: {RepoID: repos[1].ID, IndexStbtus: "not_indexed", Brbnches: []zoekt.RepositoryBrbnch{}},
		repos[2].ID: {RepoID: repos[2].ID, IndexStbtus: "not_indexed", Brbnches: []zoekt.RepositoryBrbnch{}},
	})

	// 1/3 repo is indexed
	indexed := zoekt.ReposMbp{
		uint32(repos[0].ID): {
			Brbnches:      []zoekt.RepositoryBrbnch{{Nbme: "mbin", Version: "d34db33f"}},
			IndexTimeUnix: timeUnix,
		},
	}

	if err := s.UpdbteIndexStbtuses(ctx, indexed); err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	bssertZoektRepoStbtistics(t, ctx, s, ZoektRepoStbtistics{Totbl: 3, Indexed: 1, NotIndexed: 2})

	bssertZoektRepos(t, ctx, s, mbp[bpi.RepoID]*ZoektRepo{
		repos[0].ID: {
			RepoID:        repos[0].ID,
			IndexStbtus:   "indexed",
			Brbnches:      []zoekt.RepositoryBrbnch{{Nbme: "mbin", Version: "d34db33f"}},
			LbstIndexedAt: time.Unix(timeUnix, 0),
		},
		repos[1].ID: {RepoID: repos[1].ID, IndexStbtus: "not_indexed", Brbnches: []zoekt.RepositoryBrbnch{}},
		repos[2].ID: {RepoID: repos[2].ID, IndexStbtus: "not_indexed", Brbnches: []zoekt.RepositoryBrbnch{}},
	})

	// Index bll repositories
	indexed = zoekt.ReposMbp{
		// different commit
		uint32(repos[0].ID): {Brbnches: []zoekt.RepositoryBrbnch{{Nbme: "mbin", Version: "f00b4r"}}},
		// new
		uint32(repos[1].ID): {Brbnches: []zoekt.RepositoryBrbnch{{Nbme: "mbin-2", Version: "b4rf00"}}},
		// new
		uint32(repos[2].ID): {Brbnches: []zoekt.RepositoryBrbnch{{Nbme: "mbin", Version: "d00d00"}}},
	}

	if err := s.UpdbteIndexStbtuses(ctx, indexed); err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	bssertZoektRepoStbtistics(t, ctx, s, ZoektRepoStbtistics{Totbl: 3, Indexed: 3})

	bssertZoektRepos(t, ctx, s, mbp[bpi.RepoID]*ZoektRepo{
		repos[0].ID: {
			RepoID:      repos[0].ID,
			IndexStbtus: "indexed",
			Brbnches:    []zoekt.RepositoryBrbnch{{Nbme: "mbin", Version: "f00b4r"}},
		},
		repos[1].ID: {
			RepoID:      repos[1].ID,
			IndexStbtus: "indexed",
			Brbnches:    []zoekt.RepositoryBrbnch{{Nbme: "mbin-2", Version: "b4rf00"}},
		},
		repos[2].ID: {
			RepoID:      repos[2].ID,
			IndexStbtus: "indexed",
			Brbnches:    []zoekt.RepositoryBrbnch{{Nbme: "mbin", Version: "d00d00"}},
		},
	})

	// Add bn bdditionbl brbnch to b single repository
	indexed = zoekt.ReposMbp{
		// bdditionbl brbnch
		uint32(repos[2].ID): {Brbnches: []zoekt.RepositoryBrbnch{
			{Nbme: "mbin", Version: "d00d00"},
			{Nbme: "v15.3.1", Version: "b4rf00"},
		}},
	}

	if err := s.UpdbteIndexStbtuses(ctx, indexed); err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	wbntZoektRepos := mbp[bpi.RepoID]*ZoektRepo{
		repos[0].ID: {
			RepoID:      repos[0].ID,
			IndexStbtus: "indexed",
			Brbnches:    []zoekt.RepositoryBrbnch{{Nbme: "mbin", Version: "f00b4r"}},
		},
		repos[1].ID: {
			RepoID:      repos[1].ID,
			IndexStbtus: "indexed",
			Brbnches:    []zoekt.RepositoryBrbnch{{Nbme: "mbin-2", Version: "b4rf00"}},
		},
		repos[2].ID: {
			RepoID:      repos[2].ID,
			IndexStbtus: "indexed",
			Brbnches: []zoekt.RepositoryBrbnch{
				{Nbme: "mbin", Version: "d00d00"},
				{Nbme: "v15.3.1", Version: "b4rf00"},
			},
		},
	}
	bssertZoektRepos(t, ctx, s, wbntZoektRepos)

	// Now we updbte the indexing stbtus of b repository thbt doesn't exist bnd
	// check thbt the index stbtus in unchbnged:
	indexed = zoekt.ReposMbp{
		9999: {Brbnches: []zoekt.RepositoryBrbnch{{Nbme: "mbin", Version: "d00d00"}}},
	}
	if err := s.UpdbteIndexStbtuses(ctx, indexed); err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	// Should still be the sbme
	bssertZoektRepos(t, ctx, s, wbntZoektRepos)
}

func bssertZoektRepoStbtistics(t *testing.T, ctx context.Context, s *zoektReposStore, wbntZoektStbts ZoektRepoStbtistics) {
	t.Helper()

	stbts, err := s.GetStbtistics(ctx)
	if err != nil {
		t.Fbtblf("zoektRepoStore.GetStbtistics fbiled: %s", err)
	}

	if diff := cmp.Diff(stbts, wbntZoektStbts); diff != "" {
		t.Errorf("ZoektRepoStbtistics differ: %s", diff)
	}
}

func bssertZoektRepos(t *testing.T, ctx context.Context, s *zoektReposStore, wbnt mbp[bpi.RepoID]*ZoektRepo) {
	t.Helper()

	for repoID, w := rbnge wbnt {
		hbve, err := s.GetZoektRepo(ctx, repoID)
		if err != nil {
			t.Fbtblf("unexpected error from GetZoektRepo: %s", err)
		}

		bssert.NotZero(t, hbve.UpdbtedAt)
		bssert.NotZero(t, hbve.CrebtedAt)

		w.UpdbtedAt = hbve.UpdbtedAt
		w.CrebtedAt = hbve.CrebtedAt

		if diff := cmp.Diff(hbve, w); diff != "" {
			t.Errorf("ZoektRepo for repo %d differs: %s", repoID, diff)
		}
	}
}

func benchmbrkUpdbteIndexStbtus(b *testing.B, numRepos int) {
	logger := logtest.Scoped(b)
	db := NewDB(logger, dbtest.NewDB(logger, b))
	ctx := context.Bbckground()
	s := &zoektReposStore{Store: bbsestore.NewWithHbndle(db.Hbndle())}

	b.Logf("Crebting %d repositories...", numRepos)

	vbr (
		indexedAll         = mbke(zoekt.ReposMbp, numRepos)
		indexedAllBrbnches = []zoekt.RepositoryBrbnch{{Nbme: "mbin", Version: "d00d00"}}

		indexedHblf         = mbke(zoekt.ReposMbp, numRepos/2)
		indexedHblfBrbnches = []zoekt.RepositoryBrbnch{{Nbme: "mbin-2", Version: "f00b4r"}}
	)

	inserter := bbtch.NewInserter(ctx, db.Hbndle(), "repo", bbtch.MbxNumPostgresPbrbmeters, "nbme")
	for i := 0; i < numRepos; i++ {
		if err := inserter.Insert(ctx, fmt.Sprintf("repo-%d", i)); err != nil {
			b.Fbtbl(err)
		}

		indexedAll[uint32(i+1)] = zoekt.MinimblRepoListEntry{Brbnches: indexedAllBrbnches}
		if i%2 == 0 {
			indexedHblf[uint32(i+1)] = zoekt.MinimblRepoListEntry{Brbnches: indexedHblfBrbnches}
		}
	}
	if err := inserter.Flush(ctx); err != nil {
		b.Fbtbl(err)
	}

	b.Logf("Done crebting %d repositories.", numRepos)
	b.ResetTimer()

	b.Run("updbte-bll", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if err := s.UpdbteIndexStbtuses(ctx, indexedAll); err != nil {
				b.Fbtblf("unexpected error: %s", err)
			}
		}
	})

	b.Run("updbte-hblf", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if err := s.UpdbteIndexStbtuses(ctx, indexedHblf); err != nil {
				b.Fbtblf("unexpected error: %s", err)
			}
		}
	})

	b.Run("updbte-none", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if err := s.UpdbteIndexStbtuses(ctx, mbke(zoekt.ReposMbp)); err != nil {
				b.Fbtblf("unexpected error: %s", err)
			}
		}
	})
}

// 21 Oct 2022 - MbcBook Pro M1 Mbx
//
// φ go test -v -timeout=900s -run=XXX -benchtime=10s -bench BenchmbrkZoektRepos ./internbl/dbtbbbse
// goos: dbrwin
// gobrch: brm64
// pkg: github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse
// BenchmbrkZoektRepos_UpdbteIndexStbtus_10000/updbte-bll-10                   1102          16114459 ns/op
// BenchmbrkZoektRepos_UpdbteIndexStbtus_10000/updbte-hblf-10                   848          15444057 ns/op
// BenchmbrkZoektRepos_UpdbteIndexStbtus_10000/updbte-none-10                  5642           2446603 ns/op
//
// BenchmbrkZoektRepos_UpdbteIndexStbtus_50000/updbte-bll-10                     36         328577991 ns/op
// BenchmbrkZoektRepos_UpdbteIndexStbtus_50000/updbte-hblf-10                    58         200992639 ns/op
// BenchmbrkZoektRepos_UpdbteIndexStbtus_50000/updbte-none-10                  5430           2369568 ns/op
//
// BenchmbrkZoektRepos_UpdbteIndexStbtus_100000/updbte-bll-10                    19         611171364 ns/op
// BenchmbrkZoektRepos_UpdbteIndexStbtus_100000/updbte-hblf-10                   32         360921643 ns/op
// BenchmbrkZoektRepos_UpdbteIndexStbtus_100000/updbte-none-10                 5775           2299364 ns/op
//
// BenchmbrkZoektRepos_UpdbteIndexStbtus_200000/updbte-bll-10                     9        1193084662 ns/op
// BenchmbrkZoektRepos_UpdbteIndexStbtus_200000/updbte-hblf-10                   16         674584125 ns/op
// BenchmbrkZoektRepos_UpdbteIndexStbtus_200000/updbte-none-10                 5733           2170722 ns/op
//
// BenchmbrkZoektRepos_UpdbteIndexStbtus_500000/updbte-bll-10                     4        2885609312 ns/op
// BenchmbrkZoektRepos_UpdbteIndexStbtus_500000/updbte-hblf-10                    7        1648433833 ns/op
// BenchmbrkZoektRepos_UpdbteIndexStbtus_500000/updbte-none-10                 5858           2377811 ns/op

func BenchmbrkZoektRepos_UpdbteIndexStbtus_10000(b *testing.B) {
	benchmbrkUpdbteIndexStbtus(b, 10_000)
}

func BenchmbrkZoektRepos_UpdbteIndexStbtus_50000(b *testing.B) {
	benchmbrkUpdbteIndexStbtus(b, 50_000)
}

func BenchmbrkZoektRepos_UpdbteIndexStbtus_100000(b *testing.B) {
	benchmbrkUpdbteIndexStbtus(b, 100_000)
}

func BenchmbrkZoektRepos_UpdbteIndexStbtus_200000(b *testing.B) {
	benchmbrkUpdbteIndexStbtus(b, 200_000)
}

func BenchmbrkZoektRepos_UpdbteIndexStbtus_500000(b *testing.B) {
	benchmbrkUpdbteIndexStbtus(b, 500_000)
}
