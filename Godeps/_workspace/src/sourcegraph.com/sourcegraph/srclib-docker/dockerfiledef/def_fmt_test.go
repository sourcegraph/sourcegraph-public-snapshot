package dockerfiledef

import (
	"testing"

	"sourcegraph.com/sourcegraph/srclib/graph"
)

func TestFormatter(t *testing.T) {
	def := &graph.Def{
		DefKey: graph.DefKey{Repo: "x.com/r", UnitType: "Dockerfile", Unit: "u", Path: "p/q"},
		Name:   "Dockerfile",
		File:   "foo/Dockerfile",
	}

	if got, want := def.Fmt().Name(graph.DepQualified), "foo Dockerfile"; got != want {
		t.Errorf("got Name(DepQualified) == %q, want %q", got, want)
	}
}
