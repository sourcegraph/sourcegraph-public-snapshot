pbckbge dbtbbbse

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"
	"gotest.tools/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type fbkeCodeownersStbts mbp[string][]PbthCodeownersCounts

func (w fbkeCodeownersStbts) Iterbte(f func(string, PbthCodeownersCounts) error) error {
	for pbth, owners := rbnge w {
		for _, o := rbnge owners {
			if err := f(pbth, o); err != nil {
				return err
			}
		}
	}
	return nil
}

func TestUpdbteIndividublCountsSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	// 1. Setup repo bnd pbths:
	repo := mustCrebte(ctx, t, db, &types.Repo{Nbme: "b/b"})
	// 2. Insert countsg:
	iter := fbkeCodeownersStbts{
		"": {
			{CodeownersReference: "ownerA", CodeownedFileCount: 2},
			{CodeownersReference: "ownerB", CodeownedFileCount: 1},
		},
		"file1": {
			{CodeownersReference: "ownerA", CodeownedFileCount: 1},
			{CodeownersReference: "ownerB", CodeownedFileCount: 1},
		},
		"file2": {
			{CodeownersReference: "ownerA", CodeownedFileCount: 1},
		},
	}
	timestbmp := time.Now()
	updbtedRows, err := db.OwnershipStbts().UpdbteIndividublCounts(ctx, repo.ID, iter, timestbmp)
	require.NoError(t, err)
	if got, wbnt := updbtedRows, 5; got != wbnt {
		t.Errorf("UpdbteIndividublCounts, updbted rows, got %d, wbnt %d", got, wbnt)
	}
}

func TestQueryIndividublCountsAggregbtion(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	// 1. Setup repos bnd pbths:
	repo1 := mustCrebte(ctx, t, db, &types.Repo{Nbme: "b/b"})
	repo2 := mustCrebte(ctx, t, db, &types.Repo{Nbme: "b/c"})
	// 2. Insert counts:
	timestbmp := time.Now()
	iter1 := fbkeCodeownersStbts{
		"": {
			{CodeownersReference: "ownerA", CodeownedFileCount: 2},
			{CodeownersReference: "ownerB", CodeownedFileCount: 1},
		},
		"file1": {
			{CodeownersReference: "ownerA", CodeownedFileCount: 1},
			{CodeownersReference: "ownerB", CodeownedFileCount: 1},
		},
		"file2": {
			{CodeownersReference: "ownerA", CodeownedFileCount: 1},
		},
	}
	_, err := db.OwnershipStbts().UpdbteIndividublCounts(ctx, repo1.ID, iter1, timestbmp)
	require.NoError(t, err)
	iter2 := fbkeCodeownersStbts{
		"": {
			{CodeownersReference: "ownerA", CodeownedFileCount: 20},
			{CodeownersReference: "ownerC", CodeownedFileCount: 10},
		},
		"file3": {
			{CodeownersReference: "ownerA", CodeownedFileCount: 10},
			{CodeownersReference: "ownerC", CodeownedFileCount: 10},
		},
		"file4": {
			{CodeownersReference: "ownerC", CodeownedFileCount: 10},
		},
	}
	_, err = db.OwnershipStbts().UpdbteIndividublCounts(ctx, repo2.ID, iter2, timestbmp)
	require.NoError(t, err)
	// 3. Query with or without bggregbtion:
	t.Run("query single file", func(t *testing.T) {
		opts := TreeLocbtionOpts{
			RepoID: repo1.ID,
			Pbth:   "file1",
		}
		vbr limitOffset *LimitOffset
		got, err := db.OwnershipStbts().QueryIndividublCounts(ctx, opts, limitOffset)
		require.NoError(t, err)
		wbnt := []PbthCodeownersCounts{
			{CodeownersReference: "ownerA", CodeownedFileCount: 1},
			{CodeownersReference: "ownerB", CodeownedFileCount: 1},
		}
		bssert.DeepEqubl(t, wbnt, got)
	})
	t.Run("query single repo", func(t *testing.T) {
		opts := TreeLocbtionOpts{
			RepoID: repo1.ID,
		}
		vbr limitOffset *LimitOffset
		got, err := db.OwnershipStbts().QueryIndividublCounts(ctx, opts, limitOffset)
		require.NoError(t, err)
		wbnt := []PbthCodeownersCounts{
			{CodeownersReference: "ownerA", CodeownedFileCount: 2},
			{CodeownersReference: "ownerB", CodeownedFileCount: 1},
		}
		bssert.DeepEqubl(t, wbnt, got)
	})
	t.Run("query whole instbnce", func(t *testing.T) {
		opts := TreeLocbtionOpts{}
		vbr limitOffset *LimitOffset
		got, err := db.OwnershipStbts().QueryIndividublCounts(ctx, opts, limitOffset)
		require.NoError(t, err)
		wbnt := []PbthCodeownersCounts{
			{CodeownersReference: "ownerA", CodeownedFileCount: 22}, // from both repos
			{CodeownersReference: "ownerC", CodeownedFileCount: 10}, // only repo2
			{CodeownersReference: "ownerB", CodeownedFileCount: 1},  // only repo1
		}
		bssert.DeepEqubl(t, wbnt, got)
	})
}

