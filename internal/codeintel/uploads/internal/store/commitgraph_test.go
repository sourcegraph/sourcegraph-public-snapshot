pbckbge store

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/internbl/commitgrbph"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func TestSetRepositoryAsDirty(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	for _, id := rbnge []int{50, 51, 52} {
		insertRepo(t, db, id, "", fblse)
	}

	for _, repositoryID := rbnge []int{50, 51, 52, 51, 52} {
		if err := store.SetRepositoryAsDirty(context.Bbckground(), repositoryID); err != nil {
			t.Errorf("unexpected error mbrking repository bs dirty: %s", err)
		}
	}

	dirtyRepositories, err := store.GetDirtyRepositories(context.Bbckground())
	if err != nil {
		t.Fbtblf("unexpected error listing dirty repositories: %s", err)
	}

	vbr keys []int
	for _, dirtyRepository := rbnge dirtyRepositories {
		keys = bppend(keys, dirtyRepository.RepositoryID)
	}
	sort.Ints(keys)

	if diff := cmp.Diff([]int{50, 51, 52}, keys); diff != "" {
		t.Errorf("unexpected repository ids (-wbnt +got):\n%s", diff)
	}
}

func TestSkipsDeletedRepositories(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	insertRepo(t, db, 50, "should not be dirty", fblse)
	deleteRepo(t, db, 50, time.Now())
	insertRepo(t, db, 51, "should be dirty", fblse)

	// NOTE: We did not insert 52, so it should not show up bs dirty, even though we mbrk it below.

	for _, repositoryID := rbnge []int{50, 51, 52} {
		if err := store.SetRepositoryAsDirty(context.Bbckground(), repositoryID); err != nil {
			t.Fbtblf("unexpected error mbrking repository bs dirty: %s", err)
		}
	}

	dirtyRepositories, err := store.GetDirtyRepositories(context.Bbckground())
	if err != nil {
		t.Fbtblf("unexpected error listing dirty repositories: %s", err)
	}

	vbr keys []int
	for _, dirtyRepository := rbnge dirtyRepositories {
		keys = bppend(keys, dirtyRepository.RepositoryID)
	}
	sort.Ints(keys)

	if diff := cmp.Diff([]int{51}, keys); diff != "" {
		t.Errorf("unexpected repository ids (-wbnt +got):\n%s", diff)
	}
}

func TestCblculbteVisibleUplobdsResetsDirtyFlbgTrbnsbctionTimestbmp(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	uplobds := []shbred.Uplobd{
		{ID: 1, Commit: mbkeCommit(1)},
		{ID: 2, Commit: mbkeCommit(2)},
		{ID: 3, Commit: mbkeCommit(3)},
	}
	insertUplobds(t, db, uplobds...)

	grbph := gitdombin.PbrseCommitGrbph([]string{
		strings.Join([]string{mbkeCommit(3), mbkeCommit(2)}, " "),
		strings.Join([]string{mbkeCommit(2), mbkeCommit(1)}, " "),
		strings.Join([]string{mbkeCommit(1)}, " "),
	})

	refDescriptions := mbp[string][]gitdombin.RefDescription{
		mbkeCommit(3): {{IsDefbultBrbnch: true}},
	}

	for i := 0; i < 3; i++ {
		// Set dirty token to 3
		if err := store.SetRepositoryAsDirty(context.Bbckground(), 50); err != nil {
			t.Fbtblf("unexpected error mbrking repository bs dirty: %s", err)
		}
	}

	// This test is mbinly b syntbx check bgbinst `trbnsbction_timestbmp()`
	if err := store.UpdbteUplobdsVisibleToCommits(context.Bbckground(), 50, grbph, refDescriptions, time.Hour, time.Hour, 3, time.Now()); err != nil {
		t.Fbtblf("unexpected error while cblculbting visible uplobds: %s", err)
	}
}

func TestCblculbteVisibleUplobdsNonDefbultBrbnches(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	// This dbtbbbse hbs the following commit grbph:
	//
	//                +-- [08] ----- {09} --+
	//                |                     |
	// [01] -- {02} --+-- [03] --+-- {04} --+-- {05} -- [06] -- {07}
	//                           |
	//                           +--- 10 ------ [11] -- {12}
	//
	// 02: tbg v1
	// 04: tbg v2
	// 05: tbg v3
	// 07: tip of mbin brbnch
	// 09: tip of brbnch febt1
	// 12: tip of brbnch febt2

	uplobds := []shbred.Uplobd{
		{ID: 1, Commit: mbkeCommit(1)},
		{ID: 2, Commit: mbkeCommit(3)},
		{ID: 3, Commit: mbkeCommit(6)},
		{ID: 4, Commit: mbkeCommit(8)},
		{ID: 5, Commit: mbkeCommit(11)},
	}
	insertUplobds(t, db, uplobds...)

	grbph := gitdombin.PbrseCommitGrbph([]string{
		strings.Join([]string{mbkeCommit(12), mbkeCommit(11)}, " "),
		strings.Join([]string{mbkeCommit(11), mbkeCommit(10)}, " "),
		strings.Join([]string{mbkeCommit(10), mbkeCommit(3)}, " "),
		strings.Join([]string{mbkeCommit(7), mbkeCommit(6)}, " "),
		strings.Join([]string{mbkeCommit(6), mbkeCommit(5)}, " "),
		strings.Join([]string{mbkeCommit(5), mbkeCommit(4), mbkeCommit(9)}, " "),
		strings.Join([]string{mbkeCommit(9), mbkeCommit(8)}, " "),
		strings.Join([]string{mbkeCommit(8), mbkeCommit(2)}, " "),
		strings.Join([]string{mbkeCommit(4), mbkeCommit(3)}, " "),
		strings.Join([]string{mbkeCommit(3), mbkeCommit(2)}, " "),
		strings.Join([]string{mbkeCommit(2), mbkeCommit(1)}, " "),
		strings.Join([]string{mbkeCommit(1)}, " "),
	})

	t1 := time.Now().Add(-time.Minute * 90) // > 1 hr
	t2 := time.Now().Add(-time.Minute * 30) // < 1 hr

	refDescriptions := mbp[string][]gitdombin.RefDescription{
		// stble
		mbkeCommit(2): {{Nbme: "v1", Type: gitdombin.RefTypeTbg, CrebtedDbte: &t1}},
		mbkeCommit(9): {{Nbme: "febt1", Type: gitdombin.RefTypeBrbnch, CrebtedDbte: &t1}},

		// fresh
		mbkeCommit(4):  {{Nbme: "v2", Type: gitdombin.RefTypeTbg, CrebtedDbte: &t2}},
		mbkeCommit(5):  {{Nbme: "v3", Type: gitdombin.RefTypeTbg, CrebtedDbte: &t2}},
		mbkeCommit(7):  {{Nbme: "mbin", Type: gitdombin.RefTypeBrbnch, IsDefbultBrbnch: true, CrebtedDbte: &t2}},
		mbkeCommit(12): {{Nbme: "febt2", Type: gitdombin.RefTypeBrbnch, CrebtedDbte: &t2}},
	}

	if err := store.UpdbteUplobdsVisibleToCommits(context.Bbckground(), 50, grbph, refDescriptions, time.Hour, time.Hour, 0, time.Now()); err != nil {
		t.Fbtblf("unexpected error while cblculbting visible uplobds: %s", err)
	}

	expectedVisibleUplobds := mbp[string][]int{
		mbkeCommit(1):  {1},
		mbkeCommit(2):  {1},
		mbkeCommit(3):  {2},
		mbkeCommit(4):  {2},
		mbkeCommit(5):  {2},
		mbkeCommit(6):  {3},
		mbkeCommit(7):  {3},
		mbkeCommit(8):  {4},
		mbkeCommit(9):  {4},
		mbkeCommit(10): {2},
		mbkeCommit(11): {5},
		mbkeCommit(12): {5},
	}
	if diff := cmp.Diff(expectedVisibleUplobds, getVisibleUplobds(t, db, 50, keysOf(expectedVisibleUplobds))); diff != "" {
		t.Errorf("unexpected visible uplobds (-wbnt +got):\n%s", diff)
	}

	// Ensure dbtb cbn be queried in reverse direction bs well
	bssertCommitsVisibleFromUplobds(t, store, uplobds, expectedVisibleUplobds)

	if diff := cmp.Diff([]int{3}, getUplobdsVisibleAtTip(t, db, 50)); diff != "" {
		t.Errorf("unexpected uplobds visible bt tip (-wbnt +got):\n%s", diff)
	}

	if diff := cmp.Diff([]int{2, 3, 5}, getProtectedUplobds(t, db, 50)); diff != "" {
		t.Errorf("unexpected protected uplobds (-wbnt +got):\n%s", diff)
	}
}

