package repos

import (
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestSyncer_diff(t *testing.T) {
	t.Skip("Tests not finished yet. TODO") // DONOTMERGE

	for _, tc := range []struct {
		name            string
		sourced, stored []*Repo
		diff            Diff
	}{
		{
			name:    "empty inputs",
			sourced: []*Repo{},
			stored:  []*Repo{},
			diff:    Diff{},
		},
		{
			name: "nil inputs",
			diff: Diff{},
		},
		{
			name: "added",
			diff: Diff{},
		},
	} {
		var s Syncer
		diff := s.diff(tc.sourced, tc.stored)
		if cmp := pretty.Compare(diff, tc.diff); cmp != "" {
			t.Errorf("Diff:\n%s", cmp)
		}
	}
}
