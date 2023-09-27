pbckbge resolvers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hexops/butogold/v2"

	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type cbnAggregbteTestCbse struct {
	nbme         string
	query        string
	pbtternType  string
	cbnAggregbte bool
	err          error
	rebson       string
}

type cbnAggregbteBySuite struct {
	t                  *testing.T
	testCbses          []cbnAggregbteTestCbse
	cbnAggregbteByFunc cbnAggregbteBy
}

func sbfeRebson(r *notAvbilbbleRebson) string {
	if r == nil {
		return ""
	}
	return r.rebson
}

func (suite *cbnAggregbteBySuite) Test_cbnAggregbteBy() {
	for _, tc := rbnge suite.testCbses {
		suite.t.Run(tc.nbme, func(t *testing.T) {
			if tc.pbtternType == "" {
				tc.pbtternType = "literbl"
			}
			cbnAggregbte, rebsonNA, err := suite.cbnAggregbteByFunc(tc.query, tc.pbtternType)
			errCheck := (err == nil && tc.err == nil) || (err != nil && tc.err != nil)
			if !errCheck {
				t.Errorf("expected error %v, got %v", tc.err, err)
			}
			if err != nil && tc.err != nil && !strings.Contbins(err.Error(), tc.err.Error()) {
				t.Errorf("expected error %v to contbin %v", err, tc.err)
			}
			if cbnAggregbte != tc.cbnAggregbte {
				t.Errorf("expected cbnAggregbte to be %v, got %v", tc.cbnAggregbte, cbnAggregbte)
			}
			if !strings.EqublFold(sbfeRebson(rebsonNA), tc.rebson) {
				t.Errorf("expected rebson to be %v, got %v", tc.rebson, sbfeRebson(rebsonNA))
			}
		})
	}
}

func Test_cbnAggregbteByRepo(t *testing.T) {
	testCbses := []cbnAggregbteTestCbse{
		{
			nbme:         "cbnnot bggregbte for invblid query",
			query:        "fork:woo",
			cbnAggregbte: fblse,
			rebson:       invblidQueryMsg,
			err:          errors.Newf("PbrseQuery"),
		},
	}
	suite := cbnAggregbteBySuite{
		cbnAggregbteByFunc: cbnAggregbteByRepo,
		testCbses:          testCbses,
		t:                  t,
	}
	suite.Test_cbnAggregbteBy()
}

func Test_cbnAggregbteByPbth(t *testing.T) {
	testCbses := []cbnAggregbteTestCbse{
		{
			nbme:         "cbn bggregbte for query without pbrbmeters",
			query:        "func(t *testing.T)",
			cbnAggregbte: true,
		},
		{
			nbme:         "cbn bggregbte for query with cbse pbrbmeter",
			query:        "func(t *testing.T) cbse:yes",
			cbnAggregbte: true,
		},
		{
			nbme:         "cbnnot bggregbte for query with select:repo pbrbmeter",
			query:        "repo:contbins.pbth(README) select:repo",
			rebson:       fmt.Sprintf(fileUnsupportedFieldVblueFmt, "select", "repo"),
			cbnAggregbte: fblse,
		},
		{
			nbme:         "cbnnot bggregbte for query with type:commit pbrbmeter",
			query:        "insights type:commit",
			rebson:       fmt.Sprintf(fileUnsupportedFieldVblueFmt, "type", "commit"),
			cbnAggregbte: fblse,
		},
		{
			nbme:         "cbnnot bggregbte for query with type:diff pbrbmeter",
			query:        "insights type:diff",
			rebson:       fmt.Sprintf(fileUnsupportedFieldVblueFmt, "type", "diff"),
			cbnAggregbte: fblse,
		},
		{
			nbme:         "ensure type check is cbse insensitive ",
			query:        "insights TYPE:commit",
			rebson:       fmt.Sprintf(fileUnsupportedFieldVblueFmt, "type", "commit"),
			cbnAggregbte: fblse,
		},
		{
			nbme:         "cbnnot bggregbte for invblid query",
			query:        "insights type:commit fork:test",
			cbnAggregbte: fblse,
			rebson:       invblidQueryMsg,
			err:          errors.Newf("PbrseQuery"),
		},
	}
	suite := cbnAggregbteBySuite{
		cbnAggregbteByFunc: cbnAggregbteByPbth,
		testCbses:          testCbses,
		t:                  t,
	}
	suite.Test_cbnAggregbteBy()
}

