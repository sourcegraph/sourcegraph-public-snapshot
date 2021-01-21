package db

import (
	"context"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

func TestOrgs_ValidNames(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()

	for _, test := range usernamesForTests {
		t.Run(test.name, func(t *testing.T) {
			valid := true
			if _, err := GlobalOrgs.Create(ctx, test.name, nil); err != nil {
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
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()

	org, err := GlobalOrgs.Create(ctx, "a", nil)
	if err != nil {
		t.Fatal(err)
	}

	if count, err := GlobalOrgs.Count(ctx, OrgsListOptions{}); err != nil {
		t.Fatal(err)
	} else if want := 1; count != want {
		t.Errorf("got %d, want %d", count, want)
	}

	if err := GlobalOrgs.Delete(ctx, org.ID); err != nil {
		t.Fatal(err)
	}

	if count, err := GlobalOrgs.Count(ctx, OrgsListOptions{}); err != nil {
		t.Fatal(err)
	} else if want := 0; count != want {
		t.Errorf("got %d, want %d", count, want)
	}
}

func TestOrgs_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	ctx := context.Background()

	displayName := "a"
	org, err := GlobalOrgs.Create(ctx, "a", &displayName)
	if err != nil {
		t.Fatal(err)
	}

	// Delete org.
	if err := GlobalOrgs.Delete(ctx, org.ID); err != nil {
		t.Fatal(err)
	}

	// Org no longer exists.
	_, err = GlobalOrgs.GetByID(ctx, org.ID)
	if _, ok := err.(*OrgNotFoundError); !ok {
		t.Errorf("got error %v, want *OrgNotFoundError", err)
	}
	orgs, err := GlobalOrgs.List(ctx, &OrgsListOptions{Query: "a"})
	if err != nil {
		t.Fatal(err)
	}
	if len(orgs) > 0 {
		t.Errorf("got %d orgs, want 0", len(orgs))
	}

	// Can't delete already-deleted org.
	err = GlobalOrgs.Delete(ctx, org.ID)
	if _, ok := err.(*OrgNotFoundError); !ok {
		t.Errorf("got error %v, want *OrgNotFoundError", err)
	}
}
