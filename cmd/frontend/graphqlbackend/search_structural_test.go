package graphqlbackend

import (
	"testing"

	zoektquery "github.com/google/zoekt/query"
)

func TestStructuralPatToZoektQuery(t *testing.T) {
	cases := []struct {
		Name     string
		Pattern  string
		Function func(string, bool) (zoektquery.Q, error)
		Want     string
	}{
		{
			Name:     "Just a hole",
			Pattern:  ":[1]",
			Function: StructuralPatToRegexpQuery,
			Want:     `(and case_regex:"()((?s:.))*?()")`,
		},
		{
			Name:     "Adjacent holes",
			Pattern:  ":[1]:[2]:[3]",
			Function: StructuralPatToRegexpQuery,
			Want:     `(and case_regex:"()((?s:.))*?()((?s:.))*?()((?s:.))*?()")`,
		},
		{
			Name:     "Substring between holes",
			Pattern:  ":[1] substring :[2]",
			Function: StructuralPatToRegexpQuery,
			Want:     `(and case_regex:"()((?s:.))*?([\\t-\\n\\f-\\r ]+substring[\\t-\\n\\f-\\r ]+)((?s:.))*?()")`,
		},
		{
			Name:     "Substring before and after different hole kinds",
			Pattern:  "prefix :[[1]] :[2.] suffix",
			Function: StructuralPatToRegexpQuery,
			Want:     `(and case_regex:"(prefix[\\t-\\n\\f-\\r ]+)((?s:.))*?([\\t-\\n\\f-\\r ]+)((?s:.))*?([\\t-\\n\\f-\\r ]+suffix)")`,
		},
		{
			Name:     "Substrings covering all hole kinds.",
			Pattern:  `1. :[1] 2. :[[2]] 3. :[3.] 4. :[4\n] 5. :[ ] 6. :[ 6] done.`,
			Function: StructuralPatToRegexpQuery,
			Want:     `(and case_regex:"(1\\.[\\t-\\n\\f-\\r ]+)((?s:.))*?([\\t-\\n\\f-\\r ]+2\\.[\\t-\\n\\f-\\r ]+)((?s:.))*?([\\t-\\n\\f-\\r ]+3\\.[\\t-\\n\\f-\\r ]+)((?s:.))*?([\\t-\\n\\f-\\r ]+4\\.[\\t-\\n\\f-\\r ]+)((?s:.))*?([\\t-\\n\\f-\\r ]+5\\.[\\t-\\n\\f-\\r ]+)((?s:.))*?([\\t-\\n\\f-\\r ]+6\\.[\\t-\\n\\f-\\r ]+)((?s:.))*?([\\t-\\n\\f-\\r ]+done\\.)")`,
		},
		{
			Name:     "Substrings across multiple lines.",
			Pattern:  ``,
			Function: StructuralPatToRegexpQuery,
			Want:     `(and case_regex:"()")`,
		},
		{
			Name:     "Allow alphanumeric identifiers in holes",
			Pattern:  "sub :[alphanum_ident_123] string",
			Function: StructuralPatToRegexpQuery,
			Want:     `(and case_regex:"(sub[\\t-\\n\\f-\\r ]+)((?s:.))*?([\\t-\\n\\f-\\r ]+string)")`,
		},

		{
			Name:     "Whitespace separated holes",
			Pattern:  ":[1] :[2]",
			Function: StructuralPatToRegexpQuery,
			Want:     `(and case_regex:"()((?s:.))*?([\\t-\\n\\f-\\r ]+)((?s:.))*?()")`,
		},
		{
			Name:     "Expect newline separated pattern",
			Pattern:  "ParseInt(:[stuff], :[x]) if err ",
			Function: StructuralPatToRegexpQuery,
			Want:     `(and case_regex:"(ParseInt\\()((?s:.))*?(,[\\t-\\n\\f-\\r ]+)((?s:.))*?(\\)[\\t-\\n\\f-\\r ]+if[\\t-\\n\\f-\\r ]+err[\\t-\\n\\f-\\r ]+)")`,
		},
		{
			Name: "Contiguous whitespace is replaced by regex",
			Pattern: `ParseInt(:[stuff],    :[x])
             if err `,
			Function: StructuralPatToRegexpQuery,
			Want:     `(and case_regex:"(ParseInt\\()((?s:.))*?(,[\\t-\\n\\f-\\r ]+)((?s:.))*?(\\)[\\t-\\n\\f-\\r ]+if[\\t-\\n\\f-\\r ]+err[\\t-\\n\\f-\\r ]+)")`,
		},
	}
	for _, tt := range cases {
		t.Run(tt.Name, func(t *testing.T) {
			got, _ := tt.Function(tt.Pattern, false)
			if got.String() != tt.Want {
				t.Fatalf("mismatched queries\ngot  %s\nwant %s", got.String(), tt.Want)
			}
		})
	}
}
