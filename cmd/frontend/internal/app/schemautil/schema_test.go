package schemautil

import (
	"reflect"
	"testing"

	"github.com/google/go-querystring/query"
)

func TestSchemaMatchesExceptListAndSortOptions(t *testing.T) {
	type schemaWithNoListOptions struct{ Foo string }

	tests := []struct {
		a, b      interface{}
		wantMatch bool
	}{
		{
			schemaWithNoListOptions{Foo: "x"},
			schemaWithNoListOptions{Foo: "x"},
			true,
		},
		{
			schemaWithNoListOptions{Foo: "x"},
			schemaWithNoListOptions{Foo: "y"},
			false,
		},
	}
	for _, test := range tests {
		beforeA, _ := query.Values(test.a)
		beforeB, _ := query.Values(test.b)

		match := SchemaMatchesExceptListAndSortOptions(test.a, test.b)
		if match != test.wantMatch {
			t.Errorf("got match == %v (want %v) for schemas %+v and %+v", match, test.wantMatch, test.a, test.b)
			continue
		}

		afterA, _ := query.Values(test.a)
		afterB, _ := query.Values(test.b)
		if !reflect.DeepEqual(beforeA, afterA) {
			t.Errorf("a was modified: before %v, after %v", beforeA, afterA)
		}
		if !reflect.DeepEqual(beforeB, afterB) {
			t.Errorf("b was modified: before %v, after %v", beforeB, afterB)
		}
	}
}
