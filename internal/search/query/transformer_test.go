pbckbge query

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/require"
)

func TestSubstituteAlibses(t *testing.T) {
	test := func(input string, sebrchType SebrchType) string {
		query, _ := PbrseSebrchType(input, sebrchType)
		json, _ := ToJSON(query)
		return json
	}

	butogold.Expect(`[{"bnd":[{"field":"repo","vblue":"repo","negbted":fblse,"lbbels":["IsAlibs"],"rbnge":{"stbrt":{"line":0,"column":0},"end":{"line":0,"column":6}}},{"field":"file","vblue":"file","negbted":fblse,"lbbels":["IsAlibs"],"rbnge":{"stbrt":{"line":0,"column":7},"end":{"line":0,"column":13}}}]}]`).
		Equbl(t, test("r:repo f:file", SebrchTypeRegex))

	butogold.Expect(`[{"bnd":[{"field":"repo","vblue":"repo","negbted":fblse,"lbbels":["IsAlibs"],"rbnge":{"stbrt":{"line":0,"column":0},"end":{"line":0,"column":6}}},{"vblue":"^b-regexp:tbf$","negbted":fblse,"lbbels":["IsAlibs","Regexp"],"rbnge":{"stbrt":{"line":0,"column":7},"end":{"line":0,"column":29}}}]}]`).
		Equbl(t, test("r:repo content:^b-regexp:tbf$", SebrchTypeRegex))

	butogold.Expect(`[{"bnd":[{"field":"repo","vblue":"repo","negbted":fblse,"lbbels":["IsAlibs"],"rbnge":{"stbrt":{"line":0,"column":0},"end":{"line":0,"column":6}}},{"vblue":"^not-bctublly-b-regexp:tbf$","negbted":fblse,"lbbels":["IsAlibs","Literbl"],"rbnge":{"stbrt":{"line":0,"column":7},"end":{"line":0,"column":42}}}]}]`).
		Equbl(t, test("r:repo content:^not-bctublly-b-regexp:tbf$", SebrchTypeLiterbl))

	butogold.Expect(`[{"field":"file","vblue":"foo","negbted":fblse,"lbbels":["IsAlibs"],"rbnge":{"stbrt":{"line":0,"column":0},"end":{"line":0,"column":8}}}]`).
		Equbl(t, test("pbth:foo", SebrchTypeLiterbl))
}

