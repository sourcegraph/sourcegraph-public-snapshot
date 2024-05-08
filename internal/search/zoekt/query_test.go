package zoekt

import (
	"testing"

	"github.com/hexops/autogold/v2"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

func TestQueryToZoektQuery(t *testing.T) {
	cases := []struct {
		Name            string
		Type            search.IndexedRequestType
		Query           string
		Features        search.Features
		WantZoektOutput autogold.Value
	}{
		{
			Name:            "substr",
			Type:            search.TextRequest,
			Query:           `foo patterntype:regexp`,
			WantZoektOutput: autogold.Expect(`substr:"foo"`),
		},
		{
			Name:            "symbol substr",
			Type:            search.SymbolRequest,
			Query:           `foo patterntype:regexp type:symbol`,
			WantZoektOutput: autogold.Expect(`sym:substr:"foo"`),
		},
		{
			Name:            "regex",
			Type:            search.TextRequest,
			Query:           `(foo).*?(bar) patterntype:regexp`,
			WantZoektOutput: autogold.Expect(`regex:"foo(?-s:.)*?bar"`),
		},
		{
			Name:            "path",
			Type:            search.TextRequest,
			Query:           `foo file:\.go$ file:\.yaml$ -file:\bvendor\b patterntype:regexp`,
			WantZoektOutput: autogold.Expect(`(and substr:"foo" file_regex:"\\.go(?m:$)" file_regex:"\\.yaml(?m:$)" (not file_regex:"\\bvendor\\b"))`),
		},
		{
			Name:            "case",
			Type:            search.TextRequest,
			Query:           `foo case:yes patterntype:regexp file:\.go$ file:yaml`,
			WantZoektOutput: autogold.Expect(`(and case_substr:"foo" case_file_regex:"\\.go(?m:$)" case_file_substr:"yaml")`),
		},
		{
			Name:            "casepath",
			Type:            search.TextRequest,
			Query:           `foo case:yes file:\.go$ file:\.yaml$ -file:\bvendor\b patterntype:regexp`,
			WantZoektOutput: autogold.Expect(`(and case_substr:"foo" case_file_regex:"\\.go(?m:$)" case_file_regex:"\\.yaml(?m:$)" (not case_file_regex:"\\bvendor\\b"))`),
		},
		{
			Name:            "path matches only",
			Type:            search.TextRequest,
			Query:           `test type:path`,
			WantZoektOutput: autogold.Expect(`file_substr:"test"`),
		},
		{
			Name:            "content matches only",
			Type:            search.TextRequest,
			Query:           `test type:file patterntype:literal`,
			WantZoektOutput: autogold.Expect(`content_substr:"test"`),
		},
		{
			Name:            "content and path matches",
			Type:            search.TextRequest,
			Query:           `test`,
			WantZoektOutput: autogold.Expect(`substr:"test"`),
		},
		{
			Name:            "Just file",
			Type:            search.TextRequest,
			Query:           `file:\.go$`,
			WantZoektOutput: autogold.Expect(`file_regex:"\\.go(?m:$)"`),
		},
		{
			Name:            "Languages get passed as file filter",
			Type:            search.TextRequest,
			Query:           `file:\.go$ lang:go`,
			WantZoektOutput: autogold.Expect(`(and file_regex:"(?i:\\.GO)(?m:$)" file_regex:"\\.go(?m:$)")`),
		},
		{
			Name:            "Languages still use case_insensitive in case sensitivity mode (Go)",
			Type:            search.TextRequest,
			Query:           `file:\.go$ lang:go case:true`,
			WantZoektOutput: autogold.Expect(`(and file_regex:"(?i:\\.GO)(?m:$)" case_file_regex:"\\.go(?m:$)")`),
		},
		{
			Name:            "Languages still use case_insensitive in case sensitivity mode (Typescript)",
			Type:            search.TextRequest,
			Query:           `lang:typescript case:true`,
			WantZoektOutput: autogold.Expect(`file_regex:"(?i:\\.)(?:(?i:TS)(?m:$)|(?i:CTS)(?m:$)|(?i:MTS)(?m:$))"`),
		},
		{
			Name:  "Language get passed as lang: query",
			Type:  search.TextRequest,
			Query: `lang:go`,
			Features: search.Features{
				ContentBasedLangFilters: true,
			},
			WantZoektOutput: autogold.Expect(`lang:Go`),
		},
		{
			Name:  "Multiple languages get passed as lang queries",
			Type:  search.TextRequest,
			Query: `lang:go lang:typescript`,
			Features: search.Features{
				ContentBasedLangFilters: true,
			},
			WantZoektOutput: autogold.Expect(`(and lang:Go lang:TypeScript)`),
		},
		{
			Name:  "Excluded languages get passed as lang: query",
			Type:  search.TextRequest,
			Query: `lang:go -lang:typescript -lang:markdown`,
			Features: search.Features{
				ContentBasedLangFilters: true,
			},
			WantZoektOutput: autogold.Expect(`(and lang:Go (not lang:TypeScript) (not lang:Markdown))`),
		},
		{
			Name:  "Mixed file and lang filters",
			Type:  search.TextRequest,
			Query: `file:\.go$ lang:go lang:typescript`,
			Features: search.Features{
				ContentBasedLangFilters: true,
			},
			WantZoektOutput: autogold.Expect(`(and lang:Go lang:TypeScript file_regex:"\\.go(?m:$)")`),
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

			tt.WantZoektOutput.Equal(t, got.String())
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
