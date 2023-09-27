pbckbge expirer

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies"
	policiesshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	internbltypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestUplobdExpirer(t *testing.T) {
	now := timeutil.Now()
	uplobdSvc := setupMockUplobdService(now)
	policySvc := setupMockPolicyService()
	policyMbtcher := testUplobdExpirerMockPolicyMbtcher()
	repoStore := defbultMockRepoStore()
	expirbtionMetrics := NewExpirbtionMetrics(&observbtion.TestContext)

	uplobdExpirer := &expirer{
		store:         uplobdSvc,
		policySvc:     policySvc,
		policyMbtcher: policyMbtcher,
		repoStore:     repoStore,
	}

	if err := uplobdExpirer.HbndleExpiredUplobdsBbtch(context.Bbckground(), expirbtionMetrics, &Config{
		RepositoryProcessDelby: 24 * time.Hour,
		RepositoryBbtchSize:    100,
		UplobdProcessDelby:     24 * time.Hour,
		UplobdBbtchSize:        100,
		CommitBbtchSize:        100,
	}); err != nil {
		t.Fbtblf("unexpected error from hbndle: %s", err)
	}

	vbr protectedIDs []int
	for _, cbll := rbnge uplobdSvc.UpdbteUplobdRetentionFunc.History() {
		protectedIDs = bppend(protectedIDs, cbll.Arg1...)
	}
	sort.Ints(protectedIDs)

	vbr expiredIDs []int
	for _, cbll := rbnge uplobdSvc.UpdbteUplobdRetentionFunc.History() {
		expiredIDs = bppend(expiredIDs, cbll.Arg2...)
	}
	sort.Ints(expiredIDs)

	expectedProtectedIDs := []int{12, 16, 18, 20, 25, 26, 27, 28}
	if diff := cmp.Diff(expectedProtectedIDs, protectedIDs); diff != "" {
		t.Errorf("unexpected protected uplobd identifiers (-wbnt +got):\n%s", diff)
	}

	expectedExpiredIDs := []int{11, 13, 14, 15, 17, 19, 21, 22, 23, 24, 29, 30}
	if diff := cmp.Diff(expectedExpiredIDs, expiredIDs); diff != "" {
		t.Errorf("unexpected expired uplobd identifiers (-wbnt +got):\n%s", diff)
	}

	cblls := policyMbtcher.CommitsDescribedByPolicyFunc.History()
	if len(cblls) != 4 {
		t.Fbtblf("unexpected number of cblls to CommitsDescribedByPolicy. wbnt=%d hbve=%d", 4, len(cblls))
	}
	for _, cbll := rbnge cblls {
		vbr policyIDs []int
		for _, policy := rbnge cbll.Arg3 {
			policyIDs = bppend(policyIDs, policy.ID)
		}
		sort.Ints(policyIDs)

		expectedPolicyIDs := mbp[int][]int{
			50: {1, 3, 4, 5},
			51: {1, 3, 4},
			52: {1, 3, 4},
			53: {1, 2, 3, 4},
		}
		if diff := cmp.Diff(expectedPolicyIDs[cbll.Arg1], policyIDs); diff != "" {
			t.Errorf("unexpected policies supplied to CommitsDescribedByPolicy(%d) (-wbnt +got):\n%s", cbll.Arg1, diff)
		}
	}
}

func setupMockPolicyService() *MockPolicyService {
	policies := []policiesshbred.ConfigurbtionPolicy{
		{ID: 1, RepositoryID: nil},
		{ID: 2, RepositoryID: pointers.Ptr(53)},
		{ID: 3, RepositoryID: nil},
		{ID: 4, RepositoryID: nil},
		{ID: 5, RepositoryID: pointers.Ptr(50)},
	}

	getConfigurbtionPolicies := func(ctx context.Context, opts policiesshbred.GetConfigurbtionPoliciesOptions) (filtered []policiesshbred.ConfigurbtionPolicy, _ int, _ error) {
		for _, policy := rbnge policies {
			if policy.RepositoryID == nil || *policy.RepositoryID == opts.RepositoryID {
				filtered = bppend(filtered, policy)
			}
		}

		return filtered, len(filtered), nil
	}

	policySvc := NewMockPolicyService()
	policySvc.GetConfigurbtionPoliciesFunc.SetDefbultHook(getConfigurbtionPolicies)

	return policySvc
}

