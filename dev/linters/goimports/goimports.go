package goimports

import (
	"fmt"

	"github.com/golangci/gofmt/goimports"
	"golang.org/x/tools/go/analysis"
)

// Copied from: https://sourcegraph.com/github.com/sluongng/nogo-analyzer/-/blob/goci-lint/goimports/analyzer.go
var Analyzer = &analysis.Analyzer{
	Name: "goimports",
	Doc:  "Command goimports updates your Go import lines, adding missing ones and removing unreferenced ones.",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	var fileNames []string
	for _, f := range pass.Files {
		pos := pass.Fset.PositionFor(f.Pos(), false)
		fileNames = append(fileNames, pos.Filename)
	}

	for _, f := range fileNames {
		diff, err := goimports.Run(f)
		if err != nil {
			return nil, fmt.Errorf("could not run goimports: %w", err)
		}

		if diff == nil {
			continue
		}

		pass.Report(analysis.Diagnostic{
			Pos:     1,
			Message: fmt.Sprintf("\n%s", diff),
		})
	}

	return nil, nil
}
