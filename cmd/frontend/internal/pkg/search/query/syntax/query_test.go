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

func TestQuery_EscapeImpossibleCaretsDollars(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{in: "", want: ""},
		{in: "a", want: "a"},
		{in: "^", want: "^"},
		{in: "^^", want: `^\^`},
		{in: "^^^", want: `^\^\^`},
		{in: "$", want: "$"},
		{in: "$$", want: `\$$`},
		{in: "$$$", want: `\$\$$`},
		{in: "^$", want: "^$"},
		{in: "^^$$", want: `^\^\$$`},
		{in: "^^^$$$", want: `^\^\^\$\$$`},
		{in: "$^", want: `\$\^`},
		{in: "$ ^", want: `\$ \^`},
		{in: "a$", want: "a$"},
		{in: "a$b", want: `a\$b`},
		{in: "^a$", want: "^a$"},
		{in: "a^b", want: `a\^b`},
		{in: "a^$b a^$b", want: `a\^\$b a\^\$b`},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			q, err := Parse(c.in)
			if err != nil {
				t.Fatal(err)
			}
			q2 := q.EscapeImpossibleCaretsDollars()
			q2s := q2.String()
			if q2s != c.want {
				t.Errorf(`new query is "%s", want "%s"`, q2s, c.want)
			}
		})
	}
}
