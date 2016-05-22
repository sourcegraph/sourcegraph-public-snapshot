package schemautil

import (
	"reflect"
	"testing"

	"github.com/google/go-querystring/query"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

func TestSchemaMatches(t *testing.T) {
	tests := []struct {
		a, b      interface{}
		wantMatch bool
	}{
		{
			sourcegraph.BuildListOptions{Queued: true},
			sourcegraph.BuildListOptions{Queued: false},
			false,
		},
		{
			sourcegraph.BuildListOptions{Queued: true},
			sourcegraph.BuildListOptions{Queued: true},
			true,
		},
		{
			sourcegraph.BuildListOptions{Queued: true, ListOptions: sourcegraph.ListOptions{Page: 5}},
			sourcegraph.BuildListOptions{Queued: true, ListOptions: sourcegraph.ListOptions{Page: 10}},
			false,
		},
	}
	for _, test := range tests {
		match := schemaMatches(test.a, test.b)
		if match != test.wantMatch {
			t.Errorf("got match == %v (want %v) for schemas %+v and %+v", match, test.wantMatch, test.a, test.b)
			continue
		}
	}
}

func TestSchemaMatchesExceptListAndSortOptions(t *testing.T) {
	type schemaWithNoListOptions struct{ Foo string }

	tests := []struct {
		a, b      interface{}
		wantMatch bool
	}{
		{
			sourcegraph.BuildListOptions{Queued: true},
			sourcegraph.BuildListOptions{Queued: false},
			false,
		},
		{
			&sourcegraph.BuildListOptions{Queued: true},
			&sourcegraph.BuildListOptions{Queued: true},
			true,
		},
		{
			sourcegraph.BuildListOptions{Queued: true, ListOptions: sourcegraph.ListOptions{Page: 5}},
			sourcegraph.BuildListOptions{Queued: true, ListOptions: sourcegraph.ListOptions{Page: 10}},
			true,
		},
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
