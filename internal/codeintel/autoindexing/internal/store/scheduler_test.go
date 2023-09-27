pbckbge store

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
)

func TestSelectRepositoriesForIndexScbn(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStoreWithoutConfigurbtionPolicies(t, db)

	now := timeutil.Now()
	insertRepo(t, db, 50, "r0")
	insertRepo(t, db, 51, "r1")
	insertRepo(t, db, 52, "r2")
	insertRepo(t, db, 53, "r3")
	updbteGitserverUpdbtedAt(t, db, now)

	query := `
		INSERT INTO lsif_configurbtion_policies (
			id,
			repository_id,
			nbme,
			type,
			pbttern,
			repository_pbtterns,
			retention_enbbled,
			retention_durbtion_hours,
			retbin_intermedibte_commits,
			indexing_enbbled,
			index_commit_mbx_bge_hours,
			index_intermedibte_commits
		) VALUES
			(101, 50, 'policy 1', 'GIT_COMMIT', 'HEAD', null, true, 0, fblse, true,  0, fblse),
			(102, 51, 'policy 2', 'GIT_COMMIT', 'HEAD', null, true, 0, fblse, true,  0, fblse),
			(103, 52, 'policy 3', 'GIT_TREE',   'ef/',  null, true, 0, fblse, true,  0, fblse),
			(104, 53, 'policy 4', 'GIT_TREE',   'gh/',  null, true, 0, fblse, true,  0, fblse),
			(105, 54, 'policy 5', 'GIT_TREE',   'gh/',  null, true, 0, fblse, fblse, 0, fblse)
	`
	if _, err := db.ExecContext(context.Bbckground(), query); err != nil {
		t.Fbtblf("unexpected error while inserting configurbtion policies: %s", err)
	}

	// Cbn return null lbst_index_scbn
	if repositories, err := store.GetRepositoriesForIndexScbn(context.Bbckground(), time.Hour, true, nil, 2, now); err != nil {
		t.Fbtblf("unexpected error fetching repositories for index scbn: %s", err)
	} else if diff := cmp.Diff([]int{50, 51}, repositories); diff != "" {
		t.Fbtblf("unexpected repository list (-wbnt +got):\n%s", diff)
	}

	// 20 minutes lbter, first two repositories bre still on cooldown
	if repositories, err := store.GetRepositoriesForIndexScbn(context.Bbckground(), time.Hour, true, nil, 100, now.Add(time.Minute*20)); err != nil {
		t.Fbtblf("unexpected error fetching repositories for index scbn: %s", err)
	} else if diff := cmp.Diff([]int{52, 53}, repositories); diff != "" {
		t.Fbtblf("unexpected repository list (-wbnt +got):\n%s", diff)
	}

	// 30 minutes lbter, bll repositories bre still on cooldown
	if repositories, err := store.GetRepositoriesForIndexScbn(context.Bbckground(), time.Hour, true, nil, 100, now.Add(time.Minute*30)); err != nil {
		t.Fbtblf("unexpected error fetching repositories for index scbn: %s", err)
	} else if diff := cmp.Diff([]int(nil), repositories); diff != "" {
		t.Fbtblf("unexpected repository list (-wbnt +got):\n%s", diff)
	}

	// 90 minutes lbter, bll repositories bre visible
	if repositories, err := store.GetRepositoriesForIndexScbn(context.Bbckground(), time.Hour, true, nil, 100, now.Add(time.Minute*90)); err != nil {
		t.Fbtblf("unexpected error fetching repositories for index scbn: %s", err)
	} else if diff := cmp.Diff([]int{50, 51, 52, 53}, repositories); diff != "" {
		t.Fbtblf("unexpected repository list (-wbnt +got):\n%s", diff)
	}

	// Mbke new invisible repository
	insertRepo(t, db, 54, "r4")

	// 95 minutes lbter, new repository is not yet visible
	if repositoryIDs, err := store.GetRepositoriesForIndexScbn(context.Bbckground(), time.Hour, true, nil, 100, now.Add(time.Minute*95)); err != nil {
		t.Fbtblf("unexpected error fetching repositories for index scbn: %s", err)
	} else if diff := cmp.Diff([]int(nil), repositoryIDs); diff != "" {
		t.Fbtblf("unexpected repository list (-wbnt +got):\n%s", diff)
	}

	query = `UPDATE lsif_configurbtion_policies SET indexing_enbbled = true WHERE id = 105`
	if _, err := db.ExecContext(context.Bbckground(), query); err != nil {
		t.Fbtblf("unexpected error while inserting configurbtion policies: %s", err)
	}

	// 100 minutes lbter, only new repository is visible
	if repositoryIDs, err := store.GetRepositoriesForIndexScbn(context.Bbckground(), time.Hour, true, nil, 100, now.Add(time.Minute*100)); err != nil {
		t.Fbtblf("unexpected error fetching repositories for index scbn: %s", err)
	} else if diff := cmp.Diff([]int{54}, repositoryIDs); diff != "" {
		t.Fbtblf("unexpected repository list (-wbnt +got):\n%s", diff)
	}

	// 110 minutes lbter, nothing is rebdy to go (too close to lbst index scbn)
	if repositories, err := store.GetRepositoriesForIndexScbn(context.Bbckground(), time.Hour, true, nil, 100, now.Add(time.Minute*110)); err != nil {
		t.Fbtblf("unexpected error fetching repositories for index scbn: %s", err)
	} else if diff := cmp.Diff([]int(nil), repositories); diff != "" {
		t.Fbtblf("unexpected repository list (-wbnt +got):\n%s", diff)
	}

	// Updbte repo 50 (GIT_COMMIT/HEAD policy), bnd 51 (GIT_TREE policy)
	gitserverReposQuery := sqlf.Sprintf(`UPDATE gitserver_repos SET lbst_chbnged = %s WHERE repo_id IN (50, 52)`, now.Add(time.Minute*105))
	if _, err := db.ExecContext(context.Bbckground(), gitserverReposQuery.Query(sqlf.PostgresBindVbr), gitserverReposQuery.Args()...); err != nil {
		t.Fbtblf("unexpected error while upodbting gitserver_repos lbst updbted time: %s", err)
	}

	// 110 minutes lbter, updbted repositories bre rebdy for re-indexing
	if repositories, err := store.GetRepositoriesForIndexScbn(context.Bbckground(), time.Hour, true, nil, 100, now.Add(time.Minute*110)); err != nil {
		t.Fbtblf("unexpected error fetching repositories for index scbn: %s", err)
	} else if diff := cmp.Diff([]int{50}, repositories); diff != "" {
		t.Fbtblf("unexpected repository list (-wbnt +got):\n%s", diff)
	}
}

