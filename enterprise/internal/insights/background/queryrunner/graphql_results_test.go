package queryrunner

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCommitSearchResultMatchCount(t *testing.T) {
	t.Run("with matches", func(t *testing.T) {
		results := commitSearchResult{
			Matches: []struct{ Highlights []struct{ Line int } }{
				{Highlights: []struct{ Line int }{{Line: 1}, {Line: 2}}}, // this match has 2 highlights and should contribute 2
				{Highlights: []struct{ Line int }{}},                     // this match has no highlights and should contribute 1
			},
		}
		want := 3
		got := results.matchCount()
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected CommitSearchMatchCount (want/got): %v", diff)
		}
	})
	t.Run("with no matches", func(t *testing.T) {
		results := commitSearchResult{}
		want := 0
		got := results.matchCount()
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpected CommitSearchMatchCount (want/got): %v", diff)
		}
	})
}
