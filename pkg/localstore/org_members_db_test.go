package localstore

import (
	"fmt"
	"reflect"
	"testing"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

func TestOrgMembers_CreateMembershipInOrgsForAllUsers(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := testContext()

	// Create fixtures.
	org1, err := Orgs.Create(ctx, "org1", "")
	if err != nil {
		t.Fatal(err)
	}
	org2, err := Orgs.Create(ctx, "org2", "")
	if err != nil {
		t.Fatal(err)
	}
	org3, err := Orgs.Create(ctx, "org3", "")
	if err != nil {
		t.Fatal(err)
	}
	user1, err := Users.Create(ctx, "authid1", "a1@example.com", "u1", "", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Users.Create(ctx, "authid2", "a2@example.com", "u2", "", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := OrgMembers.Create(ctx, org1.ID, user1.Auth0ID); err != nil {
		t.Fatal(err)
	}

	check := func() error {
		want := map[string][]string{
			"org1": []string{"authid1", "authid2"},
			"org2": []string{},
			"org3": []string{"authid1", "authid2"},
		}
		got := map[string][]string{}
		for _, org := range []*sourcegraph.Org{org1, org2, org3} {
			members, err := OrgMembers.GetByOrgID(ctx, org.ID)
			if err != nil {
				return err
			}
			if len(members) == 0 {
				got[org.Name] = []string{}
			}
			for _, member := range members {
				got[org.Name] = append(got[org.Name], member.UserID)
			}
		}
		if !reflect.DeepEqual(got, want) {
			return fmt.Errorf("got membership %+v, want %+v", got, want)
		}
		return nil
	}

	// Try twice; it should be idempotent.
	if err := OrgMembers.CreateMembershipInOrgsForAllUsers(ctx, nil, []string{"org1", "org3"}); err != nil {
		t.Fatal(err)
	}
	if err := check(); err != nil {
		t.Fatal(err)
	}
	if err := OrgMembers.CreateMembershipInOrgsForAllUsers(ctx, nil, []string{"org1", "org3"}); err != nil {
		t.Fatal(err)
	}
	if err := check(); err != nil {
		t.Fatal(err)
	}

	// Passing an org that does not exist should not be an error.
	if err := OrgMembers.CreateMembershipInOrgsForAllUsers(ctx, nil, []string{"doesntexist"}); err != nil {
		t.Fatal(err)
	}
	if err := check(); err != nil {
		t.Fatal(err)
	}

	// An empty list shouldn't be an error.
	if err := OrgMembers.CreateMembershipInOrgsForAllUsers(ctx, nil, []string{}); err != nil {
		t.Fatal(err)
	}
	if err := check(); err != nil {
		t.Fatal(err)
	}
}