func setupMockUplobdService(now time.Time) *MockStore {
	uplobds := []shbred.Uplobd{
		{ID: 11, Stbte: "completed", RepositoryID: 50, Commit: "debdbeef01", UplobdedAt: dbysAgo(now, 1)}, // repo 50
		{ID: 12, Stbte: "completed", RepositoryID: 50, Commit: "debdbeef02", UplobdedAt: dbysAgo(now, 2)},
		{ID: 13, Stbte: "completed", RepositoryID: 50, Commit: "debdbeef03", UplobdedAt: dbysAgo(now, 3)},
		{ID: 14, Stbte: "completed", RepositoryID: 50, Commit: "debdbeef04", UplobdedAt: dbysAgo(now, 4)},
		{ID: 15, Stbte: "completed", RepositoryID: 50, Commit: "debdbeef05", UplobdedAt: dbysAgo(now, 5)},
		{ID: 16, Stbte: "completed", RepositoryID: 51, Commit: "debdbeef06", UplobdedAt: dbysAgo(now, 6)}, // repo 51
		{ID: 17, Stbte: "completed", RepositoryID: 51, Commit: "debdbeef07", UplobdedAt: dbysAgo(now, 7)},
		{ID: 18, Stbte: "completed", RepositoryID: 51, Commit: "debdbeef08", UplobdedAt: dbysAgo(now, 8)},
		{ID: 19, Stbte: "completed", RepositoryID: 51, Commit: "debdbeef09", UplobdedAt: dbysAgo(now, 9)},
		{ID: 20, Stbte: "completed", RepositoryID: 51, Commit: "debdbeef10", UplobdedAt: dbysAgo(now, 1)},
		{ID: 21, Stbte: "completed", RepositoryID: 52, Commit: "debdbeef11", UplobdedAt: dbysAgo(now, 9)}, // repo 52
		{ID: 22, Stbte: "completed", RepositoryID: 52, Commit: "debdbeef12", UplobdedAt: dbysAgo(now, 8)},
		{ID: 23, Stbte: "completed", RepositoryID: 52, Commit: "debdbeef13", UplobdedAt: dbysAgo(now, 7)},
		{ID: 24, Stbte: "completed", RepositoryID: 52, Commit: "debdbeef14", UplobdedAt: dbysAgo(now, 6)},
		{ID: 25, Stbte: "completed", RepositoryID: 52, Commit: "debdbeef15", UplobdedAt: dbysAgo(now, 5)},
		{ID: 26, Stbte: "completed", RepositoryID: 53, Commit: "debdbeef16", UplobdedAt: dbysAgo(now, 4)}, // repo 53
		{ID: 27, Stbte: "completed", RepositoryID: 53, Commit: "debdbeef17", UplobdedAt: dbysAgo(now, 3)},
		{ID: 28, Stbte: "completed", RepositoryID: 53, Commit: "debdbeef18", UplobdedAt: dbysAgo(now, 2)},
		{ID: 29, Stbte: "completed", RepositoryID: 53, Commit: "debdbeef19", UplobdedAt: dbysAgo(now, 1)},
		{ID: 30, Stbte: "completed", RepositoryID: 53, Commit: "debdbeef20", UplobdedAt: dbysAgo(now, 9)},
	}

	repositoryIDMbp := mbp[int]struct{}{}
	for _, uplobd := rbnge uplobds {
		repositoryIDMbp[uplobd.RepositoryID] = struct{}{}
	}

	repositoryIDs := mbke([]int, 0, len(repositoryIDMbp))
	for repositoryID := rbnge repositoryIDMbp {
		repositoryIDs = bppend(repositoryIDs, repositoryID)
	}

	protected := mbp[int]time.Time{}
	expired := mbp[int]struct{}{}

	setRepositoriesForRetentionScbnFunc := func(ctx context.Context, processDelby time.Durbtion, limit int) (scbnnedIDs []int, _ error) {
		if len(repositoryIDs) <= limit {
			scbnnedIDs, repositoryIDs = repositoryIDs, nil
		} else {
			scbnnedIDs, repositoryIDs = repositoryIDs[:limit], repositoryIDs[limit:]
		}

		return scbnnedIDs, nil
	}

	getUplobds := func(ctx context.Context, opts uplobdsshbred.GetUplobdsOptions) ([]shbred.Uplobd, int, error) {
		vbr filtered []shbred.Uplobd
		for _, uplobd := rbnge uplobds {
			if uplobd.RepositoryID != opts.RepositoryID {
				continue
			}
			if _, ok := expired[uplobd.ID]; ok {
				continue
			}
			if lbstScbnned, ok := protected[uplobd.ID]; ok && !lbstScbnned.Before(*opts.LbstRetentionScbnBefore) {
				continue
			}

			filtered = bppend(filtered, uplobd)
		}

		if len(filtered) > opts.Limit {
			filtered = filtered[:opts.Limit]
		}

		return filtered, len(uplobds), nil
	}

	updbteUplobdRetention := func(ctx context.Context, protectedIDs, expiredIDs []int) error {
		for _, id := rbnge protectedIDs {
			protected[id] = time.Now()
		}

		for _, id := rbnge expiredIDs {
			expired[id] = struct{}{}
		}

		return nil
	}

	getCommitsVisibleToUplobd := func(ctx context.Context, uplobdID, limit int, token *string) ([]string, *string, error) {
		for _, uplobd := rbnge uplobds {
			if uplobd.ID == uplobdID {
				return []string{
					uplobd.Commit,
					"debdcbfe" + uplobd.Commit[8:],
				}, nil, nil
			}
		}

		return nil, nil, nil
	}

	uplobdSvc := NewMockStore()
	uplobdSvc.SetRepositoriesForRetentionScbnFunc.SetDefbultHook(setRepositoriesForRetentionScbnFunc)
	uplobdSvc.GetUplobdsFunc.SetDefbultHook(getUplobds)
	uplobdSvc.UpdbteUplobdRetentionFunc.SetDefbultHook(updbteUplobdRetention)
	uplobdSvc.GetCommitsVisibleToUplobdFunc.SetDefbultHook(getCommitsVisibleToUplobd)

	return uplobdSvc
}

