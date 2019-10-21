package syntax

import (
	"testing"
)

func TestExpr_String(t *testing.T) {
	type fields struct {
		Pos       int
		Not       bool
		Field     string
		Value     string
		ValueType TokenType
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "empty",
			fields: fields{},
			want:   "",
		},
		{
			name: "literal",
			fields: fields{
				Value:     "a",
				ValueType: TokenLiteral,
			},
			want: "a",
		},
		{
			name: "quoted",
			fields: fields{
				Value:     `"a"`,
				ValueType: TokenQuoted,
			},
			want: `"a"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := Expr{
				Pos:       tt.fields.Pos,
				Not:       tt.fields.Not,
				Field:     tt.fields.Field,
				Value:     tt.fields.Value,
				ValueType: tt.fields.ValueType,
			}
			if got := e.String(); got != tt.want {
				t.Errorf("Expr.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQuery_WithErrorsQuoted(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: ""},
		{in: "a", want: "a"},
		{in: "f:foo bar", want: `f:foo bar`},
		{in: "f:foo b(ar", want: `f:foo "b(ar"`},
		{in: "f:foo b(ar b[az", want: `f:foo "b(ar" "b[az"`},
		{name: "invalid regex in field", in: `f:(a`, want: `"f:(a"`},
		{name: "invalid regex in negated field", in: `-f:(a`, want: `"-f:(a"`},
	}
	for _, c := range cases {
		name := c.name
		if name == "" {
			name = c.in
		}
		t.Run(name, func(t *testing.T) {
			q := ParseAllowingErrors(c.in)
			q2 := q.WithErrorsQuoted()
			q2s := q2.String()
			if q2s != c.want {
				t.Errorf(`output is '%s', want '%s'`, q2s, c.want)
			}
		})
	}
}

func TestParseTreeWithQuotedSearchPattern(t *testing.T) {
	tests := []struct {
		input      string
		searchType string
		want       string
	}{
		{"", "literal", ""},
		{" ", "literal", ""},
		{"  ", "literal", ""},
		{`a`, "literal", `"a"`},
		{` a`, "literal", `"a"`},
		{`a `, "literal", `"a"`},
		{` a `, "literal", `"a"`},
		{` a b`, "literal", `"a b"`},
		{`a  b`, "literal", `"a  b"`},
		{"a\tb", "literal", "\"a\tb\""},
		{`:`, "literal", `":"`},
		{`:=`, "literal", `":="`},
		{`:= range`, "literal", `":= range"`},
		{"`", "literal", "\"`\""},
		{`'`, "literal", `"'"`},
		{`:`, "literal", `":"`},
		{"f:a", "literal", "f:a"},
		{`"f:a"`, "literal", `"\"f:a\""`},
		{"r:b r:c", "literal", "r:b r:c"},
		{"r:b -r:c", "literal", "r:b -r:c"},
		{"patterntype:regexp", "literal", ""},
		{"patterntype:literal", "literal", ""},
		{"patterntype:literal", "regexp", ""},
		{"patterntype:regexp patterntype:literal .*", "literal", `".*"`},
		{"patterntype:regexp patterntype:literal .*", "literal", `".*"`},
		{"patterntype:regexp patterntype:literal .*", "regexp", `".*"`},
		{`patterntype:regexp "patterntype:literal"`, "literal", `"patterntype:literal"`},
		{`patterntype:regexp "patterntype:regexp"`, "literal", `"patterntype:regexp"`},
		{`patterntype:literal "patterntype:regexp"`, "literal", `"\"patterntype:regexp\""`},
		{"patterntype:regexp .*", "literal", ".*"},
		{"patterntype:regexp .* ", "literal", ".*"},
		{"patterntype:regexp .* .*", "literal", ".* .*"},
		{"patterntype:regexp .*  .*", "literal", ".*  .*"},
		{"patterntype:regexp .*\t.*", "literal", ".*\t.*"},
		{".* patterntype:regexp .*", "literal", ".*  .*"},
		{".* patterntype:regexp", "literal", ".*"},
		{"patterntype:regexp .*", "literal", ".*"},
		{"patterntype:regexp .* ", "literal", ".*"},
		{"patterntype:regexp .* .*", "literal", ".* .*"},
		{"patterntype:regexp .*  .*", "literal", ".*  .*"},
		{"patterntype:regexp .*\t.*", "literal", ".*\t.*"},
		{".* patterntype:regexp .*", "literal", ".*  .*"},
		{".* patterntype:regexp", "literal", ".*"},
		{"patterntype:literal .*", "literal", `".*"`},
		{"patterntype:literal .*", "regexp", `".*"`},
		{`lang:go func main`, "literal", `lang:go "func main"`},
		{`lang:go func  main`, "literal", `lang:go "func  main"`},
		{`func main lang:go`, "literal", `lang:go "func main"`},
		{`func  main lang:go`, "literal", `lang:go "func  main"`},
		{`func lang:go main`, "literal", `lang:go "func  main"`},
		// Searching for \n in literal mode brings back literal matches for backslash followed by n.
		{`\n`, "literal", `"\\n"`},
		{`\t`, "literal", `"\\t"`},
		{`\`, "literal", `"\\"`},
		{`foo\d "bar*"`, "literal", `"foo\\d \"bar*\""`},
		{`\d`, "literal", `"\\d"`},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			parseTree, _ := Parse(test.input)
			if test.searchType == "literal" {
				parseTree = parseTree.WithQuotedSearchPattern()
			}
			outValue := parseTree.Values("")
			var out string
			if len(outValue) == 0 {
				out = ""
			} else {
				out = outValue[0]
			}
			if out != test.want {
				t.Errorf("input %q with searchType %q = %q. want %q", test.input, test.searchType, out, test.want)
			}
		})
	}
}
