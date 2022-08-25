package resolvers

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type canAggregateTestCase struct {
	name         string
	query        string
	patternType  string
	canAggregate bool
	err          error
}

type canAggregateBySuite struct {
	t                  *testing.T
	testCases          []canAggregateTestCase
	canAggregateByFunc canAggregateBy
}

func (suite *canAggregateBySuite) Test_canAggregateBy() {
	for _, tc := range suite.testCases {
		suite.t.Run(tc.name, func(t *testing.T) {
			if tc.patternType == "" {
				tc.patternType = "literal"
			}
			canAggregate, err := suite.canAggregateByFunc(tc.query, tc.patternType)
			errCheck := (err == nil && tc.err == nil) || (err != nil && tc.err != nil)
			if !errCheck {
				t.Errorf("expected error %v, got %v", tc.err, err)
			}
			if canAggregate != tc.canAggregate {
				t.Errorf("expected canAggregate to be %v, got %v", tc.canAggregate, canAggregate)
			}
		})
	}
}

func Test_canAggregateByRepo(t *testing.T) {
	testCases := []canAggregateTestCase{
		{
			name:         "cannot aggregate for invalid query",
			query:        "fork:woo",
			canAggregate: false,
			err:          errors.Newf("ParseQuery"),
		},
	}
	suite := canAggregateBySuite{
		canAggregateByFunc: canAggregateByRepo,
		testCases:          testCases,
		t:                  t,
	}
	suite.Test_canAggregateBy()
}

func Test_canAggregateByPath(t *testing.T) {
	testCases := []canAggregateTestCase{
		{
			name:         "can aggregate for query without parameters",
			query:        "func(t *testing.T)",
			canAggregate: true,
		},
		{
			name:         "can aggregate for query with case parameter",
			query:        "func(t *testing.T) case:yes",
			canAggregate: true,
		},
		{
			name:         "cannot aggregate for query with select:repo parameter",
			query:        "repo:contains.path(README) select:repo",
			canAggregate: false,
		},
		{
			name:         "cannot aggregate for query with type:commit parameter",
			query:        "insights type:commit",
			canAggregate: false,
		},
		{
			name:         "cannot aggregate for invalid query",
			query:        "insights type:commit fork:test",
			canAggregate: false,
			err:          errors.Newf("ParseQuery"),
		},
	}
	suite := canAggregateBySuite{
		canAggregateByFunc: canAggregateByPath,
		testCases:          testCases,
		t:                  t,
	}
	suite.Test_canAggregateBy()
}

func Test_canAggregateByAuthor(t *testing.T) {
	testCases := []canAggregateTestCase{
		{
			name:         "cannot aggregate for query without parameters",
			query:        "func(t *testing.T)",
			canAggregate: false,
		},
		{
			name:         "cannot aggregate for query with case parameter",
			query:        "func(t *testing.T) case:yes",
			canAggregate: false,
		},
		{
			name:         "cannot aggregate for query with select:repo parameter",
			query:        "repo:contains.path(README) select:repo",
			canAggregate: false,
		},
		{
			name:         "can aggregate for query with type:commit parameter",
			query:        "repo:contains.path(README) select:repo type:commit fix",
			canAggregate: true,
		},
		{
			name:         "can aggregate for query with select:commit parameter",
			query:        "repo:contains.path(README) select:commit fix",
			canAggregate: true,
		},
		{
			name:         "can aggregate for query with type:diff parameter",
			query:        "repo:contains.path(README) type:diff fix",
			canAggregate: true,
		},
		{
			name:         "can aggregate for weird query with type:diff select:commit",
			query:        "type:diff select:commit insights",
			canAggregate: true,
		},
		{
			name:         "cannot aggregate for invalid query",
			query:        "type:diff fork:leo",
			canAggregate: false,
			err:          errors.Newf("ParseQuery"),
		},
	}
	suite := canAggregateBySuite{
		canAggregateByFunc: canAggregateByAuthor,
		testCases:          testCases,
		t:                  t,
	}
	suite.Test_canAggregateBy()
}

