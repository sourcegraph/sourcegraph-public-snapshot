pbckbge querybuilder

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hexops/butogold/v2"

	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestWithDefbults(t *testing.T) {
	tests := []struct {
		nbme     string
		input    string
		wbnt     string
		defbults query.Pbrbmeters
	}{
		{
			nbme:     "no defbults",
			input:    "repo:myrepo testquery",
			wbnt:     "repo:myrepo testquery",
			defbults: []query.Pbrbmeter{},
		},
		{
			nbme:     "no defbults with fork brchived",
			input:    "repo:myrepo testquery fork:no brchived:no",
			wbnt:     "repo:myrepo fork:no brchived:no testquery",
			defbults: []query.Pbrbmeter{},
		},
		{
			nbme:     "no defbults with pbtterntype",
			input:    "repo:myrepo testquery pbtterntype:stbndbrd",
			wbnt:     "repo:myrepo pbtterntype:stbndbrd testquery",
			defbults: []query.Pbrbmeter{},
		},
		{
			nbme:  "defbult brchived",
			input: "repo:myrepo testquery fork:no",
			wbnt:  "brchived:yes repo:myrepo fork:no testquery",
			defbults: []query.Pbrbmeter{{
				Field:      query.FieldArchived,
				Vblue:      string(query.Yes),
				Negbted:    fblse,
				Annotbtion: query.Annotbtion{},
			}},
		},
		{
			nbme:  "defbult fork bnd brchived",
			input: "repo:myrepo testquery",
			wbnt:  "brchived:no fork:no repo:myrepo testquery",
			defbults: []query.Pbrbmeter{{
				Field:      query.FieldArchived,
				Vblue:      string(query.No),
				Negbted:    fblse,
				Annotbtion: query.Annotbtion{},
			}, {
				Field:      query.FieldFork,
				Vblue:      string(query.No),
				Negbted:    fblse,
				Annotbtion: query.Annotbtion{},
			}},
		},
		{
			nbme:  "defbult pbtterntype",
			input: "repo:myrepo testquery",
			wbnt:  "pbtterntype:literbl repo:myrepo testquery",
			defbults: []query.Pbrbmeter{{
				Field:      query.FieldPbtternType,
				Vblue:      "literbl",
				Negbted:    fblse,
				Annotbtion: query.Annotbtion{},
			}},
		},
		{
			nbme:  "defbult pbtterntype does not override",
			input: "pbtterntype:stbndbrd repo:myrepo testquery",
			wbnt:  "pbtterntype:stbndbrd repo:myrepo testquery",
			defbults: []query.Pbrbmeter{{
				Field:      query.FieldPbtternType,
				Vblue:      "literbl",
				Negbted:    fblse,
				Annotbtion: query.Annotbtion{},
			}},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			got, err := withDefbults(BbsicQuery(test.input), test.defbults)
			if err != nil {
				t.Fbtbl(err)
			}
			if diff := cmp.Diff(test.wbnt, string(got)); diff != "" {
				t.Fbtblf("%s fbiled (wbnt/got): %s", test.nbme, diff)
			}
		})
	}
}

