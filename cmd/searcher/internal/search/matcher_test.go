package search

import (
	"regexp/syntax" //nolint:depguard // using the grafana fork of regexp clashes with zoekt, which uses the std regexp/syntax.
	"testing"

	"github.com/grafana/regexp"
)

func TestLongestLiteral(t *testing.T) {
	cases := map[string]string{
		"foo":       "foo",
		"FoO":       "FoO",
		"(?m:^foo)": "foo",
		"(?m:^FoO)": "FoO",
		"[Z]":       "Z",

		`\wddSuballocation\(dump`:    "ddSuballocation(dump",
		`\wfoo(\dlongest\wbam)\dbar`: "longest",

		`(foo\dlongest\dbar)`:  "longest",
		`(foo\dlongest\dbar)+`: "longest",
		`(foo\dlongest\dbar)*`: "",

		"(foo|bar)":     "",
		"[A-Z]":         "",
		"[^A-Z]":        "",
		"[abB-Z]":       "",
		"([abB-Z]|FoO)": "",
		`[@-\[]`:        "",
		`\S`:            "",
	}

	metaLiteral := "AddSuballocation(dump->guid(), system_allocator_name)"
	cases[regexp.QuoteMeta(metaLiteral)] = metaLiteral

	for expr, want := range cases {
		re, err := syntax.Parse(expr, syntax.Perl)
		if err != nil {
			t.Fatal(expr, err)
		}
		re = re.Simplify()
		got := longestLiteral(re)
		if want != got {
			t.Errorf("longestLiteral(%q) == %q != %q", expr, got, want)
		}
	}
}
