package query

import (
	"testing"

	"github.com/hexops/autogold"
)

// `ComputeMatchContext` implements the `ComputeResult` interface so we should be able to use it for
// unit tests, but because `GroupByCaptureMatch` takes a slice, the equivalence doesn't work.
// This test helper makes it work.
func mockComputeResults(cmc []ComputeMatchContext) []ComputeResult {
	results := make([]ComputeResult, 0, len(cmc))
	for _, c := range cmc {
		results = append(results, c)
	}
	return results
}

func TestGroupByCaptureMatch(t *testing.T) {
	defaultEnvironmentEntry := ComputeEnvironmentEntry{
		Value: "func ()",
	}
	t.Run("no compute results returns no group", func(t *testing.T) {
		results := mockComputeResults([]ComputeMatchContext{})
		autogold.Want("no compute results returns no group", []GroupedResults{}).Equal(t, GroupByCaptureMatch(results))
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
		autogold.Want("single result single match one value returns single group", []GroupedResults{
			{
				Value: "func ()",
				Count: 1,
			},
		}).Equal(t, GroupByCaptureMatch(results))
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
		autogold.Want("single result single match multiple same value returns single group", []GroupedResults{
			{
				Value: "func ()",
				Count: 2,
			},
		}).Equal(t, GroupByCaptureMatch(results))
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
		autogold.Want("single result multiple matches same value returns single group", []GroupedResults{
			{
				Value: "func ()",
				Count: 2,
			},
		}).Equal(t, GroupByCaptureMatch(results))
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
		autogold.Want("multiple results multiple matches multiple values returns correct groups", []GroupedResults{
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
		}).Equal(t, GroupByCaptureMatch(results))
	})
}
