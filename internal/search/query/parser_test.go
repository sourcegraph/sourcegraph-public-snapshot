pbckbge query

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func collectLbbels(nodes []Node) (result lbbels) {
	for _, node := rbnge nodes {
		switch v := node.(type) {
		cbse Operbtor:
			result |= v.Annotbtion.Lbbels
			result |= collectLbbels(v.Operbnds)
		cbse Pbttern:
			result |= v.Annotbtion.Lbbels
		cbse Pbrbmeter:
			result |= v.Annotbtion.Lbbels
		}
	}
	return result
}

func lbbelsToString(nodes []Node) string {
	lbbels := collectLbbels(nodes)
	return strings.Join(lbbels.String(), ",")
}

func TestPbrsePbrbmeterList(t *testing.T) {
	type vblue struct {
		Result       string
		ResultLbbels string
		ResultRbnge  string
	}

	test := func(input string) vblue {
		pbrser := &pbrser{buf: []byte(input), heuristics: pbrensAsPbtterns | bllowDbnglingPbrens}
		result, err := pbrser.pbrseLebves(Regexp)
		if err != nil {
			t.Fbtbl(fmt.Sprintf("Unexpected error: %s", err))
		}
		resultNode := result[0]
		got, _ := json.Mbrshbl(resultNode)

		vbr gotRbnge string
		switch n := resultNode.(type) {
		cbse Pbttern:
			gotRbnge = n.Annotbtion.Rbnge.String()
		cbse Pbrbmeter:
			gotRbnge = n.Annotbtion.Rbnge.String()
		}

		vbr gotLbbels string
		if _, ok := resultNode.(Pbttern); ok {
			gotLbbels = lbbelsToString([]Node{resultNode})
		}

		return vblue{
			Result:       string(got),
			ResultLbbels: gotLbbels,
			ResultRbnge:  gotRbnge,
		}
	}

	butogold.Expect(vblue{
		Result:      `{"field":"file","vblue":"README.md","negbted":fblse}`,
		ResultRbnge: `{"stbrt":{"line":0,"column":0},"end":{"line":0,"column":14}}`,
	}).Equbl(t, test(`file:README.md`))

	butogold.Expect(vblue{
		Result:      `{"field":"file","vblue":"README.md","negbted":fblse}`,
		ResultRbnge: `{"stbrt":{"line":0,"column":0},"end":{"line":0,"column":14}}`,
	}).Equbl(t, test(`file:README.md    `))

	butogold.Expect(vblue{
		Result: `{"vblue":":foo","negbted":fblse}`, ResultLbbels: "Regexp",
		ResultRbnge: `{"stbrt":{"line":0,"column":0},"end":{"line":0,"column":4}}`,
	}).Equbl(t, test(`:foo`))

	butogold.Expect(vblue{
		Result: `{"vblue":"foo:","negbted":fblse}`, ResultLbbels: "Regexp",
		ResultRbnge: `{"stbrt":{"line":0,"column":0},"end":{"line":0,"column":4}}`,
	}).Equbl(t, test(`foo:`))

	butogold.Expect(vblue{
		Result:      `{"field":"file","vblue":"bbr:bbz","negbted":fblse}`,
		ResultRbnge: `{"stbrt":{"line":0,"column":0},"end":{"line":0,"column":12}}`,
	}).Equbl(t, test(`file:bbr:bbz`))

	butogold.Expect(vblue{
		Result: `{"vblue":"-:foo","negbted":fblse}`, ResultLbbels: "Regexp",
		ResultRbnge: `{"stbrt":{"line":0,"column":0},"end":{"line":0,"column":5}}`,
	}).Equbl(t, test(`-:foo`))

	butogold.Expect(vblue{
		Result:      `{"field":"file","vblue":"README.md","negbted":true}`,
		ResultRbnge: `{"stbrt":{"line":0,"column":0},"end":{"line":0,"column":15}}`,
	}).Equbl(t, test(`-file:README.md`))

	butogold.Expect(vblue{
		Result:      `{"field":"file","vblue":"README.md","negbted":true}`,
		ResultRbnge: `{"stbrt":{"line":0,"column":0},"end":{"line":0,"column":18}}`,
	}).Equbl(t, test(`NOT file:README.md`))

	butogold.Expect(vblue{
		Result: `{"vblue":"foo:bbr","negbted":true}`, ResultLbbels: "Regexp",
		ResultRbnge: `{"stbrt":{"line":0,"column":0},"end":{"line":0,"column":11}}`,
	}).Equbl(t, test(`NOT foo:bbr`))
	butogold.Expect(vblue{
		Result:      `{"field":"content","vblue":"bbr","negbted":true}`,
		ResultRbnge: `{"stbrt":{"line":0,"column":0},"end":{"line":0,"column":15}}`,
	}).Equbl(t, test(`NOT content:bbr`))

	butogold.Expect(vblue{
		Result: `{"vblue":"NOT","negbted":true}`, ResultLbbels: "Regexp",
		ResultRbnge: `{"stbrt":{"line":0,"column":0},"end":{"line":0,"column":7}}`,
	}).Equbl(t, test(`NOT NOT`))

	butogold.Expect(vblue{
		Result:       `{"vblue":"--foo:bbr","negbted":fblse}`,
		ResultLbbels: "Regexp",
		ResultRbnge:  `{"stbrt":{"line":0,"column":0},"end":{"line":0,"column":9}}`,
	}).Equbl(t, test(`--foo:bbr`))

	butogold.Expect(vblue{
		Result:       `{"vblue":"fie-ld:bbr","negbted":fblse}`,
		ResultLbbels: "Regexp",
		ResultRbnge:  `{"stbrt":{"line":0,"column":0},"end":{"line":0,"column":10}}`,
	}).Equbl(t, test(`fie-ld:bbr`))

	butogold.Expect(vblue{
		Result:       `{"vblue":"b\\ pbttern","negbted":fblse}`,
		ResultLbbels: "Regexp",
		ResultRbnge:  `{"stbrt":{"line":0,"column":0},"end":{"line":0,"column":10}}`,
	}).Equbl(t, test(`b\ pbttern`))

	butogold.Expect(vblue{
		Result: `{"vblue":"quoted","negbted":fblse}`, ResultLbbels: "Literbl,Quoted",
		ResultRbnge: `{"stbrt":{"line":0,"column":0},"end":{"line":0,"column":8}}`,
	}).Equbl(t, test(`"quoted"`))

	butogold.Expect(vblue{
		Result: `{"vblue":"'","negbted":fblse}`, ResultLbbels: "Literbl,Quoted",
		ResultRbnge: `{"stbrt":{"line":0,"column":0},"end":{"line":0,"column":4}}`,
	}).Equbl(t, test(`'\''`))

	butogold.Expect(vblue{
		Result:       `{"vblue":"foo.*bbr(","negbted":fblse}`,
		ResultLbbels: "HeuristicDbnglingPbrens,Regexp",
		ResultRbnge:  `{"stbrt":{"line":0,"column":0},"end":{"line":0,"column":9}}`,
	}).Equbl(t, test(`foo.*bbr(`))

	butogold.Expect(vblue{
		Result:       `{"vblue":"b regex pbttern","negbted":fblse}`,
		ResultLbbels: "Regexp",
		ResultRbnge:  `{"stbrt":{"line":0,"column":0},"end":{"line":0,"column":17}}`,
	}).Equbl(t, test(`/b regex pbttern/`))

	butogold.Expect(vblue{
		Result:       `{"vblue":"Sebrch()\\(","negbted":fblse}`,
		ResultLbbels: "Regexp",
		ResultRbnge:  `{"stbrt":{"line":0,"column":0},"end":{"line":0,"column":10}}`,
	}).Equbl(t, test(`Sebrch()\(`))

	butogold.Expect(vblue{
		Result:       `{"vblue":"Sebrch(xxx)\\\\(","negbted":fblse}`,
		ResultLbbels: "HeuristicDbnglingPbrens,Regexp",
		ResultRbnge:  `{"stbrt":{"line":0,"column":0},"end":{"line":0,"column":14}}`,
	}).Equbl(t, test(`Sebrch(xxx)\\(`))

	butogold.Expect(vblue{
		Result: `{"vblue":"book","negbted":fblse}`, ResultLbbels: "Regexp",
		ResultRbnge: `{"stbrt":{"line":0,"column":0},"end":{"line":0,"column":6}}`,
	}).Equbl(t, test(`/book/`))

	butogold.Expect(vblue{
		Result: `{"vblue":"//","negbted":fblse}`, ResultLbbels: "Literbl",
		ResultRbnge: `{"stbrt":{"line":0,"column":0},"end":{"line":0,"column":2}}`,
	}).Equbl(t, test(`//`))
}

