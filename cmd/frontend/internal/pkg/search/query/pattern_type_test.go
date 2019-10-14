package query

import (
	"reflect"
	"testing"
)

func TestHandlePatternType_Literal(t *testing.T) {
	tests := []struct {
		input           string
		defaultToRegexp bool
		want            string
	}{
		{"", false, ""},
		{" ", false, ""},
		{"  ", false, ""},
		{`a`, false, `"a"`},
		{` a`, false, `"a"`},
		{`a `, false, `"a"`},
		{` a `, false, `"a"`},
		{` a b`, false, `"a b"`},
		{`a  b`, false, `"a  b"`},
		{"a\tb", false, "\"a\tb\""},
		{`:`, false, `":"`},
		{`:=`, false, `":="`},
		{`:= range`, false, `":= range"`},
		{"`", false, "\"`\""},
		{`'`, false, `"'"`},
		{`:`, false, `":"`},
		{"f:a", false, "f:a"},
		{`"f:a"`, false, `"\"f:a\""`},
		{"r:b r:c", false, "r:b r:c"},
		{"r:b -r:c", false, "r:b -r:c"},
		{"patterntype:regex", false, ""},
		{"patterntype:regexp", false, ""},
		{"patterntype:literal", false, ""},
		{"patterntype:literal", true, ""},
		{"patterntype:regexp patterntype:literal .*", false, `".*"`},
		{"patterntype:regexp patterntype:literal .*", false, `".*"`},
		{"patterntype:regexp patterntype:literal .*", true, `".*"`},
		{`patterntype:regexp "patterntype:literal"`, false, `"patterntype:literal"`},
		{`patterntype:regexp "patterntype:regexp"`, false, `"patterntype:regexp"`},
		{`patterntype:literal "patterntype:regexp"`, false, `"\"patterntype:regexp\""`},
		{"patterntype:regexp .*", false, ".*"},
		{"patterntype:regexp .* ", false, ".*"},
		{"patterntype:regexp .* .*", false, ".* .*"},
		{"patterntype:regexp .*  .*", false, ".*  .*"},
		{"patterntype:regexp .*\t.*", false, ".*\t.*"},
		{".* patterntype:regexp .*", false, ".*  .*"},
		{".* patterntype:regexp", false, ".*"},
		{"patterntype:regexp .*", false, ".*"},
		{"patterntype:regexp .* ", false, ".*"},
		{"patterntype:regexp .* .*", false, ".* .*"},
		{"patterntype:regexp .*  .*", false, ".*  .*"},
		{"patterntype:regexp .*\t.*", false, ".*\t.*"},
		{".* patterntype:regexp .*", false, ".*  .*"},
		{".* patterntype:regexp", false, ".*"},
		{"patterntype:literal .*", false, `".*"`},
		{"patterntype:literal .*", true, `".*"`},
		{`lang:go func main`, false, `lang:go "func main"`},
		{`lang:go func  main`, false, `lang:go "func  main"`},
		{`func main lang:go`, false, `lang:go "func main"`},
		{`func  main lang:go`, false, `lang:go "func  main"`},
		{`func lang:go main`, false, `lang:go "func  main"`},
		// Searching for \n in literal mode brings back literal matches for backslash followed by n.
		{`\n`, false, `"\\n"`},
		{`\t`, false, `"\\t"`},
		{`\`, false, `"\\"`},
		{`foo\d "bar*"`, false, `"foo\\d \"bar*\""`},
		{`\d`, false, `"\\d"`},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			out := HandlePatternType(test.input, test.defaultToRegexp)
			if out != test.want {
				t.Errorf("handlePatternType (%q), with defaultToRegexp %t = %q, want %q", test.input, test.defaultToRegexp, out, test.want)
			}
		})
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
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

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			toks := tokenize(test.input)
			if !reflect.DeepEqual(toks, test.want) {
				t.Errorf("tokenize(`%s`) = %s, want `%s`", test.input, toks, test.want)
			}
		})
	}
}
