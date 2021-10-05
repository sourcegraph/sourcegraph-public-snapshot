package resolvers

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/graph-gophers/graphql-go/relay"
)

func TestMarshalUnmarshalID(t *testing.T) {
	dbid := dashboardID{
		IdType: "test",
		Arg:    1,
	}

	out := relay.MarshalID("dashboard", dbid)

	var final dashboardID

	err := relay.UnmarshalSpec(out, &final)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(dbid, final); diff != "" {
		t.Errorf("mismatch %v", diff)
	}
}
