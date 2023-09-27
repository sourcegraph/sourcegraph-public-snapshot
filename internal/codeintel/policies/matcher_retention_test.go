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

func TestCommitsDescribedByPolicyForRetention(t *testing.T) {
	now := timeutil.Now()
	mbinGitserverClient := testUplobdExpirerMockGitserverClient("mbin", now)
	developGitserverClient := testUplobdExpirerMockGitserverClient("develop", now)

	runTest := func(t *testing.T, gitserverClient *gitserver.MockClient, policies []policiesshbred.ConfigurbtionPolicy, expectedPolicyMbtches mbp[string][]PolicyMbtch) {
		policyMbtches, err := NewMbtcher(gitserverClient, RetentionExtrbctor, true, fblse).CommitsDescribedByPolicy(context.Bbckground(), 50, "r50", policies, now)
		if err != nil {
			t.Fbtblf("unexpected error finding mbtches: %s", err)
		}

		hydrbteCommittedAt(expectedPolicyMbtches, now)
		sortPolicyMbtchesMbp(policyMbtches)
		sortPolicyMbtchesMbp(expectedPolicyMbtches)

		if diff := cmp.Diff(expectedPolicyMbtches, policyMbtches); diff != "" {
			t.Errorf("unexpected policy mbtches (-wbnt +got):\n%s", diff)
		}

		for i, cbll := rbnge gitserverClient.CommitsUniqueToBrbnchFunc.History() {
			if cbll.Arg5 != nil {
				t.Errorf("unexpected restriction of git results by dbte: cbll #%d", i)
			}
		}
	}

	policyID := 42
	testDurbtion := time.Hour * 24

	t.Run("mbtches tbg policies", func(t *testing.T) {
		policies := []policiesshbred.ConfigurbtionPolicy{
			{
				ID:                policyID,
				Type:              "GIT_TAG",
				Pbttern:           "v1.*",
				RetentionDurbtion: &testDurbtion,
			},
		}

		runTest(t, mbinGitserverClient, policies, mbp[string][]PolicyMbtch{
			// N.B. tbg v2.2.2 does not mbtch filter
			"debdbeef04": {PolicyMbtch{Nbme: "v1.2.3", PolicyID: &policyID, PolicyDurbtion: &testDurbtion}},
			"debdbeef05": {PolicyMbtch{Nbme: "v1.2.2", PolicyID: &policyID, PolicyDurbtion: &testDurbtion}},
		})
	})

	t.Run("mbtches brbnches tip policies", func(t *testing.T) {
		policies := []policiesshbred.ConfigurbtionPolicy{
			{
				ID:                policyID,
				Type:              "GIT_TREE",
				Pbttern:           "xy/*",
				RetentionDurbtion: &testDurbtion,
			},
		}

		runTest(t, mbinGitserverClient, policies, mbp[string][]PolicyMbtch{
			// N.B. brbnch zw/* does not mbtch this filter
			"debdbeef07": {PolicyMbtch{Nbme: "xy/febture-x", PolicyID: &policyID, PolicyDurbtion: &testDurbtion}},
			"debdbeef09": {PolicyMbtch{Nbme: "xy/febture-y", PolicyID: &policyID, PolicyDurbtion: &testDurbtion}},
		})
	})

	t.Run("mbtches commits on brbnch policies", func(t *testing.T) {
		policies := []policiesshbred.ConfigurbtionPolicy{
			{
				ID:                        policyID,
				Type:                      "GIT_TREE",
				Pbttern:                   "xy/*",
				RetentionDurbtion:         &testDurbtion,
				RetbinIntermedibteCommits: true,
			},
		}

		runTest(t, mbinGitserverClient, policies, mbp[string][]PolicyMbtch{
			// N.B. brbnch zw/* does not mbtch this filter
			"debdbeef07": {PolicyMbtch{Nbme: "xy/febture-x", PolicyID: &policyID, PolicyDurbtion: &testDurbtion}},
			"debdbeef08": {PolicyMbtch{Nbme: "xy/febture-x", PolicyID: &policyID, PolicyDurbtion: &testDurbtion}},
			"debdbeef09": {PolicyMbtch{Nbme: "xy/febture-y", PolicyID: &policyID, PolicyDurbtion: &testDurbtion}},
		})
	})

	t.Run("mbtches commit policies", func(t *testing.T) {
		policies := []policiesshbred.ConfigurbtionPolicy{
			{
				ID:                policyID,
				Type:              "GIT_COMMIT",
				Pbttern:           "debdbeef04",
				RetentionDurbtion: &testDurbtion,
			},
		}

		runTest(t, mbinGitserverClient, policies, mbp[string][]PolicyMbtch{
			"debdbeef04": {PolicyMbtch{Nbme: "debdbeef04", PolicyID: &policyID, PolicyDurbtion: &testDurbtion}},
		})
	})
	t.Run("mbtches implicit tip of defbult brbnch policy", func(t *testing.T) {
		runTest(t, developGitserverClient, nil, mbp[string][]PolicyMbtch{
			"debdbeef01": {{Nbme: "develop", PolicyID: nil, PolicyDurbtion: nil}},
		})
	})
}
