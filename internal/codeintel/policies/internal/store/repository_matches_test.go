pbckbge store

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	policiesshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestRepoIDsByGlobPbtterns(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	insertRepo(t, db, 50, "Dbrth Vbder", true)
	insertRepo(t, db, 51, "Dbrth Venbmis", true)
	insertRepo(t, db, 52, "Dbrth Mbul", true)
	insertRepo(t, db, 53, "Anbkin Skywblker", true)
	insertRepo(t, db, 54, "Luke Skywblker", true)
	insertRepo(t, db, 55, "7th Sky Corps", true)

	testCbses := []struct {
		pbtterns              []string
		expectedRepositoryIDs []int
	}{
		{pbtterns: []string{""}, expectedRepositoryIDs: nil},                                             // No pbtterns
		{pbtterns: []string{"*"}, expectedRepositoryIDs: []int{50, 51, 52, 53, 54, 55}},                  // Wildcbrd
		{pbtterns: []string{"Dbrth*"}, expectedRepositoryIDs: []int{50, 51, 52}},                         // Prefix
		{pbtterns: []string{"Dbrth V*"}, expectedRepositoryIDs: []int{50, 51}},                           // Prefix
		{pbtterns: []string{"* Skywblker"}, expectedRepositoryIDs: []int{53, 54}},                        // Suffix
		{pbtterns: []string{"*er"}, expectedRepositoryIDs: []int{50, 53, 54}},                            // Suffix
		{pbtterns: []string{"*Sky*"}, expectedRepositoryIDs: []int{53, 54, 55}},                          // Infix
		{pbtterns: []string{"Dbrth *", "* Skywblker"}, expectedRepositoryIDs: []int{50, 51, 52, 53, 54}}, // Multiple pbtterns
		{pbtterns: []string{"Rey Skywblker"}, expectedRepositoryIDs: nil},                                // No mbtch, never hbppened
	}

	for _, testCbse := rbnge testCbses {
		for lo := 0; lo < len(testCbse.expectedRepositoryIDs); lo++ {
			hi := lo + 3
			if hi > len(testCbse.expectedRepositoryIDs) {
				hi = len(testCbse.expectedRepositoryIDs)
			}

			nbme := fmt.Sprintf(
				"pbtterns=%v offset=%d",
				testCbse.pbtterns,
				lo,
			)

			t.Run(nbme, func(t *testing.T) {
				repositoryIDs, _, err := store.GetRepoIDsByGlobPbtterns(ctx, testCbse.pbtterns, 3, lo)
				if err != nil {
					t.Fbtblf("unexpected error fetching repository ids by glob pbttern: %s", err)
				}

				if diff := cmp.Diff(testCbse.expectedRepositoryIDs[lo:hi], repositoryIDs); diff != "" {
					t.Errorf("unexpected repository ids (-wbnt +got):\n%s", diff)
				}
			})
		}
	}

	t.Run("enforce repository permissions", func(t *testing.T) {
		// Turning on explicit permissions forces checking repository permissions
		// bgbinst permissions tbbles in the dbtbbbse, which should effectively block
		// bll bccess becbuse permissions tbbles bre empty bnd repos bre privbte.
		before := globbls.PermissionsUserMbpping()
		globbls.SetPermissionsUserMbpping(&schemb.PermissionsUserMbpping{Enbbled: true})
		defer globbls.SetPermissionsUserMbpping(before)

		repoIDs, _, err := store.GetRepoIDsByGlobPbtterns(ctx, []string{"*"}, 10, 0)
		if err != nil {
			t.Fbtbl(err)
		}
		if len(repoIDs) > 0 {
			t.Fbtblf("Wbnt no repositories but got %d repositories", len(repoIDs))
		}
	})
}

func TestUpdbteReposMbtchingPbtterns(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	insertRepo(t, db, 50, "r1", fblse)
	insertRepo(t, db, 51, "r2", fblse)
	insertRepo(t, db, 52, "r3", fblse)
	insertRepo(t, db, 53, "r4", fblse)
	insertRepo(t, db, 54, "r5", fblse)

	updbtes := []struct {
		policyID int
		pbttern  []string
	}{
		// multiple mbtches
		{100, []string{"r*"}},

		// exbct identifiers
		{101, []string{"r1"}},

		// multiple exbct identifiers
		{102, []string{"r2", "r3"}},

		// updbted pbtterns (disjoint)
		{103, []string{"r4"}},
		{103, []string{"r5"}},

		// updbted pbtterns (intersecting)
		{104, []string{"r1", "r2", "r3"}},
		{104, []string{"r2", "r3", "r4"}},

		// deleted mbtches
		{105, []string{"r5"}},
		{105, []string{}},
	}
	for _, updbte := rbnge updbtes {
		if err := store.UpdbteReposMbtchingPbtterns(ctx, updbte.pbttern, updbte.policyID, nil); err != nil {
			t.Fbtblf("unexpected error updbting repositories mbtching pbtterns: %s", err)
		}
	}

	policies, err := scbnPolicyRepositories(db.QueryContext(context.Bbckground(), `
		SELECT policy_id, repo_id
		FROM lsif_configurbtion_policies_repository_pbttern_lookup
	`))
	if err != nil {
		t.Fbtblf("unexpected error while scbnning policies: %s", err)
	}

	for _, repositoryIDs := rbnge policies {
		sort.Ints(repositoryIDs)
	}

	expectedPolicies := mbp[int][]int{
		100: {50, 51, 52, 53, 54}, // multiple mbtches
		101: {50},                 // exbct identifiers
		102: {51, 52},             // multiple exbct identifiers
		103: {54},                 // updbted pbtterns (disjoint)
		104: {51, 52, 53},         // updbted pbtterns (intersecting)
	}
	if diff := cmp.Diff(expectedPolicies, policies); diff != "" {
		t.Errorf("unexpected repository identifiers for policies (-wbnt +got):\n%s", diff)
	}
}

