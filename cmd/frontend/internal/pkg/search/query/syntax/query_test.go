package syntax

import (
	"reflect"
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

func TestQuery_WithPartsQuoted(t *testing.T) {
	type fields struct {
		Input string
		Expr  []*Expr
	}
	tests := []struct {
		name   string
		fields fields
		want   *Query
	}{
		{
			name: "empty",
			fields: fields{
				Input: "",
				Expr:  nil,
			},
			want: &Query{
				Input: "",
				Expr:  nil,
			},
		},
		{
			name: "one field",
			fields: fields{
				Input: "",
				Expr: []*Expr{
					{
						Field: "f",
						Value: "a",
					},
				},
			},
			want: &Query{
				Input: "",
				Expr: []*Expr{
					{
						Value:     `"f:a"`,
						ValueType: TokenQuoted,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &Query{
				Input: tt.fields.Input,
				Expr:  tt.fields.Expr,
			}
			if got := q.WithPartsQuoted(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Query.WithPartsQuoted() = %v, want %v", got, tt.want)
			}
		})
	}
}
