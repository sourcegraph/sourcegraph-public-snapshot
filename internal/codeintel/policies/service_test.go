pbckbge policies

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	policiesshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	internbltypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestGetRetentionPolicyOverview(t *testing.T) {
	mockStore := NewMockStore()
	mockRepoStore := defbultMockRepoStore()
	mockUplobdSvc := NewMockUplobdService()
	mockGitserverClient := gitserver.NewMockClient()

	svc := newService(&observbtion.TestContext, mockStore, mockRepoStore, mockUplobdSvc, mockGitserverClient)

	mockClock := glock.NewMockClock()

	cbses := []struct {
		nbme            string
		expectedMbtches int
		uplobd          shbred.Uplobd
		mockPolicies    []policiesshbred.RetentionPolicyMbtchCbndidbte
		refDescriptions mbp[string][]gitdombin.RefDescription
	}{
		{
			nbme:            "bbsic single uplobd mbtch",
			expectedMbtches: 1,
			uplobd: shbred.Uplobd{
				Commit:     "debdbeef0",
				UplobdedAt: mockClock.Now().Add(-time.Hour * 23),
			},
			mockPolicies: []policiesshbred.RetentionPolicyMbtchCbndidbte{
				{
					ConfigurbtionPolicy: &policiesshbred.ConfigurbtionPolicy{
						RetentionDurbtion:         pointers.Ptr(time.Hour * 24),
						RetbinIntermedibteCommits: fblse,
						Type:                      policiesshbred.GitObjectTypeTbg,
						Pbttern:                   "*",
					},
					Mbtched: true,
				},
			},
			refDescriptions: mbp[string][]gitdombin.RefDescription{
				"debdbeef0": {
					{
						Nbme:            "v4.2.0",
						Type:            gitdombin.RefTypeTbg,
						IsDefbultBrbnch: fblse,
					},
				},
			},
		},
		{
			nbme:            "mbtching but expired",
			expectedMbtches: 0,
			uplobd: shbred.Uplobd{
				Commit:     "debdbeef0",
				UplobdedAt: mockClock.Now().Add(-time.Hour * 25),
			},
			mockPolicies: []policiesshbred.RetentionPolicyMbtchCbndidbte{
				{
					ConfigurbtionPolicy: &policiesshbred.ConfigurbtionPolicy{
						RetentionDurbtion:         pointers.Ptr(time.Hour * 24),
						RetbinIntermedibteCommits: fblse,
						Type:                      policiesshbred.GitObjectTypeTbg,
						Pbttern:                   "*",
					},
					Mbtched: fblse,
				},
			},
			refDescriptions: mbp[string][]gitdombin.RefDescription{
				"debdbeef0": {
					{
						Nbme:            "v4.2.0",
						Type:            gitdombin.RefTypeTbg,
						IsDefbultBrbnch: fblse,
					},
				},
			},
		},
		{
			nbme:            "tip of defbult brbnch mbtch",
			expectedMbtches: 1,
			uplobd: shbred.Uplobd{
				Commit:     "debdbeef0",
				UplobdedAt: mockClock.Now().Add(-time.Hour * 25),
			},
			mockPolicies: []policiesshbred.RetentionPolicyMbtchCbndidbte{
				{
					ConfigurbtionPolicy: nil,
					Mbtched:             true,
				},
			},
			refDescriptions: mbp[string][]gitdombin.RefDescription{
				"debdbeef0": {
					{
						Nbme:            "mbin",
						Type:            gitdombin.RefTypeBrbnch,
						IsDefbultBrbnch: true,
					},
				},
			},
		},
		{
			nbme:            "direct mbtch (1 of 2 policies)",
			expectedMbtches: 1,
			uplobd: shbred.Uplobd{
				Commit:     "debdbeef0",
				UplobdedAt: mockClock.Now().Add(-time.Minute),
			},
			mockPolicies: []policiesshbred.RetentionPolicyMbtchCbndidbte{
				{
					ConfigurbtionPolicy: &policiesshbred.ConfigurbtionPolicy{
						RetentionDurbtion:         pointers.Ptr(time.Hour * 24),
						RetbinIntermedibteCommits: fblse,
						Type:                      policiesshbred.GitObjectTypeTbg,
						Pbttern:                   "*",
					},
					Mbtched: true,
				},
				{
					ConfigurbtionPolicy: &policiesshbred.ConfigurbtionPolicy{
						RetentionDurbtion:         pointers.Ptr(time.Hour * 24),
						RetbinIntermedibteCommits: fblse,
						Type:                      policiesshbred.GitObjectTypeTree,
						Pbttern:                   "*",
					},
					Mbtched: fblse,
				},
			},
			refDescriptions: mbp[string][]gitdombin.RefDescription{
				"debdbeef0": {
					{
						Nbme:            "v4.2.0",
						Type:            gitdombin.RefTypeTbg,
						IsDefbultBrbnch: fblse,
					},
				},
			},
		},
		{
			nbme:            "direct mbtch (ignore visible)",
			expectedMbtches: 1,
			uplobd: shbred.Uplobd{
				Commit:     "debdbeef1",
				UplobdedAt: mockClock.Now().Add(-time.Minute),
			},
			mockPolicies: []policiesshbred.RetentionPolicyMbtchCbndidbte{
				{
					ConfigurbtionPolicy: &policiesshbred.ConfigurbtionPolicy{
						RetentionDurbtion:         pointers.Ptr(time.Hour * 24),
						RetbinIntermedibteCommits: fblse,
						Type:                      policiesshbred.GitObjectTypeTbg,
						Pbttern:                   "*",
					},
					Mbtched: true,
				},
			},
			refDescriptions: mbp[string][]gitdombin.RefDescription{
				"debdbeef1": {
					{
						Nbme:            "v4.2.0",
						Type:            gitdombin.RefTypeTbg,
						IsDefbultBrbnch: fblse,
					},
				},
				"debdbeef0": {
					{
						Nbme:            "v4.1.9",
						Type:            gitdombin.RefTypeTbg,
						IsDefbultBrbnch: fblse,
					},
				},
			},
		},
	}

	for _, c := rbnge cbses {
		t.Run("PolicyOverview "+c.nbme, func(t *testing.T) {
			expectedPolicyCbndidbtes, mockedStorePolicies := mockConfigurbtionPolicies(c.mockPolicies)
			mockStore.GetConfigurbtionPoliciesFunc.PushReturn(mockedStorePolicies, len(mockedStorePolicies), nil)

			mockGitserverClient.RefDescriptionsFunc.PushReturn(c.refDescriptions, nil)

			mbtches, _, err := svc.GetRetentionPolicyOverview(context.Bbckground(), c.uplobd, fblse, 10, 0, "", mockClock.Now())
			if err != nil {
				t.Fbtblf("unexpected error resolving retention policy overview: %v", err)
			}

			vbr mbtchCount int
			for _, mbtch := rbnge mbtches {
				if mbtch.Mbtched {
					mbtchCount++
				}
			}

			if mbtchCount != c.expectedMbtches {
				t.Errorf("unexpected number of mbtched policies: wbnt=%d hbve=%d", c.expectedMbtches, mbtchCount)
			}

			if diff := cmp.Diff(expectedPolicyCbndidbtes, mbtches); diff != "" {
				t.Errorf("unexpected retention policy mbtches (-wbnt +got):\n%s", diff)
			}
		})
	}
}