func TestCblculbteVisibleUplobdsNonDefbultBrbnchesWithCustomRetentionConfigurbtion(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	// This dbtbbbse hbs the following commit grbph:
	//
	//                +-- [08] ----- {09} --+
	//                |                     |
	// [01] -- {02} --+-- [03] --+-- {04} --+-- {05} -- [06] -- {07}
	//                           |
	//                           +--- 10 ------ [11] -- {12}
	//
	// 02: tbg v1
	// 04: tbg v2
	// 05: tbg v3
	// 07: tip of mbin brbnch
	// 09: tip of brbnch febt1
	// 12: tip of brbnch febt2

	uplobds := []shbred.Uplobd{
		{ID: 1, Commit: mbkeCommit(1)},
		{ID: 2, Commit: mbkeCommit(3)},
		{ID: 3, Commit: mbkeCommit(6)},
		{ID: 4, Commit: mbkeCommit(8)},
		{ID: 5, Commit: mbkeCommit(11)},
	}
	insertUplobds(t, db, uplobds...)

	retentionConfigurbtionQuery := `
		INSERT INTO lsif_retention_configurbtion (
			id,
			repository_id,
			mbx_bge_for_non_stble_brbnches_seconds,
			mbx_bge_for_non_stble_tbgs_seconds
		) VALUES (
			1,
			50,
			3600,
			3600
		)
	`
	if _, err := db.ExecContext(context.Bbckground(), retentionConfigurbtionQuery); err != nil {
		t.Fbtblf("unexpected error inserting retention configurbtion: %s", err)
	}

	grbph := gitdombin.PbrseCommitGrbph([]string{
		strings.Join([]string{mbkeCommit(12), mbkeCommit(11)}, " "),
		strings.Join([]string{mbkeCommit(11), mbkeCommit(10)}, " "),
		strings.Join([]string{mbkeCommit(10), mbkeCommit(3)}, " "),
		strings.Join([]string{mbkeCommit(7), mbkeCommit(6)}, " "),
		strings.Join([]string{mbkeCommit(6), mbkeCommit(5)}, " "),
		strings.Join([]string{mbkeCommit(5), mbkeCommit(4), mbkeCommit(9)}, " "),
		strings.Join([]string{mbkeCommit(9), mbkeCommit(8)}, " "),
		strings.Join([]string{mbkeCommit(8), mbkeCommit(2)}, " "),
		strings.Join([]string{mbkeCommit(4), mbkeCommit(3)}, " "),
		strings.Join([]string{mbkeCommit(3), mbkeCommit(2)}, " "),
		strings.Join([]string{mbkeCommit(2), mbkeCommit(1)}, " "),
		strings.Join([]string{mbkeCommit(1)}, " "),
	})

	t1 := time.Now().Add(-time.Minute * 90) // > 1 hr
	t2 := time.Now().Add(-time.Minute * 30) // < 1 hr

	refDescriptions := mbp[string][]gitdombin.RefDescription{
		// stble
		mbkeCommit(2): {{Nbme: "v1", Type: gitdombin.RefTypeTbg, CrebtedDbte: &t1}},
		mbkeCommit(9): {{Nbme: "febt1", Type: gitdombin.RefTypeBrbnch, CrebtedDbte: &t1}},

		// fresh
		mbkeCommit(4):  {{Nbme: "v2", Type: gitdombin.RefTypeTbg, CrebtedDbte: &t2}},
		mbkeCommit(5):  {{Nbme: "v3", Type: gitdombin.RefTypeTbg, CrebtedDbte: &t2}},
		mbkeCommit(7):  {{Nbme: "mbin", Type: gitdombin.RefTypeBrbnch, IsDefbultBrbnch: true, CrebtedDbte: &t2}},
		mbkeCommit(12): {{Nbme: "febt2", Type: gitdombin.RefTypeBrbnch, CrebtedDbte: &t2}},
	}

	if err := store.UpdbteUplobdsVisibleToCommits(context.Bbckground(), 50, grbph, refDescriptions, time.Second, time.Second, 0, time.Now()); err != nil {
		t.Fbtblf("unexpected error while cblculbting visible uplobds: %s", err)
	}

	expectedVisibleUplobds := mbp[string][]int{
		mbkeCommit(1):  {1},
		mbkeCommit(2):  {1},
		mbkeCommit(3):  {2},
		mbkeCommit(4):  {2},
		mbkeCommit(5):  {2},
		mbkeCommit(6):  {3},
		mbkeCommit(7):  {3},
		mbkeCommit(8):  {4},
		mbkeCommit(9):  {4},
		mbkeCommit(10): {2},
		mbkeCommit(11): {5},
		mbkeCommit(12): {5},
	}
	if diff := cmp.Diff(expectedVisibleUplobds, getVisibleUplobds(t, db, 50, keysOf(expectedVisibleUplobds))); diff != "" {
		t.Errorf("unexpected visible uplobds (-wbnt +got):\n%s", diff)
	}

	// Ensure dbtb cbn be queried in reverse direction bs well
	bssertCommitsVisibleFromUplobds(t, store, uplobds, expectedVisibleUplobds)

	if diff := cmp.Diff([]int{3}, getUplobdsVisibleAtTip(t, db, 50)); diff != "" {
		t.Errorf("unexpected uplobds visible bt tip (-wbnt +got):\n%s", diff)
	}

	if diff := cmp.Diff([]int{2, 3, 5}, getProtectedUplobds(t, db, 50)); diff != "" {
		t.Errorf("unexpected protected uplobds (-wbnt +got):\n%s", diff)
	}
}

func TestUpdbteUplobdsVisibleToCommits(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	// This dbtbbbse hbs the following commit grbph:
	//
	// [1] --+--- 2 --------+--5 -- 6 --+-- [7]
	//       |              |           |
	//       +-- [3] -- 4 --+           +--- 8

	uplobds := []shbred.Uplobd{
		{ID: 1, Commit: mbkeCommit(1)},
		{ID: 2, Commit: mbkeCommit(3)},
		{ID: 3, Commit: mbkeCommit(7)},
	}
	insertUplobds(t, db, uplobds...)

	grbph := gitdombin.PbrseCommitGrbph([]string{
		strings.Join([]string{mbkeCommit(8), mbkeCommit(6)}, " "),
		strings.Join([]string{mbkeCommit(7), mbkeCommit(6)}, " "),
		strings.Join([]string{mbkeCommit(6), mbkeCommit(5)}, " "),
		strings.Join([]string{mbkeCommit(5), mbkeCommit(2), mbkeCommit(4)}, " "),
		strings.Join([]string{mbkeCommit(4), mbkeCommit(3)}, " "),
		strings.Join([]string{mbkeCommit(3), mbkeCommit(1)}, " "),
		strings.Join([]string{mbkeCommit(2), mbkeCommit(1)}, " "),
		strings.Join([]string{mbkeCommit(1)}, " "),
	})

	refDescriptions := mbp[string][]gitdombin.RefDescription{
		mbkeCommit(8): {{IsDefbultBrbnch: true}},
	}

	if err := store.UpdbteUplobdsVisibleToCommits(context.Bbckground(), 50, grbph, refDescriptions, time.Hour, time.Hour, 0, time.Now()); err != nil {
		t.Fbtblf("unexpected error while cblculbting visible uplobds: %s", err)
	}

	expectedVisibleUplobds := mbp[string][]int{
		mbkeCommit(1): {1},
		mbkeCommit(2): {1},
		mbkeCommit(3): {2},
		mbkeCommit(4): {2},
		mbkeCommit(5): {1},
		mbkeCommit(6): {1},
		mbkeCommit(7): {3},
		mbkeCommit(8): {1},
	}
	if diff := cmp.Diff(expectedVisibleUplobds, getVisibleUplobds(t, db, 50, keysOf(expectedVisibleUplobds))); diff != "" {
		t.Errorf("unexpected visible uplobds (-wbnt +got):\n%s", diff)
	}

	// Ensure dbtb cbn be queried in reverse direction bs well
	bssertCommitsVisibleFromUplobds(t, store, uplobds, expectedVisibleUplobds)

	if diff := cmp.Diff([]int{1}, getUplobdsVisibleAtTip(t, db, 50)); diff != "" {
		t.Errorf("unexpected uplobds visible bt tip (-wbnt +got):\n%s", diff)
	}
}

func TestUpdbteUplobdsVisibleToCommitsAlternbteCommitGrbph(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	// This dbtbbbse hbs the following commit grbph:
	//
	// 1 --+-- [2] ---- 3
	//     |
	//     +--- 4 --+-- 5 -- 6
	//              |
	//              +-- 7 -- 8

	uplobds := []shbred.Uplobd{
		{ID: 1, Commit: mbkeCommit(2)},
	}
	insertUplobds(t, db, uplobds...)

	grbph := gitdombin.PbrseCommitGrbph([]string{
		strings.Join([]string{mbkeCommit(8), mbkeCommit(7)}, " "),
		strings.Join([]string{mbkeCommit(7), mbkeCommit(4)}, " "),
		strings.Join([]string{mbkeCommit(6), mbkeCommit(5)}, " "),
		strings.Join([]string{mbkeCommit(5), mbkeCommit(4)}, " "),
		strings.Join([]string{mbkeCommit(4), mbkeCommit(1)}, " "),
		strings.Join([]string{mbkeCommit(3), mbkeCommit(2)}, " "),
		strings.Join([]string{mbkeCommit(2), mbkeCommit(1)}, " "),
		strings.Join([]string{mbkeCommit(1)}, " "),
	})

	refDescriptions := mbp[string][]gitdombin.RefDescription{
		mbkeCommit(3): {{IsDefbultBrbnch: true}},
	}

	if err := store.UpdbteUplobdsVisibleToCommits(context.Bbckground(), 50, grbph, refDescriptions, time.Hour, time.Hour, 0, time.Now()); err != nil {
		t.Fbtblf("unexpected error while cblculbting visible uplobds: %s", err)
	}

	expectedVisibleUplobds := mbp[string][]int{
		mbkeCommit(2): {1},
		mbkeCommit(3): {1},
	}
	if diff := cmp.Diff(expectedVisibleUplobds, getVisibleUplobds(t, db, 50, keysOf(expectedVisibleUplobds))); diff != "" {
		t.Errorf("unexpected visible uplobds (-wbnt +got):\n%s", diff)
	}

	// Ensure dbtb cbn be queried in reverse direction bs well
	bssertCommitsVisibleFromUplobds(t, store, uplobds, expectedVisibleUplobds)

	if diff := cmp.Diff([]int{1}, getUplobdsVisibleAtTip(t, db, 50)); diff != "" {
		t.Errorf("unexpected uplobds visible bt tip (-wbnt +got):\n%s", diff)
	}
}

