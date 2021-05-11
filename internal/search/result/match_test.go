package result

import (
	"strings"
	"testing"

	"github.com/hexops/autogold"
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
)

func TestSelect(t *testing.T) {
	data := &FileMatch{
		Symbols: []*SymbolMatch{
			{Symbol: Symbol{Name: "a()", Kind: "func"}},
			{Symbol: Symbol{Name: "b()", Kind: "function"}},
			{Symbol: Symbol{Name: "var c", Kind: "variable"}},
		},
	}

	test := func(input string) string {
		selectPath, _ := filter.SelectPathFromString(input)
		symbols := data.Select(selectPath).(*FileMatch).Symbols
		var values []string
		for _, s := range symbols {
			values = append(values, s.Symbol.Name+":"+s.Symbol.Kind)
		}
		return strings.Join(values, ", ")
	}

	autogold.Want("filter any symbol", "a():func, b():function, var c:variable").Equal(t, test("symbol"))
	autogold.Want("filter symbol kind variable", "var c:variable").Equal(t, test("symbol.variable"))
}
