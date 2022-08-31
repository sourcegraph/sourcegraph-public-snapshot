package licensing

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/license"
)

func TestCheckFeature(t *testing.T) {
	license := func(tags ...string) *Info { return &Info{Info: license.Info{Tags: tags}} }

	check := func(t *testing.T, feature Feature, info *Info, wantEnabled bool) {
		t.Helper()
		got := checkFeature(info, feature) == nil
		if got != wantEnabled {
			t.Errorf("got %v, want %v", got, wantEnabled)
		}
	}

	plan := func(p Plan) string {
		return "plan:" + string(p)
	}

	t.Run(string(FeatureSSO), func(t *testing.T) {
		check(t, FeatureSSO, nil, false)

		check(t, FeatureSSO, license("starter"), false)
		check(t, FeatureSSO, license("starter", string(FeatureSSO)), true)
		check(t, FeatureSSO, license(plan(oldEnterpriseStarter)), false)
		check(t, FeatureSSO, license(plan(oldEnterprise)), true)
		check(t, FeatureSSO, license(), true)

		check(t, FeatureSSO, license(plan(team)), true)
		check(t, FeatureSSO, license(plan(enterprise0)), true)

		check(t, FeatureSSO, license(plan(business0)), true)
		check(t, FeatureSSO, license(plan(enterprise1)), true)
	})

	t.Run(string(FeatureACLs), func(t *testing.T) {
		check(t, FeatureACLs, nil, false)

		check(t, FeatureACLs, license("starter"), false)
		check(t, FeatureACLs, license(plan(oldEnterpriseStarter)), false)
		check(t, FeatureACLs, license(plan(oldEnterprise)), true)
		check(t, FeatureACLs, license(), true)

		check(t, FeatureACLs, license(plan(team)), false)
		check(t, FeatureACLs, license(plan(enterprise0)), false)
		check(t, FeatureACLs, license(plan(enterprise0), string(FeatureACLs)), true)
	})

	t.Run(string(FeatureExtensionRegistry), func(t *testing.T) {
		check(t, FeatureExtensionRegistry, nil, false)

		check(t, FeatureExtensionRegistry, license("starter"), false)
		check(t, FeatureExtensionRegistry, license(plan(oldEnterpriseStarter)), false)
		check(t, FeatureExtensionRegistry, license(plan(oldEnterprise)), true)
		check(t, FeatureExtensionRegistry, license(), true)

		check(t, FeatureExtensionRegistry, license(plan(team)), false)
		check(t, FeatureExtensionRegistry, license(plan(enterprise0)), false)
		check(t, FeatureExtensionRegistry, license(plan(enterprise0), string(FeatureExtensionRegistry)), true)
	})

	t.Run(string(FeatureRemoteExtensionsAllowDisallow), func(t *testing.T) {
		check(t, FeatureRemoteExtensionsAllowDisallow, nil, false)

		check(t, FeatureRemoteExtensionsAllowDisallow, license("starter"), false)
		check(t, FeatureRemoteExtensionsAllowDisallow, license(plan(oldEnterpriseStarter)), false)
		check(t, FeatureRemoteExtensionsAllowDisallow, license(plan(oldEnterprise)), true)
		check(t, FeatureRemoteExtensionsAllowDisallow, license(), true)

		check(t, FeatureRemoteExtensionsAllowDisallow, license(plan(team)), false)
		check(t, FeatureRemoteExtensionsAllowDisallow, license(plan(enterprise0)), false)
		check(t, FeatureRemoteExtensionsAllowDisallow, license(plan(enterprise0), string(FeatureRemoteExtensionsAllowDisallow)), true)
	})

	t.Run(string(FeatureBranding), func(t *testing.T) {
		check(t, FeatureBranding, nil, false)

		check(t, FeatureBranding, license("starter"), false)
		check(t, FeatureBranding, license(plan(oldEnterpriseStarter)), false)
		check(t, FeatureBranding, license(plan(oldEnterprise)), true)
		check(t, FeatureBranding, license(), true)

		check(t, FeatureBranding, license(plan(team)), false)
		check(t, FeatureBranding, license(plan(enterprise0)), false)
		check(t, FeatureBranding, license(plan(enterprise0), string(FeatureBranding)), true)
	})

	testBatchChanges := func(feature Feature) func(*testing.T) {
		return func(t *testing.T) {
			check(t, feature, nil, false)

			check(t, feature, license("starter"), false)
			check(t, feature, license(plan(oldEnterpriseStarter)), false)
			check(t, feature, license(plan(oldEnterprise)), true)
			check(t, feature, license(), true)

			check(t, feature, license(plan(team)), false)
			check(t, feature, license(plan(enterprise0)), false)
			check(t, feature, license(plan(enterprise0), string(feature)), true)

			check(t, feature, license(plan(business0)), true)
			check(t, feature, license(plan(enterprise1)), true)
		}
	}

	// FeatureCampaigns is deprecated but should behave like BatchChanges.
	t.Run(string(FeatureCampaigns), testBatchChanges(FeatureCampaigns))
	t.Run(string(FeatureBatchChanges), testBatchChanges(FeatureBatchChanges))

	testCodeInsights := func(feature Feature) func(*testing.T) {
		return func(t *testing.T) {
			check(t, feature, nil, false)

			check(t, feature, license("starter"), false)
			check(t, feature, license(plan(oldEnterpriseStarter)), false)
			check(t, feature, license(plan(oldEnterprise)), true)
			check(t, feature, license(), true)

			check(t, feature, license(plan(team)), false)
			check(t, feature, license(plan(enterprise0)), false)
			check(t, feature, license(plan(enterprise0), string(feature)), true)
		}
	}
	// Code Insights
	t.Run(string(FeatureCodeInsights), testCodeInsights(FeatureCodeInsights))

	t.Run(string(FeatureMonitoring), func(t *testing.T) {
		check(t, FeatureMonitoring, nil, false)

		check(t, FeatureMonitoring, license("starter"), false)
		check(t, FeatureMonitoring, license(plan(oldEnterpriseStarter)), false)
		check(t, FeatureMonitoring, license(plan(oldEnterprise)), true)
		check(t, FeatureMonitoring, license(), true)

		check(t, FeatureMonitoring, license(plan(team)), false)
		check(t, FeatureMonitoring, license(plan(enterprise0)), false)
		check(t, FeatureMonitoring, license(plan(enterprise0), string(FeatureMonitoring)), true)
	})

	t.Run(string(FeatureBackupAndRestore), func(t *testing.T) {
		check(t, FeatureBackupAndRestore, nil, false)

		check(t, FeatureBackupAndRestore, license("starter"), false)
		check(t, FeatureBackupAndRestore, license(plan(oldEnterpriseStarter)), false)
		check(t, FeatureBackupAndRestore, license(plan(oldEnterprise)), true)
		check(t, FeatureBackupAndRestore, license(), true)

		check(t, FeatureBackupAndRestore, license(plan(team)), false)
		check(t, FeatureBackupAndRestore, license(plan(enterprise0)), false)
		check(t, FeatureBackupAndRestore, license(plan(enterprise0), string(FeatureBackupAndRestore)), true)
	})
}