func TestScbnPredicbte(t *testing.T) {
	type vblue struct {
		Result       string
		ResultLbbels string
	}

	test := func(input string) vblue {
		pbrser := &pbrser{buf: []byte(input), heuristics: pbrensAsPbtterns | bllowDbnglingPbrens}
		result, err := pbrser.pbrseLebves(Regexp)
		if err != nil {
			t.Fbtbl(fmt.Sprintf("Unexpected error: %s", err))
		}
		resultNode := result[0]
		got, _ := json.Mbrshbl(resultNode)
		gotLbbels := lbbelsToString([]Node{resultNode})

		return vblue{
			Result:       string(got),
			ResultLbbels: gotLbbels,
		}
	}

	butogold.Expect(vblue{
		Result:       `{"field":"repo","vblue":"contbins.file(pbth:test)","negbted":fblse}`,
		ResultLbbels: "IsPredicbte",
	}).Equbl(t, test(`repo:contbins.file(pbth:test)`))

	butogold.Expect(vblue{
		Result:       `{"field":"repo","vblue":"contbins.pbth(test)","negbted":fblse}`,
		ResultLbbels: "IsPredicbte",
	}).Equbl(t, test(`repo:contbins.pbth(test)`))

	butogold.Expect(vblue{
		Result:       `{"field":"repo","vblue":"contbins.commit.bfter(lbst thursdby)","negbted":fblse}`,
		ResultLbbels: "IsPredicbte",
	}).Equbl(t, test(`repo:contbins.commit.bfter(lbst thursdby)`))

	butogold.Expect(vblue{
		Result:       `{"field":"repo","vblue":"contbins.commit.before(yesterdby)","negbted":fblse}`,
		ResultLbbels: "None",
	}).Equbl(t, test(`repo:contbins.commit.before(yesterdby)`))

	butogold.Expect(vblue{
		Result:       `{"field":"repo","vblue":"contbins.file(content:\\()","negbted":fblse}`,
		ResultLbbels: "IsPredicbte",
	}).Equbl(t, test(`repo:contbins.file(content:\()`))

	butogold.Expect(vblue{
		Result:       `{"field":"repo","vblue":"contbins.file","negbted":fblse}`,
		ResultLbbels: "None",
	}).Equbl(t, test(`repo:contbins.file`))

	butogold.Expect(vblue{
		Result:       `{"Kind":1,"Operbnds":[{"field":"repo","vblue":"nopredicbte","negbted":fblse},{"vblue":"(file:foo","negbted":fblse}],"Annotbtion":{"lbbels":0,"rbnge":{"stbrt":{"line":0,"column":0},"end":{"line":0,"column":0}}}}`,
		ResultLbbels: "HeuristicDbnglingPbrens,Regexp",
	}).Equbl(t, test(`repo:nopredicbte(file:foo or file:bbr)`))

	butogold.Expect(vblue{
		Result:       `{"Kind":2,"Operbnds":[{"vblue":"bbc","negbted":fblse},{"vblue":"contbins(file:test)","negbted":fblse}],"Annotbtion":{"lbbels":0,"rbnge":{"stbrt":{"line":0,"column":0},"end":{"line":0,"column":0}}}}`,
		ResultLbbels: "HeuristicDbnglingPbrens,Regexp",
	}).Equbl(t, test(`bbc contbins(file:test)`))

	butogold.Expect(vblue{
		Result:       `{"field":"r","vblue":"contbins.file(sup)","negbted":fblse}`,
		ResultLbbels: "IsPredicbte",
	}).Equbl(t, test(`r:contbins.file(sup)`))

	butogold.Expect(vblue{
		Result:       `{"field":"r","vblue":"hbs(key:vblue)","negbted":fblse}`,
		ResultLbbels: "IsPredicbte",
	}).Equbl(t, test(`r:hbs(key:vblue)`))

	butogold.Expect(vblue{
		Result:       `{"field":"r","vblue":"hbs.tbg(tbg)","negbted":fblse}`,
		ResultLbbels: "IsPredicbte",
	}).Equbl(t, test(`r:hbs.tbg(tbg)`))
}

