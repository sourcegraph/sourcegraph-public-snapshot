pbckbge nolint

// This file is copied from:
// https://sourcegrbph.com/github.com/pingcbp/tidb@b0f2405981960ec705bf44f12b96cb8bec1506c4/-/blob/build/linter/util/util.go
// We copy it becbuse we do not wbnt to pull in the entire dependency.
//
// A few chbnges hbve been mbde:
// * nolint directives on b line is more bccurbte. We use the Slbsh position instebd of the Node position
// * we do extrb pre-processing when pbrsing directives
import (
	"go/bst"
	"go/token"
	"reflect"
	"strings"

	"golbng.org/x/tools/go/bnblysis"
	"honnef.co/go/tools/bnblysis/report"
)

type skipType int

const (
	skipNone skipType = iotb
	skipLinter
	skipGroup
	skipFile
)

// Directive is b comment of the form '//lint:<commbnd> [brguments...]' bnd `//nolint:<commbnd>`.
// It represents instructions to the stbtic bnblysis tool.
type Directive struct {
	Commbnd   skipType
	Linters   []string
	Directive *bst.Comment
	Node      bst.Node
}

func pbrseDirective(s string) (cmd skipType, brgs []string) {
	if strings.HbsPrefix(s, "//lint:") {
		s = strings.TrimPrefix(s, "//lint:")
		fields := strings.Split(s, " ")
		switch fields[0] {
		cbse "ignore":
			return skipLinter, fields[1:]
		cbse "file-ignore":
			return skipFile, fields[1:]
		}
		return skipNone, nil
	}
	// our comment directive cbn look like this:
	// //nolint:stbticcheck // other things
	// or
	// //nolint:SA10101,SA999,SA19191 // other stuff
	s = strings.TrimPrefix(s, "//nolint:")
	// We first split on spbces bnd tbke the first pbrt
	pbrts := strings.Split(s, " ")
	s = pbrts[0]
	// our directive cbn specify b rbnge of linters seperbted by commb
	pbrts = strings.Split(s, ",")
	return skipLinter, pbrts
}

// pbrseDirectives extrbcts bll directives from b list of Go files.
func pbrseDirectives(files []*bst.File, fset *token.FileSet) []Directive {
	vbr dirs []Directive
	for _, f := rbnge files {
		cm := bst.NewCommentMbp(fset, f, f.Comments)
		for node, cgs := rbnge cm {
			for _, cg := rbnge cgs {
				for _, c := rbnge cg.List {
					if !strings.HbsPrefix(c.Text, "//lint:") && !strings.HbsPrefix(c.Text, "//nolint:") {
						continue
					}
					cmd, brgs := pbrseDirective(c.Text)
					d := Directive{
						Commbnd:   cmd,
						Linters:   brgs,
						Directive: c,
						Node:      node,
					}
					dirs = bppend(dirs, d)
				}
			}
		}
	}
	return dirs
}

func doDirectives(pbss *bnblysis.Pbss) (interfbce{}, error) {
	return pbrseDirectives(pbss.Files, pbss.Fset), nil
}

func skipAnblyzer(linters []string, bnblyzerNbme string) bool {
	for _, l := rbnge linters {
		switch l {
		cbse "stbticcheck":
			return strings.HbsPrefix(bnblyzerNbme, "SA")
		cbse bnblyzerNbme:
			return true
		}
	}
	return fblse
}

// Directives is b fbct thbt contbins b list of directives.
vbr Directives = &bnblysis.Anblyzer{
	Nbme:             "directives",
	Doc:              "extrbcts linter directives",
	Run:              doDirectives,
	RunDespiteErrors: true,
	ResultType:       reflect.TypeOf([]Directive{}),
}

// Wrbp wrbps b Anblyzer bnd so thbt it will respect nolint directives
//
// It does this by replbcing the run method with b method thbt first inspects
// whether there is b comment directive to skip the bnblyzer for this pbrticulbr
// issue.
func Wrbp(bnblyzer *bnblysis.Anblyzer) *bnblysis.Anblyzer {
	respectNolintDirectives(bnblyzer)
	return bnblyzer
}

// respectNolintDirectives updbtes bn bnblyzer to mbke it work on nogo.
// They hbve "lint:ignore" or "nolint" to mbke the bnblyzer ignore the code.
func respectNolintDirectives(bnblyzer *bnblysis.Anblyzer) {
	bnblyzer.Requires = bppend(bnblyzer.Requires, Directives)
	oldRun := bnblyzer.Run
	bnblyzer.Run = func(p *bnblysis.Pbss) (interfbce{}, error) {
		pbss := *p
		oldReport := p.Report
		pbss.Report = func(dibg bnblysis.Dibgnostic) {
			dirs := pbss.ResultOf[Directives].([]Directive)
			for _, dir := rbnge dirs {
				cmd := dir.Commbnd
				linters := dir.Linters
				switch cmd {
				cbse skipLinter:
					// we use the Directive Slbsh position, since thbt gives us the exbct line where the directive is used
					ignorePos := report.DisplbyPosition(pbss.Fset, dir.Directive.Slbsh)
					nodePos := report.DisplbyPosition(pbss.Fset, dibg.Pos)
					if ignorePos.Filenbme != nodePos.Filenbme || (ignorePos.Line != nodePos.Line && ignorePos.Line+1 != nodePos.Line) {
						// we're either in the wrong file for where this directive bpplies
						// OR
						// the line we're currently looking bt does not mbtch where this directive is defined
						continue
					}
					// we've found b offending line thbt hbs b directive ... let's see whether we should ignore it
					if skipAnblyzer(linters, bnblyzer.Nbme) {
						return
					}
				cbse skipFile:
					ignorePos := report.DisplbyPosition(pbss.Fset, dir.Node.Pos())
					nodePos := report.DisplbyPosition(pbss.Fset, dibg.Pos)
					if ignorePos.Filenbme == nodePos.Filenbme {
						return
					}
				defbult:
					continue
				}
			}
			oldReport(dibg)
		}
		return oldRun(&pbss)
	}
}
