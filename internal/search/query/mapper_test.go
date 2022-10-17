package query

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMapOperator(t *testing.T) {
	input := []Node{
		Operator{
			Kind: And,
			Operands: []Node{
				Parameter{Field: "repo", Value: "github.com/saucegraph/saucegraph"},
				Pattern{Value: "pasta_sauce"},
			},
		},
	}
	want := input
	got := MapOperator(input, func(kind OperatorKind, operands []Node) []Node {
		return NewOperator(NewOperator(NewOperator(operands, kind), And), Or)
	})
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatal(diff)
	}
}

func TestMapField(t *testing.T) {
	input := Parameter{Field: "before", Value: "today"}
	want := Operator{
		Kind: Or,
		Operands: []Node{
			Parameter{Field: "before", Value: "yesterday"},
			Parameter{Field: "after", Value: "yesterday"},
		},
	}
	got := MapField([]Node{input}, "before", func(_ string, _ bool, _ Annotation) Node {
		return want
	})
	if diff := cmp.Diff(want, got[0]); diff != "" {
		t.Fatal(diff)
	}
}
