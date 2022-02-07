package database

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GlobalStateStore interface {
	Get(context.Context) (*GlobalState, error)
	SiteInitialized(context.Context) (bool, error)

	// EnsureInitialized ensures the site is marked as having been initialized. If the site was already
	// initialized, it does nothing. It returns whether the site was already initialized prior to the
	// call.
	//
	// 🚨 SECURITY: Initialization is an important security measure. If a new account is created on a
	// site that is not initialized, and no other accounts exist, it is granted site admin
	// privileges. If the site *has* been initialized, then a new account is not granted site admin
	// privileges (even if all other users are deleted). This reduces the risk of (1) a site admin
	// accidentally deleting all user accounts and opening up their site to any attacker becoming a site
	// admin and (2) a bug in user account creation code letting attackers create site admin accounts.
	EnsureInitialized(context.Context) (bool, error)
}

func GlobalStateWith(other basestore.ShareableStore) GlobalStateStore {
	return &globalStateStore{Store: basestore.NewWithHandle(other.Handle())}
}

type GlobalState struct {
	SiteID      string
	Initialized bool // whether the initial site admin account has been created
}

type globalStateStore struct {
	*basestore.Store
}

func (g *globalStateStore) Get(ctx context.Context) (*GlobalState, error) {
	configuration, err := g.getConfiguration(ctx)
	if err == nil {
		return configuration, nil
	}

	if err != sql.ErrNoRows {
		return nil, errors.Wrap(err, "getConfiguration")
	}

	err = g.tryInsertNew(ctx)
	if err != nil {
		return nil, err
	}
	return g.getConfiguration(ctx)
}

func (g *globalStateStore) SiteInitialized(ctx context.Context) (bool, error) {
	var alreadyInitialized bool
	q := sqlf.Sprintf(`SELECT initialized FROM global_state LIMIT 1`)
	err := g.QueryRow(ctx, q).Scan(&alreadyInitialized)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return alreadyInitialized, err
}

func (g *globalStateStore) EnsureInitialized(ctx context.Context) (bool, error) {
	if err := g.tryInsertNew(ctx); err != nil {
		return false, err
	}

	// The "SELECT ... FOR UPDATE" prevents a race condition where two calls,
	// each in their own transaction, would see this initialized value as false
	// and then set it to true below.
	var alreadyInitialized bool
	q := sqlf.Sprintf(`SELECT initialized FROM global_state FOR UPDATE LIMIT 1`)
	err := g.QueryRow(ctx, q).Scan(&alreadyInitialized)
	if err != nil {
		return false, err
	}

	if !alreadyInitialized {
		err = g.Exec(ctx, sqlf.Sprintf("UPDATE global_state SET initialized=true"))
	}

	return alreadyInitialized, err
}

func (g *globalStateStore) getConfiguration(ctx context.Context) (*GlobalState, error) {
	var s GlobalState
	q := sqlf.Sprintf("SELECT site_id, initialized FROM global_state LIMIT 1")
	err := g.QueryRow(ctx, q).Scan(
		&s.SiteID,
		&s.Initialized,
	)
	return &s, err
}

func (g *globalStateStore) tryInsertNew(ctx context.Context) error {
	siteID, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	// In the normal case (when no users exist yet because the instance is brand new), create the row
	// with initialized=false.
	//
	// If any users exist, then set the site as initialized so that the init screen doesn't show
	// up. (It would not let the visitor initialize the site anyway, because other users exist.) The
	// most likely reason the instance would get into this state (uninitialized but has users) is
	// because previously global state had a siteID and now we ignore that (or someone ran `DELETE
	// FROM global_state;` in the PostgreSQL database). In either case, it's safe to generate a new
	// site ID and set the site as initialized.
	err = g.Exec(ctx, sqlf.Sprintf(`
	INSERT INTO global_state(
		site_id,
		initialized
	) values(
		%s,
		EXISTS (SELECT 1 FROM users WHERE deleted_at IS NULL)
	);`, siteID))
	if err != nil {
		var e *pgconn.PgError
		if errors.As(err, &e) && e.ConstraintName == "global_state_pkey" {
			// The row we were trying to insert already exists.
			// Don't treat this as an error.
			err = nil
		}
	}
	return err
}
