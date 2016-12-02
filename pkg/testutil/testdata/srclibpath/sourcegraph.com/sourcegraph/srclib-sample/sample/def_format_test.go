package sample

import (
	"testing"

	"sourcegraph.com/sourcegraph/srclib/graph"
)

func TestFormatter(t *testing.T) {
	def := &graph.Def{
		DefKey: graph.DefKey{Repo: "x.com/r", UnitType: "sample", Unit: "u", Path: "p"},
	}

	if got, want := def.Fmt().Name(graph.DepQualified), "imp.scope.name"; got != want {
		t.Errorf("got Name(DepQualified) == %q, want %q", got, want)
	}
}