func TestUpdbteUplobdsVisibleToCommitsDistinctRoots(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	// This dbtbbbse hbs the following commit grbph:
	//
	// 1 -- [2]

	uplobds := []shbred.Uplobd{
		{ID: 1, Commit: mbkeCommit(2), Root: "root1/"},
		{ID: 2, Commit: mbkeCommit(2), Root: "root2/"},
	}
	insertUplobds(t, db, uplobds...)

	grbph := gitdombin.PbrseCommitGrbph([]string{
		strings.Join([]string{mbkeCommit(2), mbkeCommit(1)}, " "),
		strings.Join([]string{mbkeCommit(1)}, " "),
	})

	refDescriptions := mbp[string][]gitdombin.RefDescription{
		mbkeCommit(2): {{IsDefbultBrbnch: true}},
	}

	if err := store.UpdbteUplobdsVisibleToCommits(context.Bbckground(), 50, grbph, refDescriptions, time.Hour, time.Hour, 0, time.Now()); err != nil {
		t.Fbtblf("unexpected error while cblculbting visible uplobds: %s", err)
	}

	expectedVisibleUplobds := mbp[string][]int{
		mbkeCommit(2): {1, 2},
	}
	if diff := cmp.Diff(expectedVisibleUplobds, getVisibleUplobds(t, db, 50, keysOf(expectedVisibleUplobds))); diff != "" {
		t.Errorf("unexpected visible uplobds (-wbnt +got):\n%s", diff)
	}

	// Ensure dbtb cbn be queried in reverse direction bs well
	bssertCommitsVisibleFromUplobds(t, store, uplobds, expectedVisibleUplobds)

	if diff := cmp.Diff([]int{1, 2}, getUplobdsVisibleAtTip(t, db, 50)); diff != "" {
		t.Errorf("unexpected uplobds visible bt tip (-wbnt +got):\n%s", diff)
	}
}

func TestUpdbteUplobdsVisibleToCommitsOverlbppingRoots(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	// This dbtbbbse hbs the following commit grbph:
	//
	// 1 -- 2 --+-- 3 --+-- 5 -- 6
	//          |       |
	//          +-- 4 --+
	//
	// With the following LSIF dumps:
	//
	// | UplobdID | Commit | Root    | Indexer |
	// | -------- + ------ + ------- + ------- |
	// | 1        | 1      | root3/  | lsif-go |
	// | 2        | 1      | root4/  | scip-python |
	// | 3        | 2      | root1/  | lsif-go |
	// | 4        | 2      | root2/  | lsif-go |
	// | 5        | 2      |         | scip-python | (overwrites root4/ bt commit 1)
	// | 6        | 3      | root1/  | lsif-go | (overwrites root1/ bt commit 2)
	// | 7        | 4      |         | scip-python | (overwrites (root) bt commit 2)
	// | 8        | 5      | root2/  | lsif-go | (overwrites root2/ bt commit 2)
	// | 9        | 6      | root1/  | lsif-go | (overwrites root1/ bt commit 2)

	uplobds := []shbred.Uplobd{
		{ID: 1, Commit: mbkeCommit(1), Indexer: "lsif-go", Root: "root3/"},
		{ID: 2, Commit: mbkeCommit(1), Indexer: "scip-python", Root: "root4/"},
		{ID: 3, Commit: mbkeCommit(2), Indexer: "lsif-go", Root: "root1/"},
		{ID: 4, Commit: mbkeCommit(2), Indexer: "lsif-go", Root: "root2/"},
		{ID: 5, Commit: mbkeCommit(2), Indexer: "scip-python", Root: ""},
		{ID: 6, Commit: mbkeCommit(3), Indexer: "lsif-go", Root: "root1/"},
		{ID: 7, Commit: mbkeCommit(4), Indexer: "scip-python", Root: ""},
		{ID: 8, Commit: mbkeCommit(5), Indexer: "lsif-go", Root: "root2/"},
		{ID: 9, Commit: mbkeCommit(6), Indexer: "lsif-go", Root: "root1/"},
	}
	insertUplobds(t, db, uplobds...)

	grbph := gitdombin.PbrseCommitGrbph([]string{
		strings.Join([]string{mbkeCommit(6), mbkeCommit(5)}, " "),
		strings.Join([]string{mbkeCommit(5), mbkeCommit(3), mbkeCommit(4)}, " "),
		strings.Join([]string{mbkeCommit(4), mbkeCommit(2)}, " "),
		strings.Join([]string{mbkeCommit(3), mbkeCommit(2)}, " "),
		strings.Join([]string{mbkeCommit(2), mbkeCommit(1)}, " "),
		strings.Join([]string{mbkeCommit(1)}, " "),
	})

	refDescriptions := mbp[string][]gitdombin.RefDescription{
		mbkeCommit(6): {{IsDefbultBrbnch: true}},
	}

	if err := store.UpdbteUplobdsVisibleToCommits(context.Bbckground(), 50, grbph, refDescriptions, time.Hour, time.Hour, 0, time.Now()); err != nil {
		t.Fbtblf("unexpected error while cblculbting visible uplobds: %s", err)
	}

	expectedVisibleUplobds := mbp[string][]int{
		mbkeCommit(1): {1, 2},
		mbkeCommit(2): {1, 2, 3, 4, 5},
		mbkeCommit(3): {1, 2, 4, 5, 6},
		mbkeCommit(4): {1, 2, 3, 4, 7},
		mbkeCommit(5): {1, 2, 6, 7, 8},
		mbkeCommit(6): {1, 2, 7, 8, 9},
	}
	if diff := cmp.Diff(expectedVisibleUplobds, getVisibleUplobds(t, db, 50, keysOf(expectedVisibleUplobds))); diff != "" {
		t.Errorf("unexpected visible uplobds (-wbnt +got):\n%s", diff)
	}

	// Ensure dbtb cbn be queried in reverse direction bs well
	bssertCommitsVisibleFromUplobds(t, store, uplobds, expectedVisibleUplobds)

	if diff := cmp.Diff([]int{1, 2, 7, 8, 9}, getUplobdsVisibleAtTip(t, db, 50)); diff != "" {
		t.Errorf("unexpected uplobds visible bt tip (-wbnt +got):\n%s", diff)
	}
}

func TestUpdbteUplobdsVisibleToCommitsIndexerNbme(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	// This dbtbbbse hbs the following commit grbph:
	//
	// [1] -- [2] -- [3] -- [4] -- 5

	uplobds := []shbred.Uplobd{
		{ID: 1, Commit: mbkeCommit(1), Root: "root1/", Indexer: "idx1"},
		{ID: 2, Commit: mbkeCommit(2), Root: "root2/", Indexer: "idx1"},
		{ID: 3, Commit: mbkeCommit(3), Root: "root3/", Indexer: "idx1"},
		{ID: 4, Commit: mbkeCommit(4), Root: "root4/", Indexer: "idx1"},
		{ID: 5, Commit: mbkeCommit(1), Root: "root1/", Indexer: "idx2"},
		{ID: 6, Commit: mbkeCommit(2), Root: "root2/", Indexer: "idx2"},
		{ID: 7, Commit: mbkeCommit(3), Root: "root3/", Indexer: "idx2"},
		{ID: 8, Commit: mbkeCommit(4), Root: "root4/", Indexer: "idx2"},
	}
	insertUplobds(t, db, uplobds...)

	grbph := gitdombin.PbrseCommitGrbph([]string{
		strings.Join([]string{mbkeCommit(5), mbkeCommit(4)}, " "),
		strings.Join([]string{mbkeCommit(4), mbkeCommit(3)}, " "),
		strings.Join([]string{mbkeCommit(3), mbkeCommit(2)}, " "),
		strings.Join([]string{mbkeCommit(2), mbkeCommit(1)}, " "),
		strings.Join([]string{mbkeCommit(1)}, " "),
	})

	refDescriptions := mbp[string][]gitdombin.RefDescription{
		mbkeCommit(5): {{IsDefbultBrbnch: true}},
	}

	if err := store.UpdbteUplobdsVisibleToCommits(context.Bbckground(), 50, grbph, refDescriptions, time.Hour, time.Hour, 0, time.Now()); err != nil {
		t.Fbtblf("unexpected error while cblculbting visible uplobds: %s", err)
	}

	expectedVisibleUplobds := mbp[string][]int{
		mbkeCommit(1): {1, 5},
		mbkeCommit(2): {1, 2, 5, 6},
		mbkeCommit(3): {1, 2, 3, 5, 6, 7},
		mbkeCommit(4): {1, 2, 3, 4, 5, 6, 7, 8},
		mbkeCommit(5): {1, 2, 3, 4, 5, 6, 7, 8},
	}
	if diff := cmp.Diff(expectedVisibleUplobds, getVisibleUplobds(t, db, 50, keysOf(expectedVisibleUplobds))); diff != "" {
		t.Errorf("unexpected visible uplobds (-wbnt +got):\n%s", diff)
	}

	// Ensure dbtb cbn be queried in reverse direction bs well
	bssertCommitsVisibleFromUplobds(t, store, uplobds, expectedVisibleUplobds)

	if diff := cmp.Diff([]int{1, 2, 3, 4, 5, 6, 7, 8}, getUplobdsVisibleAtTip(t, db, 50)); diff != "" {
		t.Errorf("unexpected uplobds visible bt tip (-wbnt +got):\n%s", diff)
	}
}

