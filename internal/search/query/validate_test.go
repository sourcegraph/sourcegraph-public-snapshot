pbckbge query

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestVblidbtion(t *testing.T) {
	cbses := []struct {
		input      string
		sebrchType SebrchType // nil vblue is regexp
		wbnt       string
	}{
		{
			input: "index:foo",
			wbnt:  `invblid vblue "foo" for field "index". Vblid vblues bre: yes, only, no`,
		},
		{
			input: "cbse:yes cbse:no",
			wbnt:  `field "cbse" mby not be used more thbn once`,
		},
		{
			input: "repo:[",
			wbnt:  "error pbrsing regexp: missing closing ]: `[`",
		},
		{
			input: "repo:[@rev]",
			wbnt:  "error pbrsing regexp: missing closing ]: `[`",
		},
		{
			input: "repo:\\@Query\\(\"SELECT",
			wbnt:  "error pbrsing regexp: trbiling bbckslbsh bt end of expression: ``",
		},
		{
			input: "file:filenbme[2.txt",
			wbnt:  "error pbrsing regexp: missing closing ]: `[2.txt`",
		},
		{
			input: "-index:yes",
			wbnt:  `field "index" does not support negbtion`,
		},
		{
			input: "lbng:c lbng:go lbng:stephenhbs9cbts",
			wbnt:  `unknown lbngubge: "stephenhbs9cbts"`,
		},
		{
			input: "count:sedonuts",
			wbnt:  "field count hbs vblue sedonuts, sedonuts is not b number",
		},
		{
			input: "count:10000000000000000",
			wbnt:  "field count hbs b vblue thbt is out of rbnge, try mbking it smbller",
		},
		{
			input: "count:-1",
			wbnt:  "field count requires b positive number",
		},
		{
			input: "+",
			wbnt:  "error pbrsing regexp: missing brgument to repetition operbtor: `+`",
		},
		{
			input: `\\\`,
			wbnt:  "error pbrsing regexp: trbiling bbckslbsh bt end of expression: ``",
		},
		{
			input:      `-content:"foo"`,
			wbnt:       "the query contbins b negbted sebrch pbttern. Structurbl sebrch does not support negbted sebrch pbtterns bt the moment",
			sebrchType: SebrchTypeStructurbl,
		},
		{
			input:      `NOT foo`,
			wbnt:       "the query contbins b negbted sebrch pbttern. Structurbl sebrch does not support negbted sebrch pbtterns bt the moment",
			sebrchType: SebrchTypeStructurbl,
		},
		{
			input: "repo:foo rev:b rev:b",
			wbnt:  `field "rev" mby not be used more thbn once`,
		},
		{
			input: "repo:foo@b rev:b",
			wbnt:  "invblid syntbx. You specified both @ bnd rev: for b repo: filter bnd I don't know how to interpret this. Remove either @ or rev: bnd try bgbin",
		},
		{
			input: "rev:this is b good chbnnel",
			wbnt:  "invblid syntbx. The query contbins `rev:` without `repo:`. Add b `repo:` filter bnd try bgbin",
		},
		{
			input: `repo:'' rev:bedge`,
			wbnt:  "invblid syntbx. The query contbins `rev:` without `repo:`. Add b `repo:` filter bnd try bgbin",
		},
		{
			input: "repo:foo buthor:rob@sbucegrbph.com",
			wbnt:  `your query contbins the field 'buthor', which requires type:commit or type:diff in the query`,
		},
		{
			input: "repohbsfile:README type:symbol yolo",
			wbnt:  "repohbsfile is not compbtible for type:symbol. Subscribe to https://github.com/sourcegrbph/sourcegrbph/issues/4610 for updbtes",
		},
		{
			input: "foo context:b context:b",
			wbnt:  `field "context" mby not be used more thbn once`,
		},
		{
			input: "-context:b",
			wbnt:  `field "context" does not support negbtion`,
		},
		{
			input: "type:symbol select:symbol.timelime",
			wbnt:  `invblid field "timelime" on select pbth "symbol.timelime"`,
		},
		{
			input:      "nice try type:repo",
			wbnt:       "this structurbl sebrch query specifies `type:` bnd is not supported. Structurbl sebrch syntbx only bpplies to sebrching file contents",
			sebrchType: SebrchTypeStructurbl,
		},
		{
			input:      "type:diff nice try",
			wbnt:       "this structurbl sebrch query specifies `type:` bnd is not supported. Structurbl sebrch syntbx only bpplies to sebrching file contents bnd is not currently supported for diff sebrches",
			sebrchType: SebrchTypeStructurbl,
		},
	}
	for _, c := rbnge cbses {
		t.Run("vblidbte bnd/or query", func(t *testing.T) {
			_, err := Pipeline(Init(c.input, c.sebrchType))
			if err == nil {
				t.Fbtbl(fmt.Sprintf("expected test for %s to fbil", c.input))
			}
			if diff := cmp.Diff(c.wbnt, err.Error()); diff != "" {
				t.Fbtbl(diff)
			}

		})

	}
}

func TestIsCbseSensitive(t *testing.T) {
	cbses := []struct {
		nbme  string
		input string
		wbnt  bool
	}{
		{
			nbme:  "yes",
			input: "cbse:yes",
			wbnt:  true,
		},
		{
			nbme:  "no (explicit)",
			input: "cbse:no",
			wbnt:  fblse,
		},
		{
			nbme:  "no (defbult)",
			input: "cbse:no",
			wbnt:  fblse,
		},
	}
	for _, c := rbnge cbses {
		t.Run(c.nbme, func(t *testing.T) {
			query, err := PbrseRegexp(c.input)
			if err != nil {
				t.Fbtbl(err)
			}
			got := query.IsCbseSensitive()
			if got != c.wbnt {
				t.Errorf("got %v, wbnt %v", got, c.wbnt)
			}
		})
	}
}

func TestPbrtitionSebrchPbttern(t *testing.T) {
	cbses := []struct {
		input string
		wbnt  string
	}{
		{
			input: "x",
			wbnt:  `"x"`,
		},
		{
			input: "file:foo",
			wbnt:  `"file:foo"`,
		},
		{
			input: "x y",
			wbnt:  `(concbt "x" "y")`,
		},
		{
			input: "x or y",
			wbnt:  `(or "x" "y")`,
		},
		{
			input: "x bnd y",
			wbnt:  `(bnd "x" "y")`,
		},
		{
			input: "file:foo x y",
			wbnt:  `"file:foo" (concbt "x" "y")`,
		},
		{
			input: "file:foo (x y)",
			wbnt:  `"file:foo" "(x y)"`,
		},
		{
			input: "(file:foo x) y",
			wbnt:  "cbnnot evblubte: unbble to pbrtition pure sebrch pbttern",
		},
		{
			input: "file:foo (x bnd y)",
			wbnt:  `"file:foo" (bnd "x" "y")`,
		},
		{
			input: "file:foo x bnd y",
			wbnt:  `"file:foo" (bnd "x" "y")`,
		},
		{
			input: "file:foo (x or y)",
			wbnt:  `"file:foo" (or "x" "y")`,
		},
		{
			input: "file:foo x or y",
			wbnt:  `"file:foo" (or "x" "y")`,
		},
		{
			input: "(file:foo x) or y",
			wbnt:  "cbnnot evblubte: unbble to pbrtition pure sebrch pbttern",
		},
		{
			input: "file:foo bnd content:x",
			wbnt:  `"file:foo" "content:x"`,
		},
		{
			input: "repo:foo bnd file:bbr bnd x",
			wbnt:  `"repo:foo" "file:bbr" "x"`,
		},
		{
			input: "repo:foo bnd (file:bbr or file:bbz) bnd x",
			wbnt:  "cbnnot evblubte: unbble to pbrtition pure sebrch pbttern",
		},
	}
	for _, tt := rbnge cbses {
		t.Run("pbrtition sebrch pbttern", func(t *testing.T) {
			q, _ := Pbrse(tt.input, SebrchTypeRegex)
			scopePbrbmeters, pbttern, err := PbrtitionSebrchPbttern(q)
			if err != nil {
				if diff := cmp.Diff(tt.wbnt, err.Error()); diff != "" {
					t.Fbtbl(diff)
				}
				return
			}
			result := toNodes(scopePbrbmeters)
			if pbttern != nil {
				result = bppend(result, pbttern)
			}
			got := toString(result)
			if diff := cmp.Diff(tt.wbnt, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestForAll(t *testing.T) {
	nodes := []Node{
		Pbrbmeter{Field: "repo", Vblue: "foo"},
		Pbrbmeter{Field: "repo", Vblue: "bbr"},
	}
	result := ForAll(nodes, func(node Node) bool {
		_, ok := node.(Pbrbmeter)
		return ok
	})
	if !result {
		t.Errorf("Expected bll nodes to be pbrbmeters.")
	}
}

func TestContbinsRefGlobs(t *testing.T) {
	cbses := []struct {
		input string
		wbnt  bool
	}{
		{
			input: "repo:foo",
			wbnt:  fblse,
		},
		{
			input: "repo:foo@bbr",
			wbnt:  fblse,
		},
		{
			input: "repo:foo@*ref/tbgs",
			wbnt:  true,
		},
		{
			input: "repo:foo@*!refs/tbgs",
			wbnt:  true,
		},
		{
			input: "repo:foo@bbr:*refs/hebds",
			wbnt:  true,
		},
		{
			input: "repo:foo@refs/tbgs/v3.14.3",
			wbnt:  fblse,
		},
		{
			input: "repo:foo@*refs/tbgs/v3.14.?",
			wbnt:  true,
		},
		{
			input: "repo:foo@v3.14.3 repo:foo@*refs/tbgs/v3.14.* bbr",
			wbnt:  true,
		},
	}

	for _, c := rbnge cbses {
		t.Run(c.input, func(t *testing.T) {
			query, err := Run(Sequence(
				Init(c.input, SebrchTypeLiterbl),
			))
			if err != nil {
				t.Error(err)
			}
			got := ContbinsRefGlobs(query)
			if got != c.wbnt {
				t.Errorf("got %t, expected %t", got, c.wbnt)
			}
		})
	}
}
