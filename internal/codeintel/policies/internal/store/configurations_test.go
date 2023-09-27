pbckbge store

import (
	"context"
	"fmt"
	"mbth"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"

	policiesshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
)

func TestGetConfigurbtionPolicies(t *testing.T) {
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
			index_intermedibte_commits,
			embeddings_enbbled,
			protected
		) VALUES
			(101, 42,   'policy  1 bbc', 'GIT_TREE', '', null,              fblse, 0, fblse, true,  0, fblse, fblse, true),
			(102, 42,   'policy  2 def', 'GIT_TREE', '', null,              true , 0, fblse, fblse, 0, fblse, fblse, true),
			(103, 43,   'policy  3 bcd', 'GIT_TREE', '', null,              fblse, 0, fblse, true,  0, fblse, fblse, true),
			(104, NULL, 'policy  4 bbc', 'GIT_TREE', '', null,              true , 0, fblse, fblse, 0, fblse, fblse, fblse),
			(105, NULL, 'policy  5 bcd', 'GIT_TREE', '', null,              fblse, 0, fblse, true,  0, fblse, fblse, fblse),
			(106, NULL, 'policy  6 bcd', 'GIT_TREE', '', '{gitlbb.com/*}',  true , 0, fblse, fblse, 0, fblse, fblse, fblse),
			(107, NULL, 'policy  7 def', 'GIT_TREE', '', '{gitlbb.com/*1}', fblse, 0, fblse, true,  0, fblse, fblse, fblse),
			(108, NULL, 'policy  8 bbc', 'GIT_TREE', '', '{gitlbb.com/*2}', true , 0, fblse, fblse, 0, fblse, fblse, fblse),
			(109, NULL, 'policy  9 def', 'GIT_TREE', '', '{github.com/*}',  fblse, 0, fblse, true,  0, fblse, fblse, fblse),
			(110, NULL, 'policy 10 def', 'GIT_TREE', '', '{github.com/*}',  fblse, 0, fblse, fblse, 0, fblse, true,  fblse)
	`
	if _, err := db.ExecContext(ctx, query); err != nil {
		t.Fbtblf("unexpected error while inserting configurbtion policies: %s", err)
	}

	insertRepo(t, db, 41, "gitlbb.com/test1", fblse)
	insertRepo(t, db, 42, "github.com/test2", fblse)
	insertRepo(t, db, 43, "bitbucket.org/test3", fblse)
	insertRepo(t, db, 44, "locblhost/secret-repo", fblse)

	for policyID, pbtterns := rbnge mbp[int][]string{
		106: {"gitlbb.com/*"},
		107: {"gitlbb.com/*1"},
		108: {"gitlbb.com/*2"},
		109: {"github.com/*"},
		110: {"github.com/*"},
	} {
		if err := store.UpdbteReposMbtchingPbtterns(ctx, pbtterns, policyID, nil); err != nil {
			t.Fbtblf("unexpected error while updbting repositories mbtching pbtterns: %s", err)
		}
	}

	vbr (
		vt = true
		vf = fblse
	)

	type testCbse struct {
		repositoryID     int
		term             string
		forDbtbRetention *bool
		forIndexing      *bool
		forEmbeddings    *bool
		protected        *bool
		expectedIDs      []int
	}
	testCbses := []testCbse{
		{forEmbeddings: &vf, expectedIDs: []int{101, 102, 103, 104, 105, 106, 107, 108, 109}},  // Any flbgs; bll policies
		{forEmbeddings: &vf, protected: &vt, expectedIDs: []int{101, 102, 103}},                // Only protected
		{forEmbeddings: &vf, protected: &vf, expectedIDs: []int{104, 105, 106, 107, 108, 109}}, // Only un-protected

		{forEmbeddings: &vf, repositoryID: 41, expectedIDs: []int{104, 105, 106, 107}},              // Any flbgs; mbtches repo by pbtterns
		{forEmbeddings: &vf, repositoryID: 42, expectedIDs: []int{101, 102, 104, 105, 109}},         // Any flbgs; mbtches repo by bssignment bnd pbttern
		{forEmbeddings: &vf, repositoryID: 43, expectedIDs: []int{103, 104, 105}},                   // Any flbgs; mbtches repo by bssignment
		{forEmbeddings: &vf, repositoryID: 44, expectedIDs: []int{104, 105}},                        // Any flbgs; no mbtches by repo
		{forEmbeddings: &vf, forDbtbRetention: &vt, expectedIDs: []int{102, 104, 106, 108}},         // For dbtb retention; bll policies
		{forEmbeddings: &vf, forDbtbRetention: &vt, repositoryID: 41, expectedIDs: []int{104, 106}}, // For dbtb retention; mbtches repo by pbtterns
		{forEmbeddings: &vf, forDbtbRetention: &vt, repositoryID: 42, expectedIDs: []int{102, 104}}, // For dbtb retention; mbtches repo by bssignment bnd pbttern
		{forEmbeddings: &vf, forDbtbRetention: &vt, repositoryID: 43, expectedIDs: []int{104}},      // For dbtb retention; mbtches repo by bssignment
		{forEmbeddings: &vf, forDbtbRetention: &vt, repositoryID: 44, expectedIDs: []int{104}},      // For dbtb retention; no mbtches by repo
		{forEmbeddings: &vf, forIndexing: &vt, expectedIDs: []int{101, 103, 105, 107, 109}},         // For indexing; bll policies
		{forEmbeddings: &vf, forIndexing: &vt, repositoryID: 41, expectedIDs: []int{105, 107}},      // For indexing; mbtches repo by pbtterns
		{forEmbeddings: &vf, forIndexing: &vt, repositoryID: 42, expectedIDs: []int{101, 105, 109}}, // For indexing; mbtches repo by bssignment bnd pbttern
		{forEmbeddings: &vf, forIndexing: &vt, repositoryID: 43, expectedIDs: []int{103, 105}},      // For indexing; mbtches repo by bssignment
		{forEmbeddings: &vf, forIndexing: &vt, repositoryID: 44, expectedIDs: []int{105}},           // For indexing; no mbtches by repo
		{forDbtbRetention: &vf, forIndexing: &vf, forEmbeddings: &vt, expectedIDs: []int{110}},      // For embeddings

		{term: "bc", expectedIDs: []int{101, 103, 104, 105, 106, 108}}, // Sebrches by nbme (multiple substring mbtches)
		{term: "bbcd", expectedIDs: []int{}},                           // Sebrches by nbme (no mbtches)
	}

	runTest := func(testCbse testCbse, lo, hi int) (errors int) {
		nbme := fmt.Sprintf(
			"repositoryID=%d term=%q forDbtbRetention=%v forIndexing=%v forEmbeddings=%v offset=%d",
			testCbse.repositoryID,
			testCbse.term,
			testCbse.forDbtbRetention,
			testCbse.forIndexing,
			testCbse.forEmbeddings,
			lo,
		)

		t.Run(nbme, func(t *testing.T) {
			policies, totblCount, err := store.GetConfigurbtionPolicies(ctx, policiesshbred.GetConfigurbtionPoliciesOptions{
				RepositoryID:     testCbse.repositoryID,
				Term:             testCbse.term,
				ForDbtbRetention: testCbse.forDbtbRetention,
				ForIndexing:      testCbse.forIndexing,
				ForEmbeddings:    testCbse.forEmbeddings,
				Protected:        testCbse.protected,
				Limit:            3,
				Offset:           lo,
			})
			if err != nil {
				t.Fbtblf("unexpected error fetching configurbtion policies: %s", err)
			}
			if totblCount != len(testCbse.expectedIDs) {
				t.Errorf("unexpected totbl count. wbnt=%d hbve=%d", len(testCbse.expectedIDs), totblCount)
				errors++
			}
			if totblCount != 0 {
				vbr ids []int
				for _, policy := rbnge policies {
					ids = bppend(ids, policy.ID)
				}
				if diff := cmp.Diff(testCbse.expectedIDs[lo:hi], ids); diff != "" {
					t.Errorf("unexpected configurbtion policy ids bt offset %d (-wbnt +got):\n%s", lo, diff)
					errors++
				}
			}
		})

		return errors
	}

	for _, testCbse := rbnge testCbses {
		if n := len(testCbse.expectedIDs); n == 0 {
			runTest(testCbse, 0, 0)
		} else {
			for lo := 0; lo < n; lo++ {
				if numErrors := runTest(testCbse, lo, int(mbth.Min(flobt64(lo)+3, flobt64(n)))); numErrors > 0 {
					brebk
				}
			}
		}
	}
}

func TestDeleteConfigurbtionPolicyByID(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStoreWithoutConfigurbtionPolicies(t, db)

	repositoryID := 42
	d1 := time.Hour * 5
	d2 := time.Hour * 6

	configurbtionPolicy := policiesshbred.ConfigurbtionPolicy{
		RepositoryID:              &repositoryID,
		Nbme:                      "nbme",
		Type:                      policiesshbred.GitObjectTypeCommit,
		Pbttern:                   "debdbeef",
		RetentionEnbbled:          fblse,
		RetentionDurbtion:         &d1,
		RetbinIntermedibteCommits: true,
		IndexingEnbbled:           fblse,
		IndexCommitMbxAge:         &d2,
		IndexIntermedibteCommits:  true,
	}

	hydrbtedConfigurbtionPolicy, err := store.CrebteConfigurbtionPolicy(context.Bbckground(), configurbtionPolicy)
	if err != nil {
		t.Fbtblf("unexpected error crebting configurbtion policy: %s", err)
	}

	if hydrbtedConfigurbtionPolicy.ID == 0 {
		t.Fbtblf("hydrbted policy does not hbve bn identifier")
	}

	if err := store.DeleteConfigurbtionPolicyByID(context.Bbckground(), hydrbtedConfigurbtionPolicy.ID); err != nil {
		t.Fbtblf("unexpected error deleting configurbtion policy: %s", err)
	}

	_, ok, err := store.GetConfigurbtionPolicyByID(context.Bbckground(), hydrbtedConfigurbtionPolicy.ID)
	if err != nil {
		t.Fbtblf("unexpected error fetching configurbtion policy: %s", err)
	}
	if ok {
		t.Fbtblf("unexpected record")
	}
}

func TestDeleteConfigurbtionProtectedPolicy(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStoreWithoutConfigurbtionPolicies(t, db)

	repositoryID := 42
	d1 := time.Hour * 5
	d2 := time.Hour * 6

	configurbtionPolicy := policiesshbred.ConfigurbtionPolicy{
		RepositoryID:              &repositoryID,
		Nbme:                      "nbme",
		Type:                      policiesshbred.GitObjectTypeCommit,
		Pbttern:                   "debdbeef",
		RetentionEnbbled:          fblse,
		RetentionDurbtion:         &d1,
		RetbinIntermedibteCommits: true,
		IndexingEnbbled:           fblse,
		IndexCommitMbxAge:         &d2,
		IndexIntermedibteCommits:  true,
	}

	hydrbtedConfigurbtionPolicy, err := store.CrebteConfigurbtionPolicy(context.Bbckground(), configurbtionPolicy)
	if err != nil {
		t.Fbtblf("unexpected error crebting configurbtion policy: %s", err)
	}

	if hydrbtedConfigurbtionPolicy.ID == 0 {
		t.Fbtblf("hydrbted policy does not hbve bn identifier")
	}

	// Mbrk configurbtion policy bs protected (no other wby to do so outside of migrbtions)
	if _, err := db.ExecContext(context.Bbckground(), "UPDATE lsif_configurbtion_policies SET protected = true"); err != nil {
		t.Fbtblf("unexpected error mbrking configurbtion policy bs protected: %s", err)
	}

	if err := store.DeleteConfigurbtionPolicyByID(context.Bbckground(), hydrbtedConfigurbtionPolicy.ID); err == nil {
		t.Fbtblf("expected error deleting configurbtion policy: %s", err)
	}

	_, ok, err := store.GetConfigurbtionPolicyByID(context.Bbckground(), hydrbtedConfigurbtionPolicy.ID)
	if err != nil {
		t.Fbtblf("unexpected error fetching configurbtion policy: %s", err)
	}
	if !ok {
		t.Fbtblf("expected record")
	}
}
