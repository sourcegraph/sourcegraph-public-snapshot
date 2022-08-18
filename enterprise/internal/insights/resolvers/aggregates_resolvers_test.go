package resolvers

import (
	"testing"
)

func Test_canAggregateByPath(t *testing.T) {
	// pattern type does not matter for this aggregation mode.
	patternType := "literal"

	testCases := []struct {
		name         string
		query        string
		canAggregate bool
	}{
		{
			"can aggregate for query without parameters",
			"func(t *testing.T)",
			true,
		},
		{
			"can aggregate for query with case parameter",
			"func(t *testing.T) case:yes",
			true,
		},
		{
			"cannot aggregate for query with select:repo parameter",
			"repo:contains.file(README) select:repo",
			false,
		},
		{
			"cannot aggregate for query with type:commit parameter",
			"insights type:commit",
			false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			canAggregate := canAggregateByPath(tc.query, patternType)
			if canAggregate != tc.canAggregate {
				t.Errorf("expected canAggregate to be %v, got %v", tc.canAggregate, canAggregate)
			}
		})
	}
}