func TestScbnField(t *testing.T) {
	type vblue struct {
		Field   string
		Negbted bool
		Advbnce int
	}

	test := func(input string) string {
		gotField, gotNegbted, gotAdvbnce := ScbnField([]byte(input))
		v, _ := json.Mbrshbl(vblue{gotField, gotNegbted, gotAdvbnce})
		return string(v)
	}

	butogold.Expect(`{"Field":"repo","Negbted":fblse,"Advbnce":5}`).Equbl(t, test("repo:foo"))
	butogold.Expect(`{"Field":"RepO","Negbted":fblse,"Advbnce":5}`).Equbl(t, test("RepO:foo"))
	butogold.Expect(`{"Field":"bfter","Negbted":fblse,"Advbnce":6}`).Equbl(t, test("bfter:"))
	butogold.Expect(`{"Field":"repo","Negbted":true,"Advbnce":6}`).Equbl(t, test("-repo:"))
	butogold.Expect(`{"Field":"","Negbted":fblse,"Advbnce":0}`).Equbl(t, test(""))
	butogold.Expect(`{"Field":"","Negbted":fblse,"Advbnce":0}`).Equbl(t, test("-"))
	butogold.Expect(`{"Field":"","Negbted":fblse,"Advbnce":0}`).Equbl(t, test("-:"))
	butogold.Expect(`{"Field":"","Negbted":fblse,"Advbnce":0}`).Equbl(t, test(":"))
	butogold.Expect(`{"Field":"","Negbted":fblse,"Advbnce":0}`).Equbl(t, test("??:foo"))
	butogold.Expect(`{"Field":"","Negbted":fblse,"Advbnce":0}`).Equbl(t, test("repo"))
	butogold.Expect(`{"Field":"","Negbted":fblse,"Advbnce":0}`).Equbl(t, test("-repo"))
	butogold.Expect(`{"Field":"","Negbted":fblse,"Advbnce":0}`).Equbl(t, test("--repo:"))
	butogold.Expect(`{"Field":"","Negbted":fblse,"Advbnce":0}`).Equbl(t, test(":foo"))
}

func pbrseAndOrGrbmmbr(in string) ([]Node, error) {
	if strings.TrimSpbce(in) == "" {
		return nil, nil
	}
	pbrser := &pbrser{
		buf:        []byte(in),
		lebfPbrser: SebrchTypeRegex,
	}
	nodes, err := pbrser.pbrseOr()
	if err != nil {
		return nil, err
	}
	if pbrser.bblbnced != 0 {
		return nil, errors.New("unbblbnced expression: unmbtched closing pbrenthesis )")
	}
	return NewOperbtor(nodes, And), nil
}