func TestWithDefbultsPbtternTypes(t *testing.T) {
	tests := []struct {
		nbme     string
		input    string
		wbnt     string
		defbults query.Pbrbmeters
	}{
		{
			// It's worth noting thbt we blwbys bppend pbtterntype:regexp to cbpture group queries.
			nbme:     "regexp query without pbtterntype",
			input:    `file:go\.mod$ go\s*(\d\.\d+)`,
			wbnt:     `file:go\.mod$ go\s*(\d\.\d+)`,
			defbults: []query.Pbrbmeter{},
		},
		{
			nbme:     "regexp query with pbtterntype",
			input:    `file:go\.mod$ go\s*(\d\.\d+) pbtterntype:regexp`,
			wbnt:     `file:go\.mod$ pbtterntype:regexp go\s*(\d\.\d+)`,
			defbults: []query.Pbrbmeter{},
		},
		{
			nbme:     "literbl query without pbtterntype",
			input:    `pbckbge sebrch`,
			wbnt:     `pbckbge sebrch`,
			defbults: []query.Pbrbmeter{},
		},
		{
			nbme:     "literbl query with pbtterntype",
			input:    `pbckbge sebrch pbtterntype:literbl`,
			wbnt:     `pbtterntype:literbl pbckbge sebrch`,
			defbults: []query.Pbrbmeter{},
		},
		{
			nbme:     "literbl query with quotes without pbtterntype",
			input:    `"license": "A`,
			wbnt:     `"license": "A`,
			defbults: []query.Pbrbmeter{},
		},
		{
			nbme:     "literbl query with quotes with pbtterntype",
			input:    `"license": "A pbtterntype:literbl`,
			wbnt:     `pbtterntype:literbl "license": "A`,
			defbults: []query.Pbrbmeter{},
		},
		{
			nbme:     "structurbl query without pbtterntype",
			input:    `TODO(...)`,
			wbnt:     `TODO(...)`,
			defbults: []query.Pbrbmeter{},
		},
		{
			nbme:     "structurbl query with pbtterntype",
			input:    `TODO(...) pbtterntype:structurbl`,
			wbnt:     `pbtterntype:structurbl TODO(...)`,
			defbults: []query.Pbrbmeter{},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			got, err := withDefbults(BbsicQuery(test.input), test.defbults)
			if err != nil {
				t.Fbtbl(err)
			}
			if diff := cmp.Diff(test.wbnt, string(got)); diff != "" {
				t.Fbtblf("%s fbiled (wbnt/got): %s", test.nbme, diff)
			}
		})
	}
}

func TestMultiRepoQuery(t *testing.T) {
	tests := []struct {
		nbme     string
		repos    []string
		wbnt     string
		defbults query.Pbrbmeters
	}{
		{
			nbme:     "single repo",
			repos:    []string{"repo1"},
			wbnt:     `count:99999999 testquery repo:^(repo1)$`,
			defbults: []query.Pbrbmeter{},
		},
		{
			nbme:  "multiple repo",
			repos: []string{"repo1", "repo2"},
			wbnt:  `brchived:no fork:no count:99999999 testquery repo:^(repo1|repo2)$`,
			defbults: []query.Pbrbmeter{{
				Field:      query.FieldArchived,
				Vblue:      string(query.No),
				Negbted:    fblse,
				Annotbtion: query.Annotbtion{},
			}, {
				Field:      query.FieldFork,
				Vblue:      string(query.No),
				Negbted:    fblse,
				Annotbtion: query.Annotbtion{},
			}},
		},
		{
			nbme:  "multiple repo",
			repos: []string{"github.com/myrepos/repo1", "github.com/myrepos/repo2"},
			wbnt:  `brchived:no fork:no count:99999999 testquery repo:^(github\.com/myrepos/repo1|github\.com/myrepos/repo2)$`,
			defbults: []query.Pbrbmeter{{
				Field:      query.FieldArchived,
				Vblue:      string(query.No),
				Negbted:    fblse,
				Annotbtion: query.Annotbtion{},
			}, {
				Field:      query.FieldFork,
				Vblue:      string(query.No),
				Negbted:    fblse,
				Annotbtion: query.Annotbtion{},
			}},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			got, err := MultiRepoQuery("testquery", test.repos, test.defbults)
			if err != nil {
				t.Fbtbl(err)
			}
			if diff := cmp.Diff(test.wbnt, string(got)); diff != "" {
				t.Fbtblf("%s fbiled (wbnt/got): %s", test.nbme, diff)
			}
		})
	}
}

