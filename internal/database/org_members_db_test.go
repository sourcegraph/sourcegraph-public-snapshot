package database

import (
	"context"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestOrgMembers_CreateMembershipInOrgsForAllUsers(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := dbtest.NewDB(t)
	ctx := context.Background()

	// Create fixtures.
	org1, err := Orgs(db).Create(ctx, "org1", nil)
	if err != nil {
		t.Fatal(err)
	}
	org2, err := Orgs(db).Create(ctx, "org2", nil)
	if err != nil {
		t.Fatal(err)
	}
	org3, err := Orgs(db).Create(ctx, "org3", nil)
	if err != nil {
		t.Fatal(err)
	}
	user1, err := Users(db).Create(ctx, NewUser{
		Email:                 "a1@example.com",
		Username:              "u1",
		Password:              "p",
		EmailVerificationCode: "c",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = Users(db).Create(ctx, NewUser{
		Email:                 "a2@example.com",
		Username:              "u2",
		Password:              "p",
		EmailVerificationCode: "c",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := OrgMembers(db).Create(ctx, org1.ID, user1.ID); err != nil {
		t.Fatal(err)
	}

	check := func() error {
		want := map[string][]int32{
			"org1": {1, 2},
			"org2": {},
			"org3": {1, 2},
		}
		got := map[string][]int32{}
		for _, org := range []*types.Org{org1, org2, org3} {
			members, err := OrgMembers(db).GetByOrgID(ctx, org.ID)
			if err != nil {
				return err
			}
			if len(members) == 0 {
				got[org.Name] = []int32{}
			}
			for _, member := range members {
				got[org.Name] = append(got[org.Name], member.UserID)
			}
		}
		if !reflect.DeepEqual(got, want) {
			return errors.Errorf("got membership %+v, want %+v", got, want)
		}
		return nil
	}

	// Try twice; it should be idempotent.
	if err := OrgMembers(db).CreateMembershipInOrgsForAllUsers(ctx, []string{"org1", "org3"}); err != nil {
		t.Fatal(err)
	}
	if err := check(); err != nil {
		t.Fatal(err)
	}
	if err := OrgMembers(db).CreateMembershipInOrgsForAllUsers(ctx, []string{"org1", "org3"}); err != nil {
		t.Fatal(err)
	}
	if err := check(); err != nil {
		t.Fatal(err)
	}

	// Passing an org that does not exist should not be an error.
	if err := OrgMembers(db).CreateMembershipInOrgsForAllUsers(ctx, []string{"doesntexist"}); err != nil {
		t.Fatal(err)
	}
	if err := check(); err != nil {
		t.Fatal(err)
	}

	// An empty list shouldn't be an error.
	if err := OrgMembers(db).CreateMembershipInOrgsForAllUsers(ctx, []string{}); err != nil {
		t.Fatal(err)
	}
	if err := check(); err != nil {
		t.Fatal(err)
	}
}

func TestOrgMembers_MemberCount(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	db := dbtest.NewDB(t)
	ctx := context.Background()
	// Create fixtures.
	org1, err := Orgs(db).Create(ctx, "org1", nil)
	if err != nil {
		t.Fatal(err)
	}
	org2, err := Orgs(db).Create(ctx, "org2", nil)
	if err != nil {
		t.Fatal(err)
	}
	org3, err := Orgs(db).Create(ctx, "org3", nil)
	if err != nil {
		t.Fatal(err)
	}
	user1, err := Users(db).Create(ctx, NewUser{
		Email:                 "a1@example.com",
		Username:              "u1",
		Password:              "p",
		EmailVerificationCode: "c",
	})
	if err != nil {
		t.Fatal(err)
	}
	user2, err := Users(db).Create(ctx, NewUser{
		Email:                 "a2@example.com",
		Username:              "u2",
		Password:              "p2",
		EmailVerificationCode: "c",
	})
	if err != nil {
		t.Fatal(err)
	}
	deletedUser, err := Users(db).Create(ctx, NewUser{
		Email:                 "deleted@example.com",
		Username:              "deleted",
		Password:              "p2",
		EmailVerificationCode: "c",
	})
	if err != nil {
		t.Fatal(err)
	}
	OrgMembers(db).Create(ctx, org1.ID, user1.ID)
	OrgMembers(db).Create(ctx, org2.ID, user1.ID)
	OrgMembers(db).Create(ctx, org2.ID, user2.ID)
	OrgMembers(db).Create(ctx, org3.ID, user1.ID)
	OrgMembers(db).Create(ctx, org3.ID, deletedUser.ID)
	err = Users(db).Delete(ctx, deletedUser.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name  string
		orgID int32
		want  int
	}{
		{"org with single member", org1.ID, 1},
		{"org with two members", org2.ID, 2},
		{"org with one deleted member", org3.ID, 1}} {
		t.Run(test.name, func(*testing.T) {
			got, err := OrgMembers(db).MemberCount(ctx, test.orgID)
			if err != nil {
				t.Fatal(err)
			}
			if test.want != got {
				t.Errorf("want %v, got %v", test.want, got)
			}
		})

	}

}
