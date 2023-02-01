package bg

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/rbac"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// UpdatePermissions is a startup process that compares the permissions in the database against those
// in the rbac schema config located in internal/rbac/schema.yaml. It ensures the permissions in the
// database are always up to date, using the schema config as it's source of truth.

// This method is called as part of the background process by the `frontend` service.
func UpdatePermissions(ctx context.Context, logger log.Logger, db database.DB) {
	scopedLog := logger.Scoped("permission_update", "Updates the permission in the database based on the rbac schema configuration.")
	err := db.WithTransact(ctx, func(tx database.DB) error {
		pstore := tx.Permissions()

		dbPerms, err := pstore.List(ctx)
		if err != nil {
			return errors.Wrap(err, "fetching permissions from database")
		}

		toBeAdded, toBeDeleted := rbac.ComparePermissions(dbPerms)
		scopedLog.Info("RBAC Permissions update", log.Int("added", len(toBeAdded)), log.Int("deleted", len(toBeDeleted)))

		if len(toBeDeleted) > 0 {
			// We delete all the permissions that need to be deleted from the database
			err = pstore.BulkDelete(ctx, toBeDeleted)
			if err != nil {
				return errors.Wrap(err, "deleting redundant permissions")
			}
		}

		if len(toBeAdded) > 0 {
			// Adding new permissions to the database
			_, err = pstore.BulkCreate(ctx, toBeAdded)
			if err != nil {
				return errors.Wrap(err, "creating new permissions")
			}
		}

		return nil
	})

	if err != nil {
		scopedLog.Error("failed to update RBAC permissions", log.Error(err))
	}
}
