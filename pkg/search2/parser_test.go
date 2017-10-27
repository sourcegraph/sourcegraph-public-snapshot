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

		"a":         {tokens: Tokens{{Value: "a"}}},
		"abc":       {tokens: Tokens{{Value: "abc"}}},
		`"a c"`:     {tokens: Tokens{{Value: "a c"}}},
		"a c":       {tokens: Tokens{{Value: "a"}, {Value: "c"}}},
		" a ":       {tokens: Tokens{{Value: "a"}}},
		` "a" `:     {tokens: Tokens{{Value: "a"}}},
		`"a" b "c"`: {tokens: Tokens{{Value: "a"}, {Value: "b"}, {Value: "c"}}},
		`"a\`:       {tokens: Tokens{{Value: "a"}}},
		`"\uzz"`:    {err: true},
		`"\uzz`:     {err: true},

		"f:a":       {tokens: Tokens{{Field: "f", Value: "a"}}},
		"f:":        {tokens: Tokens{{Field: "f", Value: ""}}},
		"f:abc":     {tokens: Tokens{{Field: "f", Value: "abc"}}},
		`f:"a c"`:   {tokens: Tokens{{Field: "f", Value: "a c"}}},
		"f1:a f2:c": {tokens: Tokens{{Field: "f1", Value: "a"}, {Field: "f2", Value: "c"}}},
		` f:"a" `:   {tokens: Tokens{{Field: "f", Value: "a"}}},
		` f:a `:     {tokens: Tokens{{Field: "f", Value: "a"}}},

		`f1:a b f2:c "d"`: {
			tokens: Tokens{
				{Field: "f1", Value: "a"},
				{Value: "b"},
				{Field: "f2", Value: "c"},
				{Value: "d"},
			},
		},

		`"ab\"\\"`:   {tokens: Tokens{{Value: `ab"\`}}},
		`"ab`:        {tokens: Tokens{{Value: "ab"}}},
		`f:"ab\"\\"`: {tokens: Tokens{{Field: "f", Value: `ab"\`}}},
		`f:"ab`:      {tokens: Tokens{{Field: "f", Value: "ab"}}},
		`f:ab"`:      {tokens: Tokens{{Field: "f", Value: `ab"`}}},

		`-`:        {tokens: Tokens{{Field: "-"}}},
		`-""`:      {tokens: Tokens{{Field: "-"}}},
		`-"`:       {tokens: Tokens{{Field: "-"}}},
		`-a`:       {tokens: Tokens{{Field: "-", Value: "a"}}},
		`--a`:      {tokens: Tokens{{Field: "-", Value: "-a"}}},
		`-f:`:      {tokens: Tokens{{Field: "-f", Value: ""}}},
		`-"a b"`:   {tokens: Tokens{{Field: "-", Value: "a b"}}},
		`-f:"a b"`: {tokens: Tokens{{Field: "-f", Value: "a b"}}},
		`-a b -f: g:`: {
			tokens: Tokens{
				{Field: "-", Value: "a"},
				{Value: "b"},
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
				t.Fatalf("got tokens %v, want %v", tokens, test.tokens)
			}
		})
	}
}
