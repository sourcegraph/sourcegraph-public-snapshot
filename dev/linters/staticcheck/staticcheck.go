//go:generate go run ./cmd/gen.go BUILD.bazel
package staticcheck

import (
	"golang.org/x/tools/go/analysis"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"

	"github.com/sourcegraph/sourcegraph/dev/linters/nolint"
)

// AllAnalyzers contains staticcheck and gosimple Analyzers
var AllAnalyzers = append(staticcheck.Analyzers, simple.Analyzers...)

var AnalyzerName = ""
var Analyzer *analysis.Analyzer = GetAnalyzerByName(AnalyzerName)

func GetAnalyzerByName(name string) *analysis.Analyzer {
	for _, a := range AllAnalyzers {
		if a.Analyzer.Name == name {
			return nolint.Wrap(a.Analyzer)
		}
	}
	return nil
}
