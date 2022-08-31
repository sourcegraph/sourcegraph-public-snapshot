package resolvers

import (
	"fmt"
	"strings"
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
	reason       string
}

type canAggregateBySuite struct {
	t                  *testing.T
	testCases          []canAggregateTestCase
	canAggregateByFunc canAggregateBy
}

func safeReason(r *notAvailableReason) string {
	if r == nil {
		return ""
	}
	return r.reason
}

func (suite *canAggregateBySuite) Test_canAggregateBy() {
	for _, tc := range suite.testCases {
		suite.t.Run(tc.name, func(t *testing.T) {
			if tc.patternType == "" {
				tc.patternType = "literal"
			}
			canAggregate, reasonNA, err := suite.canAggregateByFunc(tc.query, tc.patternType)
			errCheck := (err == nil && tc.err == nil) || (err != nil && tc.err != nil)
			if !errCheck {
				t.Errorf("expected error %v, got %v", tc.err, err)
			}
			if err != nil && tc.err != nil && !strings.Contains(err.Error(), tc.err.Error()) {
				t.Errorf("expected error %v to contain %v", err, tc.err)
			}
			if canAggregate != tc.canAggregate {
				t.Errorf("expected canAggregate to be %v, got %v", tc.canAggregate, canAggregate)
			}
			if !strings.EqualFold(safeReason(reasonNA), tc.reason) {
				t.Errorf("expected reason to be %v, got %v", tc.reason, safeReason(reasonNA))
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
			reason:       invalidQueryMsg,
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
			reason:       fmt.Sprintf(fileUnsupportedFieldValueFmt, "select", "repo"),
			canAggregate: false,
		},
		{
			name:         "cannot aggregate for query with type:commit parameter",
			query:        "insights type:commit",
			reason:       fmt.Sprintf(fileUnsupportedFieldValueFmt, "type", "commit"),
			canAggregate: false,
		},
		{
			name:         "ensure type check is case insensitive ",
			query:        "insights TYPE:commit",
			reason:       fmt.Sprintf(fileUnsupportedFieldValueFmt, "type", "commit"),
			canAggregate: false,
		},
		{
			name:         "cannot aggregate for invalid query",
			query:        "insights type:commit fork:test",
			canAggregate: false,
			reason:       invalidQueryMsg,
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
			reason:       authNotCommitDiffMsg,
			canAggregate: false,
		},
		{
			name:         "cannot aggregate for query with case parameter",
			query:        "func(t *testing.T) case:yes",
			reason:       authNotCommitDiffMsg,
			canAggregate: false,
		},
		{
			name:         "cannot aggregate for query with select:repo parameter",
			query:        "repo:contains.path(README) select:repo",
			reason:       authNotCommitDiffMsg,
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
			name:         "can aggregate for query with cased Type",
			query:        "repo:contains.path(README) TyPe:diff fix",
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
			reason:       invalidQueryMsg,
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
			reason:       cgInvalidQueryMsg,
			canAggregate: false,
		},
		{
			name:         "cannot aggregate for invalid query",
			query:        "type:diff fork:leo func(.*)",
			patternType:  "regexp",
			reason:       invalidQueryMsg,
			canAggregate: false,
			err:          errors.Newf("ParseQuery"),
		},
		{
			name:         "cannot aggregate for select:repo query",
			query:        "repo:contains.path(README) func(\\w+) select:repo",
			reason:       fmt.Sprintf(cgUnsupportedSelectFmt, "select", "repo"),
			patternType:  "regexp",
			canAggregate: false,
		},
		{
			name:         "cannot aggregate for select:file query",
			query:        "repo:contains.path(README) func(\\w+) select:file",
			patternType:  "regexp",
			reason:       fmt.Sprintf(cgUnsupportedSelectFmt, "select", "file"),
			canAggregate: false,
		},
		{
			name:         "cannot for type:repo query",
			query:        "repo:contains.path(README) func(\\w+) type:repo",
			patternType:  "regexp",
			reason:       fmt.Sprintf(cgUnsupportedSelectFmt, "type", "repo"),
			canAggregate: false,
		},
		{
			name:         "cannot aggregate for type:path query",
			query:        "repo:contains.path(README) func(\\w+) type:path",
			reason:       fmt.Sprintf(cgUnsupportedSelectFmt, "type", "path"),
			patternType:  "regexp",
			canAggregate: false,
		},
		{
			name:         "ensure type check is not case sensitive",
			query:        "repo:contains.path(README) func(\\w+) TyPe:path",
			reason:       fmt.Sprintf(cgUnsupportedSelectFmt, "type", "path"),
			patternType:  "regexp",
			canAggregate: false,
		},
		{
			name:         "cannot aggregate for query with unsupported pattern type",
			query:        "func(t *testing.T)",
			reason:       cgInvalidQueryMsg,
			patternType:  "literal",
			canAggregate: false,
		},
		{
			name:         "cannot aggregate for query with multiple steps",
			query:        "(repo:^github\\.com/sourcegraph/sourcegraph$ file:go\\.mod$ go\\s*(\\d\\.\\d+)) or (test file:insights)",
			reason:       cgMultipleQueryPatternMsg,
			patternType:  "regexp",
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
		reason      *string
	}{
		{
			name:  "invalid query returns REPO",
			query: "func fork:leo",
			want:  types.REPO_AGGREGATION_MODE,
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
		{
			name:  "query with capture group returns capture group",
			query: "repo:contains.path(README) todo(\\w+)",
			want:  types.CAPTURE_GROUP_AGGREGATION_MODE,
		},
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
			name: "unsupported regexp type:commit query returns author",
			// this query contains two non-capture group regexps so wouldn't support capture group aggregation.
			query: "type:commit TODO \\w+ [0-9]",
			want:  types.AUTHOR_AGGREGATION_MODE,
		},
		{
			name: "unsupported regexp single repo query returns path",
			// this query contains an or so wouldn't support capture group aggregation.
			query: "repo:^github\\.com/sourcegraph/sourcegraph$ TODO \\w+ or  var[0-9]",
			want:  types.PATH_AGGREGATION_MODE,
		},
		{
			name: "unsupported regexp query returns repo",
			// this query contains an or so wouldn't support capture group aggregation.
			query: "TODO \\w+ or  var[0-9]",
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
			mode := getDefaultAggregationMode(tc.query, pt)

			if mode != tc.want {
				t.Errorf("expected mode %v, got %v", tc.want, mode)
			}
		})
	}
}