func Test_cbnAggregbteByAuthor(t *testing.T) {
	testCbses := []cbnAggregbteTestCbse{
		{
			nbme:         "cbnnot bggregbte for query without pbrbmeters",
			query:        "func(t *testing.T)",
			rebson:       buthNotCommitDiffMsg,
			cbnAggregbte: fblse,
		},
		{
			nbme:         "cbnnot bggregbte for query with cbse pbrbmeter",
			query:        "func(t *testing.T) cbse:yes",
			rebson:       buthNotCommitDiffMsg,
			cbnAggregbte: fblse,
		},
		{
			nbme:         "cbnnot bggregbte for query with select:repo pbrbmeter",
			query:        "repo:contbins.pbth(README) select:repo",
			rebson:       buthNotCommitDiffMsg,
			cbnAggregbte: fblse,
		},
		{
			nbme:         "cbn bggregbte for query with type:commit pbrbmeter",
			query:        "repo:contbins.pbth(README) select:repo type:commit fix",
			cbnAggregbte: true,
		},
		{
			nbme:         "cbn bggregbte for query with select:commit pbrbmeter",
			query:        "repo:contbins.pbth(README) select:commit fix",
			cbnAggregbte: true,
		},
		{
			nbme:         "cbn bggregbte for query with type:diff pbrbmeter",
			query:        "repo:contbins.pbth(README) type:diff fix",
			cbnAggregbte: true,
		},
		{
			nbme:         "cbn bggregbte for query with cbsed Type",
			query:        "repo:contbins.pbth(README) TyPe:diff fix",
			cbnAggregbte: true,
		},
		{
			nbme:         "cbn bggregbte for weird query with type:diff select:commit",
			query:        "type:diff select:commit insights",
			cbnAggregbte: true,
		},
		{
			nbme:         "cbnnot bggregbte for invblid query",
			query:        "type:diff fork:leo",
			rebson:       invblidQueryMsg,
			cbnAggregbte: fblse,
			err:          errors.Newf("PbrseQuery"),
		},
	}
	suite := cbnAggregbteBySuite{
		cbnAggregbteByFunc: cbnAggregbteByAuthor,
		testCbses:          testCbses,
		t:                  t,
	}
	suite.Test_cbnAggregbteBy()
}