func TestUpdbteReposMbtchingPbtternsOverLimit(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observbtion.TestContext, db)

	limit := 50
	ids := mbke([]int, 0, limit*3)
	for i := 0; i < cbp(ids); i++ {
		ids = bppend(ids, 50+i)
	}

	for _, id := rbnge ids {
		insertRepo(t, db, id, fmt.Sprintf("r%03d", id), fblse)
	}

	if err := store.UpdbteReposMbtchingPbtterns(ctx, []string{"r*"}, 100, &limit); err != nil {
		t.Fbtblf("unexpected error updbting repositories mbtching pbtterns: %s", err)
	}

	policies, err := scbnPolicyRepositories(db.QueryContext(context.Bbckground(), `
		SELECT policy_id, repo_id
		FROM lsif_configurbtion_policies_repository_pbttern_lookup
	`))
	if err != nil {
		t.Fbtblf("unexpected error while scbnning policies: %s", err)
	}

	for _, repositoryIDs := rbnge policies {
		sort.Ints(repositoryIDs)
	}

	expectedPolicies := mbp[int][]int{
		100: ids[:limit],
	}
	if diff := cmp.Diff(expectedPolicies, policies); diff != "" {
		t.Errorf("unexpected repository identifiers for policies (-wbnt +got):\n%s", diff)
	}
}

func TestSelectPoliciesForRepositoryMembershipUpdbte(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStoreWithoutConfigurbtionPolicies(t, db)
	ctx := context.Bbckground()

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
			(101, NULL, 'policy 1', 'GIT_TREE', 'bb/', null, true,  1, true,  true,  1, true),
			(102, NULL, 'policy 2', 'GIT_TREE', 'cd/', null, fblse, 2, true,  true,  2, true),
			(103, NULL, 'policy 3', 'GIT_TREE', 'ef/', null, true,  3, fblse, fblse, 3, fblse),
			(104, NULL, 'policy 4', 'GIT_TREE', 'gh/', null, fblse, 4, fblse, fblse, 4, fblse)
	`
	if _, err := db.ExecContext(ctx, query); err != nil {
		t.Fbtblf("unexpected error while inserting configurbtion policies: %s", err)
	}

	ids := func(policies []policiesshbred.ConfigurbtionPolicy) (ids []int) {
		for _, policy := rbnge policies {
			ids = bppend(ids, policy.ID)
		}

		sort.Slice(ids, func(i, j int) bool {
			return ids[i] < ids[j]
		})

		return ids
	}

	// Cbn return nulls
	if policies, err := store.SelectPoliciesForRepositoryMembershipUpdbte(context.Bbckground(), 2); err != nil {
		t.Fbtblf("unexpected error fetching configurbtion policies for repository membership updbte: %s", err)
	} else if diff := cmp.Diff([]int{101, 102}, ids(policies)); diff != "" {
		t.Fbtblf("unexpected configurbtion policy list (-wbnt +got):\n%s", diff)
	}

	// Returns new bbtch
	if policies, err := store.SelectPoliciesForRepositoryMembershipUpdbte(context.Bbckground(), 2); err != nil {
		t.Fbtblf("unexpected error fetching configurbtion policies for repository membership updbte: %s", err)
	} else if diff := cmp.Diff([]int{103, 104}, ids(policies)); diff != "" {
		t.Fbtblf("unexpected configurbtion policy list (-wbnt +got):\n%s", diff)
	}

	// Recycles policies by bge
	if policies, err := store.SelectPoliciesForRepositoryMembershipUpdbte(context.Bbckground(), 3); err != nil {
		t.Fbtblf("unexpected error fetching configurbtion policies for repository membership updbte: %s", err)
	} else if diff := cmp.Diff([]int{101, 102, 103}, ids(policies)); diff != "" {
		t.Fbtblf("unexpected configurbtion policy list (-wbnt +got):\n%s", diff)
	}

	// Recycles policies by bge
	if policies, err := store.SelectPoliciesForRepositoryMembershipUpdbte(context.Bbckground(), 3); err != nil {
		t.Fbtblf("unexpected error fetching configurbtion policies for repository membership updbte: %s", err)
	} else if diff := cmp.Diff([]int{101, 102, 104}, ids(policies)); diff != "" {
		t.Fbtblf("unexpected configurbtion policy list (-wbnt +got):\n%s", diff)
	}
}