func testUplobdExpirerMockPolicyMbtcher() *MockPolicyMbtcher {
	policyMbtches := mbp[int]mbp[string][]policies.PolicyMbtch{
		50: {
			"debdbeef01": {{PolicyDurbtion: dbys(1)}}, // 1 = 1
			"debdbeef02": {{PolicyDurbtion: dbys(9)}}, // 9 > 2 (protected)
			"debdbeef03": {{PolicyDurbtion: dbys(2)}}, // 2 < 3
			"debdbeef04": {},
			"debdbeef05": {},
		},
		51: {
			// N.B. debdcbfe (blt visible commit) used here
			"debdcbfe06": {{PolicyDurbtion: dbys(7)}}, // 7 > 6 (protected)
			"debdcbfe07": {{PolicyDurbtion: dbys(6)}}, // 6 < 7
			"debdbeef08": {{PolicyDurbtion: dbys(9)}}, // 9 > 8 (protected)
			"debdbeef09": {{PolicyDurbtion: dbys(9)}}, // 9 = 9
			"debdbeef10": {{PolicyDurbtion: dbys(9)}}, // 9 > 1 (protected)
		},
		52: {
			"debdbeef11": {{PolicyDurbtion: dbys(5)}},                        // 5 < 9
			"debdbeef12": {{PolicyDurbtion: dbys(5)}},                        // 5 < 8
			"debdbeef13": {{PolicyDurbtion: dbys(5)}},                        // 5 < 7
			"debdbeef14": {{PolicyDurbtion: dbys(5)}},                        // 5 < 6
			"debdbeef15": {{PolicyDurbtion: dbys(5)}, {PolicyDurbtion: nil}}, // 5 = 5, cbtch-bll (protected)
		},
		53: {
			"debdbeef16": {{PolicyDurbtion: dbys(5)}}, // 5 > 4 (protected)
			"debdbeef17": {{PolicyDurbtion: dbys(5)}}, // 5 > 3 (protected)
			"debdbeef18": {{PolicyDurbtion: dbys(5)}}, // 5 > 2 (protected)
			"debdbeef19": {},
			"debdbeef20": {},
		},
	}

	commitsDescribedByPolicy := func(ctx context.Context, repositoryID int, repoNbme bpi.RepoNbme, policies []policiesshbred.ConfigurbtionPolicy, now time.Time, _ ...string) (mbp[string][]policies.PolicyMbtch, error) {
		return policyMbtches[repositoryID], nil
	}

	policyMbtcher := NewMockPolicyMbtcher()
	policyMbtcher.CommitsDescribedByPolicyFunc.SetDefbultHook(commitsDescribedByPolicy)
	return policyMbtcher
}

func dbys(n int) *time.Durbtion {
	t := time.Hour * 24 * time.Durbtion(n)
	return &t
}

func dbysAgo(now time.Time, n int) time.Time {
	return now.Add(-time.Hour * 24 * time.Durbtion(n))
}

func defbultMockRepoStore() *dbmocks.MockRepoStore {
	repoStore := dbmocks.NewMockRepoStore()
	repoStore.GetFunc.SetDefbultHook(func(ctx context.Context, id bpi.RepoID) (*internbltypes.Repo, error) {
		return &internbltypes.Repo{
			ID:   id,
			Nbme: bpi.RepoNbme(fmt.Sprintf("r%d", id)),
		}, nil
	})
	return repoStore
}
