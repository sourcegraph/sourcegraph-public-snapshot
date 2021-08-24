package api

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
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
			DisplayLimit:        math.MaxInt32,
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
			DisplayLimit:        math.MaxInt32,
		},
		"traced": {
			Trace: "abcd",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got := BuildProgressEvent(c)
			got.DurationMs = 0 // clear out non-deterministic field
			AssertGolden(t, "testdata/golden/"+t.Name()+".json", *updateGolden, got)
		})
	}
}

// TODO(camdencheek) this is copied out of testutil, but we can't import testutil because of its dependencies on store
func AssertGolden(t testing.TB, path string, update bool, want interface{}) {
	t.Helper()

	marshal := func(t testing.TB, v interface{}) []byte {
		t.Helper()

		switch v2 := v.(type) {
		case string:
			return []byte(v2)
		case []byte:
			return v2
		default:
			data, err := json.MarshalIndent(v, " ", " ")
			if err != nil {
				t.Fatal(err)
			}
			return data
		}
	}
	data := marshal(t, want)

	if update {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatalf("failed to update golden file %q: %s", path, err)
		}
		if err := os.WriteFile(path, data, 0o640); err != nil {
			t.Fatalf("failed to update golden file %q: %s", path, err)
		}
	}

	golden, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read golden file %q: %s", path, err)
	}

	if diff := cmp.Diff(string(golden), string(data)); diff != "" {
		t.Errorf("(-want, +got):\n%s", diff)
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

type repo struct {
	name string
}

func (r repo) Name() string {
	return r.name
}

func intPtr(i int) *int {
	return &i
}
