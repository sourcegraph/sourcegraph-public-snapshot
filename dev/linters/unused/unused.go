package unused

import (
	"fmt"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"honnef.co/go/tools/unused"
)

var Analyzer = &analysis.Analyzer{
	Name: "unused",
	Doc:  unused.Analyzer.Analyzer.Doc,
	Run: func(pass *analysis.Pass) (interface{}, error) {
		res, err := unused.Analyzer.Analyzer.Run(pass)
		if err != nil {
			return res, err
		}
		allUnused := res.(unused.Result).Unused
		for _, u := range allUnused {
			pass.Report(analysis.Diagnostic{
				Pos:     token.Pos(u.Position.Offset),
				Message: fmt.Sprintf("%s is unused", u.Name),
			})
		}
		return res, err
	},
	Requires:   unused.Analyzer.Analyzer.Requires,
	ResultType: unused.Analyzer.Analyzer.ResultType,
	FactTypes:  []analysis.Fact{},
}
