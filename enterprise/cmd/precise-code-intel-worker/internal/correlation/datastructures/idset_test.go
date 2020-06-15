package datastructures

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestIDSet(t *testing.T) {
	ids1 := IDSet{}
	ids1.Add("bonk")
	ids1.Add("quux")

	ids2 := IDSet{}
	ids2.Add("foo")
	ids2.Add("bar")
	ids2.Add("baz")
	ids2.AddAll(ids1)

	expected := []string{"bar", "baz", "bonk", "foo", "quux"}
	if diff := cmp.Diff(expected, ids2.Keys()); diff != "" {
		t.Errorf("unexpected keys (-want +got):\n%s", diff)
	}

	if value, ok := ids2.Choose(); !ok {
		t.Errorf("expected a value to be chosen")
	} else {
		if value != "bar" {
			t.Errorf("unexpected chosen value. want=%s have=%s", "bar", value)
		}
	}

	ids2.Add("alpha")
	if value, ok := ids2.Choose(); !ok {
		t.Errorf("expected a value to be chosen")
	} else {
		if value != "alpha" {
			t.Errorf("unexpected chosen value. want=%s have=%s", "alpha", value)
		}
	}
}

func TestChooseEmptyIDSet(t *testing.T) {
	ids := IDSet{}
	if _, ok := ids.Choose(); ok {
		t.Errorf("unexpected ok")
	}
}
