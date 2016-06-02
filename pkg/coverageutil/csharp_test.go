package coverageutil

import (
	"testing"
)

func TestCsharp(testing *testing.T) {
	testTokenizer(testing,
		&csharpTokenizer{},
		[]*test{
			{
				"multiline strings",
				"\"abc\ndef\"a 'a'",
				[]Token{{9, "a"}},
			},
			{
				"identifiers",
				"_a = 2;",
				[]Token{{0, "_a"}},
			},
			{
				"numeric suffixes",
				"1L 2uL 0.1E10F 0.1E-20d .1E30M",
				[]Token{},
			},
			{
				"verbatim strings",
				"@\"a\"\" b\" c",
				[]Token{{9, "c"}},
			},
		})
}
