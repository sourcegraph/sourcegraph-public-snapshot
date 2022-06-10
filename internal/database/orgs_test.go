package database

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/types/typestest"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestOrgs_ValidNames(t *testing.T) {
	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	for _, test := range usernamesForTests {
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
	db := NewDB(dbtest.NewDB(t))
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
	db := NewDB(dbtest.NewDB(t))
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
	if !errors.HasType(err, &OrgNotFoundError{}) {
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
	if !errors.HasType(err, &OrgNotFoundError{}) {
		t.Errorf("got error %v, want *OrgNotFoundError", err)
	}
}

func TestOrgs_HardDelete(t *testing.T) {
	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
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
	if !errors.HasType(err, &OrgNotFoundError{}) {
		t.Errorf("got error %v, want *OrgNotFoundError", err)
	}

	orgs, err := db.Orgs().List(ctx, &OrgsListOptions{Query: "org1"})
	require.NoError(t, err)
	if len(orgs) > 0 {
		t.Errorf("got %d orgs, want 0", len(orgs))
	}

	// Cannot hard delete an org that doesn't exist.
	err = db.Orgs().HardDelete(ctx, org.ID)
	if !errors.HasType(err, &OrgNotFoundError{}) {
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
	db := NewDB(dbtest.NewDB(t))
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

func TestOrgs_GetOrgsWithRepositoriesByUserID(t *testing.T) {
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

	createOrgMember := func(ctx context.Context, db DB, userID int32, orgID int32) {
		_, err := db.OrgMembers().Create(ctx, orgID, userID)
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	org1 := createOrg(ctx, db, "org1", "org1")
	org2 := createOrg(ctx, db, "org2", "org2")

	user := createUser(ctx, db, "user")
	createOrgMember(ctx, db, user.ID, org1.ID)
	createOrgMember(ctx, db, user.ID, org2.ID)

	service := &types.ExternalService{
		Kind:           extsvc.KindGitHub,
		Config:         `{"url": "https://github.com", "token": "abc", "repositoryQuery": ["none"]}`,
		NamespaceOrgID: org2.ID,
	}
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	if err := db.ExternalServices().Create(ctx, confGet, service); err != nil {
		t.Fatal(err)
	}
	repo := typestest.MakeGithubRepo(service)
	if err := db.Repos().Create(ctx, repo); err != nil {
		t.Fatal(err)
	}

	orgs, err := db.Orgs().GetOrgsWithRepositoriesByUserID(ctx, user.ID)
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

func TestOrgs_AddOrgsOpenBetaStats(t *testing.T) {
	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	userID := int32(42)

	type FooBar struct {
		Foo string `json:"foo"`
	}

	data, err := json.Marshal(FooBar{Foo: "bar"})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("When adding stats, returns valid UUID", func(t *testing.T) {
		id, err := db.Orgs().AddOrgsOpenBetaStats(ctx, userID, string(data))
		if err != nil {
			t.Fatal(err)
		}
		_, err = uuid.FromString(id)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Can add stats multiple times by the same user", func(t *testing.T) {
		_, err := db.Orgs().AddOrgsOpenBetaStats(ctx, userID, string(data))
		if err != nil {
			t.Fatal(err)
		}
		_, err = db.Orgs().AddOrgsOpenBetaStats(ctx, userID, string(data))
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestOrgs_UpdateOrgsOpenBetaStats(t *testing.T) {
	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	userID := int32(42)
	orgID := int32(10)
	statsID, err := db.Orgs().AddOrgsOpenBetaStats(ctx, userID, "{}")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Updates stats with orgID if the UUID exists in the DB", func(t *testing.T) {
		err := db.Orgs().UpdateOrgsOpenBetaStats(ctx, statsID, orgID)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Silently does nothing if UUID does not match any record", func(t *testing.T) {
		randomUUID, err := uuid.NewV4()
		if err != nil {
			t.Fatal(err)
		}
		err = db.Orgs().UpdateOrgsOpenBetaStats(ctx, randomUUID.String(), orgID)
		if err != nil {
			t.Fatal(err)
		}
	})
}