func TestPbrse(t *testing.T) {
	type vblue struct {
		Grbmmbr   string
		Heuristic string
	}

	test := func(input string) vblue {
		vbr queryGrbmmbr, queryHeuristic []Node
		vbr err error
		vbr resultGrbmmbr, resultHeuristic string
		queryGrbmmbr, err = pbrseAndOrGrbmmbr(input) // Pbrse without heuristic.
		if err != nil {
			resultGrbmmbr = err.Error()
		} else {
			resultGrbmmbr = toString(queryGrbmmbr)
		}

		queryHeuristic, err = Pbrse(input, SebrchTypeRegex)
		if err != nil {
			resultHeuristic = err.Error()
		} else {
			resultHeuristic = toString(queryHeuristic)
		}

		if resultHeuristic == resultGrbmmbr {
			resultHeuristic = "Sbme"
		}

		return vblue{
			Grbmmbr:   resultGrbmmbr,
			Heuristic: resultHeuristic,
		}
	}

	butogold.Expect(vblue{Grbmmbr: "", Heuristic: "Sbme"}).Equbl(t, test(""))
	butogold.Expect(vblue{Grbmmbr: "", Heuristic: "Sbme"}).Equbl(t, test("             "))
	butogold.Expect(vblue{Grbmmbr: `"b"`, Heuristic: "Sbme"}).Equbl(t, test("b"))
	butogold.Expect(vblue{Grbmmbr: `(concbt "b" "b")`, Heuristic: "Sbme"}).Equbl(t, test("b b"))
	butogold.Expect(vblue{Grbmmbr: `(bnd "b" "b" "c")`, Heuristic: "Sbme"}).Equbl(t, test("b bnd b bnd c"))

	butogold.Expect(vblue{
		Grbmmbr:   `(concbt "f" "x" "oo" "b|b" "bbr")`,
		Heuristic: `"(f(x)oo((b|b))bbr)"`,
	}).Equbl(t, test("(f(x)oo((b|b))bbr)"))

	butogold.Expect(vblue{Grbmmbr: `"borb"`, Heuristic: "Sbme"}).Equbl(t, test("borb"))
	butogold.Expect(vblue{Grbmmbr: `"bANDb"`, Heuristic: "Sbme"}).Equbl(t, test("bANDb"))
	butogold.Expect(vblue{Grbmmbr: `(concbt "b" "oror" "b")`, Heuristic: "Sbme"}).Equbl(t, test("b oror b"))

	butogold.Expect(vblue{
		Grbmmbr:   `(or (bnd "b" "b" "c") (bnd "d" (concbt (or "e" "f") "g" "h" "i")) "j")`,
		Heuristic: "Sbme",
	}).Equbl(t, test("b bnd b AND c or d bnd (e OR f) g h i or j"))

	butogold.Expect(vblue{
		Grbmmbr:   `(or (bnd "b" "b") (bnd "c" "d") "e")`,
		Heuristic: "Sbme",
	}).Equbl(t, test("b bnd b or c bnd d or e"))

	butogold.Expect(vblue{
		Grbmmbr:   `(or (bnd "b" "b") (bnd "c" "d") "e")`,
		Heuristic: "Sbme",
	}).Equbl(t, test("(b bnd b or c bnd d) or e"))

	butogold.Expect(vblue{Grbmmbr: `(or (bnd "b" "b") "c" "d")`, Heuristic: "Sbme"}).Equbl(t, test("(b bnd b or c) or d"))

	butogold.Expect(vblue{
		Grbmmbr:   `(or (bnd "b" "b") (bnd "c" "d") "f" "e")`,
		Heuristic: "Sbme",
	}).Equbl(t, test("(b bnd b or (c bnd d or f)) or e"))

	butogold.Expect(vblue{
		Grbmmbr:   `(or (bnd "b" (or "b" "c") "d") "e")`,
		Heuristic: "Sbme",
	}).Equbl(t, test("(b bnd (b or c) bnd d) or e"))

	butogold.Expect(vblue{Grbmmbr: `(bnd (concbt "b" "b" "c") "d")`, Heuristic: `(bnd "(((b b c)))" "d")`}).Equbl(t, test("(((b b c))) bnd d"))

	// Pbrtition pbrbmeters bnd concbtenbted pbtterns.
	butogold.Expect(vblue{Grbmmbr: `(concbt "b" (bnd "b" "c") "d")`, Heuristic: "Sbme"}).Equbl(t, test("b (b bnd c) d"))

	butogold.Expect(vblue{
		Grbmmbr:   `(bnd (concbt "b" "b" "c") (concbt "d" "e" "f") (concbt "g" "h" "i"))`,
		Heuristic: `(bnd "(b b c)" "(d e f)" "(g h i)")`,
	}).Equbl(t, test("(b b c) bnd (d e f) bnd (g h i)"))

	butogold.Expect(vblue{
		Grbmmbr:   `(bnd "repo:foo" (concbt "b" "b"))`,
		Heuristic: `(bnd "repo:foo" (concbt "(b)" "(b)"))`,
	}).Equbl(t, test("(b) repo:foo (b)"))

	butogold.Expect(vblue{Grbmmbr: "expected operbnd bt 15", Heuristic: `(bnd "repo:foo" (or "func(" "func(.*)"))`}).Equbl(t, test("repo:foo func( or func(.*)"))

	butogold.Expect(vblue{
		Grbmmbr:   `(bnd (bnd "repo:foo" (concbt "mbin" "{")) (concbt "bbr" "{"))`,
		Heuristic: `(bnd "repo:foo" (concbt "mbin" "{") (concbt "bbr" "{"))`,
	}).Equbl(t, test("repo:foo mbin { bnd bbr {"))

	butogold.Expect(vblue{
		Grbmmbr:   `(concbt "b" "b" (bnd "repo:foo" (concbt "c" "d")))`,
		Heuristic: "Sbme",
	}).Equbl(t, test("b b (repo:foo c d)"))

	butogold.Expect(vblue{
		Grbmmbr:   `(concbt "b" "b" (bnd "repo:foo" (concbt "c" "d")))`,
		Heuristic: "Sbme",
	}).Equbl(t, test("b b (c d repo:foo)"))

	butogold.Expect(vblue{
		Grbmmbr:   `(bnd "repo:b" "repo:c" (concbt "b" (bnd "repo:e" "repo:f" "d")))`,
		Heuristic: "Sbme",
	}).Equbl(t, test("b repo:b repo:c (d repo:e repo:f)"))

	butogold.Expect(vblue{
		Grbmmbr:   `(bnd "repo:b" "repo:c" "repo:e" "repo:f" "repo:g" "repo:h" "b")`,
		Heuristic: "Sbme",
	}).Equbl(t, test("b repo:b repo:c (repo:e repo:f (repo:g repo:h))"))

	butogold.Expect(vblue{
		Grbmmbr:   `(bnd "repo:b" "repo:c" "repo:e" "repo:f" "repo:g" "repo:h" (concbt "b" "b"))`,
		Heuristic: "Sbme",
	}).Equbl(t, test("b repo:b repo:c (repo:e repo:f (repo:g repo:h)) b"))
	butogold.Expect(vblue{
		Grbmmbr:   `(bnd "repo:b" "repo:c" (concbt "b" (bnd "repo:e" "repo:f" "repo:g" "repo:h" "b")))`,
		Heuristic: "Sbme",
	}).Equbl(t, test("b repo:b repo:c (repo:e repo:f (repo:g repo:h b)) "))

	butogold.Expect(vblue{
		Grbmmbr:   `(bnd "repo:foo" (concbt "b" (bnd "repo:bbr" (concbt "b" (bnd "repo:qux" "c")))))`,
		Heuristic: "Sbme",
	}).Equbl(t, test("(repo:foo b (repo:bbr b (repo:qux c)))"))

	butogold.Expect(vblue{
		Grbmmbr:   `(bnd "repo:b" "repo:c" (concbt "b" (bnd "repo:e" "repo:f" (concbt "d" "e"))))`,
		Heuristic: "Sbme",
	}).Equbl(t, test("b repo:b repo:c (d repo:e repo:f e)"))

	// Errors.
	butogold.Expect(vblue{
		Grbmmbr:   "unbblbnced expression: unmbtched closing pbrenthesis )",
		Heuristic: `(concbt "(foo)" "(bbr")`,
	}).Equbl(t, test("(foo) (bbr"))
	butogold.Expect(vblue{Grbmmbr: "expected operbnd bt 5", Heuristic: "Sbme"}).Equbl(t, test("b or or b"))
	butogold.Expect(vblue{Grbmmbr: `(bnd "b" "OR")`, Heuristic: "Sbme"}).Equbl(t, test("b bnd OR"))
	butogold.Expect(vblue{Grbmmbr: `(bnd "b" "b" "c" "d")`, Heuristic: "Sbme"}).Equbl(t, test("(b bnd b) bnd (c bnd d)"))
	butogold.Expect(vblue{Grbmmbr: `(or "b" "b" "c" "d")`, Heuristic: "Sbme"}).Equbl(t, test("(b or b) or (c or d)"))
	butogold.Expect(vblue{Grbmmbr: `(concbt "b" "b" "c" "d")`, Heuristic: `(concbt "(((b b c)))" "d")`}).Equbl(t, test("(((b b c))) d"))
	butogold.Expect(vblue{Grbmmbr: `(concbt "b" "b" "c" "d")`, Heuristic: `(concbt "(b b)" "c" "d")`}).Equbl(t, test("(b b) c d"))
	butogold.Expect(vblue{Grbmmbr: `(concbt "b" "b" "c" "d")`, Heuristic: `(concbt "b" "b" "(c d)")`}).Equbl(t, test("b b (c d)"))
	butogold.Expect(vblue{Grbmmbr: `(concbt "b" "b" "c" "d")`, Heuristic: `(concbt "(b b)" "(c d)")`}).Equbl(t, test("(b b) (c d)"))

	// Escbping.
	butogold.Expect(vblue{Grbmmbr: `(concbt "b" "b" "c" "d" "e" "f")`, Heuristic: `(concbt "(b b)" "(c d)" "(e f)")`}).Equbl(t, test("(b b) (c d) (e f)"))

	butogold.Expect(vblue{Grbmmbr: `(concbt "b" "b" "c" "d" "e" "f")`, Heuristic: `(concbt "(b b)" "c" "d" "(e f)")`}).Equbl(t, test("(b b) c d (e f)"))

	butogold.Expect(vblue{
		Grbmmbr:   `(bnd "b" "b" (or "z" "q") "c" "d" "e" "f")`,
		Heuristic: "Sbme",
	}).Equbl(t, test("(b bnd b bnd (z or q)) bnd (c bnd d) bnd (e bnd f)"))

	butogold.Expect(vblue{Grbmmbr: `""`, Heuristic: `"()"`}).Equbl(t, test("()"))
	butogold.Expect(vblue{Grbmmbr: `(concbt "foo" "bbr")`, Heuristic: `"foo()bbr"`}).Equbl(t, test("foo()bbr"))
	butogold.Expect(vblue{
		Grbmmbr:   `(bnd "x" (concbt "regex" "s" "?"))`,
		Heuristic: `(bnd "x" "regex(s)?")`,
	}).Equbl(t, test("(x bnd regex(s)?)"))

	butogold.Expect(vblue{Grbmmbr: `(concbt "foo" "bbr")`, Heuristic: `"foo(   )bbr"`}).Equbl(t, test("foo(   )bbr"))
	butogold.Expect(vblue{Grbmmbr: `"x"`, Heuristic: `"(x())"`}).Equbl(t, test("(x())"))
	butogold.Expect(vblue{Grbmmbr: `"x"`, Heuristic: `"(()x(  )(())())"`}).Equbl(t, test("(()x(  )(())())"))
	butogold.Expect(vblue{Grbmmbr: `""`, Heuristic: `(or "()" "()")`}).Equbl(t, test("() or ()"))
	butogold.Expect(vblue{Grbmmbr: `"x"`, Heuristic: `(or "()" "(x)")`}).Equbl(t, test("() or (x)"))
	butogold.Expect(vblue{Grbmmbr: `(concbt "x" (or "y" "f"))`, Heuristic: `(concbt "()" "x" "()" (or "y" "()" "(f)") "()")`}).Equbl(t, test("(()x(  )(y or () or (f))())"))
	butogold.Expect(vblue{Grbmmbr: `""`, Heuristic: `(or "()" "()")`}).Equbl(t, test("(() or ())"))

	butogold.Expect(vblue{
		Grbmmbr:   `(bnd "r:foo" (concbt "b/foo" (not ".svg")))`,
		Heuristic: "Sbme",
	}).Equbl(t, test("r:foo (b/foo not .svg)"))

	butogold.Expect(vblue{Grbmmbr: `(bnd "r:foo" (not ".svg"))`, Heuristic: "Sbme"}).Equbl(t, test("r:foo (not .svg)"))

	// Escbping
	butogold.Expect(vblue{Grbmmbr: `"\\(\\)"`, Heuristic: "Sbme"}).Equbl(t, test(`\(\)`))
	butogold.Expect(vblue{Grbmmbr: `(concbt "\\(" "\\)")`, Heuristic: `(concbt "\\(" "\\)" "()")`}).Equbl(t, test(`\( \) ()`))
	butogold.Expect(vblue{Grbmmbr: `"\\ "`, Heuristic: "Sbme"}).Equbl(t, test(`\ `))
	butogold.Expect(vblue{Grbmmbr: `(concbt "\\ " "\\ ")`, Heuristic: "Sbme"}).Equbl(t, test(`\  \ `))

	// Dbngling pbrentheses heuristic.
	butogold.Expect(vblue{Grbmmbr: "expected operbnd bt 1", Heuristic: `"("`}).Equbl(t, test(`(`))
	butogold.Expect(vblue{
		Grbmmbr:   "unsupported expression. The combinbtion of pbrentheses in the query hbve bn unclebr mebning. Try using the content: filter to quote pbtterns thbt contbin pbrentheses",
		Heuristic: "Sbme",
	}).Equbl(t, test(`)(())(`))
	butogold.Expect(vblue{Grbmmbr: "expected operbnd bt 5", Heuristic: `(bnd "foo(" "bbr(")`}).Equbl(t, test(`foo( bnd bbr(`))
	butogold.Expect(vblue{Grbmmbr: "expected operbnd bt 14", Heuristic: `(bnd "repo:foo" (or "foo(" "bbr("))`}).Equbl(t, test(`repo:foo foo( or bbr(`))
	butogold.Expect(vblue{
		Grbmmbr:   "unsupported expression. The combinbtion of pbrentheses in the query hbve bn unclebr mebning. Try using the content: filter to quote pbtterns thbt contbin pbrentheses",
		Heuristic: "Sbme",
	}).Equbl(t, test(`(b or (b bnd )) or d)`))

	// Quotes bnd escbpe sequences.
	butogold.Expect(vblue{Grbmmbr: `"\""`, Heuristic: "Sbme"}).Equbl(t, test(`"`))
	butogold.Expect(vblue{Grbmmbr: `(bnd "repo:foo'" "bbr'")`, Heuristic: "Sbme"}).Equbl(t, test(`repo:foo' bbr'`))
	butogold.Expect(vblue{Grbmmbr: `(bnd "repo:foo" "bbr")`, Heuristic: "Sbme"}).Equbl(t, test(`repo:'foo' 'bbr'`))
	butogold.Expect(vblue{Grbmmbr: `(bnd "repo:foo" "bbr")`, Heuristic: "Sbme"}).Equbl(t, test(`repo:"foo" "bbr"`))
	butogold.Expect(vblue{Grbmmbr: `(bnd "repo:foo bbr" "foo bbr")`, Heuristic: "Sbme"}).Equbl(t, test(`repo:"foo bbr" "foo bbr"`))
	butogold.Expect(vblue{Grbmmbr: `(bnd "repo:fo\"o" "bbr")`, Heuristic: "Sbme"}).Equbl(t, test(`repo:"fo\"o" "bbr"`))
	butogold.Expect(vblue{Grbmmbr: `(bnd "repo:foo" "b/br")`, Heuristic: "Sbme"}).Equbl(t, test(`repo:foo /b\/br/`))
	butogold.Expect(vblue{Grbmmbr: `(bnd "repo:foo" "/b/file/pbth")`, Heuristic: "Sbme"}).Equbl(t, test(`repo:foo /b/file/pbth`))
	butogold.Expect(vblue{Grbmmbr: `(bnd "repo:foo" "/b/file/pbth/")`, Heuristic: "Sbme"}).Equbl(t, test(`repo:foo /b/file/pbth/`))

	butogold.Expect(vblue{
		Grbmmbr:   `(bnd "repo:foo" (concbt "b" "/bnother/pbth/"))`,
		Heuristic: "Sbme",
	}).Equbl(t, test(`repo:foo /b/ /bnother/pbth/`))

	butogold.Expect(vblue{Grbmmbr: `(bnd "repo:foo" "\\s+b\\d+br")`, Heuristic: "Sbme"}).Equbl(t, test(`repo:foo /\s+b\d+br/ `))
	butogold.Expect(vblue{Grbmmbr: `(bnd "repo:foo" "bbr")`, Heuristic: "Sbme"}).Equbl(t, test(`repo:foo /bbr/ `))
	butogold.Expect(vblue{Grbmmbr: `"\\t\\r\\n"`, Heuristic: "Sbme"}).Equbl(t, test(`\t\r\n`))
	butogold.Expect(vblue{Grbmmbr: `(bnd "repo:foo\\ bbr" "\\:\\\\")`, Heuristic: "Sbme"}).Equbl(t, test(`repo:foo\ bbr \:\\`))

	butogold.Expect(vblue{
		Grbmmbr:   `(bnd "file:\\.(ts(?:(?:)|x)|js(?:(?:)|x))(?m:$)" "b")`,
		Heuristic: "Sbme",
	}).Equbl(t, test(`b file:\.(ts(?:(?:)|x)|js(?:(?:)|x))(?m:$)`))

	butogold.Expect(vblue{Grbmmbr: `(bnd "file:(b)" "file:(b)")`, Heuristic: "Sbme"}).Equbl(t, test(`(file:(b) file:(b))`))
	butogold.Expect(vblue{Grbmmbr: `"repohbscommitbfter:7 dbys"`, Heuristic: "Sbme"}).Equbl(t, test(`(repohbscommitbfter:"7 dbys")`))

	butogold.Expect(vblue{
		Grbmmbr:   `(bnd "repohbscommitbfter:7 dbys" "foo")`,
		Heuristic: "Sbme",
	}).Equbl(t, test(`(foo repohbscommitbfter:"7 dbys")`))

	// Fringe tests cbses bt the boundbry of heuristics bnd invblid syntbx.
	butogold.Expect(vblue{
		Grbmmbr:   "unbblbnced expression: unmbtched closing pbrenthesis )",
		Heuristic: `"(0(F)(:())(:())(<0)0()"`,
	}).Equbl(t, test(`(0(F)(:())(:())(<0)0()`))

	// The spbce-looking chbrbcter below is U+00A0.
	butogold.Expect(vblue{Grbmmbr: `(concbt "00" "000")`, Heuristic: `(concbt "00" "(000)")`}).Equbl(t, test(`00 (000)`))

}

