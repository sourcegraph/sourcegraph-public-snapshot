package lucky

import (
	"encoding/json"
	"testing"

	"github.com/hexops/autogold"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
)

func apply(input string, transform []transform) string {
	type want struct {
		Input string
		Query string
	}
	q, _ := query.ParseStandard(input)
	b, _ := query.ToBasicQuery(q)
	out := applyTransformation(b, transform)
	var queryStr string
	if out == nil {
		queryStr = "DOES NOT APPLY"
	} else {
		queryStr = query.StringHuman(out.ToParseTree())
	}
	result := want{Input: input, Query: queryStr}
	json, _ := json.MarshalIndent(result, "", "  ")
	return string(json)
}

func Test_unquotePatterns(t *testing.T) {
	rule := []transform{unquotePatterns}
	test := func(input string) string {
		return apply(input, rule)
	}

	cases := []string{
		`"monitor"`,
		`repo:^github\.com/sourcegraph/sourcegraph$ "monitor" "*Monitor"`,
		`content:"not quoted"`,
	}

	for _, c := range cases {
		t.Run("unquote patterns", func(t *testing.T) {
			autogold.Equal(t, autogold.Raw(test(c)))
		})
	}

}

func Test_unorderedPatterns(t *testing.T) {
	rule := []transform{unorderedPatterns}
	test := func(input string) string {
		return apply(input, rule)
	}

	cases := []string{
		`context:global parse func`,
	}

	for _, c := range cases {
		t.Run("AND patterns", func(t *testing.T) {
			autogold.Equal(t, autogold.Raw(test(c)))
		})
	}

}

func Test_langPatterns(t *testing.T) {
	rule := []transform{langPatterns}
	test := func(input string) string {
		return apply(input, rule)
	}

	cases := []string{
		`context:global python`,
		`context:global parse python`,
	}

	for _, c := range cases {
		t.Run("lang patterns", func(t *testing.T) {
			autogold.Equal(t, autogold.Raw(test(c)))
		})
	}

}

func Test_symbolPatterns(t *testing.T) {
	rule := []transform{symbolPatterns}
	test := func(input string) string {
		return apply(input, rule)
	}

	cases := []string{
		`context:global function`,
		`context:global parse function`,
	}

	for _, c := range cases {
		t.Run("symbol patterns", func(t *testing.T) {
			autogold.Equal(t, autogold.Raw(test(c)))
		})
	}

}

func Test_typePatterns(t *testing.T) {
	rule := []transform{typePatterns}
	test := func(input string) string {
		return apply(input, rule)
	}

	cases := []string{
		`context:global fix commit`,
		`context:global code monitor commit`,
		`context:global code or monitor commit`,
	}

	for _, c := range cases {
		t.Run("type patterns", func(t *testing.T) {
			autogold.Equal(t, autogold.Raw(test(c)))
		})
	}

}

func Test_regexpPatterns(t *testing.T) {
	rule := []transform{regexpPatterns}
	test := func(input string) string {
		return apply(input, rule)
	}

	cases := []string{
		`[a-z]+`,
		`(ab)*`,
		`c++`,
		`my.yaml.conf`,
	}

	for _, c := range cases {
		t.Run("regexp patterns", func(t *testing.T) {
			autogold.Equal(t, autogold.Raw(test(c)))
		})
	}
}

func Test_patternsToCodeHostFilters(t *testing.T) {
	rule := []transform{patternsToCodeHostFilters}
	test := func(input string) string {
		return apply(input, rule)
	}

	cases := []string{
		`https://github.com/sourcegraph/sourcegraph`,
		`https://github.com/sourcegraph`,
		`github.com/sourcegraph`,
		`https://github.com/sourcegraph/sourcegraph/blob/main/lib/README.md#L50`,
		`https://github.com/sourcegraph/sourcegraph/tree/main/lib`,
		`https://github.com/sourcegraph/sourcegraph/tree/2.12`,
		`https://github.com/sourcegraph/sourcegraph/commit/abc`,
	}

	for _, c := range cases {
		t.Run("URL patterns", func(t *testing.T) {
			autogold.Equal(t, autogold.Raw(test(c)))
		})
	}
}
