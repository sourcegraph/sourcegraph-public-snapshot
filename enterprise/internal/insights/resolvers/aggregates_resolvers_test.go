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

func Test_canAggregateByAuthor(t *testing.T) {
	// pattern type does not matter for this aggregation mode.
	patternType := "literal"

	testCases := []struct {
		name         string
		query        string
		canAggregate bool
	}{
		{
			"cannot aggregate for query without parameters",
			"func(t *testing.T)",
			false,
		},
		{
			"cannot aggregate for query with case parameter",
			"func(t *testing.T) case:yes",
			false,
		},
		{
			"cannot aggregate for query with select:repo parameter",
			"repo:contains.file(README) select:repo",
			false,
		},
		{
			"can aggregate for query with type:commit parameter",
			"repo:contains.file(README) select:repo type:commit fix",
			true,
		},
		{
			"can aggregate for query with select:commit parameter",
			"repo:contains.file(README) select:commit fix",
			true,
		},
		{
			"can aggregate for query with type:diff parameter",
			"repo:contains.file(README) type:diff fix",
			true,
		},
		{
			"can aggregate for weird query with type:diff select:commit",
			"type:diff select:commit insights",
			true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			canAggregate := canAggregateByAuthor(tc.query, patternType)
			if canAggregate != tc.canAggregate {
				t.Errorf("expected canAggregate to be %v, got %v", tc.canAggregate, canAggregate)
			}
		})
	}
}
