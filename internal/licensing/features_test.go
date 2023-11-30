package licensing

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/license"
)

func TestCheckFeature(t *testing.T) {
	licenseInfo := func(tags ...string) *Info {
		return &Info{Info: license.Info{Tags: tags, ExpiresAt: time.Now().Add(1 * time.Hour)}}
	}

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

		check(t, FeatureSSO, licenseInfo("starter"), false)
		check(t, FeatureSSO, licenseInfo("starter", string(FeatureSSO)), true)
		check(t, FeatureSSO, licenseInfo(plan(PlanOldEnterpriseStarter)), false)
		check(t, FeatureSSO, licenseInfo(plan(PlanOldEnterprise)), true)
		check(t, FeatureSSO, licenseInfo(), true)

		check(t, FeatureSSO, licenseInfo(plan(PlanTeam0)), true)
		check(t, FeatureSSO, licenseInfo(plan(PlanEnterprise0)), true)

		check(t, FeatureSSO, licenseInfo(plan(PlanBusiness0)), true)
		check(t, FeatureSSO, licenseInfo(plan(PlanEnterprise1)), true)

		check(t, FeatureSSO, licenseInfo(plan(PlanFree0)), true)
	})

	t.Run(string(FeatureExplicitPermissionsAPI), func(t *testing.T) {
		check(t, FeatureExplicitPermissionsAPI, nil, false)

		check(t, FeatureExplicitPermissionsAPI, licenseInfo("starter"), false)
		check(t, FeatureExplicitPermissionsAPI, licenseInfo("starter", string(FeatureExplicitPermissionsAPI)), true)
		check(t, FeatureExplicitPermissionsAPI, licenseInfo(plan(PlanOldEnterpriseStarter)), false)
		check(t, FeatureExplicitPermissionsAPI, licenseInfo(plan(PlanOldEnterprise)), true)
		check(t, FeatureExplicitPermissionsAPI, licenseInfo(), true)

		check(t, FeatureExplicitPermissionsAPI, licenseInfo(plan(PlanTeam0)), true)
		check(t, FeatureExplicitPermissionsAPI, licenseInfo(plan(PlanEnterprise0)), true)

		check(t, FeatureExplicitPermissionsAPI, licenseInfo(plan(PlanBusiness0)), false)
		check(t, FeatureExplicitPermissionsAPI, licenseInfo(plan(PlanEnterprise1)), true)
	})

	t.Run(string(FeatureACLs), func(t *testing.T) {
		check(t, FeatureACLs, nil, false)

		check(t, FeatureACLs, licenseInfo("starter"), false)
		check(t, FeatureACLs, licenseInfo(plan(PlanOldEnterpriseStarter)), false)
		check(t, FeatureACLs, licenseInfo(plan(PlanOldEnterprise)), true)
		check(t, FeatureACLs, licenseInfo(), true)

		check(t, FeatureACLs, licenseInfo(plan(PlanTeam0)), true)
		check(t, FeatureACLs, licenseInfo(plan(PlanEnterprise0)), true)
		check(t, FeatureACLs, licenseInfo(plan(PlanEnterprise0), string(FeatureACLs)), true)

		check(t, FeatureACLs, licenseInfo(plan(PlanBusiness0)), true)
		check(t, FeatureACLs, licenseInfo(plan(PlanEnterprise1)), true)
	})

	t.Run((&FeatureBatchChanges{}).FeatureName(), func(t *testing.T) {
		check(t, &FeatureBatchChanges{}, nil, false)

		checkAs(t, &FeatureBatchChanges{}, licenseInfo("starter"), true, &FeatureBatchChanges{MaxNumChangesets: 10})
		checkAs(t, &FeatureBatchChanges{}, licenseInfo(plan(PlanOldEnterpriseStarter)), true, &FeatureBatchChanges{MaxNumChangesets: 10})
		checkAs(t, &FeatureBatchChanges{}, licenseInfo(plan(PlanOldEnterprise)), true, &FeatureBatchChanges{Unrestricted: true})
		checkAs(t, &FeatureBatchChanges{}, licenseInfo(), true, &FeatureBatchChanges{Unrestricted: true})

		checkAs(t, &FeatureBatchChanges{}, licenseInfo(plan(PlanTeam0)), true, &FeatureBatchChanges{MaxNumChangesets: 10})
		checkAs(t, &FeatureBatchChanges{}, licenseInfo(plan(PlanEnterprise0)), true, &FeatureBatchChanges{MaxNumChangesets: 10})
		checkAs(t, &FeatureBatchChanges{}, licenseInfo(plan(PlanEnterprise0), (&FeatureBatchChanges{}).FeatureName()), true, &FeatureBatchChanges{Unrestricted: true})

		checkAs(t, &FeatureBatchChanges{}, licenseInfo(plan(PlanBusiness0)), true, &FeatureBatchChanges{Unrestricted: true})
		checkAs(t, &FeatureBatchChanges{}, licenseInfo(plan(PlanEnterprise1)), true, &FeatureBatchChanges{Unrestricted: true})
	})

	testCodeInsights := func(feature Feature) func(*testing.T) {
		return func(t *testing.T) {
			check(t, feature, nil, false)

			check(t, feature, licenseInfo("starter"), false)
			check(t, feature, licenseInfo(plan(PlanOldEnterpriseStarter)), false)
			check(t, feature, licenseInfo(plan(PlanOldEnterprise)), true)
			check(t, feature, licenseInfo(), true)

			check(t, feature, licenseInfo(plan(PlanTeam0)), false)
			check(t, feature, licenseInfo(plan(PlanEnterprise0)), false)
			check(t, feature, licenseInfo(plan(PlanEnterprise0), feature.FeatureName()), true)

			check(t, feature, licenseInfo(plan(PlanBusiness0)), true)
			check(t, feature, licenseInfo(plan(PlanEnterprise1)), true)
		}
	}
	// Code Insights
	t.Run(string(FeatureCodeInsights), testCodeInsights(FeatureCodeInsights))

	t.Run(string(FeatureAllowAirGapped), func(t *testing.T) {
		check(t, FeatureAllowAirGapped, nil, false)

		check(t, FeatureAllowAirGapped, licenseInfo("starter"), false)
		check(t, FeatureAllowAirGapped, licenseInfo(), false)
		check(t, FeatureAllowAirGapped, licenseInfo(plan(PlanOldEnterpriseStarter)), false)
		check(t, FeatureAllowAirGapped, licenseInfo(plan(PlanOldEnterprise)), false)
		check(t, FeatureAllowAirGapped, licenseInfo(plan(PlanTeam0)), false)
		check(t, FeatureAllowAirGapped, licenseInfo(plan(PlanEnterprise0)), false)
		check(t, FeatureAllowAirGapped, licenseInfo(plan(PlanBusiness0)), false)
		check(t, FeatureAllowAirGapped, licenseInfo(plan(PlanEnterprise1)), false)
		check(t, FeatureAllowAirGapped, licenseInfo(plan(PlanEnterpriseExtension)), false)
		check(t, FeatureAllowAirGapped, licenseInfo(plan(PlanFree0)), false)
		check(t, FeatureAllowAirGapped, licenseInfo(plan(PlanFree1)), false)

		check(t, FeatureAllowAirGapped, licenseInfo(plan(PlanEnterprise0), string(FeatureAllowAirGapped)), true)
		check(t, FeatureAllowAirGapped, licenseInfo(plan(PlanAirGappedEnterprise)), true)
	})
}