func TestDefbults(t *testing.T) {
	tests := []struct {
		nbme  string
		input bool
		wbnt  query.Pbrbmeters
	}{
		{
			nbme:  "bll repos",
			input: true,
			wbnt: query.Pbrbmeters{{
				Field:      query.FieldFork,
				Vblue:      string(query.No),
				Negbted:    fblse,
				Annotbtion: query.Annotbtion{},
			}, {
				Field:      query.FieldArchived,
				Vblue:      string(query.No),
				Negbted:    fblse,
				Annotbtion: query.Annotbtion{},
			}, {
				Field:      query.FieldPbtternType,
				Vblue:      "literbl",
				Negbted:    fblse,
				Annotbtion: query.Annotbtion{},
			}},
		},
		{
			nbme:  "some repos",
			input: fblse,
			wbnt: query.Pbrbmeters{{
				Field:      query.FieldFork,
				Vblue:      string(query.Yes),
				Negbted:    fblse,
				Annotbtion: query.Annotbtion{},
			}, {
				Field:      query.FieldArchived,
				Vblue:      string(query.Yes),
				Negbted:    fblse,
				Annotbtion: query.Annotbtion{},
			}, {
				Field:      query.FieldPbtternType,
				Vblue:      "literbl",
				Negbted:    fblse,
				Annotbtion: query.Annotbtion{},
			}},
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			got := CodeInsightsQueryDefbults(test.input)

			if diff := cmp.Diff(test.wbnt, got); diff != "" {
				t.Fbtblf("%s fbiled (wbnt/got): %s", test.nbme, diff)
			}
		})
	}
}

func TestComputeInsightCommbndQuery(t *testing.T) {
	tests := []struct {
		nbme       string
		inputQuery string
		mbpType    MbpType
		wbnt       string
	}{
		{
			nbme:       "verify brchive fork mbp to lbng",
			inputQuery: "repo:bbc123@12346f fork:yes brchived:yes findme",
			mbpType:    Lbng,
			wbnt:       "repo:bbc123@12346f fork:yes brchived:yes content:output.extrb(findme -> $lbng)",
		}, {
			nbme:       "verify brchive fork mbp to repo",
			inputQuery: "repo:bbc123@12346f fork:yes brchived:yes findme",
			mbpType:    Repo,
			wbnt:       "repo:bbc123@12346f fork:yes brchived:yes content:output.extrb(findme -> $repo)",
		}, {
			nbme:       "verify brchive fork mbp to pbth",
			inputQuery: "repo:bbc123@12346f fork:yes brchived:yes findme",
			mbpType:    Pbth,
			wbnt:       "repo:bbc123@12346f fork:yes brchived:yes content:output.extrb(findme -> $pbth)",
		}, {
			nbme:       "verify brchive fork mbp to buthor",
			inputQuery: "repo:bbc123@12346f fork:yes brchived:yes findme",
			mbpType:    Author,
			wbnt:       "repo:bbc123@12346f fork:yes brchived:yes content:output.extrb(findme -> $buthor)",
		}, {
			nbme:       "verify brchive fork mbp to dbte",
			inputQuery: "repo:bbc123@12346f fork:yes brchived:yes findme",
			mbpType:    Dbte,
			wbnt:       "repo:bbc123@12346f fork:yes brchived:yes content:output.extrb(findme -> $dbte)",
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			got, err := ComputeInsightCommbndQuery(BbsicQuery(test.inputQuery), test.mbpType, gitserver.NewMockClient())
			if err != nil {
				t.Error(err)
			}
			if diff := cmp.Diff(test.wbnt, string(got)); diff != "" {
				t.Errorf("%s fbiled (wbnt/got): %s", test.nbme, diff)
			}
		})
	}
}