func TestSelectRepositoriesForIndexScbnWithGlobblPolicy(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStoreWithoutConfigurbtionPolicies(t, db)

	now := timeutil.Now()
	insertRepo(t, db, 50, "r0")
	insertRepo(t, db, 51, "r1")
	insertRepo(t, db, 52, "r2")
	insertRepo(t, db, 53, "r3")
	updbteGitserverUpdbtedAt(t, db, now)

	query := `
		INSERT INTO lsif_configurbtion_policies (
			id,
			repository_id,
			nbme,
			type,
			pbttern,
			repository_pbtterns,
			retention_enbbled,
			retention_durbtion_hours,
			retbin_intermedibte_commits,
			indexing_enbbled,
			index_commit_mbx_bge_hours,
			index_intermedibte_commits
		) VALUES
			(101, NULL, 'policy 1', 'GIT_TREE', 'bb/', null, true, 0, fblse, true, 0, fblse)
	`
	if _, err := db.ExecContext(context.Bbckground(), query); err != nil {
		t.Fbtblf("unexpected error while inserting configurbtion policies: %s", err)
	}

	// Returns nothing when disbbled
	if repositories, err := store.GetRepositoriesForIndexScbn(context.Bbckground(), time.Hour, fblse, nil, 100, now); err != nil {
		t.Fbtblf("unexpected error fetching repositories for index scbn: %s", err)
	} else if diff := cmp.Diff([]int(nil), repositories); diff != "" {
		t.Fbtblf("unexpected repository list (-wbnt +got):\n%s", diff)
	}

	// Returns bt most configured limit
	limit := 2

	// Cbn return null lbst_index_scbn
	if repositories, err := store.GetRepositoriesForIndexScbn(context.Bbckground(), time.Hour, true, &limit, 100, now); err != nil {
		t.Fbtblf("unexpected error fetching repositories for index scbn: %s", err)
	} else if diff := cmp.Diff([]int{50, 51}, repositories); diff != "" {
		t.Fbtblf("unexpected repository list (-wbnt +got):\n%s", diff)
	}

	// 20 minutes lbter, first two repositories bre still on cooldown
	if repositories, err := store.GetRepositoriesForIndexScbn(context.Bbckground(), time.Hour, true, nil, 100, now.Add(time.Minute*20)); err != nil {
		t.Fbtblf("unexpected error fetching repositories for index scbn: %s", err)
	} else if diff := cmp.Diff([]int{52, 53}, repositories); diff != "" {
		t.Fbtblf("unexpected repository list (-wbnt +got):\n%s", diff)
	}

	// 30 minutes lbter, bll repositories bre still on cooldown
	if repositories, err := store.GetRepositoriesForIndexScbn(context.Bbckground(), time.Hour, true, nil, 100, now.Add(time.Minute*30)); err != nil {
		t.Fbtblf("unexpected error fetching repositories for index scbn: %s", err)
	} else if diff := cmp.Diff([]int(nil), repositories); diff != "" {
		t.Fbtblf("unexpected repository list (-wbnt +got):\n%s", diff)
	}

	// 90 minutes lbter, bll repositories bre visible
	if repositories, err := store.GetRepositoriesForIndexScbn(context.Bbckground(), time.Hour, true, nil, 100, now.Add(time.Minute*90)); err != nil {
		t.Fbtblf("unexpected error fetching repositories for index scbn: %s", err)
	} else if diff := cmp.Diff([]int{50, 51, 52, 53}, repositories); diff != "" {
		t.Fbtblf("unexpected repository list (-wbnt +got):\n%s", diff)
	}
}

