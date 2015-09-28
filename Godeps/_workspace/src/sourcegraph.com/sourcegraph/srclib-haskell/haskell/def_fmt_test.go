package haskell

import (
	"testing"

	"sourcegraph.com/sourcegraph/srclib/graph"
)

func TestFormatter(t *testing.T) {
	def := &graph.Def{
		DefKey: graph.DefKey{Repo: "x.com/r", UnitType: "HaskellPackage", Unit: "u", Path: "p/q"},
	}

	if got, want := def.Fmt().Name(graph.DepQualified), "p::q"; got != want {
		t.Errorf("got Name(DepQualified) == %q, want %q", got, want)
	}
}
