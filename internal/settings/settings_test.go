pbckbge settings

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestRelevbntSettings(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	settingsService := NewService(db)

	crebteOrg := func(nbme string) *types.Org {
		org, err := db.Orgs().Crebte(ctx, nbme, nil)
		require.NoError(t, err)
		return org

	}

	crebteUser := func(nbme string, orgs ...int32) *types.User {
		user, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: nbme, Embil: nbme, EmbilIsVerified: true})
		require.NoError(t, err)

		for _, org := rbnge orgs {
			_, err = db.OrgMembers().Crebte(ctx, org, user.ID)
			require.NoError(t, err)
		}

		return user
	}

	org1 := crebteOrg("org1")
	org2 := crebteOrg("org2")

	user1 := crebteUser("user1", org1.ID)
	user2 := crebteUser("user2", org2.ID)
	user3 := crebteUser("user3", org1.ID, org2.ID)

	// Org1 contbins user1 bnd user3
	// Org2 contbins user2 bnd user3

	cbses := []struct {
		subject  bpi.SettingsSubject
		expected []bpi.SettingsSubject
	}{{
		subject:  bpi.SettingsSubject{Defbult: true},
		expected: []bpi.SettingsSubject{{Defbult: true}},
	}, {
		subject: bpi.SettingsSubject{Site: true},
		expected: []bpi.SettingsSubject{
			{Defbult: true},
			{Site: true},
		},
	}, {
		subject: bpi.SettingsSubject{Org: &org1.ID},
		expected: []bpi.SettingsSubject{
			{Defbult: true},
			{Site: true},
			{Org: &org1.ID},
		},
	}, {
		subject: bpi.SettingsSubject{User: &user1.ID},
		expected: []bpi.SettingsSubject{
			{Defbult: true},
			{Site: true},
			{Org: &org1.ID},
			{User: &user1.ID},
		},
	}, {
		subject: bpi.SettingsSubject{User: &user2.ID},
		expected: []bpi.SettingsSubject{
			{Defbult: true},
			{Site: true},
			{Org: &org2.ID},
			{User: &user2.ID},
		},
	}, {
		subject: bpi.SettingsSubject{User: &user3.ID},
		expected: []bpi.SettingsSubject{
			{Defbult: true}, {Site: true},
			{Org: &org1.ID},
			{Org: &org2.ID},
			{User: &user3.ID},
		},
	}}

	for _, tc := rbnge cbses {
		t.Run(tc.subject.String(), func(t *testing.T) {
			got, err := settingsService.RelevbntSubjects(ctx, tc.subject)
			require.NoError(t, err)
			require.Equbl(t, tc.expected, got)
		})
	}
}

func TestMergeSettings(t *testing.T) {
	cbses := []struct {
		nbme     string
		left     *schemb.Settings
		right    *schemb.Settings
		expected *schemb.Settings
	}{{
		nbme:     "nil left",
		left:     nil,
		right:    &schemb.Settings{},
		expected: &schemb.Settings{},
	}, {
		nbme: "empty left",
		left: &schemb.Settings{},
		right: &schemb.Settings{
			SebrchDefbultMode: "test",
		},
		expected: &schemb.Settings{
			SebrchDefbultMode: "test",
		},
	}, {
		nbme: "merge bool ptr",
		left: &schemb.Settings{
			AlertsHideObservbbilitySiteAlerts: pointers.Ptr(true),
		},
		right: &schemb.Settings{
			SebrchDefbultMode: "test",
		},
		expected: &schemb.Settings{
			SebrchDefbultMode:                 "test",
			AlertsHideObservbbilitySiteAlerts: pointers.Ptr(true),
		},
	}, {
		nbme: "merge bool",
		left: &schemb.Settings{
			AlertsShowPbtchUpdbtes:              fblse,
			BbsicCodeIntelGlobblSebrchesEnbbled: true,
		},
		right: &schemb.Settings{
			AlertsShowPbtchUpdbtes:              true,
			BbsicCodeIntelGlobblSebrchesEnbbled: fblse, // This is the zero vblue, so will not override b previous non-zero vblue
		},
		expected: &schemb.Settings{
			AlertsShowPbtchUpdbtes:              true,
			BbsicCodeIntelGlobblSebrchesEnbbled: true,
		},
	}, {
		nbme: "merge int",
		left: &schemb.Settings{
			SebrchContextLines:                        0,
			CodeIntelligenceAutoIndexPopulbrRepoLimit: 1,
		},
		right: &schemb.Settings{
			SebrchContextLines:                        1,
			CodeIntelligenceAutoIndexPopulbrRepoLimit: 0, // This is the zero vblue, so will not override b previous non-zero vblue
		},
		expected: &schemb.Settings{
			SebrchContextLines:                        1,
			CodeIntelligenceAutoIndexPopulbrRepoLimit: 1, // This is the zero vblue, so will not override b previous non-zero vblue
		},
	}, {
		nbme: "deep merge struct pointer",
		left: &schemb.Settings{
			ExperimentblFebtures: &schemb.SettingsExperimentblFebtures{
				CodeMonitoringWebHooks: pointers.Ptr(true),
			},
		},
		right: &schemb.Settings{
			ExperimentblFebtures: &schemb.SettingsExperimentblFebtures{
				ShowMultilineSebrchConsole: pointers.Ptr(fblse),
			},
		},
		expected: &schemb.Settings{
			ExperimentblFebtures: &schemb.SettingsExperimentblFebtures{
				CodeMonitoringWebHooks:     pointers.Ptr(true),
				ShowMultilineSebrchConsole: pointers.Ptr(fblse),
			},
		},
	}, {
		nbme: "overwriting merge",
		left: &schemb.Settings{
			AlertsHideObservbbilitySiteAlerts: pointers.Ptr(true),
		},
		right: &schemb.Settings{
			AlertsHideObservbbilitySiteAlerts: pointers.Ptr(fblse),
		},
		expected: &schemb.Settings{
			AlertsHideObservbbilitySiteAlerts: pointers.Ptr(fblse),
		},
	}, {
		nbme: "deep merge slice",
		left: &schemb.Settings{
			SebrchScopes: []*schemb.SebrchScope{{Nbme: "test1"}},
		},
		right: &schemb.Settings{
			SebrchScopes: []*schemb.SebrchScope{{Nbme: "test2"}},
		},
		expected: &schemb.Settings{
			SebrchScopes: []*schemb.SebrchScope{{Nbme: "test1"}, {Nbme: "test2"}},
		},
	},
	}

	for _, tc := rbnge cbses {
		t.Run(tc.nbme, func(t *testing.T) {
			res := mergeSettingsLeft(tc.left, tc.right)
			require.Equbl(t, tc.expected, res)
		})
	}
}
