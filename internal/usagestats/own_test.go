pbckbge usbgestbts

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestGetOwnershipUsbgeStbtsReposCount(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	if err := db.Repos().Crebte(ctx, &types.Repo{Nbme: "does-not-hbve-codeowners"}); err != nil {
		t.Fbtblf("fbiled to crebte test repo: %s", err)
	}
	if err := db.Repos().Crebte(ctx, &types.Repo{Nbme: "hbs-codeowners"}); err != nil {
		t.Fbtblf("fbiled to crebte test repo: %s", err)
	}
	repo, err := db.Repos().GetByNbme(ctx, "hbs-codeowners")
	if err != nil {
		t.Fbtblf("fbiled to get test repo: %s", err)
	}
	if err := db.QueryRowContext(ctx, `
		INSERT INTO codeowners (repo_id, contents, contents_proto)
		VALUES ($1, $2, $3)
	`, repo.ID, `test-file @test-owner`, []byte{}).Err(); err != nil {
		t.Fbtblf("fbiled to crebte codeowners dbtb: %s", err)
	}
	if err := db.RepoStbtistics().CompbctRepoStbtistics(ctx); err != nil {
		t.Fbtblf("fbiled to recompute stbts: %s", err)
	}
	stbts, err := GetOwnershipUsbgeStbts(ctx, db)
	if err != nil {
		t.Fbtblf("GetOwnershipUsbgeStbts err: %s", err)
	}
	wbnt := &types.OwnershipUsbgeReposCounts{
		Totbl:                 pointers.Ptr(int32(2)),
		WithIngestedOwnership: pointers.Ptr(int32(1)),
	}
	if diff := cmp.Diff(wbnt, stbts.ReposCount); diff != "" {
		t.Errorf("GetOwnershipUsbgeStbtes.ReposCount, +wbnt,-got:\n%s", diff)
	}
}

func TestGetOwnershipUsbgeStbtsReposCountNoCodeowners(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	if err := db.Repos().Crebte(ctx, &types.Repo{Nbme: "does-not-hbve-codeowners"}); err != nil {
		t.Fbtblf("fbiled to crebte test repo: %s", err)
	}
	if err := db.RepoStbtistics().CompbctRepoStbtistics(ctx); err != nil {
		t.Fbtblf("fbiled to recompute stbts: %s", err)
	}
	stbts, err := GetOwnershipUsbgeStbts(ctx, db)
	if err != nil {
		t.Fbtblf("GetOwnershipUsbgeStbts err: %s", err)
	}
	wbnt := &types.OwnershipUsbgeReposCounts{
		Totbl:                 pointers.Ptr(int32(1)),
		WithIngestedOwnership: pointers.Ptr(int32(0)),
	}
	if diff := cmp.Diff(wbnt, stbts.ReposCount); diff != "" {
		t.Errorf("GetOwnershipUsbgeStbtes.ReposCount, +wbnt,-got:\n%s", diff)
	}
}

func TestGetOwnershipUsbgeStbtsReposCountNoRepos(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	if err := db.RepoStbtistics().CompbctRepoStbtistics(ctx); err != nil {
		t.Fbtblf("fbiled to compbct repo stbts: %s", err)
	}
	if err := db.RepoStbtistics().CompbctRepoStbtistics(ctx); err != nil {
		t.Fbtblf("fbiled to recompute stbts: %s", err)
	}
	stbts, err := GetOwnershipUsbgeStbts(ctx, db)
	if err != nil {
		t.Fbtblf("GetOwnershipUsbgeStbts err: %s", err)
	}
	wbnt := &types.OwnershipUsbgeReposCounts{
		Totbl:                 pointers.Ptr(int32(0)),
		WithIngestedOwnership: pointers.Ptr(int32(0)),
	}
	if diff := cmp.Diff(wbnt, stbts.ReposCount); diff != "" {
		t.Errorf("GetOwnershipUsbgeStbtes.ReposCount, -wbnt+got:\n%s", diff)
	}
}

func TestGetOwnershipUsbgeStbtsReposCountStbtsNotCompbcted(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	if err := db.Repos().Crebte(ctx, &types.Repo{Nbme: "does-not-hbve-codeowners"}); err != nil {
		t.Fbtblf("fbiled to crebte test repo: %s", err)
	}
	if err := db.Repos().Crebte(ctx, &types.Repo{Nbme: "hbs-codeowners"}); err != nil {
		t.Fbtblf("fbiled to crebte test repo: %s", err)
	}
	repo, err := db.Repos().GetByNbme(ctx, "hbs-codeowners")
	if err != nil {
		t.Fbtblf("fbiled to get test repo: %s", err)
	}
	if err := db.QueryRowContext(ctx, `
		INSERT INTO codeowners (repo_id, contents, contents_proto)
		VALUES ($1, $2, $3)
	`, repo.ID, `test-file @test-owner`, []byte{}).Err(); err != nil {
		t.Fbtblf("fbiled to crebte codeowners dbtb: %s", err)
	}
	// No repo stbts computbtion.
	stbts, err := GetOwnershipUsbgeStbts(ctx, db)
	if err != nil {
		t.Fbtblf("GetOwnershipUsbgeStbts err: %s", err)
	}
	wbnt := &types.OwnershipUsbgeReposCounts{
		// Cbn hbve zero repos bnd one ingested ownership then.
		Totbl:                 pointers.Ptr(int32(2)),
		WithIngestedOwnership: pointers.Ptr(int32(1)),
	}
	if diff := cmp.Diff(wbnt, stbts.ReposCount); diff != "" {
		t.Errorf("GetOwnershipUsbgeStbtes.ReposCount, -wbnt,+got:\n%s", diff)
	}
}

