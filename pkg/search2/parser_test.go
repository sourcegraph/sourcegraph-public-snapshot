package search2

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := map[string]struct {
		tokens Tokens
		err    bool
	}{
		"": {tokens: Tokens{}},

		"a":         {tokens: Tokens{{Value: Value{Value: "a"}}}},
		"abc":       {tokens: Tokens{{Value: Value{Value: "abc"}}}},
		`"a c"`:     {tokens: Tokens{{Value: Value{Value: "a c", Quoted: true}}}},
		"a c":       {tokens: Tokens{{Value: Value{Value: "a"}}, {Value: Value{Value: "c"}}}},
		" a ":       {tokens: Tokens{{Value: Value{Value: "a"}}}},
		` "a" `:     {tokens: Tokens{{Value: Value{Value: "a", Quoted: true}}}},
		"(a:b)":     {tokens: Tokens{{Value: Value{Value: "(a:b)"}}}},
		`"a" b "c"`: {tokens: Tokens{{Value: Value{Value: "a", Quoted: true}}, {Value: Value{Value: "b"}}, {Value: Value{Value: "c", Quoted: true}}}},
		`"a\`:       {tokens: Tokens{{Value: Value{Value: `a`, Quoted: true}}}},
		`"\uzz"`:    {err: true},
		`"\uzz`:     {err: true},

		"f:a":       {tokens: Tokens{{Field: "f", Value: Value{Value: "a"}}}},
		"f:a:b":     {tokens: Tokens{{Field: "f", Value: Value{Value: "a:b"}}}},
		"f:":        {tokens: Tokens{{Field: "f", Value: Value{Value: ""}}}},
		"f:abc":     {tokens: Tokens{{Field: "f", Value: Value{Value: "abc"}}}},
		`f:"a c"`:   {tokens: Tokens{{Field: "f", Value: Value{Value: "a c", Quoted: true}}}},
		"f1:a f2:c": {tokens: Tokens{{Field: "f1", Value: Value{Value: "a"}}, {Field: "f2", Value: Value{Value: "c"}}}},
		` f:"a" `:   {tokens: Tokens{{Field: "f", Value: Value{Value: "a", Quoted: true}}}},
		` f:a `:     {tokens: Tokens{{Field: "f", Value: Value{Value: "a"}}}},

		`f1:a b f2:c "d"`: {
			tokens: Tokens{
				{Field: "f1", Value: Value{Value: "a"}},
				{Value: Value{Value: "b"}},
				{Field: "f2", Value: Value{Value: "c"}},
				{Value: Value{Value: "d", Quoted: true}},
			},
		},

		`"ab\"\\"`:   {tokens: Tokens{{Value: Value{Value: `ab"\`, Quoted: true}}}},
		`"ab`:        {tokens: Tokens{{Value: Value{Value: `ab`, Quoted: true}}}},
		`f:"ab\"\\"`: {tokens: Tokens{{Field: "f", Value: Value{Value: `ab"\`, Quoted: true}}}},
		`f:"ab`:      {tokens: Tokens{{Field: "f", Value: Value{Value: "ab", Quoted: true}}}},
		`f:ab"`:      {tokens: Tokens{{Field: "f", Value: Value{Value: `ab"`}}}},

		`-`:        {tokens: Tokens{{Field: "-"}}},
		`-""`:      {tokens: Tokens{{Field: "-", Value: Value{Quoted: true}}}},
		`-"`:       {tokens: Tokens{{Field: "-", Value: Value{Quoted: true}}}},
		`-a`:       {tokens: Tokens{{Field: "-", Value: Value{Value: "a"}}}},
		`--a`:      {tokens: Tokens{{Field: "-", Value: Value{Value: "-a"}}}},
		`-f:`:      {tokens: Tokens{{Field: "-f", Value: Value{Value: ""}}}},
		`-"a b"`:   {tokens: Tokens{{Field: "-", Value: Value{Value: "a b", Quoted: true}}}},
		`-f:"a b"`: {tokens: Tokens{{Field: "-f", Value: Value{Value: "a b", Quoted: true}}}},
		`-a b -f: g:`: {
			tokens: Tokens{
				{Field: "-", Value: Value{Value: "a"}},
				{Value: Value{Value: "b"}},
				{Field: "-f"},
				{Field: "g"},
			},
		},
	}
	for query, test := range tests {
		t.Run(query, func(t *testing.T) {
			tokens, err := Parse(query)
			if (err != nil) != (test.err) {
				t.Fatalf("got error %v, want %v", err, test.err)
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(tokens, test.tokens) {
				t.Fatalf("got tokens %#v, want %#v", tokens, test.tokens)
			}
		})
	}
}
