package datastructures

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDisjointIDSet(t *testing.T) {
	s := DisjointIDSet{}
	s.Union("1", "2")
	s.Union("3", "4")
	s.Union("1", "3")
	s.Union("5", "6")

	if diff := cmp.Diff([]string{"1", "2", "3", "4"}, s.ExtractSet("1").Keys()); diff != "" {
		t.Errorf("unexpected keys (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{"1", "2", "3", "4"}, s.ExtractSet("2").Keys()); diff != "" {
		t.Errorf("unexpected keys (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{"1", "2", "3", "4"}, s.ExtractSet("3").Keys()); diff != "" {
		t.Errorf("unexpected keys (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{"1", "2", "3", "4"}, s.ExtractSet("4").Keys()); diff != "" {
		t.Errorf("unexpected keys (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{"5", "6"}, s.ExtractSet("5").Keys()); diff != "" {
		t.Errorf("unexpected keys (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{"5", "6"}, s.ExtractSet("6").Keys()); diff != "" {
		t.Errorf("unexpected keys (-want +got):\n%s", diff)
	}
}