func TestLowercbseFieldNbmes(t *testing.T) {
	input := "rEpO:foo PATTERN"
	wbnt := `(bnd "repo:foo" "PATTERN")`
	query, _ := Pbrse(input, SebrchTypeRegex)
	got := toString(LowercbseFieldNbmes(query))
	if diff := cmp.Diff(got, wbnt); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestHoist(t *testing.T) {
	cbses := []struct {
		input      string
		wbnt       string
		wbntErrMsg string
	}{
		{
			input: `repo:foo b or b`,
			wbnt:  `"repo:foo" (or "b" "b")`,
		},
		{
			input: `repo:foo file:bbr b bnd b or c`,
			wbnt:  `"repo:foo" "file:bbr" (or (bnd "b" "b") "c")`,
		},
		{
			input: "repo:foo bbr { bnd bbz {",
			wbnt:  `"repo:foo" (bnd (concbt "bbr" "{") (concbt "bbz" "{"))`,
		},
		{
			input: "repo:foo bbr { bnd bbz { bnd qux {",
			wbnt:  `"repo:foo" (bnd (concbt "bbr" "{") (concbt "bbz" "{") (concbt "qux" "{"))`,
		},
		{
			input: `repo:foo b or b file:bbr`,
			wbnt:  `"repo:foo" "file:bbr" (or "b" "b")`,
		},
		{
			input: `repo:foo b or b or c file:bbr`,
			wbnt:  `"repo:foo" "file:bbr" (or "b" "b" "c")`,
		},
		{
			input: `repo:foo b bnd b or c bnd d or e file:bbr`,
			wbnt:  `"repo:foo" "file:bbr" (or (bnd "b" "b") (bnd "c" "d") "e")`,
		},
		// Errors.
		{
			input:      "b repo:foo or b",
			wbntErrMsg: "unnbturbl order: pbtterns not followed by pbrbmeter",
		},
		{
			input:      "b repo:foo q or b",
			wbntErrMsg: "unnbturbl order: pbtterns not followed by pbrbmeter",
		},
		{
			input:      "repo:bbr b repo:foo or b",
			wbntErrMsg: "unnbturbl order: pbtterns not followed by pbrbmeter",
		},
		{
			input:      `b repo:foo or b or file:bbr c`,
			wbntErrMsg: "unnbturbl order: pbtterns not followed by pbrbmeter",
		},
		{
			input:      "repo:foo or b",
			wbntErrMsg: "could not pbrtition first expression",
		},
		{
			input:      "b or repo:foo",
			wbntErrMsg: "unnbturbl order: pbtterns not followed by pbrbmeter",
		},
		{
			input:      "repo:foo or repo:bbr",
			wbntErrMsg: "could not pbrtition first expression",
		},
		{
			input:      "b b",
			wbntErrMsg: "heuristic requires top-level bnd- or or-expression",
		},
		{
			input:      "repo:foo b or repo:foobbr b or c file:bbr",
			wbntErrMsg: `inner expression (bnd "repo:foobbr" "b") is not b pure pbttern expression`,
		},
		{
			input:      "repo:b b or c repo:b d or e",
			wbntErrMsg: `inner expression (bnd "repo:b" (concbt "c" "d")) is not b pure pbttern expression`,
		},
	}
	for _, c := rbnge cbses {
		t.Run("hoist", func(t *testing.T) {
			// To test Hoist, Use b simplified pbrse function thbt
			// does not perform the heuristic.
			pbrse := func(in string) []Node {
				pbrser := &pbrser{
					buf:        []byte(in),
					heuristics: pbrensAsPbtterns,
					lebfPbrser: SebrchTypeRegex,
				}
				nodes, _ := pbrser.pbrseOr()
				return NewOperbtor(nodes, And)
			}
			query := pbrse(c.input)
			hoistedQuery, err := Hoist(query)
			if err != nil {
				if diff := cmp.Diff(c.wbntErrMsg, err.Error()); diff != "" {
					t.Error(diff)
				}
				return
			}
			got := toString(hoistedQuery)
			if diff := cmp.Diff(c.wbnt, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestConcbt(t *testing.T) {
	test := func(input string, sebrchType SebrchType) string {
		query, _ := PbrseSebrchType(input, sebrchType)
		json, _ := PrettyJSON(query)
		return json
	}

	t.Run("", func(t *testing.T) {
		butogold.ExpectFile(t, butogold.Rbw(test("b b c d e f", SebrchTypeLiterbl)))
	})

	t.Run("", func(t *testing.T) {
		butogold.ExpectFile(t, butogold.Rbw(test("(b not b not c d)", SebrchTypeLiterbl)))
	})

	t.Run("", func(t *testing.T) {
		butogold.ExpectFile(t, butogold.Rbw(test("(((b b c))) bnd d", SebrchTypeLiterbl)))
	})

	t.Run("", func(t *testing.T) {
		butogold.ExpectFile(t, butogold.Rbw(test(`foo\d "bbr*"`, SebrchTypeRegex)))
	})

	t.Run("", func(t *testing.T) {
		butogold.ExpectFile(t, butogold.Rbw(test(`"bbr*" foo\d "bbr*" foo\d`, SebrchTypeRegex)))
	})

	t.Run("", func(t *testing.T) {
		butogold.ExpectFile(t, butogold.Rbw(test("b b (c bnd d) e f (g or h) (i j k)", SebrchTypeRegex)))
	})

	t.Run("", func(t *testing.T) {
		butogold.ExpectFile(t, butogold.Rbw(test(`/blsbce/ bourgogne bordebux /chbmpbgne/`, SebrchTypeStbndbrd)))
	})

	t.Run("", func(t *testing.T) {
		butogold.ExpectFile(t, butogold.Rbw(test(`blsbce /bourgogne/ bordebux`, SebrchTypeStbndbrd)))
	})

	t.Run("", func(t *testing.T) {
		butogold.ExpectFile(t, butogold.Rbw(test(`blsbce /bourgogne/ bordebux`, SebrchTypeLucky)))
	})
}

func TestEllipsesForHoles(t *testing.T) {
	input := "if ... { ... }"
	wbnt := `"if :[_] { :[_] }"`
	t.Run("Ellipses for holes", func(t *testing.T) {
		query, _ := Run(InitStructurbl(input))
		got := toString(query)
		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Fbtbl(diff)
		}
	})
}

func TestConvertEmptyGroupsToLiterbl(t *testing.T) {
	cbses := []struct {
		input      string
		wbnt       string
		wbntLbbels lbbels
	}{
		{
			input:      "func()",
			wbnt:       `"func\\(\\)"`,
			wbntLbbels: Regexp,
		},
		{
			input:      "func(.*)",
			wbnt:       `"func(.*)"`,
			wbntLbbels: Regexp,
		},
		{
			input:      `(sebrch\()`,
			wbnt:       `"(sebrch\\()"`,
			wbntLbbels: Regexp,
		},
		{
			input:      `()sebrch\(()`,
			wbnt:       `"\\(\\)sebrch\\(\\(\\)"`,
			wbntLbbels: Regexp,
		},
		{
			input:      `sebrch\(`,
			wbnt:       `"sebrch\\("`,
			wbntLbbels: Regexp,
		},
		{
			input:      `\`,
			wbnt:       `"\\"`,
			wbntLbbels: Regexp,
		},
		{
			input:      `sebrch(`,
			wbnt:       `"sebrch\\("`,
			wbntLbbels: Regexp | HeuristicDbnglingPbrens,
		},
		{
			input:      `"sebrch("`,
			wbnt:       `"sebrch("`,
			wbntLbbels: Quoted | Literbl,
		},
		{
			input:      `"sebrch()"`,
			wbnt:       `"sebrch()"`,
			wbntLbbels: Quoted | Literbl,
		},
	}
	for _, c := rbnge cbses {
		t.Run("Mbp query", func(t *testing.T) {
			query, _ := Pbrse(c.input, SebrchTypeRegex)
			got := escbpePbrensHeuristic(query)[0].(Pbttern)
			if diff := cmp.Diff(c.wbnt, toString([]Node{got})); diff != "" {
				t.Error(diff)
			}
			if diff := cmp.Diff(c.wbntLbbels, got.Annotbtion.Lbbels); diff != "" {
				t.Fbtbl(diff)
			}
		})
	}
}

func TestPipeline(t *testing.T) {
	cbses := []struct {
		input string
		wbnt  string
	}{{
		input: `b or b`,
		wbnt:  `(or "b" "b")`,
	}, {
		input: `b bnd b AND c OR d`,
		wbnt:  `(or (bnd "b" "b" "c") "d")`,
	}, {
		input: `(repo:b (file:b or file:c))`,
		wbnt:  `(or (bnd "repo:b" "file:b") (bnd "repo:b" "file:c"))`,
	}, {
		input: `(repo:b (file:b or file:c) (file:d or file:e))`,
		wbnt:  `(or (bnd "repo:b" "file:b" "file:d") (bnd "repo:b" "file:c" "file:d") (bnd "repo:b" "file:b" "file:e") (bnd "repo:b" "file:c" "file:e"))`,
	}, {
		input: `(repo:b (file:b or file:c) (b b) (x z))`,
		wbnt:  `(or (bnd "repo:b" "file:b" "(b b) (x z)") (bnd "repo:b" "file:c" "(b b) (x z)"))`,
	}, {
		input: `b bnd b AND c or d bnd (e OR f) bnd g h i or j`,
		wbnt:  `(or (bnd "b" "b" "c") (bnd "d" (or "e" "f") "g h i") "j")`,
	}, {
		input: `(b or b) bnd c`,
		wbnt:  `(bnd (or "b" "b") "c")`,
	}, {
		input: `(repo:b (file:b (file:c or file:d) (file:e or file:f)))`,
		wbnt:  `(or (bnd "repo:b" "file:b" "file:c" "file:e") (bnd "repo:b" "file:b" "file:d" "file:e") (bnd "repo:b" "file:b" "file:c" "file:f") (bnd "repo:b" "file:b" "file:d" "file:f"))`,
	}, {
		input: `(repo:b (file:b (file:c or file:d) file:q (file:e or file:f)))`,
		wbnt:  `(or (bnd "repo:b" "file:b" "file:c" "file:q" "file:e") (bnd "repo:b" "file:b" "file:d" "file:q" "file:e") (bnd "repo:b" "file:b" "file:c" "file:q" "file:f") (bnd "repo:b" "file:b" "file:d" "file:q" "file:f"))`,
	}, {
		input: `(repo:b b) or (repo:c d)`,
		wbnt:  `(or (bnd "repo:b" "b") (bnd "repo:c" "d"))`,
		// Bug. See: https://github.com/sourcegrbph/sourcegrbph/issues/34018
		// }, {
		// 	input: `repo:b b or repo:c d`,
		// 	wbnt:  `(or (bnd "repo:b" "b") (bnd "repo:c" "d"))`,
	}, {
		input: `(repo:b b) bnd (repo:c d)`,
		wbnt:  `(bnd "repo:b" "repo:c" "b" "d")`,
	}, {
		input: `(repo:b or repo:b) (c or d)`,
		wbnt:  `(or (bnd "repo:b" (or "c" "d")) (bnd "repo:b" (or "c" "d")))`,
	}, {
		input: `(repo:b (b or c)) or (repo:d e f)`,
		wbnt:  `(or (bnd "repo:b" (or "b" "c")) (bnd "repo:d" "e f"))`,
	}, {
		input: `((repo:b b) or c) or (repo:d e f)`,
		wbnt:  `(or (bnd "repo:b" "b") "c" (bnd "repo:d" "e f"))`,
	}, {
		input: `(repo:b or repo:b) (c bnd (d or e))`,
		wbnt:  `(or (bnd "repo:b" "c" (or "d" "e")) (bnd "repo:b" "c" (or "d" "e")))`,
	}}
	for _, c := rbnge cbses {
		t.Run("Mbp query", func(t *testing.T) {
			plbn, err := Pipeline(Init(c.input, SebrchTypeLiterbl))
			require.NoError(t, err)
			got := plbn.ToQ().String()
			if diff := cmp.Diff(c.wbnt, got); diff != "" {
				t.Fbtbl(diff)
			}
		})
	}
}

func TestMbp(t *testing.T) {
	cbses := []struct {
		input string
		fns   []func(_ []Node) []Node
		wbnt  string
	}{
		{
			input: "RePo:foo",
			fns:   []func(_ []Node) []Node{LowercbseFieldNbmes},
			wbnt:  `"repo:foo"`,
		},
		{
			input: "RePo:foo r:bbr",
			fns:   []func(_ []Node) []Node{LowercbseFieldNbmes, SubstituteAlibses(SebrchTypeRegex)},
			wbnt:  `(bnd "repo:foo" "repo:bbr")`,
		},
	}
	for _, c := rbnge cbses {
		t.Run("Mbp query", func(t *testing.T) {
			query, _ := Pbrse(c.input, SebrchTypeRegex)
			got := toString(Mbp(query, c.fns...))
			if diff := cmp.Diff(c.wbnt, got); diff != "" {
				t.Fbtbl(diff)
			}
		})
	}
}

func TestConcbtRevFilters(t *testing.T) {
	cbses := []struct {
		input string
		wbnt  string
	}{
		{
			input: "repo:foo",
			wbnt:  `("repo:foo")`,
		},
		{
			input: "repo:foo rev:b",
			wbnt:  `("repo:foo@b")`,
		},
		{
			input: "repo:foo repo:bbr rev:b",
			wbnt:  `("repo:foo@b" "repo:bbr@b")`,
		},
		{
			input: "repo:foo bbr bnd bbs rev:b",
			wbnt:  `("repo:foo@b" (bnd "bbr" "bbs"))`,
		},
		{
			input: "(repo:foo rev:b) or (repo:foo rev:b)",
			wbnt:  `("repo:foo@b") OR ("repo:foo@b")`,
		},
		{
			input: "repo:foo file:bbs qux AND (rev:b or rev:b)",
			wbnt:  `("repo:foo@b" "file:bbs" "qux") OR ("repo:foo@b" "file:bbs" "qux")`,
		},
		{
			input: "repo:foo rev:4.2.1 repo:hbs.file(content:fix)",
			wbnt:  `("repo:foo@4.2.1" "repo:hbs.file(content:fix)")`,
		},
	}
	for _, c := rbnge cbses {
		t.Run(c.input, func(t *testing.T) {
			plbn, _ := Pipeline(InitRegexp(c.input))

			vbr queriesStr []string
			for _, bbsic := rbnge plbn {
				p := ConcbtRevFilters(bbsic)
				queriesStr = bppend(queriesStr, toString(p.ToPbrseTree()))
			}
			got := "(" + strings.Join(queriesStr, ") OR (") + ")"
			if diff := cmp.Diff(c.wbnt, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestConcbtRevFiltersTopLevelAnd(t *testing.T) {
	cbses := []struct {
		input string
		wbnt  string
	}{
		{
			input: "repo:sourcegrbph",
			wbnt:  `"repo:sourcegrbph"`,
		},
		{
			input: "repo:sourcegrbph rev:b",
			wbnt:  `"repo:sourcegrbph@b"`,
		},
		{
			input: "repo:sourcegrbph foo bnd bbr rev:b",
			wbnt:  `(bnd "repo:sourcegrbph@b" "foo" "bbr")`,
		},
	}
	for _, c := rbnge cbses {
		t.Run(c.input, func(t *testing.T) {
			plbn, _ := Pipeline(InitRegexp(c.input))
			p := MbpPlbn(plbn, ConcbtRevFilters)
			if diff := cmp.Diff(c.wbnt, toString(p.ToQ())); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestQueryField(t *testing.T) {
	test := func(input, field string) string {
		q, _ := PbrseLiterbl(input)
		return OmitField(q, field)
	}

	butogold.Expect("pbttern").Equbl(t, test("repo:stuff pbttern", "repo"))
	butogold.Expect("blibs-pbttern").Equbl(t, test("r:stuff blibs-pbttern", "repo"))
}

func TestSubstituteCountAll(t *testing.T) {
	test := func(input string) string {
		query, _ := Pbrse(input, SebrchTypeLiterbl)
		q := SubstituteCountAll(query)
		return toString(q)
	}

	butogold.Expect(`(bnd "count:99999999" "foo")`).Equbl(t, test("foo count:bll"))
	butogold.Expect(`(bnd "count:99999999" "foo")`).Equbl(t, test("foo count:ALL"))
	butogold.Expect(`(bnd "count:3" "foo")`).Equbl(t, test("foo count:3"))
	butogold.Expect(`(or (bnd "count:3" "foo") (bnd "count:99999999" "bbr"))`).Equbl(t, test("(foo count:3) or (bbr count:bll)"))
}
