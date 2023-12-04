package resolvers

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/insights/store"

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

func TestHasPermissionForGrants(t *testing.T) {
	t.Run("No grants returns true", func(t *testing.T) {
		got := hasPermissionForGrants([]store.DashboardGrant{}, []int{1, 2}, []int{1})
		if got != true {
			t.Errorf("should have permission for zero grants")
		}
	})
	t.Run("Returns false for unmatched user grant", func(t *testing.T) {
		userId := 3
		got := hasPermissionForGrants([]store.DashboardGrant{{UserID: &userId}}, []int{1, 2}, []int{1})
		if got != false {
			t.Errorf("should return false for no user permission")
		}
	})
	t.Run("Returns false for unmatched org grant", func(t *testing.T) {
		orgId := 3
		got := hasPermissionForGrants([]store.DashboardGrant{{OrgID: &orgId}}, []int{1, 2}, []int{1})
		if got != false {
			t.Errorf("should return false for no org permission")
		}
	})
	t.Run("Returns true for matched user permission", func(t *testing.T) {
		userId := 2
		got := hasPermissionForGrants([]store.DashboardGrant{{UserID: &userId}}, []int{1, 2}, []int{1})
		if got != true {
			t.Errorf("should return true for matched user permission")
		}
	})
	t.Run("Returns true for matched org permission", func(t *testing.T) {
		orgId := 1
		got := hasPermissionForGrants([]store.DashboardGrant{{OrgID: &orgId}}, []int{1, 2}, []int{1})
		if got != true {
			t.Errorf("should return true for matched org permission")
		}
	})
	t.Run("Returns true for matched user and org permissions", func(t *testing.T) {
		userId := 1
		userId2 := 5
		orgId := 1
		global := true
		got := hasPermissionForGrants([]store.DashboardGrant{{UserID: &userId}, {UserID: &userId2}, {OrgID: &orgId}, {Global: &global}}, []int{5, 1, 2}, []int{1, 3})
		if got != true {
			t.Errorf("should return true for matched user and org permission")
		}
	})
}