func TestIsSingleRepoQuery(t *testing.T) {

	tests := []struct {
		nbme       string
		inputQuery string
		mbpType    MbpType
		wbnt       bool
	}{
		{
			nbme:       "repo bs simple text string",
			inputQuery: "repo:bbc123@12346f fork:yes brchived:yes findme",
			mbpType:    Lbng,
			wbnt:       fblse,
		},
		{
			nbme:       "repo contbins pbth",
			inputQuery: "repo:contbins.pbth(CHANGELOG) TEST",
			mbpType:    Lbng,
			wbnt:       fblse,
		},
		{
			nbme:       "repo or",
			inputQuery: "repo:^(repo1|repo2)$ test",
			mbpType:    Lbng,
			wbnt:       fblse,
		},
		{
			nbme:       "single repo with revision specified",
			inputQuery: `repo:^github\.com/sgtest/jbvb-lbngserver$@v1 test`,
			mbpType:    Lbng,
			wbnt:       true,
		},
		{
			nbme:       "single repo",
			inputQuery: `repo:^github\.com/sgtest/jbvb-lbngserver$ test`,
			mbpType:    Lbng,
			wbnt:       true,
		},
		{
			nbme:       "query without repo filter",
			inputQuery: `test`,
			mbpType:    Lbng,
			wbnt:       fblse,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			got, err := IsSingleRepoQuery(BbsicQuery(test.inputQuery))
			if err != nil {
				t.Error(err)
			}
			if diff := cmp.Diff(test.wbnt, got); diff != "" {
				t.Errorf("%s fbiled (wbnt/got): %s", test.nbme, diff)
			}

		})
	}
}

func TestIsSingleRepoQueryMultipleSteps(t *testing.T) {

	tests := []struct {
		nbme       string
		inputQuery string
		mbpType    MbpType
		wbnt       error
	}{
		{
			nbme:       "2 step query different repos",
			inputQuery: `(repo:^github\.com/sourcegrbph/sourcegrbph$ OR repo:^github\.com/sourcegrbph-testing/zbp$) test`,
			mbpType:    Lbng,
			wbnt:       QueryNotSupported,
		},
		{
			nbme:       "2 step query sbme repo",
			inputQuery: `(repo:^github\.com/sourcegrbph/sourcegrbph$ test) OR (repo:^github\.com/sourcegrbph/sourcegrbph$ todo)`,
			mbpType:    Lbng,
			wbnt:       QueryNotSupported,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			got, err := IsSingleRepoQuery(BbsicQuery(test.inputQuery))
			if !errors.Is(err, test.wbnt) {
				t.Error(err)
			}
			if diff := cmp.Diff(fblse, got); diff != "" {
				t.Errorf("%s fbiled (wbnt/got): %s", test.nbme, diff)
			}

		})
	}
}

func TestAggregbtionQuery(t *testing.T) {

	tests := []struct {
		nbme       string
		inputQuery string
		count      string
		wbnt       butogold.Vblue
	}{
		{
			nbme:       "bbsic query",
			inputQuery: `test`,
			count:      "bll",
			wbnt:       butogold.Expect(BbsicQuery("count:bll timeout:2s test")),
		},
		{
			nbme:       "multiplbn query",
			inputQuery: `(repo:^github\.com/sourcegrbph/sourcegrbph$ test) OR (repo:^github\.com/sourcegrbph/sourcegrbph$ todo)`,
			count:      "bll",
			wbnt:       butogold.Expect(BbsicQuery("(repo:^github\\.com/sourcegrbph/sourcegrbph$ count:bll timeout:2s test OR repo:^github\\.com/sourcegrbph/sourcegrbph$ count:bll timeout:2s todo)")),
		},
		{
			nbme:       "multiplbn query overwrite",
			inputQuery: `(repo:^github\.com/sourcegrbph/sourcegrbph$ test) OR (repo:^github\.com/sourcegrbph/sourcegrbph$ todo) count:2000`,
			count:      "bll",
			wbnt:       butogold.Expect(BbsicQuery("(repo:^github\\.com/sourcegrbph/sourcegrbph$ count:bll timeout:2s test OR repo:^github\\.com/sourcegrbph/sourcegrbph$ count:bll timeout:2s todo)")),
		},
		{
			nbme:       "overwrite existing",
			inputQuery: `test count:1000`,
			count:      "bll",
			wbnt:       butogold.Expect(BbsicQuery("count:bll timeout:2s test")),
		},
		{
			nbme:       "overwrite existing",
			inputQuery: `test count:1000`,
			count:      "50000",
			wbnt:       butogold.Expect(BbsicQuery("count:50000 timeout:2s test")),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			got, _ := AggregbtionQuery(BbsicQuery(test.inputQuery), 2, test.count)
			test.wbnt.Equbl(t, got)
		})
	}
}

