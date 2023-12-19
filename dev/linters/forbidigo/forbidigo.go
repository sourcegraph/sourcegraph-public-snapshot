package forbidigo

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/ashanbrown/forbidigo/forbidigo"
	"github.com/ashanbrown/forbidigo/pkg/analyzer"
	"github.com/gobwas/glob"
	"golang.org/x/tools/go/analysis"

	"github.com/sourcegraph/sourcegraph/dev/linters/nolint"
)

// Analyzer is the analyzer nogo should use
var Analyzer = nolint.Wrap(analyzer.NewAnalyzer())

// patterns the patterns forbigigo should ban if they match.
// JSON based off the pattern struct:
// https://sourcegraph.com/github.com/ashanbrown/forbidigo@57bf5a72a4f5c3c28dce5a8aebb37a9be36bece7/-/blob/forbidigo/patterns.go?L13-29
var patterns = []struct {
	Pattern     string
	ExcludeFile func(path string) bool
}{
	{
		Pattern: fmt.Sprintf(`{"p": "^fmt\\.Errorf$", "msg": "%s should be used instead"}`, "`errors.Newf`"),
	},
	{
		Pattern: `{"p": "^(os\\.Exit|signal\\.Notify|logger\\.Fatal|log\\.Fatal)$", "msg": "it breaks dev/sg/internal/interrupt"}`,
		ExcludeFile: func(path string) bool {
			match := glob.MustCompile("**dev/sg/**.go", '/')
			return !match.Match(path) || (strings.Contains(path, "dev/sg/interrupt") || strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, "dev/sg/linters/go_checks.go"))
		},
	},
}

var config = struct {
	IgnorePermitDirective bool
	ExcludeGodocExamples  bool
	AnalyzeTypes          bool
}{
	IgnorePermitDirective: true,
	ExcludeGodocExamples:  true,
	AnalyzeTypes:          true,
}

func init() {
	// We replace run here with our own runAnalysis since the one from NewAnalyzer
	// doesn't allow us to specify patterns ...
	Analyzer.Run = runAnalysis
}

// runAnalysis is copied from forbigigo and slightly modified
func runAnalysis(pass *analysis.Pass) (interface{}, error) {
	for _, group := range patterns {
		linter, err := forbidigo.NewLinter([]string{group.Pattern},
			forbidigo.OptionIgnorePermitDirectives(config.IgnorePermitDirective),
			forbidigo.OptionExcludeGodocExamples(config.ExcludeGodocExamples),
			forbidigo.OptionAnalyzeTypes(config.AnalyzeTypes),
		)
		if err != nil {
			return nil, err
		}

		nodes := make([]ast.Node, 0, len(pass.Files))
		for _, f := range pass.Files {
			if group.ExcludeFile != nil && group.ExcludeFile(pass.Fset.Position(f.Package).Filename) {
				continue
			}
			nodes = append(nodes, f)
		}

		runConfig := forbidigo.RunConfig{Fset: pass.Fset, TypesInfo: pass.TypesInfo}
		issues, err := linter.RunWithConfig(runConfig, nodes...)
		if err != nil {
			return nil, err
		}

		for _, i := range issues {
			diag := analysis.Diagnostic{
				Pos:      i.Pos(),
				Message:  i.Details(),
				Category: "restriction",
			}
			pass.Report(diag)
		}
	}
	return nil, nil
}
