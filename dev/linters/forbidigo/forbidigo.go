package forbidigo

import (
	"fmt"
	"go/ast"

	"github.com/ashanbrown/forbidigo/forbidigo"
	"github.com/ashanbrown/forbidigo/pkg/analyzer"
	"golang.org/x/tools/go/analysis"
)

// Analyzer is the analyzer nogo should use
var Analyzer = analyzer.NewAnalyzer()

var defaultPatterns = []string{
	"^fmt\\.Errorf$",
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
	linter, err := forbidigo.NewLinter(defaultPatterns,
		forbidigo.OptionIgnorePermitDirectives(config.IgnorePermitDirective),
		forbidigo.OptionExcludeGodocExamples(config.ExcludeGodocExamples),
		forbidigo.OptionAnalyzeTypes(config.AnalyzeTypes),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to configure linter: %w", err)
	}
	nodes := make([]ast.Node, 0, len(pass.Files))
	for _, f := range pass.Files {
		nodes = append(nodes, f)
	}
	runConfig := forbidigo.RunConfig{Fset: pass.Fset}
	if config.AnalyzeTypes {
		runConfig.TypesInfo = pass.TypesInfo
	}
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
	return nil, nil
}