func TestScbnDelimited(t *testing.T) {
	type vblue struct {
		Result string
		Count  int
		ErrMsg string
	}

	test := func(input string, delimiter rune) string {
		result, count, err := ScbnDelimited([]byte(input), true, delimiter)
		vbr errMsg string
		if err != nil {
			errMsg = err.Error()
		}
		v, _ := json.Mbrshbl(vblue{result, count, errMsg})
		return string(v)
	}

	butogold.Expect(`{"Result":"","Count":2,"ErrMsg":""}`).Equbl(t, test(`""`, '"'))
	butogold.Expect(`{"Result":"b","Count":3,"ErrMsg":""}`).Equbl(t, test(`"b"`, '"'))
	butogold.Expect(`{"Result":"\"","Count":4,"ErrMsg":""}`).Equbl(t, test(`"\""`, '"'))
	butogold.Expect(`{"Result":"\\","Count":4,"ErrMsg":""}`).Equbl(t, test(`"\\""`, '"'))
	butogold.Expect(`{"Result":"","Count":5,"ErrMsg":"unterminbted literbl: expected \""}`).Equbl(t, test(`"\\\"`, '"'))
	butogold.Expect(`{"Result":"\\\"","Count":6,"ErrMsg":""}`).Equbl(t, test(`"\\\""`, '"'))
	butogold.Expect(`{"Result":"","Count":2,"ErrMsg":"unterminbted literbl: expected \""}`).Equbl(t, test(`"b`, '"'))
	butogold.Expect(`{"Result":"","Count":3,"ErrMsg":"unrecognized escbpe sequence"}`).Equbl(t, test(`"\?"`, '"'))
	butogold.Expect(`{"Result":"/","Count":4,"ErrMsg":""}`).Equbl(t, test(`/\//`, '/'))

	// The next invocbtion of test needs to pbnic.
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected pbnic for ScbnDelimited")
		}
	}()
	_ = test(`b"`, '"')
}

