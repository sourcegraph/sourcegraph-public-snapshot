package search2

import (
	"reflect"
	"testing"
)

func TestTokens_Extract(t *testing.T) {
	tests := map[string]struct {
		tokens        Tokens
		fieldAliases  map[Field][]Field
		fieldValues   map[Field]Values
		unknownFields []Field
	}{
		"simple": {
			tokens: Tokens{
				{Value: Value{Value: "a", Quoted: true}},
				{Field: "x", Value: Value{Value: "b"}},
				{Field: "xx", Value: Value{Value: "c"}},
				{Field: "y", Value: Value{Value: "d"}},
			},
			fieldAliases:  map[Field][]Field{"": {}, "xx": {"x"}},
			fieldValues:   map[Field]Values{"": {{Value: "a", Quoted: true}}, "xx": {{Value: "b"}, {Value: "c"}}},
			unknownFields: []Field{"y"},
		},
		"minus": {
			tokens: Tokens{
				{Value: Value{Value: "a"}},
				{Field: "-", Value: Value{Value: "b"}},
				{Field: "x", Value: Value{Value: "c"}},
				{Field: "-x", Value: Value{Value: "d"}},
				{Field: "xx", Value: Value{Value: "e"}},
				{Field: "-xx", Value: Value{Value: "f"}},
				{Field: "y", Value: Value{Value: "g"}},
				{Field: "-y", Value: Value{Value: "h"}},
			},
			fieldAliases:  map[Field][]Field{"": {}, "-": {}, "xx": {"x"}, "-xx": {"-x"}},
			fieldValues:   map[Field]Values{"": {{Value: "a"}}, "-": {{Value: "b"}}, "xx": {{Value: "c"}, {Value: "e"}}, "-xx": {{Value: "d"}, {Value: "f"}}},
			unknownFields: []Field{"y", "-y"},
		},
	}
	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			fieldValues, unknownFields := test.tokens.Extract(test.fieldAliases)
			if !reflect.DeepEqual(fieldValues, test.fieldValues) {
				t.Errorf("got fieldValues %q, want %q", fieldValues, test.fieldValues)
			}
			if !reflect.DeepEqual(unknownFields, test.unknownFields) {
				t.Errorf("got unknownFields %q, want %q", unknownFields, test.unknownFields)
			}
		})
	}
}
