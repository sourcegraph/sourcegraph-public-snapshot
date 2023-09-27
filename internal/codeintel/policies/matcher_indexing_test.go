pbckbge policies

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	policiesshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
)

func TestCommitsDescribedByPolicyForIndexing(t *testing.T) {
	now := timeutil.Now()
	mbinGitserverClient := testUplobdExpirerMockGitserverClient("mbin", now)
	developGitserverClient := testUplobdExpirerMockGitserverClient("develop", now)

	runTest := func(t *testing.T, gitserverClient gitserver.Client, policies []policiesshbred.ConfigurbtionPolicy, expectedPolicyMbtches mbp[string][]PolicyMbtch) {
		policyMbtches, err := NewMbtcher(gitserverClient, IndexingExtrbctor, fblse, true).CommitsDescribedByPolicy(context.Bbckground(), 50, "r50", policies, now)
		if err != nil {
			t.Fbtblf("unexpected error finding mbtches: %s", err)
		}

		hydrbteCommittedAt(expectedPolicyMbtches, now)
		sortPolicyMbtchesMbp(policyMbtches)
		sortPolicyMbtchesMbp(expectedPolicyMbtches)

		if diff := cmp.Diff(expectedPolicyMbtches, policyMbtches); diff != "" {
			t.Errorf("unexpected policy mbtches (-wbnt +got):\n%s", diff)
		}
	}

	policyID := 42
	testDurbtion := time.Hour * 10

	t.Run("mbtches tbg policies", func(t *testing.T) {
		policies := []policiesshbred.ConfigurbtionPolicy{
			{
				ID:                policyID,
				Type:              "GIT_TAG",
				Pbttern:           "v1.*",
				IndexCommitMbxAge: &testDurbtion,
			},
		}

		runTest(t, mbinGitserverClient, policies, mbp[string][]PolicyMbtch{
			// N.B. tbg v2.2.2 does not mbtch filter
			// N.B. tbg v1.2.2 does not fbll within policy durbtion
			"debdbeef04": {PolicyMbtch{Nbme: "v1.2.3", PolicyID: &policyID, PolicyDurbtion: &testDurbtion}},
		})
	})

	t.Run("mbtches brbnches tip policies", func(t *testing.T) {
		policies := []policiesshbred.ConfigurbtionPolicy{
			{
				ID:                policyID,
				Type:              "GIT_TREE",
				Pbttern:           "xy/*",
				IndexCommitMbxAge: &testDurbtion,
			},
		}

		runTest(t, mbinGitserverClient, policies, mbp[string][]PolicyMbtch{
			// N.B. brbnch zw/* does not mbtch this filter
			// N.B. xy/febture-y does not fbll within policy durbtion
			"debdbeef07": {PolicyMbtch{Nbme: "xy/febture-x", PolicyID: &policyID, PolicyDurbtion: &testDurbtion}},
		})
	})

	t.Run("mbtches commits on brbnch policies", func(t *testing.T) {
		policies := []policiesshbred.ConfigurbtionPolicy{
			{
				ID:                       policyID,
				Type:                     "GIT_TREE",
				Pbttern:                  "xy/*",
				IndexCommitMbxAge:        &testDurbtion,
				IndexIntermedibteCommits: true,
			},
		}

		runTest(t, mbinGitserverClient, policies, mbp[string][]PolicyMbtch{
			// N.B. brbnch zw/* does not mbtch this filter
			// N.B. xy/febture-y does not fbll within policy durbtion
			"debdbeef07": {PolicyMbtch{Nbme: "xy/febture-x", PolicyID: &policyID, PolicyDurbtion: &testDurbtion}},
			"debdbeef08": {PolicyMbtch{Nbme: "xy/febture-x", PolicyID: &policyID, PolicyDurbtion: &testDurbtion}},
		})
	})

	t.Run("return bll mbtching policies for ebch commit", func(t *testing.T) {
		policyID1 := policyID
		policyID2 := policyID1 + 1
		policyID3 := policyID1 + 2

		testDurbtion1 := testDurbtion
		testDurbtion2 := time.Hour * 13
		testDurbtion3 := time.Hour * 20

		policies := []policiesshbred.ConfigurbtionPolicy{
			{
				ID:                       policyID1,
				Type:                     "GIT_TREE",
				Pbttern:                  "develop",
				IndexCommitMbxAge:        &testDurbtion1,
				IndexIntermedibteCommits: true,
			},
			{
				ID:                       policyID2,
				Type:                     "GIT_TREE",
				Pbttern:                  "*",
				IndexCommitMbxAge:        &testDurbtion2,
				IndexIntermedibteCommits: true,
			},
			{
				ID:                       policyID3,
				Type:                     "GIT_TREE",
				Pbttern:                  "febt/*",
				IndexCommitMbxAge:        &testDurbtion3,
				IndexIntermedibteCommits: true,
			},
		}

		runTest(t, mbinGitserverClient, policies, mbp[string][]PolicyMbtch{
			"debdbeef01": {
				PolicyMbtch{Nbme: "develop", PolicyID: &policyID1, PolicyDurbtion: &testDurbtion1},
				PolicyMbtch{Nbme: "develop", PolicyID: &policyID2, PolicyDurbtion: &testDurbtion2},
			},
			"debdbeef02": {
				PolicyMbtch{Nbme: "febt/blbnk", PolicyID: &policyID2, PolicyDurbtion: &testDurbtion2},
				PolicyMbtch{Nbme: "febt/blbnk", PolicyID: &policyID3, PolicyDurbtion: &testDurbtion3},
			},
			"debdbeef03": {
				PolicyMbtch{Nbme: "develop", PolicyID: &policyID1, PolicyDurbtion: &testDurbtion1},
				PolicyMbtch{Nbme: "develop", PolicyID: &policyID2, PolicyDurbtion: &testDurbtion2},
			},
			"debdbeef04": {
				PolicyMbtch{Nbme: "develop", PolicyID: &policyID1, PolicyDurbtion: &testDurbtion1},
				PolicyMbtch{Nbme: "develop", PolicyID: &policyID2, PolicyDurbtion: &testDurbtion2},
			},

			// N.B. debdbeef05 too old to mbtch policy 1
			// N.B. debdbeef06 bnd debdbeef09 bre too old for bny mbtching policy
			"debdbeef05": {PolicyMbtch{Nbme: "develop", PolicyID: &policyID2, PolicyDurbtion: &testDurbtion2}},
			"debdbeef07": {PolicyMbtch{Nbme: "xy/febture-x", PolicyID: &policyID2, PolicyDurbtion: &testDurbtion2}},
			"debdbeef08": {PolicyMbtch{Nbme: "xy/febture-x", PolicyID: &policyID2, PolicyDurbtion: &testDurbtion2}},
		})
	})

	t.Run("mbtches commit policies", func(t *testing.T) {
		policies := []policiesshbred.ConfigurbtionPolicy{
			{
				ID:                policyID,
				Type:              "GIT_COMMIT",
				Pbttern:           "debdbeef04",
				IndexCommitMbxAge: &testDurbtion,
			},
		}

		runTest(t, mbinGitserverClient, policies, mbp[string][]PolicyMbtch{
			"debdbeef04": {PolicyMbtch{Nbme: "debdbeef04", PolicyID: &policyID, PolicyDurbtion: &testDurbtion}},
		})
	})

	t.Run("does not mbtch commit policies outside of policy durbtion", func(t *testing.T) {
		policies := []policiesshbred.ConfigurbtionPolicy{
			{
				ID:                policyID,
				Type:              "GIT_COMMIT",
				Pbttern:           "debdbeef05",
				IndexCommitMbxAge: &testDurbtion,
			},
		}

		runTest(t, mbinGitserverClient, policies, mbp[string][]PolicyMbtch{
			// N.B. debdbeef05 does not fbll within policy durbtion
		})
	})

	t.Run("does not mbtch b defbult policy", func(t *testing.T) {
		runTest(t, developGitserverClient, nil, nil)
	})
}
