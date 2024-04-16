package settings

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestRelevantSettings(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	settingsService := NewService(db)

	createOrg := func(name string) *types.Org {
		org, err := db.Orgs().Create(ctx, name, nil)
		require.NoError(t, err)
		return org

	}

	createUser := func(name string, orgs ...int32) *types.User {
		user, err := db.Users().Create(ctx, database.NewUser{Username: name, Email: name, EmailIsVerified: true})
		require.NoError(t, err)

		for _, org := range orgs {
			_, err = db.OrgMembers().Create(ctx, org, user.ID)
			require.NoError(t, err)
		}

		return user
	}

	org1 := createOrg("org1")
	org2 := createOrg("org2")

	user1 := createUser("user1", org1.ID)
	user2 := createUser("user2", org2.ID)
	user3 := createUser("user3", org1.ID, org2.ID)

	// Org1 contains user1 and user3
	// Org2 contains user2 and user3

	cases := []struct {
		subject  api.SettingsSubject
		expected []api.SettingsSubject
	}{{
		subject:  api.SettingsSubject{Default: true},
		expected: []api.SettingsSubject{{Default: true}},
	}, {
		subject: api.SettingsSubject{Site: true},
		expected: []api.SettingsSubject{
			{Default: true},
			{Site: true},
		},
	}, {
		subject: api.SettingsSubject{Org: &org1.ID},
		expected: []api.SettingsSubject{
			{Default: true},
			{Site: true},
			{Org: &org1.ID},
		},
	}, {
		subject: api.SettingsSubject{User: &user1.ID},
		expected: []api.SettingsSubject{
			{Default: true},
			{Site: true},
			{Org: &org1.ID},
			{User: &user1.ID},
		},
	}, {
		subject: api.SettingsSubject{User: &user2.ID},
		expected: []api.SettingsSubject{
			{Default: true},
			{Site: true},
			{Org: &org2.ID},
			{User: &user2.ID},
		},
	}, {
		subject: api.SettingsSubject{User: &user3.ID},
		expected: []api.SettingsSubject{
			{Default: true}, {Site: true},
			{Org: &org1.ID},
			{Org: &org2.ID},
			{User: &user3.ID},
		},
	}}

	for _, tc := range cases {
		t.Run(tc.subject.String(), func(t *testing.T) {
			got, err := settingsService.RelevantSubjects(ctx, tc.subject)
			require.NoError(t, err)
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestMergeSettings(t *testing.T) {
	cases := []struct {
		name     string
		left     *schema.Settings
		right    *schema.Settings
		expected *schema.Settings
	}{{
		name:     "nil left",
		left:     nil,
		right:    &schema.Settings{},
		expected: &schema.Settings{},
	}, {
		name: "empty left",
		left: &schema.Settings{},
		right: &schema.Settings{
			SearchDefaultMode: "test",
		},
		expected: &schema.Settings{
			SearchDefaultMode: "test",
		},
	}, {
		name: "merge bool ptr",
		left: &schema.Settings{
			AlertsHideObservabilitySiteAlerts: pointers.Ptr(true),
		},
		right: &schema.Settings{
			SearchDefaultMode: "test",
		},
		expected: &schema.Settings{
			SearchDefaultMode:                 "test",
			AlertsHideObservabilitySiteAlerts: pointers.Ptr(true),
		},
	}, {
		name: "merge bool",
		left: &schema.Settings{
			AlertsShowPatchUpdates:              false,
			BasicCodeIntelGlobalSearchesEnabled: true,
		},
		right: &schema.Settings{
			AlertsShowPatchUpdates:              true,
			BasicCodeIntelGlobalSearchesEnabled: false, // This is the zero value, so will not override a previous non-zero value
		},
		expected: &schema.Settings{
			AlertsShowPatchUpdates:              true,
			BasicCodeIntelGlobalSearchesEnabled: true,
		},
	}, {
		name: "merge int",
		left: &schema.Settings{
			SearchContextLines:     nil,
			HistoryDefaultPageSize: 1,
		},
		right: &schema.Settings{
			SearchContextLines:     pointers.Ptr(1),
			HistoryDefaultPageSize: 0, // This is the zero value, so will not override a previous non-zero value
		},
		expected: &schema.Settings{
			SearchContextLines:     pointers.Ptr(1),
			HistoryDefaultPageSize: 1, // This is the zero value, so will not override a previous non-zero value
		},
	}, {
		name: "deep merge struct pointer",
		left: &schema.Settings{
			ExperimentalFeatures: &schema.SettingsExperimentalFeatures{
				CodeMonitoringWebHooks: pointers.Ptr(true),
			},
		},
		right: &schema.Settings{
			ExperimentalFeatures: &schema.SettingsExperimentalFeatures{
				ShowMultilineSearchConsole: pointers.Ptr(false),
			},
		},
		expected: &schema.Settings{
			ExperimentalFeatures: &schema.SettingsExperimentalFeatures{
				CodeMonitoringWebHooks:     pointers.Ptr(true),
				ShowMultilineSearchConsole: pointers.Ptr(false),
			},
		},
	}, {
		name: "overwriting merge",
		left: &schema.Settings{
			AlertsHideObservabilitySiteAlerts: pointers.Ptr(true),
		},
		right: &schema.Settings{
			AlertsHideObservabilitySiteAlerts: pointers.Ptr(false),
		},
		expected: &schema.Settings{
			AlertsHideObservabilitySiteAlerts: pointers.Ptr(false),
		},
	}, {
		name: "deep merge slice",
		left: &schema.Settings{
			SearchScopes: []*schema.SearchScope{{Name: "test1"}},
		},
		right: &schema.Settings{
			SearchScopes: []*schema.SearchScope{{Name: "test2"}},
		},
		expected: &schema.Settings{
			SearchScopes: []*schema.SearchScope{{Name: "test1"}, {Name: "test2"}},
		},
	},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := mergeSettingsLeft(tc.left, tc.right)
			require.Equal(t, tc.expected, res)
		})
	}
}
