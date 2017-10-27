package search2

import (
	"reflect"
	"testing"
)

func TestTokens_Extract(t *testing.T) {
	tests := map[string]struct {
		tokens        Tokens
		fieldAliases  map[Field][]Field
		fieldValues   map[Field][]string
		unknownFields []Field
	}{
		"simple": {
			tokens: Tokens{
				{Value: "a"},
				{Field: "x", Value: "b"},
				{Field: "xx", Value: "c"},
				{Field: "y", Value: "d"},
			},
			fieldAliases:  map[Field][]Field{"": {}, "xx": {"x"}},
			fieldValues:   map[Field][]string{"": {"a"}, "xx": {"b", "c"}},
			unknownFields: []Field{"y"},
		},
		"minus": {
			tokens: Tokens{
				{Value: "a"},
				{Field: "-", Value: "b"},
				{Field: "x", Value: "c"},
				{Field: "-x", Value: "d"},
				{Field: "xx", Value: "e"},
				{Field: "-xx", Value: "f"},
				{Field: "y", Value: "g"},
				{Field: "-y", Value: "h"},
			},
			fieldAliases:  map[Field][]Field{"": {}, "-": {}, "xx": {"x"}, "-xx": {"-x"}},
			fieldValues:   map[Field][]string{"": {"a"}, "-": {"b"}, "xx": {"c", "e"}, "-xx": {"d", "f"}},
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
