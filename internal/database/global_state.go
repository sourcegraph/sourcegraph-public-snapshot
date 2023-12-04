package database

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GlobalStateStore interface {
	Get(context.Context) (GlobalState, error)
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

func scanGlobalState(s dbutil.Scanner) (value GlobalState, err error) {
	err = s.Scan(&value.SiteID, &value.Initialized)
	return
}

var scanFirstGlobalState = basestore.NewFirstScanner(scanGlobalState)

type globalStateStore struct {
	*basestore.Store
}

func (g *globalStateStore) Transact(ctx context.Context) (*globalStateStore, error) {
	tx, err := g.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &globalStateStore{Store: tx}, nil
}

func (g *globalStateStore) Get(ctx context.Context) (GlobalState, error) {
	if err := g.initializeDBState(ctx); err != nil {
		return GlobalState{}, err
	}

	state, found, err := scanFirstGlobalState(g.Query(ctx, sqlf.Sprintf(globalStateGetQuery)))
	if err != nil {
		return GlobalState{}, err
	}
	if !found {
		return GlobalState{}, errors.New("expected global_state to be initialized - no rows found")
	}

	return state, nil
}

var globalStateSiteIDFragment = `
SELECT site_id FROM global_state ORDER BY ctid LIMIT 1
`

var globalStateInitializedFragment = `
SELECT coalesce(bool_or(gs.initialized), false) FROM global_state gs
`

var globalStateGetQuery = fmt.Sprintf(`
SELECT (%s) AS site_id, (%s) AS initialized
`,
	globalStateSiteIDFragment,
	globalStateInitializedFragment,
)

func (g *globalStateStore) SiteInitialized(ctx context.Context) (bool, error) {
	alreadyInitialized, _, err := basestore.ScanFirstBool(g.Query(ctx, sqlf.Sprintf(globalStateSiteInitializedQuery)))
	return alreadyInitialized, err
}

var globalStateSiteInitializedQuery = globalStateInitializedFragment

func (g *globalStateStore) EnsureInitialized(ctx context.Context) (_ bool, err error) {
	if err := g.initializeDBState(ctx); err != nil {
		return false, err
	}

	tx, err := g.Transact(ctx)
	if err != nil {
		return false, err
	}
	defer func() { err = tx.Done(err) }()

	alreadyInitialized, err := tx.SiteInitialized(ctx)
	if err != nil {
		return false, err
	}

	if err := tx.Exec(ctx, sqlf.Sprintf(globalStateEnsureInitializedQuery)); err != nil {
		return false, err
	}

	return alreadyInitialized, nil
}

var globalStateEnsureInitializedQuery = `
UPDATE global_state SET initialized = true
`

func (g *globalStateStore) initializeDBState(ctx context.Context) (err error) {
	tx, err := g.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()
	if err := tx.Exec(ctx, sqlf.Sprintf(globalStateInitializeDBStateUpdateQuery)); err != nil {
		return err
	}
	if err := tx.Exec(ctx, sqlf.Sprintf(globalStateInitializeDBStatePruneQuery)); err != nil {
		return err
	}

	siteID, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	return tx.Exec(ctx, sqlf.Sprintf(globalStateInitializeDBStateInsertIfNotExistsQuery, siteID))
}

var globalStateInitializeDBStateUpdateQuery = fmt.Sprintf(`
UPDATE global_state SET initialized = (%s)
`,
	globalStateInitializedFragment,
)

var globalStateInitializeDBStatePruneQuery = fmt.Sprintf(`
DELETE FROM global_state WHERE site_id NOT IN (%s)
`,
	globalStateSiteIDFragment,
)

var globalStateInitializeDBStateInsertIfNotExistsQuery = `
INSERT INTO global_state(
	site_id,
	initialized
)
SELECT
	%s AS site_id,
	EXISTS (
		SELECT 1
		FROM users
		WHERE deleted_at IS NULL
	) AS initialized
WHERE
	NOT EXISTS (
		SELECT 1 FROM global_state
	)
`