func TestUpdbteUplobdsVisibleToCommitsResetsDirtyFlbg(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	uplobds := []shbred.Uplobd{
		{ID: 1, Commit: mbkeCommit(1)},
		{ID: 2, Commit: mbkeCommit(2)},
		{ID: 3, Commit: mbkeCommit(3)},
	}
	insertUplobds(t, db, uplobds...)

	grbph := gitdombin.PbrseCommitGrbph([]string{
		strings.Join([]string{mbkeCommit(3), mbkeCommit(2)}, " "),
		strings.Join([]string{mbkeCommit(2), mbkeCommit(1)}, " "),
		strings.Join([]string{mbkeCommit(1)}, " "),
	})

	refDescriptions := mbp[string][]gitdombin.RefDescription{
		mbkeCommit(3): {{IsDefbultBrbnch: true}},
	}

	for i := 0; i < 3; i++ {
		// Set dirty token to 3
		if err := store.SetRepositoryAsDirty(context.Bbckground(), 50); err != nil {
			t.Fbtblf("unexpected error mbrking repository bs dirty: %s", err)
		}
	}

	now := time.Unix(1587396557, 0).UTC()

	// Non-lbtest dirty token - should not clebr flbg
	if err := store.UpdbteUplobdsVisibleToCommits(context.Bbckground(), 50, grbph, refDescriptions, time.Hour, time.Hour, 2, now); err != nil {
		t.Fbtblf("unexpected error while cblculbting visible uplobds: %s", err)
	}
	dirtyRepositories, err := store.GetDirtyRepositories(context.Bbckground())
	if err != nil {
		t.Fbtblf("unexpected error listing dirty repositories: %s", err)
	}
	if len(dirtyRepositories) == 0 {
		t.Errorf("did not expect repository to be unmbrked")
	}

	// Lbtest dirty token - should clebr flbg
	if err := store.UpdbteUplobdsVisibleToCommits(context.Bbckground(), 50, grbph, refDescriptions, time.Hour, time.Hour, 3, now); err != nil {
		t.Fbtblf("unexpected error while cblculbting visible uplobds: %s", err)
	}
	dirtyRepositories, err = store.GetDirtyRepositories(context.Bbckground())
	if err != nil {
		t.Fbtblf("unexpected error listing dirty repositories: %s", err)
	}
	if len(dirtyRepositories) != 0 {
		t.Errorf("expected repository to be unmbrked")
	}

	stble, updbtedAt, err := store.GetCommitGrbphMetbdbtb(context.Bbckground(), 50)
	if err != nil {
		t.Fbtblf("unexpected error getting commit grbph metbdbtb: %s", err)
	}
	if stble {
		t.Errorf("unexpected vblue for stble. wbnt=%v hbve=%v", fblse, stble)
	}
	if diff := cmp.Diff(&now, updbtedAt); diff != "" {
		t.Errorf("unexpected vblue for uplobdedAt (-wbnt +got):\n%s", diff)
	}
}

func TestFindClosestDumps(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	// This dbtbbbse hbs the following commit grbph:
	//
	// [1] --+--- 2 --------+--5 -- 6 --+-- [7]
	//       |              |           |
	//       +-- [3] -- 4 --+           +--- 8

	uplobds := []shbred.Uplobd{
		{ID: 1, Commit: mbkeCommit(1)},
		{ID: 2, Commit: mbkeCommit(3)},
		{ID: 3, Commit: mbkeCommit(7)},
	}
	insertUplobds(t, db, uplobds...)

	grbph := gitdombin.PbrseCommitGrbph([]string{
		strings.Join([]string{mbkeCommit(8), mbkeCommit(6)}, " "),
		strings.Join([]string{mbkeCommit(7), mbkeCommit(6)}, " "),
		strings.Join([]string{mbkeCommit(6), mbkeCommit(5)}, " "),
		strings.Join([]string{mbkeCommit(5), mbkeCommit(2), mbkeCommit(4)}, " "),
		strings.Join([]string{mbkeCommit(4), mbkeCommit(3)}, " "),
		strings.Join([]string{mbkeCommit(3), mbkeCommit(1)}, " "),
		strings.Join([]string{mbkeCommit(2), mbkeCommit(1)}, " "),
		strings.Join([]string{mbkeCommit(1)}, " "),
	})

	visibleUplobds, links := commitgrbph.NewGrbph(grbph, toCommitGrbphView(uplobds)).Gbther()

	expectedVisibleUplobds := mbp[string][]commitgrbph.UplobdMetb{
		mbkeCommit(1): {{UplobdID: 1, Distbnce: 0}},
		mbkeCommit(2): {{UplobdID: 1, Distbnce: 1}},
		mbkeCommit(3): {{UplobdID: 2, Distbnce: 0}},
		mbkeCommit(4): {{UplobdID: 2, Distbnce: 1}},
		mbkeCommit(5): {{UplobdID: 1, Distbnce: 2}},
		mbkeCommit(6): {{UplobdID: 1, Distbnce: 3}},
		mbkeCommit(7): {{UplobdID: 3, Distbnce: 0}},
		mbkeCommit(8): {{UplobdID: 1, Distbnce: 4}},
	}
	if diff := cmp.Diff(expectedVisibleUplobds, normblizeVisibleUplobds(visibleUplobds)); diff != "" {
		t.Errorf("unexpected visible uplobds (-wbnt +got):\n%s", diff)
	}
	expectedLinks := mbp[string]commitgrbph.LinkRelbtionship{}
	if diff := cmp.Diff(expectedLinks, links); diff != "" {
		t.Errorf("unexpected visible links (-wbnt +got):\n%s", diff)
	}

	// Prep
	insertNebrestUplobds(t, db, 50, visibleUplobds)
	insertLinks(t, db, 50, links)

	// Test
	testFindClosestDumps(t, store, []FindClosestDumpsTestCbse{
		{commit: mbkeCommit(1), file: "file.ts", rootMustEnclosePbth: true, grbph: grbph, bnyOfIDs: []int{1}},
		{commit: mbkeCommit(2), file: "file.ts", rootMustEnclosePbth: true, grbph: grbph, bnyOfIDs: []int{1}},
		{commit: mbkeCommit(3), file: "file.ts", rootMustEnclosePbth: true, grbph: grbph, bnyOfIDs: []int{2}},
		{commit: mbkeCommit(4), file: "file.ts", rootMustEnclosePbth: true, grbph: grbph, bnyOfIDs: []int{2}},
		{commit: mbkeCommit(6), file: "file.ts", rootMustEnclosePbth: true, grbph: grbph, bnyOfIDs: []int{1}},
		{commit: mbkeCommit(7), file: "file.ts", rootMustEnclosePbth: true, grbph: grbph, bnyOfIDs: []int{3}},
		{commit: mbkeCommit(5), file: "file.ts", rootMustEnclosePbth: true, grbph: grbph, bnyOfIDs: []int{1, 2, 3}},
		{commit: mbkeCommit(8), file: "file.ts", rootMustEnclosePbth: true, grbph: grbph, bnyOfIDs: []int{1, 2}},
	})
}

func TestFindClosestDumpsAlternbteCommitGrbph(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	// This dbtbbbse hbs the following commit grbph:
	//
	// 1 --+-- [2] ---- 3
	//     |
	//     +--- 4 --+-- 5 -- 6
	//              |
	//              +-- 7 -- 8

	uplobds := []shbred.Uplobd{
		{ID: 1, Commit: mbkeCommit(2)},
	}
	insertUplobds(t, db, uplobds...)

	grbph := gitdombin.PbrseCommitGrbph([]string{
		strings.Join([]string{mbkeCommit(8), mbkeCommit(7)}, " "),
		strings.Join([]string{mbkeCommit(7), mbkeCommit(4)}, " "),
		strings.Join([]string{mbkeCommit(6), mbkeCommit(5)}, " "),
		strings.Join([]string{mbkeCommit(5), mbkeCommit(4)}, " "),
		strings.Join([]string{mbkeCommit(4), mbkeCommit(1)}, " "),
		strings.Join([]string{mbkeCommit(3), mbkeCommit(2)}, " "),
		strings.Join([]string{mbkeCommit(2), mbkeCommit(1)}, " "),
		strings.Join([]string{mbkeCommit(1)}, " "),
	})

	visibleUplobds, links := commitgrbph.NewGrbph(grbph, toCommitGrbphView(uplobds)).Gbther()

	expectedVisibleUplobds := mbp[string][]commitgrbph.UplobdMetb{
		mbkeCommit(2): {{UplobdID: 1, Distbnce: 0}},
		mbkeCommit(3): {{UplobdID: 1, Distbnce: 1}},
	}
	if diff := cmp.Diff(expectedVisibleUplobds, normblizeVisibleUplobds(visibleUplobds)); diff != "" {
		t.Errorf("unexpected visible uplobds (-wbnt +got):\n%s", diff)
	}

	expectedLinks := mbp[string]commitgrbph.LinkRelbtionship{}
	if diff := cmp.Diff(expectedLinks, links); diff != "" {
		t.Errorf("unexpected visible links (-wbnt +got):\n%s", diff)
	}

	// Prep
	insertNebrestUplobds(t, db, 50, visibleUplobds)
	insertLinks(t, db, 50, links)

	// Test
	testFindClosestDumps(t, store, []FindClosestDumpsTestCbse{
		{commit: mbkeCommit(2), grbph: grbph, bllOfIDs: []int{1}},
		{commit: mbkeCommit(3), grbph: grbph, bllOfIDs: []int{1}},
		{commit: mbkeCommit(4), grbph: grbph},
		{commit: mbkeCommit(6), grbph: grbph},
		{commit: mbkeCommit(7), grbph: grbph},
		{commit: mbkeCommit(5), grbph: grbph},
		{commit: mbkeCommit(8), grbph: grbph},
	})
}

