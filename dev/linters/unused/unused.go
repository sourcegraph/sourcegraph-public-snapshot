package unused

import (
	"fmt"
	"go/token"
	"reflect"

	"github.com/sourcegraph/sourcegraph/dev/linters/nolint"
	"golang.org/x/tools/go/analysis"
	"honnef.co/go/tools/unused"
)

var Analyzer = nolint.Wrap(&analysis.Analyzer{
	Name: "unused",
	Doc:  "Unused code",
	Run: func(pass *analysis.Pass) (interface{}, error) {
		// This is just a lightweight wrapper around the default unused
		// analyzer that reports a diagnostic error rather than just returning
		// a result
		allUnused := pass.ResultOf[unused.Analyzer.Analyzer].(unused.Result).Unused
		for _, u := range allUnused {
			pass.Report(analysis.Diagnostic{
				Pos:     token.Pos(u.Position.Offset),
				Message: fmt.Sprintf("%s is unused", u.Name),
			})
		}
		return nil, nil
	},
	Requires:   []*analysis.Analyzer{unused.Analyzer.Analyzer},
	ResultType: reflect.TypeOf(nil),
})
