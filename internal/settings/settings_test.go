package settings

import (
	"testing"

	"github.com/stretchr/testify/require"

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
			SearchDefaultMode: "test",
		},
		expected: &schema.Settings{
			SearchDefaultMode: "test",
		},
	}, {
		name: "merge bool ptr",
		left: &schema.Settings{
			AlertsHideObservabilitySiteAlerts: boolPtr(true),
		},
		right: &schema.Settings{
			SearchDefaultMode: "test",
		},
		expected: &schema.Settings{
			SearchDefaultMode:                 "test",
			AlertsHideObservabilitySiteAlerts: boolPtr(true),
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
				CodeMonitoringWebHooks: boolPtr(true),
			},
		},
		right: &schema.Settings{
			ExperimentalFeatures: &schema.SettingsExperimentalFeatures{
				ShowMultilineSearchConsole: boolPtr(false),
			},
		},
		expected: &schema.Settings{
			ExperimentalFeatures: &schema.SettingsExperimentalFeatures{
				CodeMonitoringWebHooks:     boolPtr(true),
				ShowMultilineSearchConsole: boolPtr(false),
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
	},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := mergeSettingsLeft(tc.left, tc.right)
			require.Equal(t, tc.expected, res)
		})
	}
}