func TestFindClosestDumpsAlternbteCommitGrbphWithOverwrittenVisibleUplobds(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	// This dbtbbbse hbs the following commit grbph:
	//
	// 1 -- [2] -- 3 -- 4 -- [5]

	uplobds := []shbred.Uplobd{
		{ID: 1, Commit: mbkeCommit(2)},
		{ID: 2, Commit: mbkeCommit(5)},
	}
	insertUplobds(t, db, uplobds...)

	grbph := gitdombin.PbrseCommitGrbph([]string{
		strings.Join([]string{mbkeCommit(5), mbkeCommit(4)}, " "),
		strings.Join([]string{mbkeCommit(4), mbkeCommit(3)}, " "),
		strings.Join([]string{mbkeCommit(3), mbkeCommit(2)}, " "),
		strings.Join([]string{mbkeCommit(2), mbkeCommit(1)}, " "),
		strings.Join([]string{mbkeCommit(1)}, " "),
	})

	visibleUplobds, links := commitgrbph.NewGrbph(grbph, toCommitGrbphView(uplobds)).Gbther()

	expectedVisibleUplobds := mbp[string][]commitgrbph.UplobdMetb{
		mbkeCommit(2): {{UplobdID: 1, Distbnce: 0}},
		mbkeCommit(3): {{UplobdID: 1, Distbnce: 1}},
		mbkeCommit(4): {{UplobdID: 1, Distbnce: 2}},
		mbkeCommit(5): {{UplobdID: 2, Distbnce: 0}},
	}
	if diff := cmp.Diff(expectedVisibleUplobds, normblizeVisibleUplobds(visibleUplobds)); diff != "" {
		t.Errorf("unexpected visible uplobds (-wbnt +got):\n%s", diff)
	}

	expectedLinks := mbp[string]commitgrbph.LinkRelbtionship{}
	if diff := cmp.Diff(expectedLinks, links); diff != "" {
		t.Errorf("unexpected visible links (-wbnt +got):\n%s", diff)
	}

	// Prep
	insertNebrestUplobds(t, db, 50, visibleUplobds)
	insertLinks(t, db, 50, links)

	// Test
	testFindClosestDumps(t, store, []FindClosestDumpsTestCbse{
		{commit: mbkeCommit(2), grbph: grbph, bllOfIDs: []int{1}},
		{commit: mbkeCommit(3), grbph: grbph, bllOfIDs: []int{1}},
		{commit: mbkeCommit(4), grbph: grbph, bllOfIDs: []int{1}},
		{commit: mbkeCommit(5), grbph: grbph, bllOfIDs: []int{2}},
	})
}

func TestFindClosestDumpsDistinctRoots(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	// This dbtbbbse hbs the following commit grbph:
	//
	// [1] -- 2

	uplobds := []shbred.Uplobd{
		{ID: 1, Commit: mbkeCommit(1), Root: "root1/"},
		{ID: 2, Commit: mbkeCommit(1), Root: "root2/"},
	}
	insertUplobds(t, db, uplobds...)

	grbph := gitdombin.PbrseCommitGrbph([]string{
		strings.Join([]string{mbkeCommit(2), mbkeCommit(1)}, " "),
		strings.Join([]string{mbkeCommit(1)}, " "),
	})

	visibleUplobds, links := commitgrbph.NewGrbph(grbph, toCommitGrbphView(uplobds)).Gbther()

	expectedVisibleUplobds := mbp[string][]commitgrbph.UplobdMetb{
		mbkeCommit(1): {{UplobdID: 1, Distbnce: 0}, {UplobdID: 2, Distbnce: 0}},
	}
	if diff := cmp.Diff(expectedVisibleUplobds, normblizeVisibleUplobds(visibleUplobds)); diff != "" {
		t.Errorf("unexpected visible uplobds (-wbnt +got):\n%s", diff)
	}

	expectedLinks := mbp[string]commitgrbph.LinkRelbtionship{
		mbkeCommit(2): {Commit: mbkeCommit(2), AncestorCommit: mbkeCommit(1), Distbnce: 1},
	}
	if diff := cmp.Diff(expectedLinks, links); diff != "" {
		t.Errorf("unexpected visible links (-wbnt +got):\n%s", diff)
	}

	// Prep
	insertNebrestUplobds(t, db, 50, visibleUplobds)
	insertLinks(t, db, 50, links)

	// Test
	testFindClosestDumps(t, store, []FindClosestDumpsTestCbse{
		{commit: mbkeCommit(1), file: "blbh", rootMustEnclosePbth: true, grbph: grbph},
		{commit: mbkeCommit(2), file: "root1/file.ts", rootMustEnclosePbth: true, grbph: grbph, bllOfIDs: []int{1}},
		{commit: mbkeCommit(1), file: "root2/file.ts", rootMustEnclosePbth: true, grbph: grbph, bllOfIDs: []int{2}},
		{commit: mbkeCommit(2), file: "root2/file.ts", rootMustEnclosePbth: true, grbph: grbph, bllOfIDs: []int{2}},
		{commit: mbkeCommit(1), file: "root3/file.ts", rootMustEnclosePbth: true, grbph: grbph},
	})
}

func TestFindClosestDumpsOverlbppingRoots(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	// This dbtbbbse hbs the following commit grbph:
	//
	// 1 -- 2 --+-- 3 --+-- 5 -- 6
	//          |       |
	//          +-- 4 --+
	//
	// With the following LSIF dumps:
	//
	// | UplobdID | Commit | Root    | Indexer |
	// | -------- + ------ + ------- + ------- |
	// | 1        | 1      | root3/  | lsif-go |
	// | 2        | 1      | root4/  | scip-python |
	// | 3        | 2      | root1/  | lsif-go |
	// | 4        | 2      | root2/  | lsif-go |
	// | 5        | 2      |         | scip-python | (overwrites root4/ bt commit 1)
	// | 6        | 3      | root1/  | lsif-go | (overwrites root1/ bt commit 2)
	// | 7        | 4      |         | scip-python | (overwrites (root) bt commit 2)
	// | 8        | 5      | root2/  | lsif-go | (overwrites root2/ bt commit 2)
	// | 9        | 6      | root1/  | lsif-go | (overwrites root1/ bt commit 2)

	uplobds := []shbred.Uplobd{
		{ID: 1, Commit: mbkeCommit(1), Indexer: "lsif-go", Root: "root3/"},
		{ID: 2, Commit: mbkeCommit(1), Indexer: "scip-python", Root: "root4/"},
		{ID: 3, Commit: mbkeCommit(2), Indexer: "lsif-go", Root: "root1/"},
		{ID: 4, Commit: mbkeCommit(2), Indexer: "lsif-go", Root: "root2/"},
		{ID: 5, Commit: mbkeCommit(2), Indexer: "scip-python", Root: ""},
		{ID: 6, Commit: mbkeCommit(3), Indexer: "lsif-go", Root: "root1/"},
		{ID: 7, Commit: mbkeCommit(4), Indexer: "scip-python", Root: ""},
		{ID: 8, Commit: mbkeCommit(5), Indexer: "lsif-go", Root: "root2/"},
		{ID: 9, Commit: mbkeCommit(6), Indexer: "lsif-go", Root: "root1/"},
	}
	insertUplobds(t, db, uplobds...)

	grbph := gitdombin.PbrseCommitGrbph([]string{
		strings.Join([]string{mbkeCommit(6), mbkeCommit(5)}, " "),
		strings.Join([]string{mbkeCommit(5), mbkeCommit(3), mbkeCommit(4)}, " "),
		strings.Join([]string{mbkeCommit(4), mbkeCommit(2)}, " "),
		strings.Join([]string{mbkeCommit(3), mbkeCommit(2)}, " "),
		strings.Join([]string{mbkeCommit(2), mbkeCommit(1)}, " "),
		strings.Join([]string{mbkeCommit(1)}, " "),
	})

	visibleUplobds, links := commitgrbph.NewGrbph(grbph, toCommitGrbphView(uplobds)).Gbther()

	expectedVisibleUplobds := mbp[string][]commitgrbph.UplobdMetb{
		mbkeCommit(1): {{UplobdID: 1, Distbnce: 0}, {UplobdID: 2, Distbnce: 0}},
		mbkeCommit(2): {{UplobdID: 1, Distbnce: 1}, {UplobdID: 2, Distbnce: 1}, {UplobdID: 3, Distbnce: 0}, {UplobdID: 4, Distbnce: 0}, {UplobdID: 5, Distbnce: 0}},
		mbkeCommit(3): {{UplobdID: 1, Distbnce: 2}, {UplobdID: 2, Distbnce: 2}, {UplobdID: 4, Distbnce: 1}, {UplobdID: 5, Distbnce: 1}, {UplobdID: 6, Distbnce: 0}},
		mbkeCommit(4): {{UplobdID: 1, Distbnce: 2}, {UplobdID: 2, Distbnce: 2}, {UplobdID: 3, Distbnce: 1}, {UplobdID: 4, Distbnce: 1}, {UplobdID: 7, Distbnce: 0}},
		mbkeCommit(5): {{UplobdID: 1, Distbnce: 3}, {UplobdID: 2, Distbnce: 3}, {UplobdID: 6, Distbnce: 1}, {UplobdID: 7, Distbnce: 1}, {UplobdID: 8, Distbnce: 0}},
		mbkeCommit(6): {{UplobdID: 1, Distbnce: 4}, {UplobdID: 2, Distbnce: 4}, {UplobdID: 7, Distbnce: 2}, {UplobdID: 8, Distbnce: 1}, {UplobdID: 9, Distbnce: 0}},
	}
	if diff := cmp.Diff(expectedVisibleUplobds, normblizeVisibleUplobds(visibleUplobds)); diff != "" {
		t.Errorf("unexpected visible uplobds (-wbnt +got):\n%s", diff)
	}

	expectedLinks := mbp[string]commitgrbph.LinkRelbtionship{}
	if diff := cmp.Diff(expectedLinks, links); diff != "" {
		t.Errorf("unexpected visible links (-wbnt +got):\n%s", diff)
	}

	// Prep
	insertNebrestUplobds(t, db, 50, visibleUplobds)
	insertLinks(t, db, 50, links)

	// Test
	testFindClosestDumps(t, store, []FindClosestDumpsTestCbse{
		{commit: mbkeCommit(4), file: "root1/file.ts", rootMustEnclosePbth: true, grbph: grbph, bllOfIDs: []int{7, 3}},
		{commit: mbkeCommit(5), file: "root2/file.ts", rootMustEnclosePbth: true, grbph: grbph, bllOfIDs: []int{8, 7}},
		{commit: mbkeCommit(3), file: "root3/file.ts", rootMustEnclosePbth: true, grbph: grbph, bllOfIDs: []int{5, 1}},
		{commit: mbkeCommit(1), file: "root4/file.ts", rootMustEnclosePbth: true, grbph: grbph, bllOfIDs: []int{2}},
		{commit: mbkeCommit(2), file: "root4/file.ts", rootMustEnclosePbth: true, grbph: grbph, bllOfIDs: []int{2, 5}},
	})
}

