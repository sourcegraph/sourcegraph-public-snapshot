package datastructures

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDefaultIDSetMap(t *testing.T) {
	m := DefaultIDSetMap{}
	m.GetOrCreate(50).Add(51)
	m.GetOrCreate(50).Add(52)
	m.GetOrCreate(51).Add(53)
	m.GetOrCreate(51).Add(54)

	keys := []int{}
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	if diff := cmp.Diff([]int{50, 51}, keys); diff != "" {
		t.Errorf("unexpected keys (-want +got):\n%s", diff)
	}

	if diff := cmp.Diff([]int{51, 52}, m[50].Identifiers()); diff != "" {
		t.Errorf("unexpected keys (-want +got):\n%s", diff)
	}

	if diff := cmp.Diff([]int{53, 54}, m[51].Identifiers()); diff != "" {
		t.Errorf("unexpected keys (-want +got):\n%s", diff)
	}
}