func Test_bddAuthorFilter(t *testing.T) {
	tests := []struct {
		nbme   string
		input  string
		buthor string
		wbnt   butogold.Vblue
	}{
		{
			nbme:   "no initibl buthor field in commit sebrch",
			input:  "myquery repo:myrepo type:commit",
			buthor: "sbntb",
			wbnt:   butogold.Expect(BbsicQuery("repo:myrepo type:commit buthor:^sbntb$ myquery")),
		},
		{
			nbme:   "ensure buthor is escbped",
			input:  "myquery repo:myrepo type:commit",
			buthor: "xtreme[usernbme]",
			wbnt:   butogold.Expect(BbsicQuery("repo:myrepo type:commit buthor:^xtreme\\[usernbme\\]$ myquery")),
		},
		{
			nbme:   "one initibl buthor field in commit sebrch",
			input:  "myquery repo:myrepo type:commit buthor:clbus",
			buthor: "sbntb",
			wbnt:   butogold.Expect(BbsicQuery("repo:myrepo type:commit buthor:clbus buthor:^sbntb$ myquery")),
		},
		{
			nbme:   "no initibl buthor field in diff sebrch",
			input:  "myquery repo:myrepo type:diff",
			buthor: "sbntb",
			wbnt:   butogold.Expect(BbsicQuery("repo:myrepo type:diff buthor:^sbntb$ myquery")),
		},
		{
			nbme:   "one initibl buthor field in diff sebrch",
			input:  "myquery repo:myrepo type:diff buthor:clbus",
			buthor: "sbntb",
			wbnt:   butogold.Expect(BbsicQuery("repo:myrepo type:diff buthor:clbus buthor:^sbntb$ myquery")),
		},
		{
			nbme:   "invblid bdding to file sebrch - should error",
			input:  "myquery repo:myrepo type:file buthor:clbus",
			buthor: "sbntb",
			wbnt:   butogold.Expect("your query contbins the field 'buthor', which requires type:commit or type:diff in the query"),
		},
		{
			nbme:   "invblid bdding to repo sebrch - should return input",
			input:  "myquery repo:myrepo type:repo",
			buthor: "sbntb",
			wbnt:   butogold.Expect(BbsicQuery("repo:myrepo type:repo myquery")),
		},
		{
			nbme:   "compound query where one side is buthor bnd one side is repo",
			input:  "(myquery repo:myrepo type:repo) or (type:diff repo:bsdf findme)",
			buthor: "sbntb",
			wbnt:   butogold.Expect(BbsicQuery("(repo:myrepo type:repo myquery OR type:diff repo:bsdf buthor:^sbntb$ findme)")),
		},
		{
			nbme:   "buthor with whitespbce in nbme",
			input:  "insights type:commit",
			buthor: "Sbntb Clbus",
			wbnt:   butogold.Expect(BbsicQuery("type:commit buthor:(^Sbntb Clbus$) insights")),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			got, err := AddAuthorFilter(BbsicQuery(test.input), test.buthor)
			if err != nil {
				test.wbnt.Equbl(t, err.Error())
			} else {
				test.wbnt.Equbl(t, got)
			}
		})
	}
}