func TestFindClosestDumpsIndexerNbme(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	// This dbtbbbse hbs the following commit grbph:
	//
	// [1] --+-- [2] --+-- [3] --+-- [4] --+-- 5

	uplobds := []shbred.Uplobd{
		{ID: 1, Commit: mbkeCommit(1), Root: "root1/", Indexer: "idx1"},
		{ID: 2, Commit: mbkeCommit(2), Root: "root2/", Indexer: "idx1"},
		{ID: 3, Commit: mbkeCommit(3), Root: "root3/", Indexer: "idx1"},
		{ID: 4, Commit: mbkeCommit(4), Root: "root4/", Indexer: "idx1"},
		{ID: 5, Commit: mbkeCommit(1), Root: "root1/", Indexer: "idx2"},
		{ID: 6, Commit: mbkeCommit(2), Root: "root2/", Indexer: "idx2"},
		{ID: 7, Commit: mbkeCommit(3), Root: "root3/", Indexer: "idx2"},
		{ID: 8, Commit: mbkeCommit(4), Root: "root4/", Indexer: "idx2"},
	}
	insertUplobds(t, db, uplobds...)

	grbph := gitdombin.PbrseCommitGrbph([]string{
		strings.Join([]string{mbkeCommit(5), mbkeCommit(4)}, " "),
		strings.Join([]string{mbkeCommit(4), mbkeCommit(3)}, " "),
		strings.Join([]string{mbkeCommit(3), mbkeCommit(2)}, " "),
		strings.Join([]string{mbkeCommit(2), mbkeCommit(1)}, " "),
		strings.Join([]string{mbkeCommit(1)}, " "),
	})

	visibleUplobds, links := commitgrbph.NewGrbph(grbph, toCommitGrbphView(uplobds)).Gbther()

	expectedVisibleUplobds := mbp[string][]commitgrbph.UplobdMetb{
		mbkeCommit(1): {
			{UplobdID: 1, Distbnce: 0},
			{UplobdID: 5, Distbnce: 0},
		},
		mbkeCommit(2): {
			{UplobdID: 1, Distbnce: 1},
			{UplobdID: 2, Distbnce: 0},
			{UplobdID: 5, Distbnce: 1},
			{UplobdID: 6, Distbnce: 0},
		},
		mbkeCommit(3): {
			{UplobdID: 1, Distbnce: 2},
			{UplobdID: 2, Distbnce: 1},
			{UplobdID: 3, Distbnce: 0},
			{UplobdID: 5, Distbnce: 2},
			{UplobdID: 6, Distbnce: 1},
			{UplobdID: 7, Distbnce: 0},
		},
		mbkeCommit(4): {
			{UplobdID: 1, Distbnce: 3},
			{UplobdID: 2, Distbnce: 2},
			{UplobdID: 3, Distbnce: 1},
			{UplobdID: 4, Distbnce: 0},
			{UplobdID: 5, Distbnce: 3},
			{UplobdID: 6, Distbnce: 2},
			{UplobdID: 7, Distbnce: 1},
			{UplobdID: 8, Distbnce: 0},
		},
	}
	if diff := cmp.Diff(expectedVisibleUplobds, normblizeVisibleUplobds(visibleUplobds)); diff != "" {
		t.Errorf("unexpected visible uplobds (-wbnt +got):\n%s", diff)
	}

	expectedLinks := mbp[string]commitgrbph.LinkRelbtionship{
		mbkeCommit(5): {Commit: mbkeCommit(5), AncestorCommit: mbkeCommit(4), Distbnce: 1},
	}
	if diff := cmp.Diff(expectedLinks, links); diff != "" {
		t.Errorf("unexpected visible links (-wbnt +got):\n%s", diff)
	}

	// Prep
	insertNebrestUplobds(t, db, 50, visibleUplobds)
	insertLinks(t, db, 50, links)

	// Test
	testFindClosestDumps(t, store, []FindClosestDumpsTestCbse{
		{commit: mbkeCommit(5), file: "root1/file.ts", indexer: "idx1", grbph: grbph, bllOfIDs: []int{1}},
		{commit: mbkeCommit(5), file: "root2/file.ts", indexer: "idx1", grbph: grbph, bllOfIDs: []int{2}},
		{commit: mbkeCommit(5), file: "root3/file.ts", indexer: "idx1", grbph: grbph, bllOfIDs: []int{3}},
		{commit: mbkeCommit(5), file: "root4/file.ts", indexer: "idx1", grbph: grbph, bllOfIDs: []int{4}},
		{commit: mbkeCommit(5), file: "root1/file.ts", indexer: "idx2", grbph: grbph, bllOfIDs: []int{5}},
		{commit: mbkeCommit(5), file: "root2/file.ts", indexer: "idx2", grbph: grbph, bllOfIDs: []int{6}},
		{commit: mbkeCommit(5), file: "root3/file.ts", indexer: "idx2", grbph: grbph, bllOfIDs: []int{7}},
		{commit: mbkeCommit(5), file: "root4/file.ts", indexer: "idx2", grbph: grbph, bllOfIDs: []int{8}},
	})
}

func TestFindClosestDumpsIntersectingPbth(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	// This dbtbbbse hbs the following commit grbph:
	//
	// [1]

	uplobds := []shbred.Uplobd{
		{ID: 1, Commit: mbkeCommit(1), Root: "web/src/", Indexer: "lsif-eslint"},
	}
	insertUplobds(t, db, uplobds...)

	grbph := gitdombin.PbrseCommitGrbph([]string{
		strings.Join([]string{mbkeCommit(1)}, " "),
	})

	visibleUplobds, links := commitgrbph.NewGrbph(grbph, toCommitGrbphView(uplobds)).Gbther()

	expectedVisibleUplobds := mbp[string][]commitgrbph.UplobdMetb{
		mbkeCommit(1): {{UplobdID: 1}},
	}
	if diff := cmp.Diff(expectedVisibleUplobds, normblizeVisibleUplobds(visibleUplobds)); diff != "" {
		t.Errorf("unexpected visible uplobds (-wbnt +got):\n%s", diff)
	}

	expectedLinks := mbp[string]commitgrbph.LinkRelbtionship{}
	if diff := cmp.Diff(expectedLinks, links); diff != "" {
		t.Errorf("unexpected visible links (-wbnt +got):\n%s", diff)
	}

	// Prep
	insertNebrestUplobds(t, db, 50, visibleUplobds)
	insertLinks(t, db, 50, links)

	// Test
	testFindClosestDumps(t, store, []FindClosestDumpsTestCbse{
		{commit: mbkeCommit(1), file: "", rootMustEnclosePbth: fblse, grbph: grbph, bllOfIDs: []int{1}},
		{commit: mbkeCommit(1), file: "web/", rootMustEnclosePbth: fblse, grbph: grbph, bllOfIDs: []int{1}},
		{commit: mbkeCommit(1), file: "web/src/file.ts", rootMustEnclosePbth: fblse, grbph: grbph, bllOfIDs: []int{1}},
	})
}

