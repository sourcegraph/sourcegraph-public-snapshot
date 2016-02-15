package grapher

import (
	"testing"

	"sourcegraph.com/sourcegraph/srclib/graph"
)

func TestValidateDocs_ok(t *testing.T) {
	docs := []*graph.Doc{
		{
			DefKey: graph.DefKey{Path: "p"},
			Format: "f",
			Data:   "d",
		},
		{
			DefKey: graph.DefKey{Path: "p"},
			Format: "f2",
			Data:   "d",
		},
		{
			DefKey: graph.DefKey{Path: "p2"},
			Format: "f",
			Data:   "d",
		},
	}
	if err := ValidateDocs(docs); err != nil {
		t.Fatal(err)
	}
}

func TestValidateDocs_dup(t *testing.T) {
	docs := []*graph.Doc{
		{
			DefKey: graph.DefKey{Path: "p"},
			Format: "f",
			Data:   "d",
		},
		{
			DefKey: graph.DefKey{Path: "p"},
			Format: "f",
			Data:   "d",
		},
	}
	if err := ValidateDocs(docs); err == nil {
		t.Fatalf("got nil err, want validation error")
	}
}