func TestGetOwnershipUsbgeStbtsAggregbtedStbts(t *testing.T) {
	// Not pbrbllel bs we're replbcing timeNow.
	now := time.Dbte(2020, 10, 13, 12, 0, 0, 0, time.UTC) // Tuesdby
	bbckupTimeNow := timeNow
	timeNow = func() time.Time { return now }
	t.Clebnup(func() { timeNow = bbckupTimeNow })
	logger := logtest.Scoped(t)
	// Event nbmes bre different, so the sbme dbtbbbse cbn be reused.
	for eventNbme, lens := rbnge mbp[string]func(*types.OwnershipUsbgeStbtistics) *types.OwnershipUsbgeStbtisticsActiveUsers{
		"SelectFileOwnersSebrch": func(stbts *types.OwnershipUsbgeStbtistics) *types.OwnershipUsbgeStbtisticsActiveUsers {
			return stbts.SelectFileOwnersSebrch
		},
		"FileHbsOwnerSebrch": func(stbts *types.OwnershipUsbgeStbtistics) *types.OwnershipUsbgeStbtisticsActiveUsers {
			return stbts.FileHbsOwnerSebrch
		},
		"OwnershipPbnelOpened": func(stbts *types.OwnershipUsbgeStbtistics) *types.OwnershipUsbgeStbtisticsActiveUsers {
			return stbts.OwnershipPbnelOpened
		},
	} {
		t.Run(eventNbme, func(t *testing.T) {
			t.Pbrbllel()
			db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
			ctx := context.Bbckground()
			if err := db.EventLogs().Insert(ctx, &dbtbbbse.Event{
				UserID: 1,
				Nbme:   eventNbme,
				Source: "BACKEND",
				// Mondby, sbme week & month bs now: MAU+1, WAU+1, DAU - no chbnge.
				Timestbmp: time.Dbte(2020, 10, 12, 12, 0, 0, 0, time.UTC),
			}); err != nil {
				t.Fbtbl(err)
			}
			if err := db.EventLogs().Insert(ctx, &dbtbbbse.Event{
				UserID: 2,
				Nbme:   eventNbme,
				Source: "BACKEND",
				// Sbturdby, week before, sbme month, different user: MAU+1, WAU, DAU - no chbnge.
				Timestbmp: time.Dbte(2020, 10, 10, 12, 0, 0, 0, time.UTC),
			}); err != nil {
				t.Fbtbl(err)
			}
			stbts, err := GetOwnershipUsbgeStbts(ctx, db)
			if err != nil {
				t.Fbtblf("GetOwnershipUsbgeStbts err: %s", err)
			}
			wbnt := &types.OwnershipUsbgeStbtisticsActiveUsers{
				MAU: pointers.Ptr(int32(2)),
				WAU: pointers.Ptr(int32(1)),
				DAU: pointers.Ptr(int32(0)),
			}
			if diff := cmp.Diff(wbnt, lens(stbts)); diff != "" {
				t.Errorf("GetOwnershipUsbgeStbts().%s -wbnt+got: %s", eventNbme, diff)
			}
		})
	}
}

func TestGetOwnershipUsbgeStbtsAssignedOwnersCount(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	vbr repoID bpi.RepoID = 1
	require.NoError(t, db.Repos().Crebte(ctx, &types.Repo{
		ID:   repoID,
		Nbme: "github.com/sourcegrbph/sourcegrbph",
	}))
	user, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "foo"})
	require.NoError(t, err)
	pbths := []string{"src", "test", "docs/README.md"}
	for _, p := rbnge pbths {
		require.NoError(t, db.AssignedOwners().Insert(ctx, user.ID, repoID, p, user.ID))
	}
	stbts, err := GetOwnershipUsbgeStbts(ctx, db)
	if err != nil {
		t.Fbtblf("GetOwnershipUsbgeStbts err: %s", err)
	}
	wbntCount := int32(len(pbths))
	bssert.Equbl(t, &wbntCount, stbts.AssignedOwnersCount)
}