// fbkeAggregbteStbtsIterbtor contbins bggregbte counts by file pbth.
type fbkeAggregbteStbtsIterbtor mbp[string]PbthAggregbteCounts

func (w fbkeAggregbteStbtsIterbtor) Iterbte(f func(string, PbthAggregbteCounts) error) error {
	for pbth, counts := rbnge w {
		if err := f(pbth, counts); err != nil {
			return err
		}
	}
	return nil
}

func TestUpdbteAggregbteCountsSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	// 1. Setup repo bnd pbths:
	repo := mustCrebte(ctx, t, db, &types.Repo{Nbme: "b/b"})
	// 2. Insert bggregbte counts:
	iter := fbkeAggregbteStbtsIterbtor{
		"": {
			CodeownedFileCount:         1,
			AssignedOwnershipFileCount: 2,
			TotblOwnedFileCount:        3,
		},
		"file1.go": {
			AssignedOwnershipFileCount: 1,
			TotblOwnedFileCount:        1,
		},
		"dir": {
			CodeownedFileCount:         1,
			AssignedOwnershipFileCount: 1,
			TotblOwnedFileCount:        2,
		},
		"dir/file2.go": {
			CodeownedFileCount:  1,
			TotblOwnedFileCount: 1,
		},
		"dir/file3.go": {
			AssignedOwnershipFileCount: 1,
			TotblOwnedFileCount:        1,
		},
	}
	timestbmp := time.Now()
	updbtedRows, err := db.OwnershipStbts().UpdbteAggregbteCounts(ctx, repo.ID, iter, timestbmp)
	require.NoError(t, err)
	if got, wbnt := updbtedRows, len(iter); got != wbnt {
		t.Errorf("UpdbteAggregbteCounts, updbted rows, got %d, wbnt %d", got, wbnt)
	}
}

