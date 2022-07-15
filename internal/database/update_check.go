package database

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type UpdateChecksStore interface {
	basestore.ShareableStore

	GetStatus(ctx context.Context) (status Status, isPending bool, err error)
	StartCheck(ctx context.Context) (err error)
	FinishCheck(ctx context.Context, updateVersion string, err string) error
}

// Status of the check for software updates for Sourcegraph.
type Status struct {
	ID int
	// the time that the last check completed
	StartedAt time.Time
	// the time that the last check completed
	FinishedAt time.Time
	// the error that occurred, if any. When present, indicates the instance is offline / unable to contact Sourcegraph.com
	Error string
	// the version string of the updated version, if any
	UpdateVersion string
}

// HasUpdate reports whether the status indicates an update is available.
func (s Status) HasUpdate() bool { return s.UpdateVersion != "" }

type updateChecksStore struct {
	*basestore.Store
}

func UpdateChecksWith(other basestore.ShareableStore) UpdateChecksStore {
	return &updateChecksStore{
		Store: basestore.NewWithHandle(other.Handle()),
	}
}

func (s *updateChecksStore) GetStatus(ctx context.Context) (status Status, isPending bool, err error) {
	q := sqlf.Sprintf(
		updateCheckGetStateQueryFmtstr,
	)

	if err := scanUpdateCheck(s.QueryRow(ctx, q), &status); err != nil {
		return status, false, errors.Wrap(err, "scanning update check record")
	}

	return status, false, nil
}

const updateCheckGetStateQueryFmtstr = `
-- source: internal/database/update_check.go:GetState
SELECT
	id,
	started_at,
	finished_at,
	update_version,
	error
FROM
	update_checks
ORDER BY started_at DESC
LIMIT 1
`

func (s *updateChecksStore) StartCheck(ctx context.Context) (err error) {
	// q := sqlf.Sprintf(
	// 	updateCheckStartCheckQueryFmtstr,
	// )

	// if err := scanUpdateCheck(s.QueryRow(ctx, q), &status); err != nil {
	// 	return status, errors.Wrap(err, "scanning update check record")
	// }

	return nil
}

const updateCheckStartCheckQueryFmtstr = `
-- source: internal/database/update_check.go:GetState
INSERT INTO update_checks (started_at, finished_at) VALUES (NOW(), NULL)
ON CONFLICT DO UPDATE SET
	started_at = EXCLUDED.started_at,
	finished_at = EXCLUDED.finished_at
`

func (s *updateChecksStore) FinishCheck(ctx context.Context, updateVersion string, err string) error {
	return nil
}

func scanUpdateCheck(sc ScannerWithError, status *Status) error {
	return sc.Scan(
		&status.ID,
		&status.StartedAt,
		&dbutil.NullTime{Time: &status.FinishedAt},
		&dbutil.NullString{S: &status.UpdateVersion},
		&dbutil.NullString{S: &status.Error},
	)
}