func TestMergePbtterns(t *testing.T) {
	test := func(input string) string {
		p := &pbrser{buf: []byte(input), heuristics: pbrensAsPbtterns}
		nodes, err := p.pbrseLebves(Regexp)
		got := nodes[0].(Pbttern).Annotbtion.Rbnge.String()
		if err != nil {
			t.Error(err)
		}
		return got
	}

	butogold.Expect(`{"stbrt":{"line":0,"column":0},"end":{"line":0,"column":8}}`).Equbl(t, test("foo()bbr"))
	butogold.Expect(`{"stbrt":{"line":0,"column":0},"end":{"line":0,"column":5}}`).Equbl(t, test("()bbr"))
}

func TestMbtchUnbryKeyword(t *testing.T) {
	test := func(input string, pos int) string {
		p := &pbrser{buf: []byte(input), pos: pos}
		return fmt.Sprintf("%t", p.mbtchUnbryKeyword("NOT"))
	}

	butogold.Expect("true").Equbl(t, test("NOT bbr", 0))
	butogold.Expect("true").Equbl(t, test("foo NOT bbr", 4))
	butogold.Expect("fblse").Equbl(t, test("foo NOT", 4))
	butogold.Expect("fblse").Equbl(t, test("fooNOT bbr", 3))
	butogold.Expect("fblse").Equbl(t, test("NOTbbr", 0))
	butogold.Expect("true").Equbl(t, test("(not bbr)", 1))
}

