package search2

import (
	"reflect"
	"testing"
)

func TestTokens_String(t *testing.T) {
	tests := map[string]Tokens{
		"foo":     {{Value: Value{Value: "foo"}}},
		`"foo"`:   {{Value: Value{Value: "foo", Quoted: true}}},
		"-foo":    {{Field: "-", Value: Value{Value: "foo"}}},
		`-"foo"`:  {{Field: "-", Value: Value{Value: "foo", Quoted: true}}},
		`foo bar`: {{Value: Value{Value: "foo"}}, {Value: Value{Value: "bar"}}},
	}
	for want, tokens := range tests {
		t.Run(want, func(t *testing.T) {
			text := tokens.String()
			if text != want {
				t.Errorf("got %q, want %q", text, want)
			}
		})
	}
}

func TestTokens_Normalize(t *testing.T) {
	tokens := Tokens{
		{Value: Value{Value: "a", Quoted: true}},
		{Field: "x", Value: Value{Value: "b"}},
		{Field: "xx", Value: Value{Value: "c"}},
		{Field: "y", Value: Value{Value: "d"}},
	}
	want := Tokens{
		{Value: Value{Value: "a", Quoted: true}},
		{Field: "xx", Value: Value{Value: "b"}},
		{Field: "xx", Value: Value{Value: "c"}},
		{Field: "y", Value: Value{Value: "d"}},
	}
	tokens.Normalize(map[Field][]Field{"xx": {"x"}})
	if !reflect.DeepEqual(tokens, want) {
		t.Errorf("got %q, want %q", tokens, want)
	}
}

func TestTokens_Extract(t *testing.T) {
	tests := map[string]struct {
		tokens        Tokens
		fieldAliases  map[Field][]Field
		fieldValues   map[Field]Values
		normTokens    []Token
		unknownFields []Field
	}{
		"simple": {
			tokens: Tokens{
				{Value: Value{Value: "a", Quoted: true}},
				{Field: "xx", Value: Value{Value: "b"}},
				{Field: "xx", Value: Value{Value: "c"}},
				{Field: "y", Value: Value{Value: "d"}},
			},
			fieldValues: map[Field]Values{
				"":   {{Value: "a", Quoted: true}},
				"xx": {{Value: "b"}, {Value: "c"}},
				"y":  {{Value: "d"}},
			},
		},
		"minus": {
			tokens: Tokens{
				{Value: Value{Value: "a"}},
				{Field: "-", Value: Value{Value: "b"}},
				{Field: "xx", Value: Value{Value: "c"}},
				{Field: "-xx", Value: Value{Value: "d"}},
				{Field: "xx", Value: Value{Value: "e"}},
				{Field: "-xx", Value: Value{Value: "f"}},
				{Field: "y", Value: Value{Value: "g"}},
				{Field: "-y", Value: Value{Value: "h"}},
			},
			fieldValues: map[Field]Values{
				"":    {{Value: "a"}},
				"-":   {{Value: "b"}},
				"xx":  {{Value: "c"}, {Value: "e"}},
				"-xx": {{Value: "d"}, {Value: "f"}},
				"y":   {{Value: "g"}},
				"-y":  {{Value: "h"}},
			},
		},
	}
	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			fieldValues := test.tokens.Extract()
			if !reflect.DeepEqual(fieldValues, test.fieldValues) {
				t.Errorf("got fieldValues %q, want %q", fieldValues, test.fieldValues)
			}
		})
	}
}
