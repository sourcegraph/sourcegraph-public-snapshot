package exhaustruct

import (
	"github.com/GaijinEntertainment/go-exhaustruct/v3/analyzer"
	"github.com/sourcegraph/sourcegraph/dev/linters/nolint"
	"golang.org/x/tools/go/analysis"
)

var Analyzer *analysis.Analyzer = nolint.Wrap(createAnalyzer())

func createAnalyzer() *analysis.Analyzer {
	var include []string
	var exclude []string

	analyzer, err := analyzer.NewAnalyzer(include, exclude)
	if err != nil {
		panic(err)
	}

	return analyzer
}