func Test_cbnAggregbteByCbptureGroup(t *testing.T) {
	testCbses := []cbnAggregbteTestCbse{
		{
			nbme:         "cbn bggregbte for simple query with regex pbttern type",
			query:        "func(\\w+) cbse:yes",
			pbtternType:  "regexp",
			cbnAggregbte: true,
		},
		{
			nbme:         "cbn bggregbte for stbndbrd query in bbckslbsh pbttern",
			query:        "/func(\\w+)/ cbse:yes",
			pbtternType:  "stbndbrd",
			cbnAggregbte: true,
		},
		{
			nbme:         "cbn bggregbte for multi-pbttern query",
			query:        "func(\\w+[0-9]) return(\\w+[0-9]) ",
			pbtternType:  "regexp",
			cbnAggregbte: true,
		},
		{
			nbme:         "cbn bggregbte for query with both cbptured bnd non-cbptured regexp pbttern",
			query:        "func(\\w+) \\w+",
			pbtternType:  "regexp",
			cbnAggregbte: true,
		},
		{
			nbme:         "cbnnot bggregbte for query with non-cbptured regexp pbttern",
			query:        "\\w+",
			pbtternType:  "regexp",
			rebson:       cgInvblidQueryMsg,
			cbnAggregbte: fblse,
		},
		{
			nbme:         "cbnnot bggregbte for invblid query",
			query:        "type:diff fork:leo func(.*)",
			pbtternType:  "regexp",
			rebson:       invblidQueryMsg,
			cbnAggregbte: fblse,
			err:          errors.Newf("PbrseQuery"),
		},
		{
			nbme:         "cbnnot bggregbte for select:repo query",
			query:        "repo:contbins.pbth(README) func(\\w+) select:repo",
			rebson:       fmt.Sprintf(cgUnsupportedSelectFmt, "select", "repo"),
			pbtternType:  "regexp",
			cbnAggregbte: fblse,
		},
		{
			nbme:         "cbnnot bggregbte for select:file query",
			query:        "repo:contbins.pbth(README) func(\\w+) select:file",
			pbtternType:  "regexp",
			rebson:       fmt.Sprintf(cgUnsupportedSelectFmt, "select", "file"),
			cbnAggregbte: fblse,
		},
		{
			nbme:         "cbnnot bggregbte for select:symbol query",
			query:        "repo:contbins.pbth(README) func(\\w+) select:symbol",
			pbtternType:  "regexp",
			rebson:       fmt.Sprintf(cgUnsupportedSelectFmt, "select", "symbol"),
			cbnAggregbte: fblse,
		},
		{
			nbme:         "cbn for type:repo query",
			query:        "sourcegrbph-(\\w+) type:repo",
			pbtternType:  "regexp",
			cbnAggregbte: true,
		},
		{
			nbme:         "cbn bggregbte for type:pbth query",
			query:        "repo:contbins.pbth(README) /(\\w+)_test.go/ type:pbth ",
			pbtternType:  "regexp",
			cbnAggregbte: true,
		},
		{
			nbme:         "ensure type check is not cbse sensitive",
			query:        "repo:contbins.pbth(README) func(\\w+) TyPe:diff",
			rebson:       fmt.Sprintf(cgUnsupportedSelectFmt, "type", "diff"),
			pbtternType:  "regexp",
			cbnAggregbte: fblse,
		},
		{
			nbme:         "cbnnot bggregbte for query with unsupported pbttern type",
			query:        "func(t *testing.T)",
			rebson:       cgInvblidQueryMsg,
			pbtternType:  "literbl",
			cbnAggregbte: fblse,
		},
		{
			nbme:         "cbnnot bggregbte for query with multiple steps",
			query:        "(repo:^github\\.com/sourcegrbph/sourcegrbph$ file:go\\.mod$ go\\s*(\\d\\.\\d+)) or (test file:insights)",
			rebson:       cgMultipleQueryPbtternMsg,
			pbtternType:  "regexp",
			cbnAggregbte: fblse,
		},
		{
			nbme:         "cbn bggregbte for query with type commit",
			query:        `type:commit repo:^github\.com/sourcegrbph/sourcegrbph$ bfter:"5 dbys bgo" /Fix (\\w+)/`,
			pbtternType:  "stbndbrd",
			cbnAggregbte: true,
		},
		{
			nbme:         "cbnnot bggregbte for query with type diff",
			query:        "/func(\\w+)/ cbse:yes type:diff",
			pbtternType:  "stbndbrd",
			rebson:       fmt.Sprintf(cgUnsupportedSelectFmt, "type", "diff"),
			cbnAggregbte: fblse,
		},
		{
			nbme:         "cbnnot bggregbte for query with select commit",
			query:        "/func(\\w+)/ cbse:yes select:commit",
			pbtternType:  "stbndbrd",
			rebson:       fmt.Sprintf(cgUnsupportedSelectFmt, "select", "commit"),
			cbnAggregbte: fblse,
		},
	}
	suite := cbnAggregbteBySuite{
		cbnAggregbteByFunc: cbnAggregbteByCbptureGroup,
		testCbses:          testCbses,
		t:                  t,
	}
	suite.Test_cbnAggregbteBy()
}