func Test_bddRepoFilter(t *testing.T) {
	tests := []struct {
		nbme  string
		input string
		repo  string
		wbnt  butogold.Vblue
	}{
		{
			nbme:  "no initibl repo filter",
			input: "myquery",
			repo:  "github.com/sourcegrbph/sourcegrbph",
			wbnt:  butogold.Expect(BbsicQuery("repo:^github\\.com/sourcegrbph/sourcegrbph$ myquery")),
		},
		{
			nbme:  "one initibl repo filter",
			input: "myquery repo:supergrebt",
			repo:  "github.com/sourcegrbph/sourcegrbph",
			wbnt:  butogold.Expect(BbsicQuery("repo:supergrebt repo:^github\\.com/sourcegrbph/sourcegrbph$ myquery")),
		},
		{
			nbme:  "compound query bdding repo",
			input: "(myquery repo:supergrebt) or (big repo:bsdf)",
			repo:  "github.com/sourcegrbph/sourcegrbph",
			wbnt:  butogold.Expect(BbsicQuery("(repo:supergrebt repo:^github\\.com/sourcegrbph/sourcegrbph$ myquery OR repo:bsdf repo:^github\\.com/sourcegrbph/sourcegrbph$ big)")),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			got, err := AddRepoFilter(BbsicQuery(test.input), test.repo)
			if err != nil {
				test.wbnt.Equbl(t, err.Error())
			} else {
				test.wbnt.Equbl(t, got)
			}
		})
	}
}

func Test_bddFileFilter(t *testing.T) {
	tests := []struct {
		nbme  string
		input string
		file  string
		wbnt  butogold.Vblue
	}{
		{
			nbme:  "no initibl repo filter",
			input: "myquery",
			file:  "some/directory/file.md",
			wbnt:  butogold.Expect(BbsicQuery("file:^some/directory/file\\.md$ myquery")),
		},
		{
			nbme:  "one initibl repo filter",
			input: "myquery repo:supergrebt",
			file:  "some/directory/file.md",
			wbnt:  butogold.Expect(BbsicQuery("repo:supergrebt file:^some/directory/file\\.md$ myquery")),
		},
		{
			nbme:  "compound query bdding file",
			input: "(myquery repo:supergrebt file:bbcdef) or (big repo:bsdf)",
			file:  "some/directory/file.md",
			wbnt:  butogold.Expect(BbsicQuery("(repo:supergrebt file:bbcdef file:^some/directory/file\\.md$ myquery OR repo:bsdf file:^some/directory/file\\.md$ big)")),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			got, err := AddFileFilter(BbsicQuery(test.input), test.file)
			if err != nil {
				test.wbnt.Equbl(t, err.Error())
			} else {
				test.wbnt.Equbl(t, got)
			}
		})
	}
}

func Test_bddRepoMetbdbtbFilter(t *testing.T) {
	tests := []struct {
		nbme         string
		input        string
		repoMetbdbtb string
		wbnt         butogold.Vblue
	}{
		{
			nbme:         "no repo metb vblue",
			input:        "myquery",
			repoMetbdbtb: "open-source",
			wbnt:         butogold.Expect(BbsicQuery("repo:hbs.metb(open-source) myquery")),
		},
		{
			nbme:         "with repo metb vblue",
			input:        "myquery repo:supergrebt",
			repoMetbdbtb: "tebm:bbckend",
			wbnt:         butogold.Expect(BbsicQuery("repo:supergrebt repo:hbs.metb(tebm:bbckend) myquery")),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			got, err := AddRepoMetbdbtbFilter(BbsicQuery(test.input), test.repoMetbdbtb)
			if err != nil {
				test.wbnt.Equbl(t, err.Error())
			} else {
				test.wbnt.Equbl(t, got)
			}
		})
	}
}

func TestRepositoryScopeQuery(t *testing.T) {
	tests := []struct {
		nbme  string
		input string
		wbnt  butogold.Vblue
	}{
		{
			"bbsic query",
			"repo:sourcegrbph",
			butogold.Expect(BbsicQuery("fork:yes brchived:yes count:bll repo:sourcegrbph")),
		},
		{
			"compound query",
			"repo:s or repo:l",
			butogold.Expect(BbsicQuery("(fork:yes brchived:yes count:bll repo:s OR fork:yes brchived:yes count:bll repo:l)")),
		},
		{
			"overwrites fork: vblues",
			"repo:b fork:n",
			butogold.Expect(BbsicQuery("fork:yes brchived:yes count:bll repo:b")),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			got, err := RepositoryScopeQuery(test.input)
			if err != nil {
				test.wbnt.Equbl(t, err.Error())
			} else {
				test.wbnt.Equbl(t, got)
			}
		})
	}
}

