package command_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/command"
)

func TestFlatten(t *testing.T) {
	actual := command.Flatten(
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
	actual := command.Intersperse("-e", []string{
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
