package backend

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/zoekt"
)

func TestHoistTopScore(t *testing.T) {
	cases := []struct {
		inputScore []float64
		wantScore  []float64
		n          int
	}{
		{
			inputScore: []float64{1, 2, 33, 99, 5, 6},
			wantScore:  []float64{99, 1, 2, 33, 5, 6},
			n:          5,
		},
		{
			inputScore: []float64{99, 1, 2, 3, 5, 6},
			wantScore:  []float64{99, 1, 2, 3, 5, 6},
			n:          5,
		},
		{
			inputScore: []float64{1, 99, 2, 3, 5, 6},
			wantScore:  []float64{99, 1, 2, 3, 5, 6},
			n:          5,
		},

		{
			inputScore: []float64{1, 2, 3, 5, 6, 99},
			wantScore:  []float64{6, 1, 2, 3, 5, 99},
			n:          5,
		},
		{
			inputScore: []float64{1, 99, 2},
			wantScore:  []float64{99, 1, 2},
			n:          5,
		},
		{
			inputScore: []float64{1, 99, 2},
			wantScore:  []float64{1, 99, 2},
			n:          1,
		},

		{
			inputScore: []float64{1, 99, 2},
			wantScore:  []float64{1, 99, 2},
			n:          0,
		},
	}

	for _, tt := range cases {
		t.Run("", func(t *testing.T) {
			fileMatches := mkFileMatch(tt.inputScore)
			hoistMaxScore(fileMatches, tt.n)

			haveScore := []float64{}
			for _, fm := range fileMatches {
				haveScore = append(haveScore, fm.Score)
			}

			if d := cmp.Diff(tt.wantScore, haveScore); d != "" {
				t.Fatalf("-want, +got:\n%s", d)
			}
		})
	}

	hoistMaxScore(nil, 5)
	hoistMaxScore([]zoekt.FileMatch{}, 5)
}

func mkFileMatch(scores []float64) []zoekt.FileMatch {
	fm := make([]zoekt.FileMatch, 0, len(scores))
	for _, score := range scores {
		fm = append(fm, zoekt.FileMatch{Score: score})
	}

	return fm
}
