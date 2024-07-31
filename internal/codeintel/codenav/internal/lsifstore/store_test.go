package lsifstore

import (
	"fmt"
	"testing"

	"github.com/hexops/autogold/v2"
	genslices "github.com/life4/genesis/slices"
	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
)

func TestIdentifyMatchingOccurrences(t *testing.T) {
	inputOccs := []*scip.Occurrence{
		{Symbol: "a#", Range: []int32{1, 1, 2}},
		{Symbol: "x#", Range: []int32{1, 1, 2}},
		{Symbol: "b#", Range: []int32{2, 1, 5}},
	}

	type testCase struct {
		matcher          shared.Matcher
		expectedOccs     autogold.Value
		expectedStrategy autogold.Value
	}

	posMatcher := shared.NewStartPositionMatcher
	rangeMatcher := shared.NewSCIPBasedMatcher
	scipRange := scip.NewRangeUnchecked

	testCases := []testCase{
		{
			matcher:          posMatcher(scip.Position{Line: 1, Character: 1}),
			expectedOccs:     autogold.Expect([]string{`"a#" @ 1:1-1:2`, `"x#" @ 1:1-1:2`}),
			expectedStrategy: autogold.Expect("single-position based"),
		},
		{
			matcher:          posMatcher(scip.Position{Line: 1, Character: 2}),
			expectedOccs:     autogold.Expect([]string{}),
			expectedStrategy: autogold.Expect("single-position based"),
		},
		{
			// Any intersection is fine
			matcher:          posMatcher(scip.Position{Line: 2, Character: 3}),
			expectedOccs:     autogold.Expect([]string{`"b#" @ 2:1-2:5`}),
			expectedStrategy: autogold.Expect("single-position based"),
		},
		{
			matcher:          rangeMatcher(scipRange([]int32{1, 1, 2}), ""),
			expectedOccs:     autogold.Expect([]string{`"a#" @ 1:1-1:2`, `"x#" @ 1:1-1:2`}),
			expectedStrategy: autogold.Expect("range based"),
		},
		{
			matcher:          rangeMatcher(scipRange([]int32{1, 1, 2}), "a#"),
			expectedOccs:     autogold.Expect([]string{`"a#" @ 1:1-1:2`}),
			expectedStrategy: autogold.Expect("range and symbol based"),
		},
		{
			// Exact range matching => no match
			matcher:          rangeMatcher(scipRange([]int32{2, 1, 3}), ""),
			expectedOccs:     autogold.Expect([]string{}),
			expectedStrategy: autogold.Expect("range based"),
		},
		{
			matcher:          rangeMatcher(scipRange([]int32{2, 1, 5}), ""),
			expectedOccs:     autogold.Expect([]string{`"b#" @ 2:1-2:5`}),
			expectedStrategy: autogold.Expect("range based"),
		},
	}

	for _, tc := range testCases {
		key := FindUsagesKey{UploadID: -1, Path: core.NewUploadRelPathUnchecked("lol"), Matcher: tc.matcher}
		gotOccs, strat := key.IdentifyMatchingOccurrences(inputOccs)
		occstrs := genslices.Map(gotOccs, func(occ *scip.Occurrence) string {
			return fmt.Sprintf("%q @ %s", occ.Symbol, scipRange(occ.Range).String())
		})
		tc.expectedOccs.Equal(t, occstrs)
		tc.expectedStrategy.Equal(t, string(strat))
	}

}