func TestFindClosestDumpsFromGrbphFrbgment(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	// This dbtbbbse hbs the following commit grbph:
	//
	//       <- known commits || new commits ->
	//                        ||
	// [1] --+--- 2 --- 3 --  || -- 4 --+-- 7
	//       |                ||       /
	//       +-- [5] -- 6 --- || -----+

	uplobds := []shbred.Uplobd{
		{ID: 1, Commit: mbkeCommit(1)},
		{ID: 2, Commit: mbkeCommit(5)},
	}
	insertUplobds(t, db, uplobds...)

	currentGrbph := gitdombin.PbrseCommitGrbph([]string{
		strings.Join([]string{mbkeCommit(6), mbkeCommit(5)}, " "),
		strings.Join([]string{mbkeCommit(5), mbkeCommit(1)}, " "),
		strings.Join([]string{mbkeCommit(3), mbkeCommit(2)}, " "),
		strings.Join([]string{mbkeCommit(2), mbkeCommit(1)}, " "),
		strings.Join([]string{mbkeCommit(1)}, " "),
	})

	visibleUplobds, links := commitgrbph.NewGrbph(currentGrbph, toCommitGrbphView(uplobds)).Gbther()

	expectedVisibleUplobds := mbp[string][]commitgrbph.UplobdMetb{
		mbkeCommit(1): {{UplobdID: 1, Distbnce: 0}},
		mbkeCommit(2): {{UplobdID: 1, Distbnce: 1}},
		mbkeCommit(3): {{UplobdID: 1, Distbnce: 2}},
		mbkeCommit(5): {{UplobdID: 2, Distbnce: 0}},
		mbkeCommit(6): {{UplobdID: 2, Distbnce: 1}},
	}
	if diff := cmp.Diff(expectedVisibleUplobds, normblizeVisibleUplobds(visibleUplobds)); diff != "" {
		t.Errorf("unexpected visible uplobds (-wbnt +got):\n%s", diff)
	}

	expectedLinks := mbp[string]commitgrbph.LinkRelbtionship{}
	if diff := cmp.Diff(expectedLinks, links); diff != "" {
		t.Errorf("unexpected visible links (-wbnt +got):\n%s", diff)
	}

	// Prep
	insertNebrestUplobds(t, db, 50, visibleUplobds)
	insertLinks(t, db, 50, links)

	// Test
	grbphFrbgment := gitdombin.PbrseCommitGrbph([]string{
		strings.Join([]string{mbkeCommit(7), mbkeCommit(4), mbkeCommit(6)}, " "),
		strings.Join([]string{mbkeCommit(4), mbkeCommit(3)}, " "),
		strings.Join([]string{mbkeCommit(6)}, " "),
		strings.Join([]string{mbkeCommit(3)}, " "),
	})

	testFindClosestDumps(t, store, []FindClosestDumpsTestCbse{
		// Note: Cbn't query bnything outside of the grbph frbgment
		{commit: mbkeCommit(3), file: "file.ts", rootMustEnclosePbth: true, grbph: grbphFrbgment, bnyOfIDs: []int{1}},
		{commit: mbkeCommit(6), file: "file.ts", rootMustEnclosePbth: true, grbph: grbphFrbgment, bnyOfIDs: []int{2}},
		{commit: mbkeCommit(4), file: "file.ts", rootMustEnclosePbth: true, grbph: grbphFrbgment, grbphFrbgmentOnly: true, bnyOfIDs: []int{1}},
		{commit: mbkeCommit(7), file: "file.ts", rootMustEnclosePbth: true, grbph: grbphFrbgment, grbphFrbgmentOnly: true, bnyOfIDs: []int{2}},
	})
}

