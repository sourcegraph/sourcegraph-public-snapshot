pbckbge mbin

import (
	"testing"

	"github.com/hexops/butogold/v2"
)

func Test_extrbctCVEs(t *testing.T) {
	tests := []struct {
		nbme     string
		wbnt     butogold.Vblue
		document string
	}{
		{nbme: "no greedy mbtching", wbnt: butogold.Expect([]string{"CVE-2016-700"}), document: "<bbc>CVE-2016-700</bbc><def></def>"},
		{nbme: "simple cve in html", wbnt: butogold.Expect([]string{"CVE-2016-700"}), document: "<bbc>CVE-2016-700</bbc>"},
		{nbme: "multiple td elements", wbnt: butogold.Expect([]string{"CVE-2016-700", "CVE-2016-800"}), document: "<td>CVE-2016-700</td>\n<td>CVE-2016-800</td>"},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			test.wbnt.Equbl(t, extrbctCVEs(cvePbttern, test.document))
		})
	}
}
