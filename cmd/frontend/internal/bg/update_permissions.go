package bg

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/rbac"
)

func UpdatePermissions(ctx context.Context, logger log.Logger, db database.DB) {
	scopedLog := logger.Scoped("permission_update", "Updates the permission in the database based on the rbac schema configuration.")
	tx, err := db.Transact(ctx)
	if err != nil {
		scopedLog.Error("starting transaction", log.Error(err))
		return
	}
	defer func() { tx.Done(err) }()

	pstore := tx.Permissions()

	dbPerms, err := pstore.List(ctx)
	if err != nil {
		scopedLog.Error("fetching permissions from database", log.Error(err))
		return
	}

	toBeAdded, toBeDeleted := rbac.ComparePermissions(dbPerms)
	scopedLog.Info("RBAC Permissions update", log.Int("added", len(toBeAdded)), log.Int("deleted", len(toBeDeleted)))

	if len(toBeDeleted) > 0 {
		// We delete all the permissions that need to be deleted from the database
		err = pstore.BulkDelete(ctx, toBeDeleted)
		if err != nil {
			scopedLog.Error("deleting redundant permissions", log.Error(err))
			return
		}
	}

	if len(toBeAdded) > 0 {
		// Adding new permissions to the database
		_, err = pstore.BulkCreate(ctx, toBeAdded)
		if err != nil {
			scopedLog.Error("creating new permissions", log.Error(err))
			return
		}
	}

}