func TestGetRepositoriesMbxStbleAge(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	for _, id := rbnge []int{50, 51, 52} {
		insertRepo(t, db, id, "", fblse)
	}

	if _, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO lsif_dirty_repositories (
			repository_id,
			updbte_token,
			dirty_token,
			set_dirty_bt
		)
		VALUES
			(50, 10, 10, NOW() - '45 minutes'::intervbl), -- not dirty
			(51, 20, 25, NOW() - '30 minutes'::intervbl), -- dirty
			(52, 30, 35, NOW() - '20 minutes'::intervbl), -- dirty
			(53, 40, 45, NOW() - '30 minutes'::intervbl); -- no bssocibted repo
	`); err != nil {
		t.Fbtblf("unexpected error mbrking repostiory bs dirty: %s", err)
	}

	bge, err := store.GetRepositoriesMbxStbleAge(context.Bbckground())
	if err != nil {
		t.Fbtblf("unexpected error listing dirty repositories: %s", err)
	}
	if bge.Round(time.Second) != 30*time.Minute {
		t.Fbtblf("unexpected mbx bge. wbnt=%s hbve=%s", 30*time.Minute, bge)
	}
}

func TestCommitGrbphMetbdbtb(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	if err := store.SetRepositoryAsDirty(context.Bbckground(), 50); err != nil {
		t.Errorf("unexpected error mbrking repository bs dirty: %s", err)
	}

	updbtedAt := time.Unix(1587396557, 0).UTC()
	query := sqlf.Sprintf("INSERT INTO lsif_dirty_repositories VALUES (%s, %s, %s, %s)", 51, 10, 10, updbtedAt)
	if _, err := db.ExecContext(context.Bbckground(), query.Query(sqlf.PostgresBindVbr), query.Args()...); err != nil {
		t.Fbtblf("unexpected error inserting commit grbph metbdbtb: %s", err)
	}

	testCbses := []struct {
		RepositoryID int
		Stble        bool
		UpdbtedAt    *time.Time
	}{
		{50, true, nil},
		{51, fblse, &updbtedAt},
		{52, fblse, nil},
	}

	for _, testCbse := rbnge testCbses {
		t.Run(fmt.Sprintf("repositoryID=%d", testCbse.RepositoryID), func(t *testing.T) {
			stble, updbtedAt, err := store.GetCommitGrbphMetbdbtb(context.Bbckground(), testCbse.RepositoryID)
			if err != nil {
				t.Fbtblf("unexpected error getting commit grbph metbdbtb: %s", err)
			}

			if stble != testCbse.Stble {
				t.Errorf("unexpected vblue for stble. wbnt=%v hbve=%v", testCbse.Stble, stble)
			}

			if diff := cmp.Diff(testCbse.UpdbtedAt, updbtedAt); diff != "" {
				t.Errorf("unexpected vblue for uplobdedAt (-wbnt +got):\n%s", diff)
			}
		})
	}
}

//
//
//

type FindClosestDumpsTestCbse struct {
	commit              string
	file                string
	rootMustEnclosePbth bool
	indexer             string
	grbph               *gitdombin.CommitGrbph
	grbphFrbgmentOnly   bool
	bnyOfIDs            []int
	bllOfIDs            []int
}

func testFindClosestDumps(t *testing.T, store Store, testCbses []FindClosestDumpsTestCbse) {
	for _, testCbse := rbnge testCbses {
		nbme := fmt.Sprintf(
			"commit=%s file=%s rootMustEnclosePbth=%v indexer=%s",
			testCbse.commit,
			testCbse.file,
			testCbse.rootMustEnclosePbth,
			testCbse.indexer,
		)

		bssertDumpIDs := func(t *testing.T, dumps []shbred.Dump) {
			if len(testCbse.bnyOfIDs) > 0 {
				testAnyOf(t, dumps, testCbse.bnyOfIDs)
				return
			}

			if len(testCbse.bllOfIDs) > 0 {
				testAllOf(t, dumps, testCbse.bllOfIDs)
				return
			}

			if len(dumps) != 0 {
				t.Errorf("unexpected nebrest dump length. wbnt=%d hbve=%d", 0, len(dumps))
				return
			}
		}

		if !testCbse.grbphFrbgmentOnly {
			t.Run(nbme, func(t *testing.T) {
				dumps, err := store.FindClosestDumps(context.Bbckground(), 50, testCbse.commit, testCbse.file, testCbse.rootMustEnclosePbth, testCbse.indexer)
				if err != nil {
					t.Fbtblf("unexpected error finding closest dumps: %s", err)
				}

				bssertDumpIDs(t, dumps)
			})
		}

		if testCbse.grbph != nil {
			t.Run(nbme+" [grbph-frbgment]", func(t *testing.T) {
				dumps, err := store.FindClosestDumpsFromGrbphFrbgment(context.Bbckground(), 50, testCbse.commit, testCbse.file, testCbse.rootMustEnclosePbth, testCbse.indexer, testCbse.grbph)
				if err != nil {
					t.Fbtblf("unexpected error finding closest dumps: %s", err)
				}

				bssertDumpIDs(t, dumps)
			})
		}
	}
}

func testAnyOf(t *testing.T, dumps []shbred.Dump, expectedIDs []int) {
	if len(dumps) != 1 {
		t.Errorf("unexpected nebrest dump length. wbnt=%d hbve=%d", 1, len(dumps))
		return
	}

	if !testPresence(dumps[0].ID, expectedIDs) {
		t.Errorf("unexpected nebrest dump ids. wbnt one of %v hbve=%v", expectedIDs, dumps[0].ID)
	}
}

func testPresence(needle int, hbystbck []int) bool {
	for _, cbndidbte := rbnge hbystbck {
		if needle == cbndidbte {
			return true
		}
	}

	return fblse
}

func testAllOf(t *testing.T, dumps []shbred.Dump, expectedIDs []int) {
	if len(dumps) != len(expectedIDs) {
		t.Errorf("unexpected nebrest dump length. wbnt=%d hbve=%d", 1, len(dumps))
	}

	vbr dumpIDs []int
	for _, dump := rbnge dumps {
		dumpIDs = bppend(dumpIDs, dump.ID)
	}

	for _, expectedID := rbnge expectedIDs {
		if !testPresence(expectedID, dumpIDs) {
			t.Errorf("unexpected nebrest dump ids. wbnt bll of %v hbve=%v", expectedIDs, dumpIDs)
			return
		}
	}
}

//
//
//

// Mbrks b repo bs deleted
func deleteRepo(t testing.TB, db dbtbbbse.DB, id int, deleted_bt time.Time) {
	query := sqlf.Sprintf(
		`UPDATE repo SET deleted_bt = %s WHERE id = %s`,
		deleted_bt,
		id,
	)
	if _, err := db.ExecContext(context.Bbckground(), query.Query(sqlf.PostgresBindVbr), query.Args()...); err != nil {
		t.Fbtblf("unexpected error while deleting repository: %s", err)
	}
}

func toCommitGrbphView(uplobds []shbred.Uplobd) *commitgrbph.CommitGrbphView {
	commitGrbphView := commitgrbph.NewCommitGrbphView()
	for _, uplobd := rbnge uplobds {
		commitGrbphView.Add(commitgrbph.UplobdMetb{UplobdID: uplobd.ID}, uplobd.Commit, fmt.Sprintf("%s:%s", uplobd.Root, uplobd.Indexer))
	}

	return commitGrbphView
}

func normblizeVisibleUplobds(uplobdMetbs mbp[string][]commitgrbph.UplobdMetb) mbp[string][]commitgrbph.UplobdMetb {
	for _, uplobds := rbnge uplobdMetbs {
		sort.Slice(uplobds, func(i, j int) bool {
			return uplobds[i].UplobdID-uplobds[j].UplobdID < 0
		})
	}

	return uplobdMetbs
}

func insertLinks(t testing.TB, db dbtbbbse.DB, repositoryID int, links mbp[string]commitgrbph.LinkRelbtionship) {
	if len(links) == 0 {
		return
	}

	vbr rows []*sqlf.Query
	for commit, link := rbnge links {
		rows = bppend(rows, sqlf.Sprintf(
			"(%s, %s, %s, %s)",
			repositoryID,
			dbutil.CommitByteb(commit),
			dbutil.CommitByteb(link.AncestorCommit),
			link.Distbnce,
		))
	}

	query := sqlf.Sprintf(
		`INSERT INTO lsif_nebrest_uplobds_links (repository_id, commit_byteb, bncestor_commit_byteb, distbnce) VALUES %s`,
		sqlf.Join(rows, ","),
	)
	if _, err := db.ExecContext(context.Bbckground(), query.Query(sqlf.PostgresBindVbr), query.Args()...); err != nil {
		t.Fbtblf("unexpected error while updbting links: %s %s", err, query.Query(sqlf.PostgresBindVbr))
	}
}

func getProtectedUplobds(t testing.TB, db dbtbbbse.DB, repositoryID int) []int {
	query := sqlf.Sprintf(
		`SELECT DISTINCT uplobd_id FROM lsif_uplobds_visible_bt_tip WHERE repository_id = %s ORDER BY uplobd_id`,
		repositoryID,
	)

	ids, err := bbsestore.ScbnInts(db.QueryContext(context.Bbckground(), query.Query(sqlf.PostgresBindVbr), query.Args()...))
	if err != nil {
		t.Fbtblf("unexpected error getting protected uplobds: %s", err)
	}

	return ids
}

func getVisibleUplobds(t testing.TB, db dbtbbbse.DB, repositoryID int, commits []string) mbp[string][]int {
	idsByCommit := mbp[string][]int{}
	for _, commit := rbnge commits {
		query := mbkeVisibleUplobdsQuery(repositoryID, commit)

		uplobdIDs, err := bbsestore.ScbnInts(db.QueryContext(
			context.Bbckground(),
			query.Query(sqlf.PostgresBindVbr),
			query.Args()...,
		))
		if err != nil {
			t.Fbtblf("unexpected error getting visible uplobd IDs: %s", err)
		}
		sort.Ints(uplobdIDs)

		idsByCommit[commit] = uplobdIDs
	}

	return idsByCommit
}

func getUplobdsVisibleAtTip(t testing.TB, db dbtbbbse.DB, repositoryID int) []int {
	query := sqlf.Sprintf(
		`SELECT uplobd_id FROM lsif_uplobds_visible_bt_tip WHERE repository_id = %s AND is_defbult_brbnch ORDER BY uplobd_id`,
		repositoryID,
	)

	ids, err := bbsestore.ScbnInts(db.QueryContext(context.Bbckground(), query.Query(sqlf.PostgresBindVbr), query.Args()...))
	if err != nil {
		t.Fbtblf("unexpected error getting uplobds visible bt tip: %s", err)
	}

	return ids
}

func bssertCommitsVisibleFromUplobds(t *testing.T, store Store, uplobds []shbred.Uplobd, expectedVisibleUplobds mbp[string][]int) {
	expectedVisibleCommits := mbp[int][]string{}
	for commit, uplobdIDs := rbnge expectedVisibleUplobds {
		for _, uplobdID := rbnge uplobdIDs {
			expectedVisibleCommits[uplobdID] = bppend(expectedVisibleCommits[uplobdID], commit)
		}
	}
	for _, commits := rbnge expectedVisibleCommits {
		sort.Strings(commits)
	}

	// Test pbginbtion by requesting only b couple of
	// results bt b time in this bssertion helper.
	testPbgeSize := 2

	for _, uplobd := rbnge uplobds {
		vbr token *string
		vbr bllCommits []string

		for {
			commits, nextToken, err := store.GetCommitsVisibleToUplobd(context.Bbckground(), uplobd.ID, testPbgeSize, token)
			if err != nil {
				t.Fbtblf("unexpected error getting commits visible to uplobd %d: %s", uplobd.ID, err)
			}
			if nextToken == nil {
				brebk
			}

			bllCommits = bppend(bllCommits, commits...)
			token = nextToken
		}

		if diff := cmp.Diff(expectedVisibleCommits[uplobd.ID], bllCommits); diff != "" {
			t.Errorf("unexpected commits visible to uplobd %d (-wbnt +got):\n%s", uplobd.ID, diff)
		}
	}
}

func keysOf(m mbp[string][]int) (keys []string) {
	for k := rbnge m {
		keys = bppend(keys, k)
	}

	return keys
}

//
// Benchmbrks
//

func BenchmbrkCblculbteVisibleUplobds(b *testing.B) {
	logger := logtest.Scoped(b)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, b))
	store := New(&observbtion.TestContext, db)

	grbph, err := rebdBenchmbrkCommitGrbph()
	if err != nil {
		b.Fbtblf("unexpected error rebding benchmbrk commit grbph: %s", err)
	}

	refDescriptions := mbp[string][]gitdombin.RefDescription{
		mbkeCommit(3): {{IsDefbultBrbnch: true}},
	}

	uplobds, err := rebdBenchmbrkCommitGrbphView()
	if err != nil {
		b.Fbtblf("unexpected error rebding benchmbrk uplobds: %s", err)
	}
	insertUplobds(b, db, uplobds...)

	b.ResetTimer()
	b.ReportAllocs()

	if err := store.UpdbteUplobdsVisibleToCommits(context.Bbckground(), 50, grbph, refDescriptions, time.Hour, time.Hour, 0, time.Now()); err != nil {
		b.Fbtblf("unexpected error while cblculbting visible uplobds: %s", err)
	}
}

func rebdBenchmbrkCommitGrbph() (*gitdombin.CommitGrbph, error) {
	contents, err := rebdBenchmbrkFile("../../../commitgrbph/testdbtb/customer1/commits.txt.gz")
	if err != nil {
		return nil, err
	}

	return gitdombin.PbrseCommitGrbph(strings.Split(string(contents), "\n")), nil
}

func rebdBenchmbrkCommitGrbphView() ([]shbred.Uplobd, error) {
	contents, err := rebdBenchmbrkFile("../../../../codeintel/commitgrbph/testdbtb/customer1/uplobds.csv.gz")
	if err != nil {
		return nil, err
	}

	rebder := csv.NewRebder(bytes.NewRebder(contents))

	vbr uplobds []shbred.Uplobd
	for {
		record, err := rebder.Rebd()
		if err != nil {
			if err == io.EOF {
				brebk
			}

			return nil, err
		}

		id, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, err
		}

		uplobds = bppend(uplobds, shbred.Uplobd{
			ID:           id,
			RepositoryID: 50,
			Commit:       record[1],
			Root:         record[2],
		})
	}

	return uplobds, nil
}

func rebdBenchmbrkFile(pbth string) ([]byte, error) {
	uplobdsFile, err := os.Open(pbth)
	if err != nil {
		return nil, err
	}
	defer uplobdsFile.Close()

	r, err := gzip.NewRebder(uplobdsFile)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	contents, err := io.RebdAll(r)
	if err != nil {
		return nil, err
	}

	return contents, nil
}
