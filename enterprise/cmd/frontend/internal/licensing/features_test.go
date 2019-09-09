package licensing

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/enterprise/pkg/license"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

func TestIsFeatureEnabled(t *testing.T) {
	check := func(t *testing.T, feature conf.Feature, licenseTags []string, wantEnabled bool) {
		t.Helper()
		got := isFeatureEnabled(license.Info{Tags: licenseTags}, feature)
		if got != wantEnabled {
			t.Errorf("got %v, want %v", got, wantEnabled)
		}
	}

	t.Run(string(conf.FeatureGuestUsers), func(t *testing.T) {
		check(t, conf.FeatureGuestUsers, EnterpriseStarterTags, false)
		check(t, conf.FeatureGuestUsers, EnterpriseTags, false)
		check(t, conf.FeatureGuestUsers, EnterprisePlusTags, false)
		check(t, conf.FeatureGuestUsers, EliteTags, true)
	})

	t.Run(string(FeatureACLs), func(t *testing.T) {
		check(t, FeatureACLs, EnterpriseStarterTags, false)
		check(t, FeatureACLs, EnterpriseTags, true)
		check(t, FeatureACLs, EnterprisePlusTags, true)
		check(t, FeatureACLs, EliteTags, true)
	})

	t.Run(string(FeatureExtensionRegistry), func(t *testing.T) {
		check(t, FeatureExtensionRegistry, EnterpriseStarterTags, false)
		check(t, FeatureExtensionRegistry, EnterpriseTags, false)
		check(t, FeatureExtensionRegistry, EnterprisePlusTags, false)
		check(t, FeatureExtensionRegistry, EliteTags, true)
		check(t, FeatureExtensionRegistry, []string{string(FeatureExtensionRegistry)}, true)
	})
}