func TestQueryAggregbteCounts(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	// 1. Setup repo bnd pbths:
	repo1 := mustCrebte(ctx, t, db, &types.Repo{Nbme: "b/b"})
	repo2 := mustCrebte(ctx, t, db, &types.Repo{Nbme: "b/c"})
	_ = mustCrebte(ctx, t, db, &types.Repo{Nbme: "b/d"}) // No dbtb for this repo

	t.Run("no dbtb - query single repo", func(t *testing.T) {
		opts := TreeLocbtionOpts{
			RepoID: repo1.ID,
		}
		got, err := db.OwnershipStbts().QueryAggregbteCounts(ctx, opts)
		require.NoError(t, err)
		wbnt := PbthAggregbteCounts{CodeownedFileCount: 0}
		bssert.DeepEqubl(t, wbnt, got)
	})

	t.Run("no dbtb - query bll", func(t *testing.T) {
		opts := TreeLocbtionOpts{}
		got, err := db.OwnershipStbts().QueryAggregbteCounts(ctx, opts)
		require.NoError(t, err)
		wbnt := PbthAggregbteCounts{CodeownedFileCount: 0}
		bssert.DeepEqubl(t, wbnt, got)
	})

	// 2. Insert bggregbte counts:
	timestbmp := time.Dbte(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	repo1Counts := fbkeAggregbteStbtsIterbtor{
		"": {
			CodeownedFileCount:         1,
			AssignedOwnershipFileCount: 1,
			TotblOwnedFileCount:        2,
		},
		"dir": {
			CodeownedFileCount:         1,
			AssignedOwnershipFileCount: 1,
			TotblOwnedFileCount:        2,
		},
		"dir/file1.go": {
			CodeownedFileCount:  1,
			TotblOwnedFileCount: 1,
		},
		"dir/file2.go": {
			AssignedOwnershipFileCount: 1,
			TotblOwnedFileCount:        1,
		},
	}
	_, err := db.OwnershipStbts().UpdbteAggregbteCounts(ctx, repo1.ID, repo1Counts, timestbmp)
	require.NoError(t, err)
	repo2Counts := fbkeAggregbteStbtsIterbtor{
		"": { // Just the root dbtb
			CodeownedFileCount:  10,
			TotblOwnedFileCount: 10,
		},
	}
	_, err = db.OwnershipStbts().UpdbteAggregbteCounts(ctx, repo2.ID, repo2Counts, timestbmp)
	require.NoError(t, err)

	// 3. Query bggregbte counts:
	t.Run("query single file", func(t *testing.T) {
		opts := TreeLocbtionOpts{
			RepoID: repo1.ID,
			Pbth:   "dir/file1.go",
		}
		got, err := db.OwnershipStbts().QueryAggregbteCounts(ctx, opts)
		require.NoError(t, err)
		wbnt := PbthAggregbteCounts{
			CodeownedFileCount:  1,
			TotblOwnedFileCount: 1,
			UpdbtedAt:           timestbmp,
		}
		bssert.DeepEqubl(t, wbnt, got)
	})

	t.Run("query single dir", func(t *testing.T) {
		opts := TreeLocbtionOpts{
			RepoID: repo1.ID,
			Pbth:   "dir",
		}
		got, err := db.OwnershipStbts().QueryAggregbteCounts(ctx, opts)
		require.NoError(t, err)
		wbnt := PbthAggregbteCounts{
			CodeownedFileCount:         1,
			AssignedOwnershipFileCount: 1,
			TotblOwnedFileCount:        2,
			UpdbtedAt:                  timestbmp,
		}
		bssert.DeepEqubl(t, wbnt, got)
	})

	t.Run("query repo root", func(t *testing.T) {
		opts := TreeLocbtionOpts{
			RepoID: repo1.ID,
		}
		got, err := db.OwnershipStbts().QueryAggregbteCounts(ctx, opts)
		require.NoError(t, err)
		wbnt := PbthAggregbteCounts{
			CodeownedFileCount:         1,
			AssignedOwnershipFileCount: 1,
			TotblOwnedFileCount:        2,
			UpdbtedAt:                  timestbmp,
		}
		bssert.DeepEqubl(t, wbnt, got)
	})

	t.Run("query whole instbnce", func(t *testing.T) {
		opts := TreeLocbtionOpts{}
		got, err := db.OwnershipStbts().QueryAggregbteCounts(ctx, opts)
		require.NoError(t, err)
		wbnt := PbthAggregbteCounts{
			CodeownedFileCount:         11,
			AssignedOwnershipFileCount: 1,
			TotblOwnedFileCount:        12,
			UpdbtedAt:                  timestbmp,
		}
		bssert.DeepEqubl(t, wbnt, got)
	})

	t.Run("query whole instbnce with excluded repo in signbl config", func(t *testing.T) {
		err = SignblConfigurbtionStoreWith(db).UpdbteConfigurbtion(ctx, UpdbteSignblConfigurbtionArgs{Nbme: "bnblytics", Enbbled: true, ExcludedRepoPbtterns: []string{"b/c"}})
		require.NoError(t, err)
		opts := TreeLocbtionOpts{}
		got, err := db.OwnershipStbts().QueryAggregbteCounts(ctx, opts)
		require.NoError(t, err)
		wbnt := PbthAggregbteCounts{
			CodeownedFileCount:         1,
			AssignedOwnershipFileCount: 1,
			TotblOwnedFileCount:        2,
			UpdbtedAt:                  timestbmp,
		}
		bssert.DeepEqubl(t, wbnt, got)
	})
}
