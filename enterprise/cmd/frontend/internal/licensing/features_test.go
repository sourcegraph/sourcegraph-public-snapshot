package licensing

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/license"
)

func TestIsFeatureEnabled(t *testing.T) {
	check := func(t *testing.T, feature Feature, licenseTags []string, wantEnabled bool) {
		t.Helper()
		got := isFeatureEnabled(license.Info{Tags: licenseTags}, feature)
		if got != wantEnabled {
			t.Errorf("got %v, want %v", got, wantEnabled)
		}
	}

	t.Run(string(FeatureACLs), func(t *testing.T) {
		check(t, FeatureACLs, EnterpriseStarterTags, false)
		check(t, FeatureACLs, EnterpriseTags, true)
	})

	t.Run(string(FeatureExtensionRegistry), func(t *testing.T) {
		check(t, FeatureExtensionRegistry, EnterpriseStarterTags, false)
		check(t, FeatureExtensionRegistry, EnterpriseTags, true)
		check(t, FeatureExtensionRegistry, []string{string(FeatureExtensionRegistry)}, true)
	})
}
