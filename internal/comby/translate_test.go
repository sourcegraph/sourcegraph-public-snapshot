package comby

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestStructuralPatToRegexpQuery(t *testing.T) {
	cases := []struct {
		Name    string
		Pattern string
		Want    string
	}{
		{
			Name:    "Just a hole",
			Pattern: ":[1]",
			Want:    `(?:.|\s)*?`,
		},
		{
			Name:    "Adjacent holes",
			Pattern: ":[1]:[2]:[3]",
			Want:    `(?:.|\s)*?`,
		},
		{
			Name:    "Substring between holes",
			Pattern: ":[1] substring :[2]",
			Want:    `(?:[\s]+substring[\s]+)`,
		},
		{
			Name:    "Substring before and after different hole kinds",
			Pattern: "prefix :[[1]] :[2.] suffix",
			Want:    `(?:prefix[\s]+)(?:.|\s)*?(?:[\s]+)(?:.|\s)*?(?:[\s]+suffix)`,
		},
		{
			Name:    "Substrings covering all hole kinds.",
			Pattern: `1. :[1] 2. :[[2]] 3. :[3.] 4. :[4\n] 5. :[ ] 6. :[ 6] done.`,
			Want:    `(?:1\.[\s]+)(?:.|\s)*?(?:[\s]+2\.[\s]+)(?:.|\s)*?(?:[\s]+3\.[\s]+)(?:.|\s)*?(?:[\s]+4\.[\s]+)(?:.|\s)*?(?:[\s]+5\.[\s]+)(?:.|\s)*?(?:[\s]+6\.[\s]+)(?:.|\s)*?(?:[\s]+done\.)`,
		},
		{
			Name:    "Allow alphanumeric identifiers in holes",
			Pattern: "sub :[alphanum_ident_123] string",
			Want:    `(?:sub[\s]+)(?:.|\s)*?(?:[\s]+string)`,
		},

		{
			Name:    "Whitespace separated holes",
			Pattern: ":[1] :[2]",
			Want:    `(?:[\s]+)`,
		},
		{
			Name:    "Expect newline separated pattern",
			Pattern: "ParseInt(:[stuff], :[x]) if err ",
			Want:    `(?:ParseInt\()(?:.|\s)*?(?:,[\s]+)(?:.|\s)*?(?:\)[\s]+if[\s]+err[\s]+)`,
		},
		{
			Name: "Contiguous whitespace is replaced by regex",
			Pattern: `ParseInt(:[stuff],    :[x])
             if err `,
			Want: `(?:ParseInt\()(?:.|\s)*?(?:,[\s]+)(?:.|\s)*?(?:\)[\s]+if[\s]+err[\s]+)`,
		},
		{
			Name:    "Regex holes extracts regex",
			Pattern: `:[x~[yo]]`,
			Want:    `(?:[yo])`,
		},
		{
			Name:    "Regex holes with escaped space",
			Pattern: `:[x~\ ]`,
			Want:    `(?:\ )`,
		},
		{
			Name:    "Shorthand",
			Pattern: ":[[1]]",
			Want:    `(?:.|\s)*?`,
		},
		{
			Name:    "Array-like preserved",
			Pattern: `[:[x]]`,
			Want:    `(?:\[)(?:.|\s)*?(?:\])`,
		},
		{
			Name:    "Shorthand",
			Pattern: ":[[1]]",
			Want:    `(?:.|\s)*?`,
		},
		{
			Name:    "Not well-formed is undefined",
			Pattern: ":[[",
			Want:    `(?::\[\[)`,
		},
		{
			Name:    "Complex regex with character class",
			Pattern: `:[chain~[^(){}\[\],]+\n( +\..*\n)+]`,
			Want:    `(?:[^(){}\[\],]+\n( +\..*\n)+)`,
		},
		{
			Name:    "Colon regex",
			Pattern: `:[~:]`,
			Want:    `(?::)`,
		},
		{
			Name:    "Colon prefix",
			Pattern: `::[version]bar`,
			Want:    `(?::)(?:.|\s)*?(?:bar)`,
		},
		{
			Name:    "Colon prefix",
			Pattern: `::::[version]bar`,
			Want:    `(?::::)(?:.|\s)*?(?:bar)`,
		},
	}
	for _, tt := range cases {
		t.Run(tt.Name, func(t *testing.T) {
			got := StructuralPatToRegexpQuery(tt.Pattern, false)
			if diff := cmp.Diff(tt.Want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}
