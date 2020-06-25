package query

import (
	"testing"
)

func TestFieldExists(t *testing.T) {
	nodes := []Node{
		Parameter{Field: "repo", Value: "foo"},
		Parameter{Field: "repo", Value: "bar", Negated: true},
		Pattern{Value: "baz"},
	}

	if !FieldExists(nodes, "repo") {
		t.Errorf("Expected field repo to be found")
	}

	if FieldExists(nodes, "baz") {
		t.Errorf("Expected field baz to not be found")
	}
}
