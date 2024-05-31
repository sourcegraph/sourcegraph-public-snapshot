package database

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var orgnamesForTests = []struct {
	name      string
	wantValid bool
}{
	{"nick", true},
	{"n1ck", true},
	{"Nick2", true},
	{"N-S", true},
	{"nick-s", true},
	{"renfred-xh", true},
	{"renfred-x-h", true},
	{"deadmau5", true},
	{"deadmau-5", true},
	{"3blindmice", true},
	{"nick.com", true},
	{"nick.com.uk", true},
	{"nick.com-put-er", true},
	{"nick-", true},
	{"777", true},
	{"7-7", true},
	{"long-butnotquitelongenoughtoreachlimit", true},

	{".nick", false},
	{"-nick", false},
	{"nick.", false},
	{"nick--s", false},
	{"nick--sny", false},
	{"nick..sny", false},
	{"nick.-sny", false},
	{"_", false},
	{"_nick", false},
	{"ke$ha", false},
	{"ni%k", false},
	{"#nick", false},
	{"@nick", false},
	{"", false},
	{"nick s", false},
	{" ", false},
	{"-", false},
	{"--", false},
	{"-s", false},
	{"レンフレッド", false},
	{"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", false},
}

func TestOrgs_ValidNames(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	for _, test := range orgnamesForTests {
		t.Run(test.name, func(t *testing.T) {
			valid := true
			if _, err := db.Orgs().Create(ctx, test.name, nil); err != nil {
				if strings.Contains(err.Error(), "org name invalid") {
					valid = false
				} else {
					t.Fatal(err)
				}
			}
			if valid != test.wantValid {
				t.Errorf("%q: got valid %v, want %v", test.name, valid, test.wantValid)
			}
		})
	}
}

func TestOrgs_Count(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	org, err := db.Orgs().Create(ctx, "a", nil)
	if err != nil {
		t.Fatal(err)
	}

	if count, err := db.Orgs().Count(ctx, OrgsListOptions{}); err != nil {
		t.Fatal(err)
	} else if want := 1; count != want {
		t.Errorf("got %d, want %d", count, want)
	}

	if err := db.Orgs().Delete(ctx, org.ID); err != nil {
		t.Fatal(err)
	}

	if count, err := db.Orgs().Count(ctx, OrgsListOptions{}); err != nil {
		t.Fatal(err)
	} else if want := 0; count != want {
		t.Errorf("got %d, want %d", count, want)
	}
}

func TestOrgs_Delete(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	displayName := "a"
	org, err := db.Orgs().Create(ctx, "a", &displayName)
	if err != nil {
		t.Fatal(err)
	}

	// Delete org.
	if err := db.Orgs().Delete(ctx, org.ID); err != nil {
		t.Fatal(err)
	}

	// Org no longer exists.
	_, err = db.Orgs().GetByID(ctx, org.ID)
	if !errors.HasType[*OrgNotFoundError](err) {
		t.Errorf("got error %v, want *OrgNotFoundError", err)
	}
	orgs, err := db.Orgs().List(ctx, &OrgsListOptions{Query: "a"})
	if err != nil {
		t.Fatal(err)
	}
	if len(orgs) > 0 {
		t.Errorf("got %d orgs, want 0", len(orgs))
	}

	// Can't delete already-deleted org.
	err = db.Orgs().Delete(ctx, org.ID)
	if !errors.HasType[*OrgNotFoundError](err) {
		t.Errorf("got error %v, want *OrgNotFoundError", err)
	}
}

func TestOrgs_HardDelete(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	displayName := "org1"
	org, err := db.Orgs().Create(ctx, "org1", &displayName)
	require.NoError(t, err)

	// Hard Delete org.
	if err := db.Orgs().HardDelete(ctx, org.ID); err != nil {
		t.Fatal(err)
	}

	// Org no longer exists.
	_, err = db.Orgs().GetByID(ctx, org.ID)
	if !errors.HasType[*OrgNotFoundError](err) {
		t.Errorf("got error %v, want *OrgNotFoundError", err)
	}

	orgs, err := db.Orgs().List(ctx, &OrgsListOptions{Query: "org1"})
	require.NoError(t, err)
	if len(orgs) > 0 {
		t.Errorf("got %d orgs, want 0", len(orgs))
	}

	// Cannot hard delete an org that doesn't exist.
	err = db.Orgs().HardDelete(ctx, org.ID)
	if !errors.HasType[*OrgNotFoundError](err) {
		t.Errorf("got error %v, want *OrgNotFoundError", err)
	}

	// Can hard delete an org that has been soft deleted.
	displayName2 := "org2"
	org2, err := db.Orgs().Create(ctx, "org2", &displayName2)
	require.NoError(t, err)

	err = db.Orgs().Delete(ctx, org2.ID)
	require.NoError(t, err)

	err = db.Orgs().HardDelete(ctx, org2.ID)
	require.NoError(t, err)
}

func TestOrgs_GetByID(t *testing.T) {
	createOrg := func(ctx context.Context, db DB, name string, displayName string) *types.Org {
		org, err := db.Orgs().Create(ctx, name, &displayName)
		if err != nil {
			t.Fatal(err)
			return nil
		}
		return org
	}

	createUser := func(ctx context.Context, db DB, name string) *types.User {
		user, err := db.Users().Create(ctx, NewUser{
			Username: name,
		})
		if err != nil {
			t.Fatal(err)
			return nil
		}
		return user
	}

	createOrgMember := func(ctx context.Context, db DB, userID int32, orgID int32) *types.OrgMembership {
		member, err := db.OrgMembers().Create(ctx, orgID, userID)
		if err != nil {
			t.Fatal(err)
			return nil
		}
		return member
	}

	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	createOrg(ctx, db, "org1", "org1")
	org2 := createOrg(ctx, db, "org2", "org2")

	user := createUser(ctx, db, "user")
	createOrgMember(ctx, db, user.ID, org2.ID)

	orgs, err := db.Orgs().GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(orgs) != 1 {
		t.Errorf("got %d orgs, want 0", len(orgs))
	}
	if orgs[0].Name != org2.Name {
		t.Errorf("got %q org Name, want %q", orgs[0].Name, org2.Name)
	}
}
