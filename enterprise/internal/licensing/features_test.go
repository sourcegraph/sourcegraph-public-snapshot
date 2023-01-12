package licensing

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/license"
)

func TestCheckFeature(t *testing.T) {
	license := func(tags ...string) *Info { return &Info{Info: license.Info{Tags: tags}} }

	check := func(t *testing.T, feature Feature, info *Info, wantEnabled bool) {
		t.Helper()
		if got := feature.Check(info) == nil; got != wantEnabled {
			t.Errorf("got enabled %v, want %v, for %q", got, wantEnabled, info)
		}
	}

	checkAs := func(t *testing.T, feature Feature, info *Info, wantEnabled bool, wantFeature Feature) {
		t.Helper()
		enabled := feature.Check(info) == nil
		if enabled != wantEnabled {
			t.Errorf("got enabled %v, want %v, for %q", enabled, wantEnabled, info)
		}
		if enabled {
			if cmp.Diff(feature, wantFeature) != "" {
				t.Errorf("got %v want %v, for %q", feature, wantFeature, info)
			}
		}
	}

	plan := func(p Plan) string {
		return "plan:" + string(p)
	}

	t.Run(string(FeatureSSO), func(t *testing.T) {
		check(t, FeatureSSO, nil, false)

		check(t, FeatureSSO, license("starter"), false)
		check(t, FeatureSSO, license("starter", string(FeatureSSO)), true)
		check(t, FeatureSSO, license(plan(PlanOldEnterpriseStarter)), false)
		check(t, FeatureSSO, license(plan(PlanOldEnterprise)), true)
		check(t, FeatureSSO, license(), true)

		check(t, FeatureSSO, license(plan(PlanTeam0)), true)
		check(t, FeatureSSO, license(plan(PlanEnterprise0)), true)

		check(t, FeatureSSO, license(plan(PlanBusiness0)), true)
		check(t, FeatureSSO, license(plan(PlanEnterprise1)), true)

		check(t, FeatureSSO, license(plan(PlanFree0)), true)
	})

	t.Run(string(FeatureExplicitPermissionsAPI), func(t *testing.T) {
		check(t, FeatureExplicitPermissionsAPI, nil, false)

		check(t, FeatureExplicitPermissionsAPI, license("starter"), false)
		check(t, FeatureExplicitPermissionsAPI, license("starter", string(FeatureExplicitPermissionsAPI)), true)
		check(t, FeatureExplicitPermissionsAPI, license(plan(PlanOldEnterpriseStarter)), false)
		check(t, FeatureExplicitPermissionsAPI, license(plan(PlanOldEnterprise)), true)
		check(t, FeatureExplicitPermissionsAPI, license(), true)

		check(t, FeatureExplicitPermissionsAPI, license(plan(PlanTeam0)), true)
		check(t, FeatureExplicitPermissionsAPI, license(plan(PlanEnterprise0)), true)

		check(t, FeatureExplicitPermissionsAPI, license(plan(PlanBusiness0)), false)
		check(t, FeatureExplicitPermissionsAPI, license(plan(PlanEnterprise1)), true)
	})

	t.Run(string(FeatureACLs), func(t *testing.T) {
		check(t, FeatureACLs, nil, false)

		check(t, FeatureACLs, license("starter"), false)
		check(t, FeatureACLs, license(plan(PlanOldEnterpriseStarter)), false)
		check(t, FeatureACLs, license(plan(PlanOldEnterprise)), true)
		check(t, FeatureACLs, license(), true)

		check(t, FeatureACLs, license(plan(PlanTeam0)), true)
		check(t, FeatureACLs, license(plan(PlanEnterprise0)), true)
		check(t, FeatureACLs, license(plan(PlanEnterprise0), string(FeatureACLs)), true)

		check(t, FeatureACLs, license(plan(PlanBusiness0)), true)
		check(t, FeatureACLs, license(plan(PlanEnterprise1)), true)
	})

	t.Run(string(FeatureExtensionRegistry), func(t *testing.T) {
		check(t, FeatureExtensionRegistry, nil, false)

		check(t, FeatureExtensionRegistry, license("starter"), false)
		check(t, FeatureExtensionRegistry, license(plan(PlanOldEnterpriseStarter)), false)
		check(t, FeatureExtensionRegistry, license(plan(PlanOldEnterprise)), true)
		check(t, FeatureExtensionRegistry, license(), true)

		check(t, FeatureExtensionRegistry, license(plan(PlanTeam0)), false)
		check(t, FeatureExtensionRegistry, license(plan(PlanEnterprise0)), false)
		check(t, FeatureExtensionRegistry, license(plan(PlanEnterprise0), string(FeatureExtensionRegistry)), true)
	})

	t.Run(string(FeatureRemoteExtensionsAllowDisallow), func(t *testing.T) {
		check(t, FeatureRemoteExtensionsAllowDisallow, nil, false)

		check(t, FeatureRemoteExtensionsAllowDisallow, license("starter"), false)
		check(t, FeatureRemoteExtensionsAllowDisallow, license(plan(PlanOldEnterpriseStarter)), false)
		check(t, FeatureRemoteExtensionsAllowDisallow, license(plan(PlanOldEnterprise)), true)
		check(t, FeatureRemoteExtensionsAllowDisallow, license(), true)

		check(t, FeatureRemoteExtensionsAllowDisallow, license(plan(PlanTeam0)), false)
		check(t, FeatureRemoteExtensionsAllowDisallow, license(plan(PlanEnterprise0)), false)
		check(t, FeatureRemoteExtensionsAllowDisallow, license(plan(PlanEnterprise0), string(FeatureRemoteExtensionsAllowDisallow)), true)
	})

	t.Run(string(FeatureBranding), func(t *testing.T) {
		check(t, FeatureBranding, nil, false)

		check(t, FeatureBranding, license("starter"), false)
		check(t, FeatureBranding, license(plan(PlanOldEnterpriseStarter)), false)
		check(t, FeatureBranding, license(plan(PlanOldEnterprise)), true)
		check(t, FeatureBranding, license(), true)

		check(t, FeatureBranding, license(plan(PlanTeam0)), false)
		check(t, FeatureBranding, license(plan(PlanEnterprise0)), false)
		check(t, FeatureBranding, license(plan(PlanEnterprise0), string(FeatureBranding)), true)
	})

	t.Run((&FeatureBatchChanges{}).FeatureName(), func(t *testing.T) {
		check(t, &FeatureBatchChanges{}, nil, false)

		checkAs(t, &FeatureBatchChanges{}, license("starter"), true, &FeatureBatchChanges{MaxNumChangesets: 10})
		checkAs(t, &FeatureBatchChanges{}, license(plan(PlanOldEnterpriseStarter)), true, &FeatureBatchChanges{MaxNumChangesets: 10})
		checkAs(t, &FeatureBatchChanges{}, license(plan(PlanOldEnterprise)), true, &FeatureBatchChanges{Unrestricted: true})
		checkAs(t, &FeatureBatchChanges{}, license(), true, &FeatureBatchChanges{Unrestricted: true})

		checkAs(t, &FeatureBatchChanges{}, license(plan(PlanTeam0)), true, &FeatureBatchChanges{MaxNumChangesets: 10})
		checkAs(t, &FeatureBatchChanges{}, license(plan(PlanEnterprise0)), true, &FeatureBatchChanges{MaxNumChangesets: 10})
		checkAs(t, &FeatureBatchChanges{}, license(plan(PlanEnterprise0), (&FeatureBatchChanges{}).FeatureName()), true, &FeatureBatchChanges{Unrestricted: true})

		checkAs(t, &FeatureBatchChanges{}, license(plan(PlanBusiness0)), true, &FeatureBatchChanges{Unrestricted: true})
		checkAs(t, &FeatureBatchChanges{}, license(plan(PlanEnterprise1)), true, &FeatureBatchChanges{Unrestricted: true})

		// Batch changes should be unrestricted if Campaigns is set.
		checkAs(t, &FeatureBatchChanges{}, license("starter", string(FeatureCampaigns)), true, &FeatureBatchChanges{Unrestricted: true})
	})

	testCodeInsights := func(feature Feature) func(*testing.T) {
		return func(t *testing.T) {
			check(t, feature, nil, false)

			check(t, feature, license("starter"), false)
			check(t, feature, license(plan(PlanOldEnterpriseStarter)), false)
			check(t, feature, license(plan(PlanOldEnterprise)), true)
			check(t, feature, license(), true)

			check(t, feature, license(plan(PlanTeam0)), false)
			check(t, feature, license(plan(PlanEnterprise0)), false)
			check(t, feature, license(plan(PlanEnterprise0), feature.FeatureName()), true)

			check(t, feature, license(plan(PlanBusiness0)), true)
			check(t, feature, license(plan(PlanEnterprise1)), true)
		}
	}
	// Code Insights
	t.Run(string(FeatureCodeInsights), testCodeInsights(FeatureCodeInsights))

	t.Run(string(FeatureMonitoring), func(t *testing.T) {
		check(t, FeatureMonitoring, nil, false)

		check(t, FeatureMonitoring, license("starter"), false)
		check(t, FeatureMonitoring, license(plan(PlanOldEnterpriseStarter)), false)
		check(t, FeatureMonitoring, license(plan(PlanOldEnterprise)), true)
		check(t, FeatureMonitoring, license(), true)

		check(t, FeatureMonitoring, license(plan(PlanTeam0)), false)
		check(t, FeatureMonitoring, license(plan(PlanEnterprise0)), false)
		check(t, FeatureMonitoring, license(plan(PlanEnterprise0), string(FeatureMonitoring)), true)
	})

	t.Run(string(FeatureBackupAndRestore), func(t *testing.T) {
		check(t, FeatureBackupAndRestore, nil, false)

		check(t, FeatureBackupAndRestore, license("starter"), false)
		check(t, FeatureBackupAndRestore, license(plan(PlanOldEnterpriseStarter)), false)
		check(t, FeatureBackupAndRestore, license(plan(PlanOldEnterprise)), true)
		check(t, FeatureBackupAndRestore, license(), true)

		check(t, FeatureBackupAndRestore, license(plan(PlanTeam0)), false)
		check(t, FeatureBackupAndRestore, license(plan(PlanEnterprise0)), false)
		check(t, FeatureBackupAndRestore, license(plan(PlanEnterprise0), string(FeatureBackupAndRestore)), true)
	})
}
