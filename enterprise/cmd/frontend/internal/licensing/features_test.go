package licensing

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/license"
)

func Test_isFeatureEnabled(t *testing.T) {
	check := func(t *testing.T, feature Feature, licenseTags []string, wantEnabled bool) {
		t.Helper()
		got := isFeatureEnabled(license.Info{Tags: licenseTags}, feature)
		if got != wantEnabled {
			t.Errorf("%q: got %v, want %v", licenseTags, got, wantEnabled)
		}
	}

	t.Run(string(FeatureCustomBranding), func(t *testing.T) {
		check(t, FeatureCustomBranding, EnterpriseBasicTags, false)
		check(t, FeatureCustomBranding, EnterpriseStarterTags, false)
		check(t, FeatureCustomBranding, EnterprisePlusTags, true)
		check(t, FeatureCustomBranding, EliteTags, true)
	})

	t.Run(string(FeatureACLs), func(t *testing.T) {
		check(t, FeatureACLs, EnterpriseBasicTags, false)
		check(t, FeatureACLs, EnterpriseStarterTags, false)
		check(t, FeatureACLs, EnterprisePlusTags, true)
		check(t, FeatureACLs, EliteTags, true)
	})

	t.Run(string(FeatureExtensionWhitelist), func(t *testing.T) {
		check(t, FeatureExtensionWhitelist, EnterpriseBasicTags, false)
		check(t, FeatureExtensionWhitelist, EnterpriseStarterTags, false)
		check(t, FeatureExtensionWhitelist, EnterprisePlusTags, true)
		check(t, FeatureExtensionWhitelist, EliteTags, true)
	})

	t.Run(string(FeatureRemoteExtensionsAllowDisallow), func(t *testing.T) {
		check(t, FeatureRemoteExtensionsAllowDisallow, EnterpriseBasicTags, false)
		check(t, FeatureRemoteExtensionsAllowDisallow, EnterpriseStarterTags, false)
		check(t, FeatureRemoteExtensionsAllowDisallow, EnterprisePlusTags, true)
		check(t, FeatureRemoteExtensionsAllowDisallow, EliteTags, true)
	})

	t.Run(string(FeatureExtensionRegistry), func(t *testing.T) {
		check(t, FeatureExtensionRegistry, EnterpriseBasicTags, false)
		check(t, FeatureExtensionRegistry, EnterpriseStarterTags, false)
		check(t, FeatureExtensionRegistry, EnterprisePlusTags, false)
		check(t, FeatureExtensionRegistry, EliteTags, true)
	})

	t.Run(string(FeatureAutomation), func(t *testing.T) {
		check(t, FeatureAutomation, EnterpriseBasicTags, false)
		check(t, FeatureAutomation, EnterpriseStarterTags, false)
		check(t, FeatureAutomation, EnterprisePlusTags, false)
		check(t, FeatureAutomation, EliteTags, true)
	})
}
