package zoekt

import (
	"cmp"
	"slices"
	"testing"

	"github.com/hexops/autogold/v2"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/result"

	zoekt "github.com/sourcegraph/zoekt/query"
)

func TestQueryToZoektQuery(t *testing.T) {
	cases := []struct {
		Name           string
		Type           search.IndexedRequestType
		Query          string
		Features       search.Features
		WantZoektQuery string
	}{
		{
			Name:           "substr",
			Type:           search.TextRequest,
			Query:          `foo patterntype:regexp`,
			WantZoektQuery: "foo case:no",
		},
		{
			Name:           "symbol substr",
			Type:           search.SymbolRequest,
			Query:          `foo patterntype:regexp type:symbol`,
			WantZoektQuery: "sym:foo case:no",
		},
		{
			Name:           "regex",
			Type:           search.TextRequest,
			Query:          `(foo).*?(bar) patterntype:regexp`,
			WantZoektQuery: "(foo).*?(bar) case:no",
		},
		{
			Name:           "path",
			Type:           search.TextRequest,
			Query:          `foo file:\.go$ file:\.yaml$ -file:\bvendor\b patterntype:regexp`,
			WantZoektQuery: `foo case:no f:\.go$ f:\.yaml$ -f:\bvendor\b`,
		},
		{
			Name:           "case",
			Type:           search.TextRequest,
			Query:          `foo case:yes patterntype:regexp file:\.go$ file:yaml`,
			WantZoektQuery: `foo case:yes f:\.go$ f:yaml`,
		},
		{
			Name:           "casepath",
			Type:           search.TextRequest,
			Query:          `foo case:yes file:\.go$ file:\.yaml$ -file:\bvendor\b patterntype:regexp`,
			WantZoektQuery: `foo case:yes f:\.go$ f:\.yaml$ -f:\bvendor\b`,
		},
		{
			Name:           "path matches only",
			Type:           search.TextRequest,
			Query:          `test type:path`,
			WantZoektQuery: `f:test`,
		},
		{
			Name:           "content matches only",
			Type:           search.TextRequest,
			Query:          `test type:file patterntype:literal`,
			WantZoektQuery: `c:test`,
		},
		{
			Name:           "content and path matches",
			Type:           search.TextRequest,
			Query:          `test`,
			WantZoektQuery: `test`,
		},
		{
			Name:           "Just file",
			Type:           search.TextRequest,
			Query:          `file:\.go$`,
			WantZoektQuery: `file:"\\.go(?m:$)"`,
		},
		{
			Name:           "Languages get passed as file filter",
			Type:           search.TextRequest,
			Query:          `file:\.go$ lang:go`,
			WantZoektQuery: `file:"\\.go(?m:$)" file:"\\.go(?m:$)"`,
		},
		{
			Name:  "Language get passed as lang: query",
			Type:  search.TextRequest,
			Query: `lang:go`,
			Features: search.Features{
				ContentBasedLangFilters: true,
			},
			WantZoektQuery: `lang:go`,
		},
		{
			Name:  "Multiple languages get passed as lang queries",
			Type:  search.TextRequest,
			Query: `lang:go lang:typescript`,
			Features: search.Features{
				ContentBasedLangFilters: true,
			},
			WantZoektQuery: `lang:go lang:typescript`,
		},
		{
			Name:  "Excluded languages get passed as lang: query",
			Type:  search.TextRequest,
			Query: `lang:go -lang:typescript -lang:markdown`,
			Features: search.Features{
				ContentBasedLangFilters: true,
			},
			WantZoektQuery: `lang:go -lang:typescript -lang:markdown`,
		},
		{
			Name:  "Mixed file and lang filters",
			Type:  search.TextRequest,
			Query: `file:\.go$ lang:go lang:typescript`,
			Features: search.Features{
				ContentBasedLangFilters: true,
			},
			WantZoektQuery: `file:"\\.go(?m:$)" lang:go lang:typescript`,
		},
	}
	for _, tt := range cases {
		t.Run(tt.Name, func(t *testing.T) {
			sourceQuery, _ := query.ParseRegexp(tt.Query)
			b, _ := query.ToBasicQuery(sourceQuery)

			types, _ := b.ToParseTree().StringValues(query.FieldType)
			resultTypes := computeResultTypes(types, b, query.SearchTypeRegex)
			got, err := QueryToZoektQuery(b, resultTypes, &tt.Features, tt.Type)
			if err != nil {
				t.Fatal("QueryToZoektQuery failed:", err)
			}

			zoektQuery, err := zoekt.Parse(tt.WantZoektQuery)
			if err != nil {
				t.Fatalf("failed to parse %q: %v", tt.WantZoektQuery, err)
			}

			if !queryEqual(got, zoektQuery) {
				t.Fatalf("mismatched queries\ngot  %s\nwant %s", got.String(), zoektQuery.String())
			}
		})
	}
}

func Test_toZoektPattern(t *testing.T) {
	test := func(input string, searchType query.SearchType, typ search.IndexedRequestType) string {
		p, err := query.Pipeline(query.Init(input, searchType))
		if err != nil {
			return err.Error()
		}
		zoektQuery, err := toZoektPattern(p[0].Pattern, false, false, false, typ)
		if err != nil {
			return err.Error()
		}
		return zoektQuery.String()
	}

	autogold.Expect(`substr:"a"`).
		Equal(t, test(`a`, query.SearchTypeLiteral, search.TextRequest))

	autogold.Expect(`(or (and substr:"a" substr:"b" (not substr:"c")) substr:"d")`).
		Equal(t, test(`a and b and not c or d`, query.SearchTypeLiteral, search.TextRequest))

	autogold.Expect(`substr:"\"func main() {\\n\""`).
		Equal(t, test(`"func main() {\n"`, query.SearchTypeLiteral, search.TextRequest))

	autogold.Expect(`substr:"func main() {\n"`).
		Equal(t, test(`"func main() {\n"`, query.SearchTypeRegex, search.TextRequest))

	autogold.Expect(`(and sym:substr:"foo" (not sym:substr:"bar"))`).
		Equal(t, test(`type:symbol (foo and not bar)`, query.SearchTypeLiteral, search.SymbolRequest))
}

func queryEqual(a, b zoekt.Q) bool {
	sortChildren := func(q zoekt.Q) zoekt.Q {
		switch s := q.(type) {
		case *zoekt.And:
			slices.SortFunc(s.Children, zoektQStringCompare)
		case *zoekt.Or:
			slices.SortFunc(s.Children, zoektQStringCompare)
		}
		return q
	}
	return zoekt.Map(a, sortChildren).String() == zoekt.Map(b, sortChildren).String()
}

func zoektQStringCompare(a, b zoekt.Q) int {
	return cmp.Compare(a.String(), b.String())
}

func computeResultTypes(types []string, b query.Basic, searchType query.SearchType) result.Types {
	var rts result.Types
	if searchType == query.SearchTypeStructural && !b.IsEmptyPattern() {
		rts = result.TypeStructural
	} else {
		if len(types) == 0 {
			rts = result.TypeFile | result.TypePath | result.TypeRepo
		} else {
			for _, t := range types {
				rts = rts.With(result.TypeFromString[t])
			}
		}
	}
	return rts
}
