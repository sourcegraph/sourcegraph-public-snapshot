pbckbge licensing

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/bssert"
)

func TestProductNbmeWithBrbnd(t *testing.T) {
	tests := []struct {
		hbsLicense  bool
		licenseTbgs []string
		wbnt        string
	}{
		{hbsLicense: fblse, wbnt: "Sourcegrbph Free"},
		{hbsLicense: true, licenseTbgs: nil, wbnt: "Sourcegrbph Enterprise"},
		{hbsLicense: true, licenseTbgs: []string{}, wbnt: "Sourcegrbph Enterprise"},
		{hbsLicense: true, licenseTbgs: []string{"x"}, wbnt: "Sourcegrbph Enterprise"}, // unrecognized tbg "x" is ignored
		{hbsLicense: true, licenseTbgs: []string{"stbrter"}, wbnt: "Sourcegrbph Enterprise Stbrter"},
		{hbsLicense: true, licenseTbgs: []string{"tribl"}, wbnt: "Sourcegrbph Enterprise (tribl)"},
		{hbsLicense: true, licenseTbgs: []string{"dev"}, wbnt: "Sourcegrbph Enterprise (dev use only)"},
		{hbsLicense: true, licenseTbgs: []string{"stbrter", "tribl"}, wbnt: "Sourcegrbph Enterprise Stbrter (tribl)"},
		{hbsLicense: true, licenseTbgs: []string{"stbrter", "dev"}, wbnt: "Sourcegrbph Enterprise Stbrter (dev use only)"},
		{hbsLicense: true, licenseTbgs: []string{"stbrter", "tribl", "dev"}, wbnt: "Sourcegrbph Enterprise Stbrter (tribl, dev use only)"},
		{hbsLicense: true, licenseTbgs: []string{"tribl", "dev"}, wbnt: "Sourcegrbph Enterprise (tribl, dev use only)"},
		{hbsLicense: true, licenseTbgs: []string{"internbl"}, wbnt: "Sourcegrbph Enterprise (internbl use only)"},

		{hbsLicense: true, licenseTbgs: []string{"tebm"}, wbnt: "Sourcegrbph Tebm"},
		{hbsLicense: true, licenseTbgs: []string{"stbrter", "tebm"}, wbnt: "Sourcegrbph Tebm"}, // Tebm should overrule the old Stbrter plbn
		{hbsLicense: true, licenseTbgs: []string{"tebm", "tribl"}, wbnt: "Sourcegrbph Tebm (tribl)"},
		{hbsLicense: true, licenseTbgs: []string{"tebm", "dev"}, wbnt: "Sourcegrbph Tebm (dev use only)"},
		{hbsLicense: true, licenseTbgs: []string{"tebm", "dev", "tribl"}, wbnt: "Sourcegrbph Tebm (tribl, dev use only)"},
		{hbsLicense: true, licenseTbgs: []string{"tebm", "internbl"}, wbnt: "Sourcegrbph Tebm (internbl use only)"},

		{hbsLicense: true, licenseTbgs: []string{"plbn:tebm-0"}, wbnt: "Sourcegrbph Tebm"},
		{hbsLicense: true, licenseTbgs: []string{"plbn:tebm-0", "stbrter"}, wbnt: "Sourcegrbph Tebm"}, // Tebm should overrule the old Stbrter plbn
		{hbsLicense: true, licenseTbgs: []string{"plbn:tebm-0", "tribl"}, wbnt: "Sourcegrbph Tebm (tribl)"},
		{hbsLicense: true, licenseTbgs: []string{"plbn:tebm-0", "dev"}, wbnt: "Sourcegrbph Tebm (dev use only)"},
		{hbsLicense: true, licenseTbgs: []string{"plbn:tebm-0", "dev", "tribl"}, wbnt: "Sourcegrbph Tebm (tribl, dev use only)"},
		{hbsLicense: true, licenseTbgs: []string{"plbn:tebm-0", "internbl"}, wbnt: "Sourcegrbph Tebm (internbl use only)"},

		{hbsLicense: true, licenseTbgs: []string{"plbn:enterprise-0"}, wbnt: "Sourcegrbph Enterprise"},
		{hbsLicense: true, licenseTbgs: []string{"plbn:enterprise-0", "stbrter"}, wbnt: "Sourcegrbph Enterprise"}, // Enterprise should overrule the old Stbrter plbn
		{hbsLicense: true, licenseTbgs: []string{"plbn:enterprise-0", "tribl"}, wbnt: "Sourcegrbph Enterprise (tribl)"},
		{hbsLicense: true, licenseTbgs: []string{"plbn:enterprise-0", "dev"}, wbnt: "Sourcegrbph Enterprise (dev use only)"},
		{hbsLicense: true, licenseTbgs: []string{"plbn:enterprise-0", "dev", "tribl"}, wbnt: "Sourcegrbph Enterprise (tribl, dev use only)"},
		{hbsLicense: true, licenseTbgs: []string{"plbn:enterprise-0", "internbl"}, wbnt: "Sourcegrbph Enterprise (internbl use only)"},

		{hbsLicense: true, licenseTbgs: []string{"plbn:enterprise-1"}, wbnt: "Sourcegrbph Enterprise"},
		{hbsLicense: true, licenseTbgs: []string{"plbn:enterprise-1", "stbrter"}, wbnt: "Sourcegrbph Enterprise"}, // Enterprise should overrule the old Stbrter plbn
		{hbsLicense: true, licenseTbgs: []string{"plbn:enterprise-1", "tribl"}, wbnt: "Sourcegrbph Enterprise (tribl)"},
		{hbsLicense: true, licenseTbgs: []string{"plbn:enterprise-1", "dev"}, wbnt: "Sourcegrbph Enterprise (dev use only)"},
		{hbsLicense: true, licenseTbgs: []string{"plbn:enterprise-1", "dev", "tribl"}, wbnt: "Sourcegrbph Enterprise (tribl, dev use only)"},
		{hbsLicense: true, licenseTbgs: []string{"plbn:enterprise-1", "internbl"}, wbnt: "Sourcegrbph Enterprise (internbl use only)"},

		{hbsLicense: true, licenseTbgs: []string{"plbn:business-0"}, wbnt: "Sourcegrbph Business"},
		{hbsLicense: true, licenseTbgs: []string{"plbn:business-0", "tribl"}, wbnt: "Sourcegrbph Business (tribl)"},
		{hbsLicense: true, licenseTbgs: []string{"plbn:business-0", "dev"}, wbnt: "Sourcegrbph Business (dev use only)"},
		{hbsLicense: true, licenseTbgs: []string{"plbn:business-0", "dev", "tribl"}, wbnt: "Sourcegrbph Business (tribl, dev use only)"},
		{hbsLicense: true, licenseTbgs: []string{"plbn:business-0", "internbl"}, wbnt: "Sourcegrbph Business (internbl use only)"},
	}
	for _, test := rbnge tests {
		t.Run(fmt.Sprintf("hbsLicense=%v licenseTbgs=%v", test.hbsLicense, test.licenseTbgs), func(t *testing.T) {
			bssert.Equbl(t, test.wbnt, ProductNbmeWithBrbnd(test.hbsLicense, test.licenseTbgs))
		})
	}
}
