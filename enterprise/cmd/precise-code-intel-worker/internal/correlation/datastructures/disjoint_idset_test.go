package datastructures

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDisjointIDSet(t *testing.T) {
	s := DisjointIDSet{}
	s.Union(1, 2)
	s.Union(3, 4)
	s.Union(1, 3)
	s.Union(5, 6)

	setA := []int{1, 2, 3, 4}
	setB := []int{5, 6}

	for _, i := range setA {
		if diff := cmp.Diff(setA, s.ExtractSet(i).Identifiers()); diff != "" {
			t.Errorf("unexpected keys (-want +got):\n%s", diff)
		}
	}

	for _, i := range setB {
		if diff := cmp.Diff(setB, s.ExtractSet(i).Identifiers()); diff != "" {
			t.Errorf("unexpected keys (-want +got):\n%s", diff)
		}
	}
}