func Test_canAggregateByCaptureGroup(t *testing.T) {
	testCases := []canAggregateTestCase{
		{
			name:         "can aggregate for simple query with regex pattern type",
			query:        "func(\\w+) case:yes",
			patternType:  "regexp",
			canAggregate: true,
		},
		{
			name:         "can aggregate for standard query in backslash pattern",
			query:        "/func(\\w+)/ case:yes",
			patternType:  "standard",
			canAggregate: true,
		},
		{
			name:         "can aggregate for multi-pattern query",
			query:        "func(\\w+[0-9]) return(\\w+[0-9]) ",
			patternType:  "regexp",
			canAggregate: true,
		},
		{
			name:         "can aggregate for query with both captured and non-captured regexp pattern",
			query:        "func(\\w+) \\w+",
			patternType:  "regexp",
			canAggregate: true,
		},
		{
			name:         "cannot aggregate for query with non-captured regexp pattern",
			query:        "\\w+",
			patternType:  "regexp",
			canAggregate: false,
		},
		{
			name:         "cannot aggregate for invalid query",
			query:        "type:diff fork:leo func(.*)",
			patternType:  "regexp",
			canAggregate: false,
			err:          errors.Newf("ParseAndValidateQuery"),
		},
		{
			name:         "cannot aggregate for select:repo query",
			query:        "repo:contains.path(README) func(\\w+) select:repo",
			patternType:  "regexp",
			canAggregate: false,
		},
		{
			name:         "cannot aggregate for select:file query",
			query:        "repo:contains.path(README) func(\\w+) select:file",
			patternType:  "regexp",
			canAggregate: false,
		},
		{
			name:         "cannot for type:repo query",
			query:        "repo:contains.path(README) func(\\w+) type:repo",
			patternType:  "regexp",
			canAggregate: false,
		},
		{
			name:         "cannot aggregate for type:path query",
			query:        "repo:contains.path(README) func(\\w+) type:path",
			patternType:  "regexp",
			canAggregate: false,
		},
		{
			name:         "cannot aggregate for query with unsupported pattern type",
			query:        "func(t *testing.T)",
			patternType:  "literal",
			canAggregate: false,
		},
	}
	suite := canAggregateBySuite{
		canAggregateByFunc: canAggregateByCaptureGroup,
		testCases:          testCases,
		t:                  t,
	}
	suite.Test_canAggregateBy()
}

func Test_getDefaultAggregationMode(t *testing.T) {
	testCases := []struct {
		name        string
		query       string
		patternType string
		want        types.SearchAggregationMode
		err         error
	}{
		{
			name:  "invalid query returns error",
			query: "func fork:leo",
			want:  "",
			err:   errors.New("ParseQuery"),
		},
		{
			name:        "literal type query does not return capture group mode",
			query:       "func([0-9]+)",
			patternType: "literal",
			want:        types.REPO_AGGREGATION_MODE,
		},
		{
			name:  "query with regex no capture group returns repo",
			query: "func [0-9] case:yes",
			want:  types.REPO_AGGREGATION_MODE,
		},
		//{
		//	name:  "query with capture group returns capture group",
		//	query: "repo:contains.path(README) todo(\\w+)",
		//	want:  types.CAPTURE_GROUP_AGGREGATION_MODE,
		//},
		{
			name:  "type:commit query returns author",
			query: "type:commit fix",
			want:  types.AUTHOR_AGGREGATION_MODE,
		},
		{
			name:  "type:diff query returns author",
			query: "type:diff fix",
			want:  types.AUTHOR_AGGREGATION_MODE,
		},
		{
			name:  "query for single repo returns path",
			query: "repo:^github\\.com/sourcegraph/sourcegraph$ insights",
			want:  types.PATH_AGGREGATION_MODE,
		},
		{
			name:  "query not for single repo returns repo",
			query: "repo:^github.com/sourcegraph insights",
			want:  types.REPO_AGGREGATION_MODE,
		},
		{
			name:  "query with repo predicate returns repo",
			query: "repo:contains.path(README) insights",
			want:  types.REPO_AGGREGATION_MODE,
		},
		{
			name:  "defaults to repo",
			query: "getDefaultAggregationMode file:insights",
			want:  types.REPO_AGGREGATION_MODE,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pt := "regexp"
			if tc.patternType != "" {
				pt = tc.patternType
			}
			mode, err := getDefaultAggregationMode(tc.query, pt)
			if (err != nil && tc.err == nil) || (err == nil && tc.err != nil) {
				t.Errorf("expected different error behavior: got %v, want %v", err, tc.err)
			}
			if mode != tc.want {
				t.Errorf("expected mode %v, got %v", tc.want, mode)
			}
		})
	}
}
