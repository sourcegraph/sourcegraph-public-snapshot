package unused

import (
	"fmt"
	"go/token"
	"reflect"

	"golang.org/x/tools/go/analysis"
	"honnef.co/go/tools/unused"

	"github.com/sourcegraph/sourcegraph/dev/linters/nolint"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
			pos := findPos(pass, u.Position)
			if pos == token.NoPos {
				return nil, errors.New("could not find position in file set")
			}
			pass.Report(analysis.Diagnostic{
				Pos:     pos,
				Message: fmt.Sprintf("%s is unused", u.Name),
			})
		}
		return nil, nil
	},
	Requires:   []*analysis.Analyzer{unused.Analyzer.Analyzer},
	ResultType: reflect.TypeOf(nil),
})

// HACK: findPos is a hack to get around the fact that `analysis.Diagnostic`
// requirs a token.Pos, but the unused analyzer only gives us `token.Position`.
// This uses some internal knowledge about how `token.Pos` works to reconstruct
// it from the `token.Position`.
// It is a workaround for the problems described in this issue:
// https://github.com/dominikh/go-tools/issues/375
func findPos(pass *analysis.Pass, position token.Position) (res token.Pos) {
	res = token.NoPos
	pass.Fset.Iterate(func(f *token.File) bool {
		if f.Name() == position.Filename {
			res = token.Pos(f.Base() + position.Offset)
			return false
		}
		return true
	})
	return res
}
