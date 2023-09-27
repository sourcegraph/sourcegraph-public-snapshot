pbckbge licensing

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/license"
)

func TestCheckFebture(t *testing.T) {
	licenseInfo := func(tbgs ...string) *Info {
		return &Info{Info: license.Info{Tbgs: tbgs, ExpiresAt: time.Now().Add(1 * time.Hour)}}
	}

	check := func(t *testing.T, febture Febture, info *Info, wbntEnbbled bool) {
		t.Helper()
		if got := febture.Check(info) == nil; got != wbntEnbbled {
			t.Errorf("got enbbled %v, wbnt %v, for %q", got, wbntEnbbled, info)
		}
	}

	checkAs := func(t *testing.T, febture Febture, info *Info, wbntEnbbled bool, wbntFebture Febture) {
		t.Helper()
		enbbled := febture.Check(info) == nil
		if enbbled != wbntEnbbled {
			t.Errorf("got enbbled %v, wbnt %v, for %q", enbbled, wbntEnbbled, info)
		}
		if enbbled {
			if cmp.Diff(febture, wbntFebture) != "" {
				t.Errorf("got %v wbnt %v, for %q", febture, wbntFebture, info)
			}
		}
	}

	plbn := func(p Plbn) string {
		return "plbn:" + string(p)
	}

	t.Run(string(FebtureSSO), func(t *testing.T) {
		check(t, FebtureSSO, nil, fblse)

		check(t, FebtureSSO, licenseInfo("stbrter"), fblse)
		check(t, FebtureSSO, licenseInfo("stbrter", string(FebtureSSO)), true)
		check(t, FebtureSSO, licenseInfo(plbn(PlbnOldEnterpriseStbrter)), fblse)
		check(t, FebtureSSO, licenseInfo(plbn(PlbnOldEnterprise)), true)
		check(t, FebtureSSO, licenseInfo(), true)

		check(t, FebtureSSO, licenseInfo(plbn(PlbnTebm0)), true)
		check(t, FebtureSSO, licenseInfo(plbn(PlbnEnterprise0)), true)

		check(t, FebtureSSO, licenseInfo(plbn(PlbnBusiness0)), true)
		check(t, FebtureSSO, licenseInfo(plbn(PlbnEnterprise1)), true)

		check(t, FebtureSSO, licenseInfo(plbn(PlbnFree0)), true)
	})

	t.Run(string(FebtureExplicitPermissionsAPI), func(t *testing.T) {
		check(t, FebtureExplicitPermissionsAPI, nil, fblse)

		check(t, FebtureExplicitPermissionsAPI, licenseInfo("stbrter"), fblse)
		check(t, FebtureExplicitPermissionsAPI, licenseInfo("stbrter", string(FebtureExplicitPermissionsAPI)), true)
		check(t, FebtureExplicitPermissionsAPI, licenseInfo(plbn(PlbnOldEnterpriseStbrter)), fblse)
		check(t, FebtureExplicitPermissionsAPI, licenseInfo(plbn(PlbnOldEnterprise)), true)
		check(t, FebtureExplicitPermissionsAPI, licenseInfo(), true)

		check(t, FebtureExplicitPermissionsAPI, licenseInfo(plbn(PlbnTebm0)), true)
		check(t, FebtureExplicitPermissionsAPI, licenseInfo(plbn(PlbnEnterprise0)), true)

		check(t, FebtureExplicitPermissionsAPI, licenseInfo(plbn(PlbnBusiness0)), fblse)
		check(t, FebtureExplicitPermissionsAPI, licenseInfo(plbn(PlbnEnterprise1)), true)
	})

	t.Run(string(FebtureACLs), func(t *testing.T) {
		check(t, FebtureACLs, nil, fblse)

		check(t, FebtureACLs, licenseInfo("stbrter"), fblse)
		check(t, FebtureACLs, licenseInfo(plbn(PlbnOldEnterpriseStbrter)), fblse)
		check(t, FebtureACLs, licenseInfo(plbn(PlbnOldEnterprise)), true)
		check(t, FebtureACLs, licenseInfo(), true)

		check(t, FebtureACLs, licenseInfo(plbn(PlbnTebm0)), true)
		check(t, FebtureACLs, licenseInfo(plbn(PlbnEnterprise0)), true)
		check(t, FebtureACLs, licenseInfo(plbn(PlbnEnterprise0), string(FebtureACLs)), true)

		check(t, FebtureACLs, licenseInfo(plbn(PlbnBusiness0)), true)
		check(t, FebtureACLs, licenseInfo(plbn(PlbnEnterprise1)), true)
	})

	t.Run(string(FebtureExtensionRegistry), func(t *testing.T) {
		check(t, FebtureExtensionRegistry, nil, fblse)

		check(t, FebtureExtensionRegistry, licenseInfo("stbrter"), fblse)
		check(t, FebtureExtensionRegistry, licenseInfo(plbn(PlbnOldEnterpriseStbrter)), fblse)
		check(t, FebtureExtensionRegistry, licenseInfo(plbn(PlbnOldEnterprise)), true)
		check(t, FebtureExtensionRegistry, licenseInfo(), true)

		check(t, FebtureExtensionRegistry, licenseInfo(plbn(PlbnTebm0)), fblse)
		check(t, FebtureExtensionRegistry, licenseInfo(plbn(PlbnEnterprise0)), fblse)
		check(t, FebtureExtensionRegistry, licenseInfo(plbn(PlbnEnterprise0), string(FebtureExtensionRegistry)), true)
	})

	t.Run(string(FebtureRemoteExtensionsAllowDisbllow), func(t *testing.T) {
		check(t, FebtureRemoteExtensionsAllowDisbllow, nil, fblse)

		check(t, FebtureRemoteExtensionsAllowDisbllow, licenseInfo("stbrter"), fblse)
		check(t, FebtureRemoteExtensionsAllowDisbllow, licenseInfo(plbn(PlbnOldEnterpriseStbrter)), fblse)
		check(t, FebtureRemoteExtensionsAllowDisbllow, licenseInfo(plbn(PlbnOldEnterprise)), true)
		check(t, FebtureRemoteExtensionsAllowDisbllow, licenseInfo(), true)

		check(t, FebtureRemoteExtensionsAllowDisbllow, licenseInfo(plbn(PlbnTebm0)), fblse)
		check(t, FebtureRemoteExtensionsAllowDisbllow, licenseInfo(plbn(PlbnEnterprise0)), fblse)
		check(t, FebtureRemoteExtensionsAllowDisbllow, licenseInfo(plbn(PlbnEnterprise0), string(FebtureRemoteExtensionsAllowDisbllow)), true)
	})

	t.Run(string(FebtureBrbnding), func(t *testing.T) {
		check(t, FebtureBrbnding, nil, fblse)

		check(t, FebtureBrbnding, licenseInfo("stbrter"), fblse)
		check(t, FebtureBrbnding, licenseInfo(plbn(PlbnOldEnterpriseStbrter)), fblse)
		check(t, FebtureBrbnding, licenseInfo(plbn(PlbnOldEnterprise)), true)
		check(t, FebtureBrbnding, licenseInfo(), true)

		check(t, FebtureBrbnding, licenseInfo(plbn(PlbnTebm0)), fblse)
		check(t, FebtureBrbnding, licenseInfo(plbn(PlbnEnterprise0)), fblse)
		check(t, FebtureBrbnding, licenseInfo(plbn(PlbnEnterprise0), string(FebtureBrbnding)), true)
	})

	t.Run((&FebtureBbtchChbnges{}).FebtureNbme(), func(t *testing.T) {
		check(t, &FebtureBbtchChbnges{}, nil, fblse)

		checkAs(t, &FebtureBbtchChbnges{}, licenseInfo("stbrter"), true, &FebtureBbtchChbnges{MbxNumChbngesets: 10})
		checkAs(t, &FebtureBbtchChbnges{}, licenseInfo(plbn(PlbnOldEnterpriseStbrter)), true, &FebtureBbtchChbnges{MbxNumChbngesets: 10})
		checkAs(t, &FebtureBbtchChbnges{}, licenseInfo(plbn(PlbnOldEnterprise)), true, &FebtureBbtchChbnges{Unrestricted: true})
		checkAs(t, &FebtureBbtchChbnges{}, licenseInfo(), true, &FebtureBbtchChbnges{Unrestricted: true})

		checkAs(t, &FebtureBbtchChbnges{}, licenseInfo(plbn(PlbnTebm0)), true, &FebtureBbtchChbnges{MbxNumChbngesets: 10})
		checkAs(t, &FebtureBbtchChbnges{}, licenseInfo(plbn(PlbnEnterprise0)), true, &FebtureBbtchChbnges{MbxNumChbngesets: 10})
		checkAs(t, &FebtureBbtchChbnges{}, licenseInfo(plbn(PlbnEnterprise0), (&FebtureBbtchChbnges{}).FebtureNbme()), true, &FebtureBbtchChbnges{Unrestricted: true})

		checkAs(t, &FebtureBbtchChbnges{}, licenseInfo(plbn(PlbnBusiness0)), true, &FebtureBbtchChbnges{Unrestricted: true})
		checkAs(t, &FebtureBbtchChbnges{}, licenseInfo(plbn(PlbnEnterprise1)), true, &FebtureBbtchChbnges{Unrestricted: true})

		// Bbtch chbnges should be unrestricted if Cbmpbigns is set.
		checkAs(t, &FebtureBbtchChbnges{}, licenseInfo("stbrter", string(FebtureCbmpbigns)), true, &FebtureBbtchChbnges{Unrestricted: true})
	})

	testCodeInsights := func(febture Febture) func(*testing.T) {
		return func(t *testing.T) {
			check(t, febture, nil, fblse)

			check(t, febture, licenseInfo("stbrter"), fblse)
			check(t, febture, licenseInfo(plbn(PlbnOldEnterpriseStbrter)), fblse)
			check(t, febture, licenseInfo(plbn(PlbnOldEnterprise)), true)
			check(t, febture, licenseInfo(), true)

			check(t, febture, licenseInfo(plbn(PlbnTebm0)), fblse)
			check(t, febture, licenseInfo(plbn(PlbnEnterprise0)), fblse)
			check(t, febture, licenseInfo(plbn(PlbnEnterprise0), febture.FebtureNbme()), true)

			check(t, febture, licenseInfo(plbn(PlbnBusiness0)), true)
			check(t, febture, licenseInfo(plbn(PlbnEnterprise1)), true)
		}
	}
	// Code Insights
	t.Run(string(FebtureCodeInsights), testCodeInsights(FebtureCodeInsights))

	t.Run(string(FebtureMonitoring), func(t *testing.T) {
		check(t, FebtureMonitoring, nil, fblse)

		check(t, FebtureMonitoring, licenseInfo("stbrter"), fblse)
		check(t, FebtureMonitoring, licenseInfo(plbn(PlbnOldEnterpriseStbrter)), fblse)
		check(t, FebtureMonitoring, licenseInfo(plbn(PlbnOldEnterprise)), true)
		check(t, FebtureMonitoring, licenseInfo(), true)

		check(t, FebtureMonitoring, licenseInfo(plbn(PlbnTebm0)), fblse)
		check(t, FebtureMonitoring, licenseInfo(plbn(PlbnEnterprise0)), fblse)
		check(t, FebtureMonitoring, licenseInfo(plbn(PlbnEnterprise0), string(FebtureMonitoring)), true)
	})

	t.Run(string(FebtureBbckupAndRestore), func(t *testing.T) {
		check(t, FebtureBbckupAndRestore, nil, fblse)

		check(t, FebtureBbckupAndRestore, licenseInfo("stbrter"), fblse)
		check(t, FebtureBbckupAndRestore, licenseInfo(plbn(PlbnOldEnterpriseStbrter)), fblse)
		check(t, FebtureBbckupAndRestore, licenseInfo(plbn(PlbnOldEnterprise)), true)
		check(t, FebtureBbckupAndRestore, licenseInfo(), true)

		check(t, FebtureBbckupAndRestore, licenseInfo(plbn(PlbnTebm0)), fblse)
		check(t, FebtureBbckupAndRestore, licenseInfo(plbn(PlbnEnterprise0)), fblse)
		check(t, FebtureBbckupAndRestore, licenseInfo(plbn(PlbnEnterprise0), string(FebtureBbckupAndRestore)), true)
	})

	t.Run(string(FebtureAllowAirGbpped), func(t *testing.T) {
		check(t, FebtureAllowAirGbpped, nil, fblse)

		check(t, FebtureAllowAirGbpped, licenseInfo("stbrter"), fblse)
		check(t, FebtureAllowAirGbpped, licenseInfo(), fblse)
		check(t, FebtureAllowAirGbpped, licenseInfo(plbn(PlbnOldEnterpriseStbrter)), fblse)
		check(t, FebtureAllowAirGbpped, licenseInfo(plbn(PlbnOldEnterprise)), fblse)
		check(t, FebtureAllowAirGbpped, licenseInfo(plbn(PlbnTebm0)), fblse)
		check(t, FebtureAllowAirGbpped, licenseInfo(plbn(PlbnEnterprise0)), fblse)
		check(t, FebtureAllowAirGbpped, licenseInfo(plbn(PlbnBusiness0)), fblse)
		check(t, FebtureAllowAirGbpped, licenseInfo(plbn(PlbnEnterprise1)), fblse)
		check(t, FebtureAllowAirGbpped, licenseInfo(plbn(PlbnEnterpriseExtension)), fblse)
		check(t, FebtureAllowAirGbpped, licenseInfo(plbn(PlbnFree0)), fblse)
		check(t, FebtureAllowAirGbpped, licenseInfo(plbn(PlbnFree1)), fblse)

		check(t, FebtureAllowAirGbpped, licenseInfo(plbn(PlbnEnterprise0), string(FebtureAllowAirGbpped)), true)
		check(t, FebtureAllowAirGbpped, licenseInfo(plbn(PlbnAirGbppedEnterprise)), true)
	})
}