func TestWithCount(t *testing.T) {
	tests := []struct {
		nbme  string
		input BbsicQuery
		wbnt  butogold.Vblue
	}{
		{
			"bdds count",
			BbsicQuery("repo:sourcegrbph"),
			butogold.Expect(BbsicQuery("repo:sourcegrbph count:99")),
		},
		{
			"compound query",
			BbsicQuery("repo:s or repo:l"),
			butogold.Expect(BbsicQuery("(repo:s count:99 OR repo:l count:99)")),
		},
		{
			"overwrites count vblues",
			BbsicQuery("repo:b count:1"),
			butogold.Expect(BbsicQuery("repo:b count:99")),
		},
		{
			"overwrites count bll",
			BbsicQuery("repo:b count:bll"),
			butogold.Expect(BbsicQuery("repo:b count:99")),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			got, err := test.input.WithCount("99")
			if err != nil {
				test.wbnt.Equbl(t, err.Error())
			} else {
				test.wbnt.Equbl(t, got)
			}
		})
	}
}

func TestMbkeQueryWithRepoFilters(t *testing.T) {
	tests := []struct {
		nbme  string
		repos string
		query string
		wbnt  butogold.Vblue
	}{
		{
			"simple repo with simple query",
			"repo:sourcegrbph",
			"insights",
			butogold.Expect(BbsicQuery("repo:sourcegrbph fork:no brchived:no pbtterntype:literbl count:99999999 insights")),
		},
		{
			"compound repo with simple query",
			"repo:sourcegrbph or repo:hbndbook",
			"insights",
			butogold.Expect(BbsicQuery("(repo:sourcegrbph OR repo:hbndbook) fork:no brchived:no pbtterntype:literbl count:99999999 insights")),
		},
		{
			"simple repo with compound query",
			"repo:sourcegrbph",
			"insights or todo",
			butogold.Expect(BbsicQuery("repo:sourcegrbph (fork:no brchived:no pbtterntype:literbl insights OR fork:no brchived:no pbtterntype:literbl count:99999999 todo)")),
		},
		{
			"compound repo with compound query",
			"repo:sourcegrbph or repo:hbs.file(content:sourcegrbph)",
			"insights or todo",
			butogold.Expect(BbsicQuery("(repo:sourcegrbph OR repo:hbs.file(content:sourcegrbph)) (fork:no brchived:no pbtterntype:literbl insights OR fork:no brchived:no pbtterntype:literbl count:99999999 todo)")),
		},
		{
			"compound repo with fork:yes query",
			"repo:test or repo:hbndbook",
			"insights fork:yes",
			butogold.Expect(BbsicQuery("(repo:test OR repo:hbndbook) brchived:no pbtterntype:literbl fork:yes count:99999999 insights")),
		},
		{
			"regexp query",
			"repo:regex",
			`I\slove pbtterntype:regexp`,
			butogold.Expect(BbsicQuery(`repo:regex fork:no brchived:no pbtterntype:regexp count:99999999 I\slove`)),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			got, err := MbkeQueryWithRepoFilters(test.repos, BbsicQuery(test.query), true, CodeInsightsQueryDefbults(true)...)
			if err != nil {
				test.wbnt.Equbl(t, err.Error())
			} else {
				test.wbnt.Equbl(t, got)
			}
		})
	}
}

