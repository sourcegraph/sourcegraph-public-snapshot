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

func TestIsReal(t *testing.T) {
	dbid := dashboardID{
		IdType: "custom",
		Arg:    5,
	}
	if !dbid.isReal() {
		t.Fatal()
	}
}

func TestIsUserDashboard(t *testing.T) {
	dbid := dashboardID{
		IdType: "user",
		Arg:    5,
	}
	if !dbid.isUser() {
		t.Fatal()
	}
}

func TestIsOrgDashboard(t *testing.T) {
	dbid := dashboardID{
		IdType: "organization",
		Arg:    5,
	}
	if !dbid.isOrg() {
		t.Fatal()
	}
}

func TestIsVirtualDashboard(t *testing.T) {
	dbid := dashboardID{
		IdType: "user",
		Arg:    5,
	}
	if !dbid.isVirtualized() {
		t.Fatal()
	}
}
