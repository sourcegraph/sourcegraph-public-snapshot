package query

import (
	"reflect"
	"testing"
)

func TestHandlePatternType(t *testing.T) {
	tcs := []struct {
		input string
		want  string
	}{
		{"", ""},
		{" ", ""},
		{"  ", ""},
		{`a`, `"a"`},
		{` a`, `"a"`},
		{`a `, `"a"`},
		{` a `, `"a"`},
		{`a b`, `"a b"`},
		{`a  b`, `"a  b"`},
		{"a\tb", "\"a\tb\""},
		{`:`, `":"`},
		{`:=`, `":="`},
		{`:= range`, `":= range"`},
		{"`", "\"`\""},
		{`'`, `"'"`},
		{`:`, `":"`},
		{"f:a", "f:a"},
		{`"f:a"`, `"\"f:a\""`},
		{"r:b r:c", "r:b r:c"},
		{"r:b -r:c", "r:b -r:c"},
		{"patternType:regex", ""},
		{"patternType:regexp", ""},
		{"patternType:literal", ""},
		{"patternType:regexp patternType:literal .*", `".*"`},
		{"patternType:regexp patternType:literal .*", `".*"`},
		{`patternType:regexp "patternType:literal"`, `"patternType:literal"`},
		{`patternType:regexp "patternType:regexp"`, `"patternType:regexp"`},
		{`patternType:literal "patternType:regexp"`, `"\"patternType:regexp\""`},
		{"patternType:regexp .*", ".*"},
		{"patternType:regexp .* ", ".*"},
		{"patternType:regexp .* .*", ".* .*"},
		{"patternType:regexp .*  .*", ".*  .*"},
		{"patternType:regexp .*\t.*", ".*\t.*"},
		{".* patternType:regexp .*", ".*  .*"},
		{".* patternType:regexp", ".*"},
		{"patternType:literal .*", `".*"`},
		{`lang:go func main`, `lang:go "func main"`},
		{`lang:go func  main`, `lang:go "func  main"`},
		{`func main lang:go`, `lang:go "func main"`},
		{`func  main lang:go`, `lang:go "func  main"`},
		{`func lang:go main`, `lang:go "func  main"`},
		// Searching for \n in literal mode brings back literal matches for backslash followed by n.
		{`\n`, `"\\n"`},
		{`\t`, `"\\t"`},
		// Searching for a backslash should also bring back literal matches for backslashes.
		{`\`, `"\\"`},
	}
	for _, tc := range tcs {
		t.Run(tc.input, func(t *testing.T) {
			out := HandlePatternType(tc.input, false)
			if out != tc.want {
				t.Errorf("handlePatternType(%q) = %q, want %q", tc.input, out, tc.want)
			}
		})
	}
}

func TestTokenize(t *testing.T) {
	tcs := []struct {
		input string
		want  []string
	}{
		{"", nil},
		{" ", []string{" "}},
		{"a", []string{"a"}},
		{" a", []string{" ", "a"}},
		{"a ", []string{"a", " "}},
		{"a b", []string{"a", " ", "b"}},
		{`"`, []string{`"`}},
		{`""`, []string{`""`}},
		{`"""`, []string{`""`, `"`}},
		{`""""`, []string{`""`, `""`}},
		{`"""""`, []string{`""`, `""`, `"`}},
		{`" ""`, []string{`" "`, `"`}},
		{`" """`, []string{`" "`, `""`}},
		{`" "" "`, []string{`" "`, `" "`}},
		{`" " "`, []string{`" "`, " ", `"`}},
		{`" " " "`, []string{`" "`, " ", `" "`}},
		{`"\""`, []string{`"\""`}},
		{`"\""`, []string{`"\""`}},
		{`"\"" "\""`, []string{`"\""`, " ", `"\""`}},
		{`f:a "r:b"`, []string{`f:a`, " ", `"r:b"`}},
		{"//", []string{"//"}},
		{"/**/", []string{"/**/"}},
	}

	for _, tc := range tcs {
		t.Run(tc.input, func(t *testing.T) {
			toks := tokenize(tc.input)
			if !reflect.DeepEqual(toks, tc.want) {
				t.Errorf("tokenize(`%s`) = %s, want `%s`", tc.input, toks, tc.want)
			}
		})
	}
}
