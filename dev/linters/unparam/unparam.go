package unparam

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/packages"
	"mvdan.cc/unparam/check"

	"github.com/sourcegraph/sourcegraph/dev/linters/nolint"
)

var Analyzer *analysis.Analyzer

func init() {
	Analyzer = nolint.Wrap(&analysis.Analyzer{
		Name:             "unparam",
		Doc:              "Reports unused function parameters and results in your code",
		Run:              run,
		Requires:         []*analysis.Analyzer{buildssa.Analyzer}, // required since unparam requires the result
		RunDespiteErrors: false,
	})

}

// Test is a test function to check that this linter works
// To check whether this linter works, remove the nolint directive
func Test(a string, b string) { //nolint:unparam
	println("USING A but not B", a)
}

// run is lifted from golangci, with the only change in how issues are reported
func run(pass *analysis.Pass) (interface{}, error) {
	ssa := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)
	ssaPkg := ssa.Pkg

	pkg := &packages.Package{
		Fset:      pass.Fset,
		Syntax:    pass.Files,
		Types:     pass.Pkg,
		TypesInfo: pass.TypesInfo,
	}

	c := &check.Checker{}
	c.CheckExportedFuncs(false)
	c.Packages([]*packages.Package{pkg})
	c.ProgramSSA(ssaPkg.Prog)

	issues, err := c.Check()
	if err != nil {
		return nil, err
	}
	for _, issue := range issues {
		pass.Report(analysis.Diagnostic{
			Pos:     issue.Pos(),
			Message: issue.Message(),
		})
	}
	return nil, nil
}
