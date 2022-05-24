package database

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestOrgStats_Upsert(t *testing.T) {
	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	org, err := db.Orgs().Create(ctx, "org1", nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Succeeds to create stats for existing org", func(t *testing.T) {
		stats, err := OrgStats(db).Upsert(ctx, org.ID, 42)
		if err != nil {
			t.Fatal(err)
		}

		if stats.OrgID != org.ID || stats.CodeHostRepoCount != 42 {
			t.Fatal("Incorrect data returned from DB write")
		}
	})

	t.Run("Succeeds to update stats for existing org", func(t *testing.T) {
		stats, err := OrgStats(db).Upsert(ctx, org.ID, 1024)
		if err != nil {
			t.Fatal(err)
		}

		if stats.OrgID != org.ID || stats.CodeHostRepoCount != 1024 {
			t.Fatal("Incorrect data returned from DB write")
		}
	})

	t.Run("Fails to update stats for non-existing org", func(t *testing.T) {
		_, err := OrgStats(db).Upsert(ctx, 42, 1)
		if err == nil {
			t.Fatal("Expected error when adding stats for non-existing organization")
		}
	})
}
