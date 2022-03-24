package query

import (
	"fmt"
	"sort"
	"testing"

	"github.com/hexops/autogold"
)

// ComputeMatchContext implements the ComputeResult interface so we should be able to use it for
// unit tests, but because GroupByCaptureMatch takes a slice, the equivalence doesn't work.
// This test helper makes it work.
func mockComputeResults(cmc []ComputeMatchContext) []ComputeResult {
	results := make([]ComputeResult, 0, len(cmc))
	for _, c := range cmc {
		results = append(results, c)
	}
	return results
}

// stringify will turn a slice of GroupedResults into a sorted slice of strings for comparison.
func stringify(results []GroupedResults) []string {
	stringified := make([]string, 0, len(results))
	for _, result := range results {
		stringified = append(stringified, fmt.Sprintf("%s: %d", result.Value, result.Count))
	}
	sort.Strings(stringified)
	return stringified
}

func TestGroupByCaptureMatch(t *testing.T) {
	defaultEnvironmentEntry := ComputeEnvironmentEntry{
		Value: "func ()",
	}
	longValue := "alpha bravo charlie delta echo foxtrot golf hotel india juliet kilo lima mike november oscar papa quebec romeo sierra"

	t.Run("no compute results returns no group", func(t *testing.T) {
		got := GroupByCaptureMatch([]ComputeResult{})
		autogold.Want("no compute results returns no group", []string{}).Equal(t, stringify(got))
	})

	t.Run("single result single match single environment returns single group", func(t *testing.T) {
		results := mockComputeResults([]ComputeMatchContext{
			{
				Matches: []ComputeMatch{
					{
						Environment: []ComputeEnvironmentEntry{defaultEnvironmentEntry},
					},
				},
			},
		})
		want := []GroupedResults{
			{
				Value: "func ()",
				Count: 1,
			},
		}
		got := GroupByCaptureMatch(results)
		autogold.Want("single result single match one value returns single group", stringify(want)).Equal(t, stringify(got))
	})

	t.Run("single result single match multiple same value returns single group", func(t *testing.T) {
		results := mockComputeResults([]ComputeMatchContext{
			{
				Matches: []ComputeMatch{
					{
						Environment: []ComputeEnvironmentEntry{
							defaultEnvironmentEntry,
							defaultEnvironmentEntry,
						},
					},
				},
			},
		})
		want := []GroupedResults{
			{
				Value: "func ()",
				Count: 2,
			},
		}
		got := GroupByCaptureMatch(results)
		autogold.Want("single result single match multiple same value returns single group", stringify(want)).Equal(t, stringify(got))
	})

	t.Run("single result multiple matches same value returns single group", func(t *testing.T) {
		results := mockComputeResults([]ComputeMatchContext{
			{
				Matches: []ComputeMatch{
					{
						Environment: []ComputeEnvironmentEntry{defaultEnvironmentEntry},
					},
					{
						Environment: []ComputeEnvironmentEntry{defaultEnvironmentEntry},
					},
				},
			},
		})
		want := []GroupedResults{
			{
				Value: "func ()",
				Count: 2,
			},
		}
		got := GroupByCaptureMatch(results)
		autogold.Want("single result multiple matches same value returns single group", stringify(want)).Equal(t, stringify(got))
	})

	t.Run("multiple results multiple matches multiple values returns correct groups", func(t *testing.T) {
		results := mockComputeResults([]ComputeMatchContext{
			{
				Matches: []ComputeMatch{
					{
						Environment: []ComputeEnvironmentEntry{
							defaultEnvironmentEntry,
							{
								Value: "func diff()",
							},
						},
					},
					{
						Environment: []ComputeEnvironmentEntry{
							{
								Value: "func charlie()",
							},
						},
					},
				},
			},
			{
				Matches: []ComputeMatch{
					{
						Environment: []ComputeEnvironmentEntry{
							{
								Value: "func charlie()",
							},
						},
					},
					{
						Environment: []ComputeEnvironmentEntry{
							defaultEnvironmentEntry,
							{
								Value: "func sam()",
							},
						},
					},
				},
			},
		})
		want := []GroupedResults{
			{
				Value: "func ()",
				Count: 2,
			},
			{
				Value: "func diff()",
				Count: 1,
			},
			{
				Value: "func charlie()",
				Count: 2,
			},
			{
				Value: "func sam()",
				Count: 1,
			},
		}
		got := GroupByCaptureMatch(results)
		autogold.Want("multiple results multiple matches multiple values returns correct groups", stringify(want)).Equal(t, stringify(got))
	})

	t.Run("results with values over max allowed length get truncated", func(t *testing.T) {
		results := mockComputeResults([]ComputeMatchContext{
			{
				Matches: []ComputeMatch{
					{
						Environment: []ComputeEnvironmentEntry{
							{
								Value: longValue,
							},
						},
					},
				},
			},
		})
		want := []GroupedResults{
			{
				Value: longValue[:capturedValueMaxLength],
				Count: 1,
			},
		}
		got := GroupByCaptureMatch(results)
		autogold.Want("results with values over max allowed length get truncated", stringify(want)).Equal(t, stringify(got))
	})

	t.Run("identical values after truncation get grouped together", func(t *testing.T) {
		results := mockComputeResults([]ComputeMatchContext{
			{
				Matches: []ComputeMatch{
					{
						Environment: []ComputeEnvironmentEntry{
							{
								Value: longValue,
							},
						},
					},
					{
						Environment: []ComputeEnvironmentEntry{
							{
								Value: longValue + " tango",
							},
						},
					},
				},
			},
		})
		want := []GroupedResults{
			{
				Value: longValue[:capturedValueMaxLength],
				Count: 2,
			},
		}
		got := GroupByCaptureMatch(results)
		autogold.Want("identical values after truncation get grouped together", stringify(want)).Equal(t, stringify(got))

	})
}