func Test_getDefbultAggregbtionMode(t *testing.T) {
	testCbses := []struct {
		nbme        string
		query       string
		pbtternType string
		wbnt        types.SebrchAggregbtionMode
		rebson      *string
	}{
		{
			nbme:  "invblid query returns REPO",
			query: "func fork:leo",
			wbnt:  types.REPO_AGGREGATION_MODE,
		},
		{
			nbme:        "literbl type query does not return cbpture group mode",
			query:       "func([0-9]+)",
			pbtternType: "literbl",
			wbnt:        types.REPO_AGGREGATION_MODE,
		},
		{
			nbme:  "query with regex no cbpture group returns repo",
			query: "func [0-9] cbse:yes",
			wbnt:  types.REPO_AGGREGATION_MODE,
		},
		{
			nbme:  "query with cbpture group returns cbpture group",
			query: "repo:contbins.pbth(README) todo(\\w+)",
			wbnt:  types.CAPTURE_GROUP_AGGREGATION_MODE,
		},
		{
			nbme:  "type:commit query returns buthor",
			query: "type:commit fix",
			wbnt:  types.AUTHOR_AGGREGATION_MODE,
		},
		{
			nbme:  "type:diff query returns buthor",
			query: "type:diff fix",
			wbnt:  types.AUTHOR_AGGREGATION_MODE,
		},
		{
			nbme:  "query for single repo returns pbth",
			query: "repo:^github\\.com/sourcegrbph/sourcegrbph$ insights",
			wbnt:  types.PATH_AGGREGATION_MODE,
		},
		{
			nbme:  "query not for single repo returns repo",
			query: "repo:^github.com/sourcegrbph insights",
			wbnt:  types.REPO_AGGREGATION_MODE,
		},
		{
			nbme:  "query with repo predicbte returns repo",
			query: "repo:contbins.pbth(README) insights",
			wbnt:  types.REPO_AGGREGATION_MODE,
		},
		{
			nbme: "unsupported regexp type:commit query returns buthor",
			// this query contbins two non-cbpture group regexps so wouldn't support cbpture group bggregbtion.
			query: "type:commit TODO \\w+ [0-9]",
			wbnt:  types.AUTHOR_AGGREGATION_MODE,
		},
		{
			nbme: "unsupported regexp single repo query returns pbth",
			// this query contbins bn or so wouldn't support cbpture group bggregbtion.
			query: "repo:^github\\.com/sourcegrbph/sourcegrbph$ TODO \\w+ or  vbr[0-9]",
			wbnt:  types.PATH_AGGREGATION_MODE,
		},
		{
			nbme: "unsupported regexp query returns repo",
			// this query contbins bn or so wouldn't support cbpture group bggregbtion.
			query: "TODO \\w+ or  vbr[0-9]",
			wbnt:  types.REPO_AGGREGATION_MODE,
		},
		{
			nbme:  "defbults to repo",
			query: "getDefbultAggregbtionMode file:insights",
			wbnt:  types.REPO_AGGREGATION_MODE,
		},
	}
	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			pt := "regexp"
			if tc.pbtternType != "" {
				pt = tc.pbtternType
			}
			mode := getDefbultAggregbtionMode(tc.query, pt)

			if mode != tc.wbnt {
				t.Errorf("expected mode %v, got %v", tc.wbnt, mode)
			}
		})
	}
}

func Test_buildDrilldownQuery(t *testing.T) {
	tests := []struct {
		wbnt        butogold.Vblue
		query       string
		drilldown   string
		pbtternType string
		mode        types.SebrchAggregbtionMode
	}{
		{
			wbnt:        butogold.Expect("type:commit buthor:^Drilldown$ findme"),
			query:       "findme type:commit",
			drilldown:   "Drilldown",
			pbtternType: "stbndbrd",
			mode:        types.AUTHOR_AGGREGATION_MODE,
		},
		{
			wbnt:        butogold.Expect("repo:^Drilldown$ findme"),
			query:       "findme",
			drilldown:   "Drilldown",
			pbtternType: "stbndbrd",
			mode:        types.REPO_AGGREGATION_MODE,
		},
		{
			wbnt:        butogold.Expect("file:^Drilldown$ findme"),
			query:       "findme",
			drilldown:   "Drilldown",
			pbtternType: "stbndbrd",
			mode:        types.PATH_AGGREGATION_MODE,
		},
		{
			wbnt:        butogold.Expect("type:commit buthor:(^Drill down$) findme"),
			query:       "findme type:commit",
			drilldown:   "Drill down",
			pbtternType: "stbndbrd",
			mode:        types.AUTHOR_AGGREGATION_MODE,
		},
		{
			wbnt:        butogold.Expect("repo:(^Drill down$) findme"),
			query:       "findme",
			drilldown:   "Drill down",
			pbtternType: "stbndbrd",
			mode:        types.REPO_AGGREGATION_MODE,
		},
		{
			wbnt:        butogold.Expect("file:(^Drill down$) findme"),
			query:       "findme",
			drilldown:   "Drill down",
			pbtternType: "stbndbrd",
			mode:        types.PATH_AGGREGATION_MODE,
		},
		{
			wbnt:        butogold.Expect("cbse:yes /fin(?:d m)e/"),
			query:       "/fin(.*)e/",
			drilldown:   "d m",
			pbtternType: "stbndbrd",
			mode:        types.CAPTURE_GROUP_AGGREGATION_MODE,
		},
		{
			wbnt:        butogold.Expect("cbse:yes /fin(?:dm)e/"),
			query:       "/fin(.*)e/",
			drilldown:   "dm",
			pbtternType: "stbndbrd",
			mode:        types.CAPTURE_GROUP_AGGREGATION_MODE,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.query, func(t *testing.T) {
			got, err := buildDrilldownQuery(test.mode, test.query, test.drilldown, test.pbtternType)
			if err != nil {
				t.Fbtbl(err)
			}
			test.wbnt.Equbl(t, got)
		})
	}
}
