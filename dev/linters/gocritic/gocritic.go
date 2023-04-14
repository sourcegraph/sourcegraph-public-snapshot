package gocritic

import (
	"fmt"
	"path/filepath"

	"github.com/go-critic/go-critic/framework/linter"
	"golang.org/x/tools/go/analysis"

	"github.com/sourcegraph/sourcegraph/dev/linters/nolint"
)

var Analyzer *analysis.Analyzer = createAnalyzer()

// DisabledLinters is a list of linter names that should be disabled and not be used during analysis
var DisabledLinters = []string{
	"appendAssign",
	"assignOp",
	"commentFormatting",
	"deprecatedComment",
	"exitAfterDefer",
	"ifElseChain",
	"singleCaseSwitch",
}

func createAnalyzer() *analysis.Analyzer {
	linters := GetEnabledLinters(DisabledLinters...)

	return nolint.Wrap(&analysis.Analyzer{
		Name: "gocritic",
		Doc:  "linter from go-critic/go-critic, modified to be runnable in nogo",
		Run:  runWithLinters(linters),
	})
}

// runWithLinters is copied from https://sourcegraph.com/github.com/go-critic/go-critic@3f8d719ce34bb78eacfdb8fef52228aff8cbdb10/-/blob/checkers/analyzer/run.go?L27
// modifications: manually specify the checkers instead of parsing it from command line args
func runWithLinters(enabledLinters []*linter.CheckerInfo) func(*analysis.Pass) (interface{}, error) {
	return func(pass *analysis.Pass) (interface{}, error) {
		ctx := linter.NewContext(pass.Fset, pass.TypesSizes)
		ctx.SetPackageInfo(pass.TypesInfo, pass.Pkg)

		checkers, err := toCheckers(ctx, enabledLinters)
		if err != nil {
			return nil, err
		}
		for _, f := range pass.Files {
			filename := filepath.Base(pass.Fset.Position(f.Pos()).Filename)
			ctx.SetFileInfo(filename, f)
			for _, c := range checkers {
				warnings := c.Check(f)
				for _, warning := range warnings {
					diag := analysis.Diagnostic{
						Pos:     warning.Pos,
						Message: fmt.Sprintf("%s: %s", c.Info.Name, warning.Text),
					}
					if warning.HasQuickFix() {
						diag.SuggestedFixes = []analysis.SuggestedFix{
							{
								Message: "suggested replacement",
								TextEdits: []analysis.TextEdit{
									{
										Pos:     warning.Suggestion.From,
										End:     warning.Suggestion.To,
										NewText: warning.Suggestion.Replacement,
									},
								},
							},
						}
					}
					pass.Report(diag)
				}
			}
		}

		return nil, nil
	}
}

func toCheckers(ctx *linter.Context, linters []*linter.CheckerInfo) ([]*linter.Checker, error) {
	checkers := []*linter.Checker{}

	for _, l := range linters {
		c, err := linter.NewChecker(ctx, l)
		if err != nil {
			return nil, err
		}
		checkers = append(checkers, c)
	}

	return checkers, nil
}

func GetEnabledLinters(disabledLinters ...string) []*linter.CheckerInfo {
	all := linter.GetCheckersInfo()

	enabled := []*linter.CheckerInfo{}
	for _, l := range all {
		if !isDisabled(l, disabledLinters) {
			enabled = append(enabled, l)
		}
	}

	return enabled
}

func isDisabled(linter *linter.CheckerInfo, disabledLinters []string) bool {
	for _, d := range disabledLinters {
		if d == linter.Name {
			return true
		}
	}
	return false
}
