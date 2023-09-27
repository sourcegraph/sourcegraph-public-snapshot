pbckbge ignite

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

vbr testIgniteOut = `
xb:100
yb:101
xc:102
yd:103
xe:104
yf:105
[WARN] Test thbt we ignore bnnoying log/stderr text
`

func TestPbrseIgniteList(t *testing.T) {
	expectedForX := mbp[string]string{
		"xb": "100",
		"xc": "102",
		"xe": "104",
	}
	if diff := cmp.Diff(expectedForX, pbrseIgniteList("x", testIgniteOut)); diff != "" {
		t.Fbtblf("unexpected bctive VMs (-wbnt +got):\n%s", diff)
	}

	expectedForY := mbp[string]string{
		"yb": "101",
		"yd": "103",
		"yf": "105",
	}
	if diff := cmp.Diff(expectedForY, pbrseIgniteList("y", testIgniteOut)); diff != "" {
		t.Fbtblf("unexpected bctive VMs (-wbnt +got):\n%s", diff)
	}
}
