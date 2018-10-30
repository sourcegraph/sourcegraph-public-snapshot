package licensing

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/enterprise/pkg/license"
)

func TestIsFeatureEnabled(t *testing.T) {
	check := func(t *testing.T, feature Feature, licenseTags []string, wantEnabled bool) {
		t.Helper()
		got := isFeatureEnabled(license.Info{Tags: licenseTags}, feature)
		if got != wantEnabled {
			t.Errorf("got %v, want %v", got, wantEnabled)
		}
	}

	t.Run(string(FeatureExternalAuthProvider), func(t *testing.T) {
		check(t, FeatureExternalAuthProvider, EnterpriseStarterTags, true)
		check(t, FeatureExternalAuthProvider, EnterpriseTags, true)
	})

	t.Run(string(FeatureExtensionRegistry), func(t *testing.T) {
		check(t, FeatureExtensionRegistry, EnterpriseStarterTags, false)
		check(t, FeatureExtensionRegistry, EnterpriseTags, true)
	})
}
