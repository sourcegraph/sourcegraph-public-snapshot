package upgradestore

import (
	"context"
	"time"

	"github.com/Masterminds/semver"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

// store manages checking and updating the version of the instance that was running prior to an ongoing
// instance upgrade or downgrade operation.
type store struct {
	db *basestore.Store
}

// New returns a new version store with the given database handle.
func New(db database.DB) *store {
	return NewWith(db.Handle())
}

// NewWith returns a new version store with the given transactable handle.
func NewWith(db basestore.TransactableHandle) *store {
	return &store{
		db: basestore.NewWithHandle(db),
	}
}

// GetFirstServiceVersion returns the first version registered for the given Sourcegraph service. This
// method will return a false-valued flag if UpdateServiceVersion has never been called for the given
// service.
func (s *store) GetFirstServiceVersion(ctx context.Context, service string) (_ string, _ bool, err error) {
	return basestore.ScanFirstString(s.db.Query(ctx, sqlf.Sprintf(getFirstServiceVersionQuery, service)))
}

const getFirstServiceVersionQuery = `
-- source: internal/version/store/store.go:GetFirstServiceVersion
SELECT first_version FROM versions WHERE service = %s
`

// ValidateUpgrade enforces our documented upgrade policy and will return an error (performing no side-effects)
// if the upgrade is between two unsupported versions. See https://docs.sourcegraph.com/#upgrading-sourcegraph.
func (s *store) ValidateUpgrade(ctx context.Context, service, version string) (err error) {
	return s.updateServiceVersion(ctx, service, version, false)
}

// UpdateServiceVersion updates the latest version for the given Sourcegraph service. This method also enforces
// our documented upgrade policy and will return an error (performing no side-effects) if the upgrade is between
// two unsupported versions. See https://docs.sourcegraph.com/#upgrading-sourcegraph.
func (s *store) UpdateServiceVersion(ctx context.Context, service, version string) (err error) {
	return s.updateServiceVersion(ctx, service, version, true)
}

func (s *store) updateServiceVersion(ctx context.Context, service, version string, update bool) (err error) {
	tx, err := s.db.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	prev, _, err := basestore.ScanFirstString(tx.Query(ctx, sqlf.Sprintf(updateServiceVersionSelectQuery, service)))
	if err != nil {
		return err
	}

	latest, _ := semver.NewVersion(version)
	previous, _ := semver.NewVersion(prev)

	if !IsValidUpgrade(previous, latest) {
		return &UpgradeError{Service: service, Previous: previous, Latest: latest}
	}

	if update {
		if err := tx.Exec(ctx, sqlf.Sprintf(updateServiceVersionSelectUpsertQuery, service, version, time.Now().UTC(), prev)); err != nil {
			return err
		}
	}

	return nil
}

const updateServiceVersionSelectQuery = `
-- source: internal/version/store/store.go:updateServiceVersion
SELECT version FROM versions WHERE service = %s
`

const updateServiceVersionSelectUpsertQuery = `
-- source: internal/version/store/store.go:updateServiceVersion
INSERT INTO versions (service, version, updated_at)
VALUES (%s, %s, %s) ON CONFLICT (service) DO
UPDATE SET (version, updated_at) = (excluded.version, excluded.updated_at)
WHERE versions.version = %s
`