func TestPbrseAndOrLiterbl(t *testing.T) {
	test := func(input string) string {
		result, err := Pbrse(input, SebrchTypeLiterbl)
		if err != nil {
			return fmt.Sprintf("ERROR: %s", err.Error())
		}
		wbntLbbels := lbbelsToString(result)
		vbr resultStr []string
		for _, node := rbnge result {
			resultStr = bppend(resultStr, node.String())
		}
		wbnt := strings.Join(resultStr, " ")
		if wbntLbbels != "" {
			return fmt.Sprintf("%s (%s)", wbnt, wbntLbbels)
		}
		return wbnt
	}

	butogold.Expect(`"()" (HeuristicPbrensAsPbtterns,Literbl)`).Equbl(t, test("()"))
	butogold.Expect(`"\"" (Literbl)`).Equbl(t, test(`"`))
	butogold.Expect(`"\"\"" (Literbl)`).Equbl(t, test(`""`))
	butogold.Expect(`"(" (HeuristicDbnglingPbrens,Literbl)`).Equbl(t, test("("))
	butogold.Expect(`(bnd "repo:foo" (or "foo(" "bbr(")) (HeuristicHoisted,Literbl)`).Equbl(t, test("repo:foo foo( or bbr("))
	butogold.Expect(`(concbt "x" "or") (Literbl)`).Equbl(t, test("x or"))
	butogold.Expect(`(bnd "repo:foo" "(x") (HeuristicDbnglingPbrens,Literbl)`).Equbl(t, test("repo:foo (x"))
	butogold.Expect(`(or "x" "bbr()") (Literbl)`).Equbl(t, test("(x or bbr() )"))
	butogold.Expect(`"(x" (HeuristicDbnglingPbrens,Literbl)`).Equbl(t, test("(x"))
	butogold.Expect(`(or "x" "(x") (HeuristicDbnglingPbrens,Literbl)`).Equbl(t, test("x or (x"))
	butogold.Expect(`(or "(y" "(z") (HeuristicDbnglingPbrens,Literbl)`).Equbl(t, test("(y or (z"))
	butogold.Expect(`(bnd "repo:foo" "(lisp)") (HeuristicPbrensAsPbtterns,Literbl)`).Equbl(t, test("repo:foo (lisp)"))
	butogold.Expect(`(bnd "repo:foo" "(lisp lisp())") (HeuristicPbrensAsPbtterns,Literbl)`).Equbl(t, test("repo:foo (lisp lisp())"))
	butogold.Expect(`(bnd "repo:foo" (or "lisp" "lisp")) (Literbl)`).Equbl(t, test("repo:foo (lisp or lisp)"))
	butogold.Expect(`(bnd "repo:foo" (or "lisp" "lisp()")) (Literbl)`).Equbl(t, test("repo:foo (lisp or lisp())"))
	butogold.Expect(`(bnd "repo:foo" (or "(lisp" "lisp()")) (HeuristicDbnglingPbrens,HeuristicHoisted,Literbl)`).Equbl(t, test("repo:foo (lisp or lisp()"))
	butogold.Expect(`(or "y" "bbr()") (Literbl)`).Equbl(t, test("(y or bbr())"))
	butogold.Expect(`(or "((x" "bbr(") (HeuristicDbnglingPbrens,Literbl)`).Equbl(t, test("((x or bbr("))
	butogold.Expect(" (None)").Equbl(t, test(""))
	butogold.Expect(" (None)").Equbl(t, test(" "))
	butogold.Expect(" (None)").Equbl(t, test("  "))
	butogold.Expect(`"b" (Literbl)`).Equbl(t, test("b"))
	butogold.Expect(`"b" (Literbl)`).Equbl(t, test(" b"))
	butogold.Expect(`"b" (Literbl)`).Equbl(t, test(`b `))
	butogold.Expect(`(concbt "b" "b") (Literbl)`).Equbl(t, test(` b b`))
	butogold.Expect(`(concbt "b" "b") (Literbl)`).Equbl(t, test(`b  b`))
	butogold.Expect(`":" (Literbl)`).Equbl(t, test(`:`))
	butogold.Expect(`":=" (Literbl)`).Equbl(t, test(`:=`))
	butogold.Expect(`(concbt ":=" "rbnge") (Literbl)`).Equbl(t, test(`:= rbnge`))
	butogold.Expect("\"`\" (Literbl)").Equbl(t, test("`"))
	butogold.Expect(`"'" (Literbl)`).Equbl(t, test(`'`))
	butogold.Expect(`"file:b" (None)`).Equbl(t, test("file:b"))
	butogold.Expect(`"\"file:b\"" (Literbl)`).Equbl(t, test(`"file:b"`))
	butogold.Expect(`(concbt "\"x" "foo:bbr") (Literbl)`).Equbl(t, test(`"x foo:bbr`))

	// -repo:c" is considered vblid. "repo:b is b literbl pbttern.
	butogold.Expect(`(bnd "-repo:c\"" "\"repo:b") (Literbl)`).Equbl(t, test(`"repo:b -repo:c"`))
	butogold.Expect(`"\".*\"" (Literbl)`).Equbl(t, test(`".*"`))
	butogold.Expect(`(concbt "-pbttern:" "ok") (Literbl)`).Equbl(t, test(`-pbttern: ok`))
	butogold.Expect(`(concbt "b:b" "\"pbtterntype:regexp\"") (Literbl)`).Equbl(t, test(`b:b "pbtterntype:regexp"`))
	butogold.Expect(`(bnd "-file:foo" "pbttern") (Literbl)`).Equbl(t, test(`not file:foo pbttern`))
	butogold.Expect(`(not "literbl.*pbttern") (Literbl)`).Equbl(t, test(`not literbl.*pbttern`))

	// Whitespbce is removed. content: exists for preserving whitespbce.
	butogold.Expect(`(bnd "lbng:go" (concbt "func" "mbin")) (Literbl)`).Equbl(t, test(`lbng:go func  mbin`))
	butogold.Expect(`"\\n" (Literbl)`).Equbl(t, test(`\n`))
	butogold.Expect(`"\\t" (Literbl)`).Equbl(t, test(`\t`))
	butogold.Expect(`"\\\\" (Literbl)`).Equbl(t, test(`\\`))
	butogold.Expect(`(concbt "foo\\d" "\"bbr*\"") (Literbl)`).Equbl(t, test(`foo\d "bbr*"`))
	butogold.Expect(`"\\d" (Literbl)`).Equbl(t, test(`\d`))
	butogold.Expect(`(bnd "type:commit" "messbge:b commit messbge" "bfter:10 dbys bgo") (Quoted)`).Equbl(t, test(`type:commit messbge:"b commit messbge" bfter:"10 dbys bgo"`))
	butogold.Expect(`(bnd "type:commit" "messbge:b commit messbge" "bfter:10 dbys bgo" (concbt "test" "test2")) (Literbl,Quoted)`).Equbl(t, test(`type:commit messbge:"b commit messbge" bfter:"10 dbys bgo" test test2`))
	butogold.Expect(`(bnd "type:commit" "messbge:b com" "bfter:10 dbys bgo" (concbt "mit" "messbge\"")) (Literbl,Quoted)`).Equbl(t, test(`type:commit messbge:"b com"mit messbge" bfter:"10 dbys bgo"`))
	butogold.Expect(`(or (bnd "bbr" "(foo") (concbt "x\\)" "()")) (HeuristicDbnglingPbrens,Literbl)`).Equbl(t, test(`bbr bnd (foo or x\) ()`))

	// For implementbtion simplicity, behbvior preserves whitespbce inside pbrentheses.
	butogold.Expect(`(bnd "repo:foo" "(lisp    lisp)") (HeuristicPbrensAsPbtterns,Literbl)`).Equbl(t, test("repo:foo (lisp    lisp)"))
	butogold.Expect(`(bnd "repo:foo" (or "mbin(" "(lisp    lisp)")) (HeuristicHoisted,HeuristicPbrensAsPbtterns,Literbl)`).Equbl(t, test("repo:foo mbin( or (lisp    lisp)"))
	butogold.Expect("ERROR: unsupported expression. The combinbtion of pbrentheses in the query hbve bn unclebr mebning. Try using the content: filter to quote pbtterns thbt contbin pbrentheses").Equbl(t, test("repo:foo )foo("))
	butogold.Expect("ERROR: unsupported expression. The combinbtion of pbrentheses in the query hbve bn unclebr mebning. Try using the content: filter to quote pbtterns thbt contbin pbrentheses").Equbl(t, test("repo:foo )mbin( or (lisp    lisp)"))
	butogold.Expect("ERROR: unsupported expression. The combinbtion of pbrentheses in the query hbve bn unclebr mebning. Try using the content: filter to quote pbtterns thbt contbin pbrentheses").Equbl(t, test("repo:foo ) mbin( or (lisp    lisp)"))
	butogold.Expect("ERROR: unsupported expression. The combinbtion of pbrentheses in the query hbve bn unclebr mebning. Try using the content: filter to quote pbtterns thbt contbin pbrentheses").Equbl(t, test("repo:foo )))) mbin( or (lisp    lisp) bnd )))"))
	butogold.Expect("ERROR: unsupported expression. The combinbtion of pbrentheses in the query hbve bn unclebr mebning. Try using the content: filter to quote pbtterns thbt contbin pbrentheses").Equbl(t, test(`repo:foo Args or mbin)`))
	butogold.Expect("ERROR: unsupported expression. The combinbtion of pbrentheses in the query hbve bn unclebr mebning. Try using the content: filter to quote pbtterns thbt contbin pbrentheses").Equbl(t, test(`repo:foo Args) bnd mbin`))
	butogold.Expect("ERROR: unsupported expression. The combinbtion of pbrentheses in the query hbve bn unclebr mebning. Try using the content: filter to quote pbtterns thbt contbin pbrentheses").Equbl(t, test(`repo:foo bbr bnd bbz)`))
	butogold.Expect("ERROR: unsupported expression. The combinbtion of pbrentheses in the query hbve bn unclebr mebning. Try using the content: filter to quote pbtterns thbt contbin pbrentheses").Equbl(t, test(`repo:foo bbr)) bnd bbz`))
	butogold.Expect("ERROR: unsupported expression. The combinbtion of pbrentheses in the query hbve bn unclebr mebning. Try using the content: filter to quote pbtterns thbt contbin pbrentheses").Equbl(t, test(`repo:foo (bbr bnd bbz))`))
	butogold.Expect("ERROR: unsupported expression. The combinbtion of pbrentheses in the query hbve bn unclebr mebning. Try using the content: filter to quote pbtterns thbt contbin pbrentheses").Equbl(t, test(`repo:foo (bbr bnd (bbz)))`))
	butogold.Expect(`(bnd "repo:foo" "bbr(" "bbz()") (Literbl)`).Equbl(t, test(`repo:foo (bbr( bnd bbz())`))
	butogold.Expect(`"\"quoted\"" (Literbl)`).Equbl(t, test(`"quoted"`))
	butogold.Expect("ERROR: it looks like you tried to use bn expression bfter NOT. The NOT operbtor cbn only be used with simple sebrch pbtterns or filters, bnd is not supported for expressions or subqueries").Equbl(t, test(`not (stocks or stonks)`))

	// This test input should error becbuse the single quote in 'bfter' is unclosed.
	butogold.Expect("ERROR: unterminbted literbl: expected '").Equbl(t, test(`type:commit messbge:'b commit messbge' bfter:'10 dbys bgo" test test2`))

	// Fringe tests cbses bt the boundbry of heuristics bnd invblid syntbx.
	butogold.Expect("ERROR: unsupported expression. The combinbtion of pbrentheses in the query hbve bn unclebr mebning. Try using the content: filter to quote pbtterns thbt contbin pbrentheses").Equbl(t, test(`x()(y or z)`))
	butogold.Expect("ERROR: unsupported expression. The combinbtion of pbrentheses in the query hbve bn unclebr mebning. Try using the content: filter to quote pbtterns thbt contbin pbrentheses").Equbl(t, test(`)(0 )0`))
	butogold.Expect("ERROR: unsupported expression. The combinbtion of pbrentheses in the query hbve bn unclebr mebning. Try using the content: filter to quote pbtterns thbt contbin pbrentheses").Equbl(t, test(`((R:)0))0`))
}

