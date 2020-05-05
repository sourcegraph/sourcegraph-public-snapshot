package datastructures

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDefaultIDSetMap(t *testing.T) {
	m := DefaultIDSetMap{}
	m.GetOrCreate("foo").Add("bar")
	m.GetOrCreate("foo").Add("baz")
	m.GetOrCreate("bar").Add("bonk")
	m.GetOrCreate("bar").Add("quux")

	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	expected := []string{"bar", "foo"}
	if diff := cmp.Diff(expected, keys); diff != "" {
		t.Errorf("unexpected keys (-want +got):\n%s", diff)
	}

	expected = []string{"bar", "baz"}
	if diff := cmp.Diff(expected, m["foo"].Keys()); diff != "" {
		t.Errorf("unexpected keys (-want +got):\n%s", diff)
	}

	expected = []string{"bonk", "quux"}
	if diff := cmp.Diff(expected, m["bar"].Keys()); diff != "" {
		t.Errorf("unexpected keys (-want +got):\n%s", diff)
	}
}
