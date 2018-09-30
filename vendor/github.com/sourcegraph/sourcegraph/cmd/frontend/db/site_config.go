package db

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db/dbconn"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

type siteConfig struct{}

func (o *siteConfig) Get(ctx context.Context) (*types.SiteConfig, error) {
	if Mocks.SiteConfig.Get != nil {
		return Mocks.SiteConfig.Get(ctx)
	}

	configuration, err := o.getConfiguration(ctx)
	if err == nil {
		return configuration, nil
	}
	err = o.tryInsertNew(ctx, dbconn.Global)
	if err != nil {
		return nil, err
	}
	return o.getConfiguration(ctx)
}

func siteInitialized(ctx context.Context) (alreadyInitialized bool, err error) {
	if err := dbconn.Global.QueryRowContext(ctx, `SELECT initialized FROM site_config LIMIT 1`).Scan(&alreadyInitialized); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return alreadyInitialized, err
}

// ensureInitialized ensures the site is marked as having been initialized. If the site was already
// initialized, it does nothing. It returns whether the site was already initialized prior to the
// call.
//
// 🚨 SECURITY: Initialization is an important security measure. If a new account is created on a
// site that is not initialized, and no other accounts exist, it is granted site admin
// privileges. If the site *has* been initialized, then a new account is not granted site admin
// privileges (even if all other users are deleted). This reduces the risk of (1) a site admin
// accidentally deleting all user accounts and opening up their site to any attacker becoming a site
// admin and (2) a bug in user account creation code letting attackers create site admin accounts.
func (o *siteConfig) ensureInitialized(ctx context.Context, dbh interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}) (alreadyInitialized bool, err error) {
	if err := o.tryInsertNew(ctx, dbh); err != nil {
		return false, err
	}

	// The "SELECT ... FOR UPDATE" prevents a race condition where two calls, each in their own transaction,
	// would see this initialized value as false and then set it to true below.
	if err := dbh.QueryRowContext(ctx, `SELECT initialized FROM site_config FOR UPDATE LIMIT 1`).Scan(&alreadyInitialized); err != nil {
		return false, err
	}

	if !alreadyInitialized {
		_, err = dbh.ExecContext(ctx, "UPDATE site_config SET initialized=true")
	}

	return alreadyInitialized, err
}

func (o *siteConfig) getConfiguration(ctx context.Context) (*types.SiteConfig, error) {
	configuration := &types.SiteConfig{}
	err := dbconn.Global.QueryRowContext(ctx, "SELECT site_id, initialized FROM site_config LIMIT 1").Scan(
		&configuration.SiteID,
		&configuration.Initialized,
	)
	return configuration, err
}

func (o *siteConfig) tryInsertNew(ctx context.Context, dbh interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}) error {
	siteID, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	_, err = dbh.ExecContext(ctx, "INSERT INTO site_config(site_id, initialized) values($1, false)", siteID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Constraint == "site_config_pkey" {
				// The row we were trying to insert already exists.
				// Don't treat this as an error.
				err = nil
			}

		}
	}
	return err
}
