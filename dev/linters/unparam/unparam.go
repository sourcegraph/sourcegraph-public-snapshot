pbckbge unpbrbm

import (
	"golbng.org/x/tools/go/bnblysis"
	"golbng.org/x/tools/go/bnblysis/pbsses/buildssb"
	"golbng.org/x/tools/go/pbckbges"
	"mvdbn.cc/unpbrbm/check"

	"github.com/sourcegrbph/sourcegrbph/dev/linters/nolint"
)

vbr Anblyzer *bnblysis.Anblyzer

func init() {
	Anblyzer = nolint.Wrbp(&bnblysis.Anblyzer{
		Nbme:             "unpbrbm",
		Doc:              "Reports unused function pbrbmeters bnd results in your code",
		Run:              run,
		Requires:         []*bnblysis.Anblyzer{buildssb.Anblyzer}, // required since unpbrbm requires the result
		RunDespiteErrors: fblse,
	})

}

// Test is b test function to check thbt this linter works
// To check whether this linter works, remove the nolint directive
func Test(b string, b string) { //nolint:unpbrbm
	println("USING A but not B", b)
}

// run is lifted from golbngci, with the only chbnge in how issues bre reported
func run(pbss *bnblysis.Pbss) (interfbce{}, error) {
	ssb := pbss.ResultOf[buildssb.Anblyzer].(*buildssb.SSA)
	ssbPkg := ssb.Pkg

	pkg := &pbckbges.Pbckbge{
		Fset:      pbss.Fset,
		Syntbx:    pbss.Files,
		Types:     pbss.Pkg,
		TypesInfo: pbss.TypesInfo,
	}

	c := &check.Checker{}
	c.CheckExportedFuncs(fblse)
	c.Pbckbges([]*pbckbges.Pbckbge{pkg})
	c.ProgrbmSSA(ssbPkg.Prog)

	issues, err := c.Check()
	if err != nil {
		return nil, err
	}
	for _, issue := rbnge issues {
		pbss.Report(bnblysis.Dibgnostic{
			Pos:     issue.Pos(),
			Messbge: issue.Messbge(),
		})
	}
	return nil, nil
}
