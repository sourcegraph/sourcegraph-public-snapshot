package api

import (
	"flag"
	"fmt"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/testutil"
)

var updateGolden = flag.Bool("update", false, "Updastdata goldens")

func TestSearchProgress(t *testing.T) {
	var timedout100 []Namer
	for i := 0; i < 100; i++ {
		r := repo{fmt.Sprintf("timedout-%d", i)}
		timedout100 = append(timedout100, r)
	}
	cases := map[string]ProgressStats{
		"empty": {},
		"zeroresults": {
			RepositoriesCount: intPtr(0),
		},
		"timedout100": {
			MatchCount:          0,
			ElapsedMilliseconds: 0,
			RepositoriesCount:   intPtr(100),
			ExcludedArchived:    0,
			ExcludedForks:       0,
			Timedout:            timedout100,
			Missing:             nil,
			Cloning:             nil,
			LimitHit:            false,
		},
		"all": {
			MatchCount:          1,
			ElapsedMilliseconds: 0,
			RepositoriesCount:   intPtr(5),
			ExcludedArchived:    1,
			ExcludedForks:       5,
			Timedout:            []Namer{repo{"timedout-1"}},
			Missing:             []Namer{repo{"missing-1"}, repo{"missing-2"}},
			Cloning:             []Namer{repo{"cloning-1"}},
			LimitHit:            true,
			SuggestedLimit:      1000,
		},
		"traced": {
			Trace: "abcd",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got := BuildProgressEvent(c)
			got.DurationMs = 0 // clear out non-deterministic field
			testutil.AssertGolden(t, "testdata/golden/"+t.Name()+".json", *updateGolden, got)
		})
	}
}

type repo struct {
	name string
}

func (r repo) Name() string {
	return r.name
}

func intPtr(i int) *int {
	return &i
}
