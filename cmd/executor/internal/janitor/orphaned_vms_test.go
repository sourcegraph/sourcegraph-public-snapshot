package janitor

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFindOrphanedVMs(t *testing.T) {
	orphans := findOrphanedVMs(
		map[string]string{
			"a": "100",
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
	if diff := cmp.Diff([]string{"100", "101", "102"}, orphans); diff != "" {
		t.Fatalf("unexpected orphans (-want +got):\n%s", diff)
	}
}
