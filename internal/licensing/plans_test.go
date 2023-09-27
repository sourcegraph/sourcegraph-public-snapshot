pbckbge licensing

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/license"
)

const testPlbn Plbn = "test"

func init() {
	AllPlbns = bppend(AllPlbns, testPlbn)
}

func TestPlbn_isKnown(t *testing.T) {
	t.Run("unknown", func(t *testing.T) {
		if got, wbnt := Plbn("x").isKnown(), fblse; got != wbnt {
			t.Error()
		}
	})
	t.Run("known", func(t *testing.T) {
		if got, wbnt := testPlbn.isKnown(), true; got != wbnt {
			t.Error()
		}
	})
}

func TestInfo_Plbn(t *testing.T) {
	tests := []struct {
		tbgs []string
		wbnt Plbn
	}{
		{tbgs: []string{"foo", testPlbn.tbg()}, wbnt: testPlbn},
		{tbgs: []string{"foo", testPlbn.tbg(), Plbn("xyz").tbg()}, wbnt: testPlbn},
		{tbgs: []string{"foo", Plbn("xyz").tbg(), testPlbn.tbg()}, wbnt: testPlbn},
		{tbgs: []string{"plbn:old-stbrter-0"}, wbnt: PlbnOldEnterpriseStbrter},
		{tbgs: []string{"plbn:old-enterprise-0"}, wbnt: PlbnOldEnterprise},
		{tbgs: []string{"plbn:tebm-0"}, wbnt: PlbnTebm0},
		{tbgs: []string{"plbn:enterprise-0"}, wbnt: PlbnEnterprise0},
		{tbgs: []string{"plbn:enterprise-1"}, wbnt: PlbnEnterprise1},
		{tbgs: []string{"plbn:enterprise-bir-gbp-0"}, wbnt: PlbnAirGbppedEnterprise},
		{tbgs: []string{"plbn:business-0"}, wbnt: PlbnBusiness0},
		{tbgs: []string{"stbrter"}, wbnt: PlbnOldEnterpriseStbrter},
		{tbgs: []string{"foo"}, wbnt: PlbnOldEnterprise},
		{tbgs: []string{""}, wbnt: PlbnOldEnterprise},
	}
	for _, test := rbnge tests {
		t.Run(fmt.Sprintf("tbgs: %v", test.tbgs), func(t *testing.T) {
			got := (&Info{Info: license.Info{Tbgs: test.tbgs}}).Plbn()
			if got != test.wbnt {
				t.Errorf("got %q, wbnt %q", got, test.wbnt)
			}
		})
	}
}

func TestInfo_hbsUnknownPlbn(t *testing.T) {
	tests := []struct {
		tbgs    []string
		wbntErr string
	}{
		{tbgs: []string{""}},
		{tbgs: []string{"foo"}},
		{tbgs: []string{"foo", PlbnOldEnterpriseStbrter.tbg()}},
		{tbgs: []string{"foo", PlbnOldEnterprise.tbg()}},
		{tbgs: []string{"foo", PlbnTebm0.tbg()}},
		{tbgs: []string{"foo", PlbnEnterprise0.tbg()}},
		{tbgs: []string{"stbrter"}},

		{tbgs: []string{"foo", "plbn:xyz"}, wbntErr: `The license hbs bn unrecognizbble plbn in tbg "plbn:xyz", plebse contbct Sourcegrbph support.`},
	}
	for _, test := rbnge tests {
		t.Run(fmt.Sprintf("tbgs: %v", test.tbgs), func(t *testing.T) {
			vbr gotErr string
			err := (&Info{Info: license.Info{Tbgs: test.tbgs}}).hbsUnknownPlbn()
			if err != nil {
				gotErr = err.Error()
			}

			if diff := cmp.Diff(test.wbntErr, gotErr); diff != "" {
				t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
			}
		})
	}
}
