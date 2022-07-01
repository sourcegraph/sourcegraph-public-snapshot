package upgradestore

import (
	"context"
	"time"

	"github.com/Masterminds/semver"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// store manages checking and updating the version of the instance that
// was running prior to an ongoing instance upgrade or downgrade operation.
type store struct {
	db *basestore.Store
	// operations *operations
}

// New returns a new version store.
func New(db database.DB, observationContext *observation.Context) *store {
	return &store{
		db: basestore.NewWithHandle(db.Handle()),
		// operations: newOperations(observationContext),
	}
}

// GetFirstServiceVersion returns the first version registered for the given Sourcegraph service. This
// method will return a false-valued flag if UpdateServiceVersion has never been called for the given
// service.
func (s *store) GetFirstServiceVersion(ctx context.Context, service string) (_ string, _ bool, err error) {
	// ctx, _, endObservation := s.operations.getFirstServiceVersion.With(ctx, &err, observation.Args{})
	// defer endObservation(1, observation.Args{})

	return basestore.ScanFirstString(s.db.Query(ctx, sqlf.Sprintf(getFirstServiceVersionQuery, service)))
}

const getFirstServiceVersionQuery = `
-- source: internal/version/store/store.go:GetFirstServiceVersion
SELECT first_version FROM versions WHERE service = %s
`

// UpdateServiceVersion updates the latest version for the given Sourcegraph service. It enforces our
// documented upgrade policy.
//
// See https://docs.sourcegraph.com/#upgrading-sourcegraph.
func (s *store) UpdateServiceVersion(ctx context.Context, service, version string) (err error) {
	// ctx, _, endObservation := s.operations.updateServiceVersion.With(ctx, &err, observation.Args{})
	// defer endObservation(1, observation.Args{})

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

	if err := tx.Exec(ctx, sqlf.Sprintf(upsertVersionQuery, service, version, time.Now().UTC(), prev)); err != nil {
		return err
	}

	return nil
}

const updateServiceVersionSelectQuery = `
-- source: internal/version/store/store.go:UpdateServiceVersion
SELECT version FROM versions WHERE service = %s
`

const upsertVersionQuery = `
-- source: internal/version/store/store.go:UpdateServiceVersion
INSERT INTO versions (service, version, updated_at)
VALUES (%s, %s, %s) ON CONFLICT (service) DO
UPDATE SET (version, updated_at) = (excluded.version, excluded.updated_at)
WHERE versions.version = %s
`
