package query

import (
	"reflect"
	"testing"
)

func TestConvertToLiteral(t *testing.T) {
	tests := []struct {
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
		{` a b`, `"a b"`},
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
		{".*", `".*"`},
		{`a:b "patterntype:regexp"`, `a:b "\"patterntype:regexp\""`},
		{`lang:go func main`, `lang:go "func main"`},
		{`lang:go func  main`, `lang:go "func  main"`},
		{`func main lang:go`, `lang:go "func main"`},
		{`func  main lang:go`, `lang:go "func  main"`},
		{`func lang:go main`, `lang:go "func  main"`},
		// Searching for \n in literal mode brings back literal matches for backslash followed by n.
		{`\n`, `"\\n"`},
		{`\t`, `"\\t"`},
		{`\`, `"\\"`},
		{`foo\d "bar*"`, `"foo\\d \"bar*\""`},
		{`\d`, `"\\d"`},
		{`type:commit message:"a commit message" after:"10 days ago"`, `message:"a commit message" after:"10 days ago" type:commit`},
		{`type:commit message:"a commit message" after:"10 days ago" test`, `message:"a commit message" after:"10 days ago" type:commit "test"`},
		{`type:commit message:"a commit message" after:"10 days ago" test test2`, `message:"a commit message" after:"10 days ago" type:commit "test test2"`},
		{`type:commit message:"a commit message" test after:"10 days ago" test2`, `message:"a commit message" after:"10 days ago" type:commit "test  test2"`},
		{`type:commit message:'a commit message' test after:'10 days ago' test2`, `message:'a commit message' after:'10 days ago' type:commit "test  test2"`},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			out := ConvertToLiteral(test.input)
			if out != test.want {
				t.Errorf("ConvertToLiteral (%q) = %q, want %q", test.input, out, test.want)
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
		{`type:commit message:"a commit message"`, []string{"message:\"a commit message\"", "type:commit", " "}},
		{`type:commit message:"a commit message" after:"10 days ago"`, []string{"message:\"a commit message\"", "after:\"10 days ago\"", "type:commit", "  "}},
		{`type:commit message:"a commit message" test after:"10 days ago"`, []string{"message:\"a commit message\"", "after:\"10 days ago\"", "type:commit", "  ", "test", " "}},
		{`type:commit message:"a commit message" after:"10 days ago" test`, []string{"message:\"a commit message\"", "after:\"10 days ago\"", "type:commit", "   ", "test"}},
		{`type:commit message:"a commit message" after:"10 days ago" test test2`, []string{"message:\"a commit message\"", "after:\"10 days ago\"", "type:commit", "   ", "test", " ", "test2"}},
		{`type:commit message:'a commit message' after:'10 days ago' test test2`, []string{"message:'a commit message'", "after:'10 days ago'", "type:commit", "   ", "test", " ", "test2"}},
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
