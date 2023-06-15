package api

import (
	"flag"
	"fmt"
	"math"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

var updateGolden = flag.Bool("update", false, "Updastdata goldens")

func TestSearchProgress(t *testing.T) {
	namer := func(ids []api.RepoID) (names []api.RepoName) {
		for _, id := range ids {
			names = append(names, api.RepoName(fmt.Sprintf("repo-%d", id)))
		}
		return names
	}

	var timedout100 []api.RepoID
	for id := api.RepoID(1); id <= 100; id++ {
		timedout100 = append(timedout100, id)
	}
	cases := map[string]ProgressStats{
		"empty": {},
		"zeroresults": {
			RepositoriesCount: pointers.Ptr(0),
		},
		"timedout100": {
			MatchCount:          0,
			ElapsedMilliseconds: 0,
			RepositoriesCount:   pointers.Ptr(100),
			ExcludedArchived:    0,
			ExcludedForks:       0,
			Timedout:            timedout100,
			Missing:             nil,
			Cloning:             nil,
			LimitHit:            false,
			DisplayLimit:        math.MaxInt32,
		},
		"all": {
			MatchCount:          1,
			ElapsedMilliseconds: 0,
			RepositoriesCount:   pointers.Ptr(5),
			BackendsMissing:     1,
			ExcludedArchived:    1,
			ExcludedForks:       5,
			Timedout:            []api.RepoID{1},
			Missing:             []api.RepoID{2, 3},
			Cloning:             []api.RepoID{4},
			LimitHit:            true,
			SuggestedLimit:      1000,
			DisplayLimit:        math.MaxInt32,
		},
		"traced": {
			Trace: "abcd",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got := BuildProgressEvent(c, namer)
			got.DurationMs = 0 // clear out non-deterministic field
			testutil.AssertGolden(t, "testdata/golden/"+t.Name()+".json", *updateGolden, got)
		})
	}
}

func TestNumber(t *testing.T) {
	cases := map[int]string{
		0:     "0",
		1:     "1",
		100:   "100",
		999:   "999",
		1000:  "1,000",
		1234:  "1,234",
		3004:  "3,004",
		3040:  "3,040",
		3400:  "3,400",
		9999:  "9,999",
		10000: "10k",
		10400: "10k",
		54321: "54k",
	}
	for n, want := range cases {
		got := number(n)
		if got != want {
			t.Errorf("number(%d) got %q want %q", n, got, want)
		}
	}
}
