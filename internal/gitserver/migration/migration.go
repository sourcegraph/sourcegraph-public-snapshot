// Package migration contains code for migrating repositories from instance to instance
// using the Rendezvous hashing algorithm.
package migration

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

var MigrationMocks, emptyMigrationMocks struct {
	GetCursor func(ctx context.Context, db dbutil.DB) (string, error)
}

// ResetMigrationMocks clears the mock functions set on Mocks (so that subsequent
// tests don't inadvertently use them).
func ResetMigrationMocks() {
	MigrationMocks = emptyMigrationMocks
}

// GetCursor is a helper function that returns the current state of the migration.
// The cursor is used to determine which is the last repository that was migrated.
// Since repositories are migrated in alphabetical order, the cursor can be used by clients
// to determine which hashing algorithm to use.
// Before the migration is run, this function returns an empty string.
func GetCursor(ctx context.Context, db dbutil.DB) (string, error) {
	if MigrationMocks.GetCursor != nil {
		return MigrationMocks.GetCursor(ctx, db)
	}
	return "", nil
}