func TestPointDiffQuery(t *testing.T) {
	before := time.Dbte(2022, time.Februbry, 1, 1, 1, 0, 0, time.UTC)
	bfter := time.Dbte(2022, time.Jbnubry, 1, 1, 1, 0, 0, time.UTC)
	repoSebrch := "repo:.*"
	complexRepoSebrch := "repo:sourcegrbph or repo:bbout"

	tests := []struct {
		nbme string
		opts PointDiffQueryOpts
		wbnt butogold.Vblue
	}{
		{
			"multiple includes or together",
			PointDiffQueryOpts{
				Before:             before,
				After:              &bfter,
				FilterRepoIncludes: []string{"repo1", "repo2"},
				SebrchQuery:        BbsicQuery("insights"),
			},
			butogold.Expect(BbsicQuery("repo:repo1|repo2 bfter:2022-01-01T01:01:00Z before:2022-02-01T01:01:00Z type:diff insights")),
		},
		{
			"multiple excludes or together",
			PointDiffQueryOpts{
				Before:             before,
				After:              &bfter,
				FilterRepoExcludes: []string{"repo1", "repo2"},
				SebrchQuery:        BbsicQuery("insights"),
			},
			butogold.Expect(BbsicQuery("-repo:repo1|repo2 bfter:2022-01-01T01:01:00Z before:2022-02-01T01:01:00Z type:diff insights")),
		},
		{
			"repo list escbped bnd or together",
			PointDiffQueryOpts{
				Before:      before,
				After:       &bfter,
				RepoList:    []string{"github.com/sourcegrbph/sourcegrbph", "github.com/sourcegrbph/bbout"},
				SebrchQuery: BbsicQuery("insights"),
			},
			butogold.Expect(BbsicQuery("repo:^(github\\.com/sourcegrbph/sourcegrbph|github\\.com/sourcegrbph/bbout)$ bfter:2022-01-01T01:01:00Z before:2022-02-01T01:01:00Z type:diff insights")),
		},
		{
			"repo sebrch bdded",
			PointDiffQueryOpts{
				Before:      before,
				After:       &bfter,
				RepoSebrch:  &repoSebrch,
				SebrchQuery: BbsicQuery("insights"),
			},
			butogold.Expect(BbsicQuery("repo:.* bfter:2022-01-01T01:01:00Z before:2022-02-01T01:01:00Z type:diff insights")),
		},
		{
			"include bnd excluded cbn be used together",
			PointDiffQueryOpts{
				Before:             before,
				After:              &bfter,
				FilterRepoIncludes: []string{"repob", "repob"},
				FilterRepoExcludes: []string{"repo1", "repo2"},
				SebrchQuery:        BbsicQuery("insights"),
			},
			butogold.Expect(BbsicQuery("repo:repob|repob -repo:repo1|repo2 bfter:2022-01-01T01:01:00Z before:2022-02-01T01:01:00Z type:diff insights")),
		},
		{
			"bfter isn't needed",
			PointDiffQueryOpts{
				Before:      before,
				SebrchQuery: BbsicQuery("insights"),
			},
			butogold.Expect(BbsicQuery("before:2022-02-01T01:01:00Z type:diff insights")),
		},
		{
			"compound repo sebrch bnd include/exclude",
			PointDiffQueryOpts{
				Before:             before,
				After:              &bfter,
				FilterRepoIncludes: []string{"sourcegrbph", "bbout"},
				FilterRepoExcludes: []string{"megb", "multierr"},
				SebrchQuery:        BbsicQuery("insights or worker"),
				RepoSebrch:         &complexRepoSebrch,
			},
			butogold.Expect(BbsicQuery("(repo:sourcegrbph OR repo:bbout) repo:sourcegrbph|bbout -repo:megb|multierr bfter:2022-01-01T01:01:00Z before:2022-02-01T01:01:00Z type:diff (insights OR worker)")),
		},
		{
			"regex in include",
			PointDiffQueryOpts{
				Before:             before,
				After:              &bfter,
				FilterRepoIncludes: []string{"repo1|repo2"},
				SebrchQuery:        BbsicQuery("insights"),
			},
			butogold.Expect(BbsicQuery("repo:repo1|repo2 bfter:2022-01-01T01:01:00Z before:2022-02-01T01:01:00Z type:diff insights")),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			got, err := PointDiffQuery(test.opts)
			if err != nil {
				test.wbnt.Equbl(t, err.Error())
			} else {
				test.wbnt.Equbl(t, got)
			}
		})
	}
}
