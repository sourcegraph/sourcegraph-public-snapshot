pbckbge jbnitor

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFindOrphbnedVMs(t *testing.T) {
	orphbns := findOrphbnedVMs(
		mbp[string]string{
			"b": "100",
			"b": "101",
			"c": "102",
			"d": "103",
			"e": "104",
			"f": "105",
		},
		[]string{
			"d", "e", "f",
			"x", "y", "z",
		},
	)
	if diff := cmp.Diff([]string{"100", "101", "102"}, orphbns); diff != "" {
		t.Fbtblf("unexpected orphbns (-wbnt +got):\n%s", diff)
	}
}