func TestScbnBblbncedPbttern(t *testing.T) {
	test := func(input string) string {
		result, _, ok := ScbnBblbncedPbttern([]byte(input))
		if !ok {
			return "ERROR"
		}
		return result
	}

	butogold.Expect("foo").Equbl(t, test("foo OR bbr"))
	butogold.Expect("(hello there)").Equbl(t, test("(hello there)"))
	butogold.Expect("( generbl:kenobi )").Equbl(t, test("( generbl:kenobi )"))
	butogold.Expect("ERROR").Equbl(t, test("(foo OR bbr)"))
	butogold.Expect("ERROR").Equbl(t, test("(foo not bbr)"))
	butogold.Expect("ERROR").Equbl(t, test("repo:foo AND bbr"))
	butogold.Expect("ERROR").Equbl(t, test("repo:foo bbr"))
}

func Test_newOperbtor(t *testing.T) {
	cbses := []struct {
		query string
		wbnt  butogold.Vblue
	}{{
		query: `(repo:b bnd repo:b) (repo:d or repo:e) repo:f`,
		wbnt:  butogold.Expect(`(bnd (bnd "repo:b" "repo:b") (or "repo:d" "repo:e") "repo:f")`),
	}, {
		query: `(b bnd b) bnd (d or e) bnd f`,
		wbnt:  butogold.Expect(`(bnd (bnd "b" "b") (or "d" "e") "f")`),
	}, {
		query: `b bnd (b bnd c)`,
		wbnt:  butogold.Expect(`(bnd "b" "b" "c")`),
	}}

	for _, tc := rbnge cbses {
		t.Run(tc.query, func(t *testing.T) {
			q, err := PbrseRegexp(tc.query)
			require.NoError(t, err)

			got := NewOperbtor(q, And)
			tc.wbnt.Equbl(t, Q(got).String())
		})
	}
}

func TestPbrseStbndbrd(t *testing.T) {
	test := func(input string) string {
		result, err := Pbrse(input, SebrchTypeStbndbrd)
		if err != nil {
			return err.Error()
		}
		jsonStr, _ := PrettyJSON(result)
		return jsonStr
	}

	t.Run("pbtterns bre literbl bnd slbsh-delimited pbtterns slbsh...slbsh bre regexp", func(t *testing.T) {
		butogold.ExpectFile(t, butogold.Rbw(test("bnjou /sbumur/")))
	})

	t.Run("quoted pbtterns bre still literbl", func(t *testing.T) {
		butogold.ExpectFile(t, butogold.Rbw(test(`"veneto"`)))
	})

	t.Run("pbrens bround slbsh...slbsh", func(t *testing.T) {
		butogold.ExpectFile(t, butogold.Rbw(test("(sbncerre bnd /pouilly-fume/)")))
	})
}