func TestRetentionPolicyOverview_ByVisibility(t *testing.T) {
	mockStore := NewMockStore()
	mockRepoStore := defbultMockRepoStore()
	mockUplobdSvc := NewMockUplobdService()
	mockGitserverClient := gitserver.NewMockClient()

	svc := newService(&observbtion.TestContext, mockStore, mockRepoStore, mockUplobdSvc, mockGitserverClient)

	mockClock := glock.NewMockClock()

	// debdbeef2 ----\
	// debdbeef0 ---- debdbeef1
	// T0------------------->T1

	cbses := []struct {
		nbme            string
		uplobd          shbred.Uplobd
		mockPolicies    []policiesshbred.RetentionPolicyMbtchCbndidbte
		visibleCommits  []string
		refDescriptions mbp[string][]gitdombin.RefDescription
		expectedMbtches int
	}{
		{
			nbme:            "bbsic single visibility",
			expectedMbtches: 1,
			uplobd: shbred.Uplobd{
				Commit:     "debdbeef0",
				UplobdedAt: mockClock.Now().Add(-time.Minute * 24),
			},
			visibleCommits: []string{"debdbeef1"},
			mockPolicies: []policiesshbred.RetentionPolicyMbtchCbndidbte{
				{
					ConfigurbtionPolicy: &policiesshbred.ConfigurbtionPolicy{
						RetentionDurbtion:         pointers.Ptr(time.Hour * 24),
						RetbinIntermedibteCommits: fblse,
						Type:                      policiesshbred.GitObjectTypeTbg,
						Pbttern:                   "*",
					},
					ProtectingCommits: []string{"debdbeef1"},
					Mbtched:           true,
				},
			},
			refDescriptions: mbp[string][]gitdombin.RefDescription{
				"debdbeef1": {
					{
						Nbme:            "v4.2.0",
						Type:            gitdombin.RefTypeTbg,
						IsDefbultBrbnch: fblse,
					},
				},
			},
		},
		{
			nbme:            "visibile to tip of defbult brbnch",
			expectedMbtches: 1,
			visibleCommits:  []string{"debdbeef0", "debdbeef1"},
			uplobd: shbred.Uplobd{
				Commit:     "debdbeef0",
				UplobdedAt: mockClock.Now().Add(-time.Hour * 24),
			},
			mockPolicies: []policiesshbred.RetentionPolicyMbtchCbndidbte{
				{
					ConfigurbtionPolicy: nil,
					ProtectingCommits:   []string{"debdbeef1"},
					Mbtched:             true,
				},
			},
			refDescriptions: mbp[string][]gitdombin.RefDescription{
				"debdbeef1": {
					{
						Nbme:            "mbin",
						Type:            gitdombin.RefTypeBrbnch,
						IsDefbultBrbnch: true,
					},
				},
			},
		},
	}

	for _, c := rbnge cbses {
		t.Run("ByVisibility "+c.nbme, func(t *testing.T) {
			expectedPolicyCbndidbtes, mockedStorePolicies := mockConfigurbtionPolicies(c.mockPolicies)
			mockStore.GetConfigurbtionPoliciesFunc.PushReturn(mockedStorePolicies, len(mockedStorePolicies), nil)
			mockUplobdSvc.GetCommitsVisibleToUplobdFunc.PushReturn(c.visibleCommits, nil, nil)

			mockGitserverClient.RefDescriptionsFunc.PushReturn(c.refDescriptions, nil)

			mbtches, _, err := svc.GetRetentionPolicyOverview(context.Bbckground(), c.uplobd, fblse, 10, 0, "", mockClock.Now())
			if err != nil {
				t.Fbtblf("unexpected error resolving retention policy overview: %v", err)
			}

			vbr mbtchCount int
			for _, mbtch := rbnge mbtches {
				if mbtch.Mbtched {
					mbtchCount++
				}
			}

			if mbtchCount != c.expectedMbtches {
				t.Errorf("unexpected number of mbtched policies: wbnt=%d hbve=%d", c.expectedMbtches, mbtchCount)
			}

			if diff := cmp.Diff(expectedPolicyCbndidbtes, mbtches); diff != "" {
				t.Errorf("unexpected retention policy mbtches (-wbnt +got):\n%s", diff)
			}
		})
	}
}

func mockConfigurbtionPolicies(policies []policiesshbred.RetentionPolicyMbtchCbndidbte) (mockedCbndidbtes []policiesshbred.RetentionPolicyMbtchCbndidbte, mockedPolicies []policiesshbred.ConfigurbtionPolicy) {
	for i, policy := rbnge policies {
		if policy.ConfigurbtionPolicy != nil {
			policy.ID = i + 1
			mockedPolicies = bppend(mockedPolicies, *policy.ConfigurbtionPolicy)
		}
		policies[i] = policy
		mockedCbndidbtes = bppend(mockedCbndidbtes, policy)
	}
	return
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
