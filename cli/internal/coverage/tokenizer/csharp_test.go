package tokenizer

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
				[]Token{{9, 2, 6, "a"}},
			},
			{
				"identifiers",
				"_a = 2;",
				[]Token{{0, 1, 3, "_a"}},
			},
			{
				"numeric suffixes",
				"1L 2uL 0.1E10F 0.1E-20d .1E30M",
				[]Token{},
			},
			{
				"verbatim strings",
				"@\"a\"\" b\" c",
				[]Token{{9, 1, 11, "c"}},
			},
			{
				"preprocessor directives",
				"#ifdef A\nfoo\n#endif\n#region License Information (GPL v3)",
				[]Token{{9, 2, 4, "foo"}},
			},
		})
}