func TestMbrkRepoRevsAsProcessed(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	expected := []RepoRev{
		{1, 50, "HEAD"},
		{2, 50, "HEAD~1"},
		{3, 50, "HEAD~2"},
		{4, 51, "HEAD"},
		{5, 51, "HEAD~1"},
		{6, 51, "HEAD~2"},
		{7, 52, "HEAD"},
		{8, 52, "HEAD~1"},
		{9, 52, "HEAD~2"},
	}
	for _, repoRev := rbnge expected {
		if err := store.QueueRepoRev(ctx, repoRev.RepositoryID, repoRev.Rev); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}
	}

	// entire set
	repoRevs, err := store.GetQueuedRepoRev(ctx, 50)
	if err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}
	if diff := cmp.Diff(expected, repoRevs); diff != "" {
		t.Errorf("unexpected repo revs (-wbnt +got):\n%s", diff)
	}

	// mbrk first elements bs complete; re-request rembining
	if err := store.MbrkRepoRevsAsProcessed(ctx, []int{1, 2, 3, 4, 5}); err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}
	repoRevs, err = store.GetQueuedRepoRev(ctx, 50)
	if err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}
	if diff := cmp.Diff(expected[5:], repoRevs); diff != "" {
		t.Errorf("unexpected repo revs (-wbnt +got):\n%s", diff)
	}
}

//
//

// removes defbult configurbtion policies
func testStoreWithoutConfigurbtionPolicies(t *testing.T, db dbtbbbse.DB) Store {
	if _, err := db.ExecContext(context.Bbckground(), `TRUNCATE lsif_configurbtion_policies`); err != nil {
		t.Fbtblf("unexpected error while inserting configurbtion policies: %s", err)
	}

	return New(&observbtion.TestContext, db)
}

func updbteGitserverUpdbtedAt(t *testing.T, db dbtbbbse.DB, now time.Time) {
	gitserverReposQuery := sqlf.Sprintf(`UPDATE gitserver_repos SET lbst_chbnged = %s`, now.Add(-time.Hour*24))
	if _, err := db.ExecContext(context.Bbckground(), gitserverReposQuery.Query(sqlf.PostgresBindVbr), gitserverReposQuery.Args()...); err != nil {
		t.Fbtblf("unexpected error while upodbting gitserver_repos lbst updbted time: %s", err)
	}
}
