package command

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFlatten(t *testing.T) {
	actual := flatten(
		"foo",
		[]string{"bar", "baz"},
		[]string{"bonk", "quux"},
	)

	expected := []string{
		"foo",
		"bar", "baz",
		"bonk", "quux",
	}
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("unexpected slice (-want +got):\n%s", diff)
	}
}

func TestIntersperse(t *testing.T) {
	actual := intersperse("-e", []string{
		"A=B",
		"C=D",
		"E=F",
	})

	expected := []string{
		"-e", "A=B",
		"-e", "C=D",
		"-e", "E=F",
	}
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("unexpected slice (-want +got):\n%s", diff)
	}
}
