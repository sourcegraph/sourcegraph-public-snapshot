package ignite

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

var testIgniteOut = `
xa:100
yb:101
xc:102
yd:103
xe:104
yf:105
[WARN] Test that we ignore annoying log/stderr text
`

func TestParseIgniteList(t *testing.T) {
	expectedForX := map[string]string{
		"xa": "100",
		"xc": "102",
		"xe": "104",
	}
	if diff := cmp.Diff(expectedForX, parseIgniteList("x", testIgniteOut)); diff != "" {
		t.Fatalf("unexpected active VMs (-want +got):\n%s", diff)
	}

	expectedForY := map[string]string{
		"yb": "101",
		"yd": "103",
		"yf": "105",
	}
	if diff := cmp.Diff(expectedForY, parseIgniteList("y", testIgniteOut)); diff != "" {
		t.Fatalf("unexpected active VMs (-want +got):\n%s", diff)
	}
}
