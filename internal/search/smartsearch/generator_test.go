package smartsearch

import (
	"encoding/json"
	"testing"

	"github.com/hexops/autogold/v2"

	"github.com/sourcegraph/sourcegraph/internal/search/query"
)

type want struct {
	Description string
	Input       string
	Query       string
}

func TestNewGenerator(t *testing.T) {
	test := func(input string, rulesNarrow, rulesWiden []rule) string {
		q, _ := query.ParseStandard(input)
		b, _ := query.ToBasicQuery(q)
		g := NewGenerator(b, rulesNarrow, rulesWiden)
		result, _ := json.MarshalIndent(generateAll(g, input), "", "  ")
		return string(result)
	}

	cases := [][2][]rule{
		{rulesNarrow, rulesWiden},
		{rulesNarrow, nil},
		{nil, rulesWiden},
	}

	for _, c := range cases {
		t.Run("rule application", func(t *testing.T) {
			autogold.ExpectFile(t, autogold.Raw(test(`go commit yikes derp`, c[0], c[1])))
		})
	}
}

func TestSkippedRules(t *testing.T) {
	test := func(input string) string {
		q, _ := query.ParseStandard(input)
		b, _ := query.ToBasicQuery(q)
		g := NewGenerator(b, rulesNarrow, rulesWiden)
		result, _ := json.MarshalIndent(generateAll(g, input), "", "  ")
		return string(result)
	}

	c := `type:diff foo bar`

	t.Run("do not apply rules for type_diff", func(t *testing.T) {
		autogold.ExpectFile(t, autogold.Raw(test(c)))
	})
}

func TestSkipInvalidQueries(t *testing.T) {
	test := func(input string) []want {
		q, _ := query.ParseStandard(input)
		b, _ := query.ToBasicQuery(q)
		g := NewGenerator(b, rulesNarrow, rulesWiden)
		return generateAll(g, input)
	}

	// The "expand URLs to filters" rule can produce a repo filter with
	// an invalid regex, like `repo:github.com/org/repo(`
	c := `github.com/org/repo(/tree/rev)`
	got := test(c)
	if len(got) != 0 {
		t.Errorf("expected no queries to be generated")
	}
}

func generateAll(g next, input string) []want {
	var autoQ *autoQuery
	generated := []want{}
	for g != nil {
		autoQ, g = g()
		generated = append(
			generated,
			want{
				Description: autoQ.description,
				Input:       input,
				Query:       query.StringHuman(autoQ.query.ToParseTree()),
			})
	}
	return generated
}
