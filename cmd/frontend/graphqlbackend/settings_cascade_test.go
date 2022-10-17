package graphqlbackend

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestMergeSettings(t *testing.T) {
	boolPtr := func(b bool) *bool {
		return &b
	}

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
			AlertsCodeHostIntegrationMessaging: "test",
		},
		expected: &schema.Settings{
			AlertsCodeHostIntegrationMessaging: "test",
		},
	}, {
		name: "merge bool ptr",
		left: &schema.Settings{
			AlertsHideObservabilitySiteAlerts: boolPtr(true),
		},
		right: &schema.Settings{
			AlertsCodeHostIntegrationMessaging: "test",
		},
		expected: &schema.Settings{
			AlertsCodeHostIntegrationMessaging: "test",
			AlertsHideObservabilitySiteAlerts:  boolPtr(true),
		},
	}, {
		name: "merge bool",
		left: &schema.Settings{
			AlertsShowPatchUpdates:    false,
			CodeHostUseNativeTooltips: true,
		},
		right: &schema.Settings{
			AlertsShowPatchUpdates:    true,
			CodeHostUseNativeTooltips: false, // This is the zero value, so will not override a previous non-zero value
		},
		expected: &schema.Settings{
			AlertsShowPatchUpdates:    true,
			CodeHostUseNativeTooltips: true,
		},
	}, {
		name: "merge int",
		left: &schema.Settings{
			SearchContextLines:                        0,
			CodeIntelligenceAutoIndexPopularRepoLimit: 1,
		},
		right: &schema.Settings{
			SearchContextLines:                        1,
			CodeIntelligenceAutoIndexPopularRepoLimit: 0, // This is the zero value, so will not override a previous non-zero value
		},
		expected: &schema.Settings{
			SearchContextLines:                        1,
			CodeIntelligenceAutoIndexPopularRepoLimit: 1, // This is the zero value, so will not override a previous non-zero value
		},
	}, {
		name: "deep merge struct pointer",
		left: &schema.Settings{
			ExperimentalFeatures: &schema.SettingsExperimentalFeatures{
				ShowSearchNotebook: boolPtr(true),
			},
		},
		right: &schema.Settings{
			ExperimentalFeatures: &schema.SettingsExperimentalFeatures{
				ShowSearchContextManagement: boolPtr(false),
			},
		},
		expected: &schema.Settings{
			ExperimentalFeatures: &schema.SettingsExperimentalFeatures{
				ShowSearchNotebook:          boolPtr(true),
				ShowSearchContextManagement: boolPtr(false),
			},
		},
	}, {
		name: "overwriting merge",
		left: &schema.Settings{
			AlertsHideObservabilitySiteAlerts: boolPtr(true),
		},
		right: &schema.Settings{
			AlertsHideObservabilitySiteAlerts: boolPtr(false),
		},
		expected: &schema.Settings{
			AlertsHideObservabilitySiteAlerts: boolPtr(false),
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
	}, {
		name: "no deep merge slice",
		left: &schema.Settings{
			Notices: []*schema.Notice{{Message: "test1"}},
		},
		right: &schema.Settings{
			Notices: []*schema.Notice{{Message: "test2"}},
		},
		expected: &schema.Settings{
			Notices: []*schema.Notice{{Message: "test2"}},
		},
	}, {
		name: "deep merge map",
		left: &schema.Settings{
			SearchRepositoryGroups: map[string][]any{
				"test1": {"test", 1},
				"test2": {"test", 2},
			},
		},
		right: &schema.Settings{
			SearchRepositoryGroups: map[string][]any{
				"test2": {"overridden", 3},
				"test3": {"merged", 4},
			},
		},
		expected: &schema.Settings{
			SearchRepositoryGroups: map[string][]any{
				"test1": {"test", 1},
				"test2": {"overridden", 3},
				"test3": {"merged", 4},
			},
		},
	}, {
		name: "deep merge insightsDashboards",
		left: &schema.Settings{
			InsightsDashboards: map[string]schema.InsightDashboard{
				"1": {Id: "1"},
				"2": {Id: "2"},
				"3": {Id: "3"},
			},
		},
		right: &schema.Settings{
			InsightsDashboards: map[string]schema.InsightDashboard{
				"2": {Id: "overridden", Title: "overridden"},
				"3": {Title: "overridden"},
				"4": {Id: "merged"},
			},
		},
		expected: &schema.Settings{
			InsightsDashboards: map[string]schema.InsightDashboard{
				"1": {Id: "1"},
				"2": {Id: "overridden", Title: "overridden"},
				"3": {Title: "overridden"},
				"4": {Id: "merged"},
			},
		},
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := mergeSettingsLeft(tc.left, tc.right)
			require.Equal(t, tc.expected, res)
		})
	}
}

func TestSubjects(t *testing.T) {
	t.Run("Default settings are included", func(t *testing.T) {
		cascade := &settingsCascade{db: database.NewMockDB(), unauthenticatedActor: true}
		subjects, err := cascade.Subjects(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(subjects) < 1 {
			t.Fatal("Expected at least 1 subject")
		}
		if subjects[0].defaultSettings == nil {
			t.Fatal("Expected the first subject to be default settings")
		}
	})
}
