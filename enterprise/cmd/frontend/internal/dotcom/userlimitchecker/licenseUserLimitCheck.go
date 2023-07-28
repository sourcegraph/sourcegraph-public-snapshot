package userlimitchecker

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// dbUserLimitChecker describes a user limit c row in the license_user_limit_check DB table
type dbUserLimitChecker struct {
	ID                         string // UUID
	LicenseID                  string // UUID
	UserCountAlertSentAt       *time.Time
	UserCountWhenEmailLastSent int
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
}

// dbUserLimitCheckers exposes user limit checkers in the license_user_limit_check DB table.
type dbUserLimitCheckers struct {
	db database.DB
}

func NewUserLimitChecker(db database.DB) dbUserLimitCheckers {
	return dbUserLimitCheckers{db: db}
}

var errLimitCheckerNotFound = errors.New("limit c not found")

func (c dbUserLimitCheckers) Create(ctx context.Context, licenseID string, userCount int) (id string, err error) {
	newUUID, err := uuid.NewRandom()
	if err != nil {
		return "", errors.Wrap(err, "new UUID")
	}
	if err = c.db.QueryRowContext(ctx, createUserLimitCheckerFmtStr,
		newUUID,
		licenseID,
		userCount,
	).Scan(&id); err != nil {
		return "", errors.Wrap(err, "could not create userLimitChecker")
	}
	return id, nil
}

const createUserLimitCheckerFmtStr = `
INSERT INTO license_user_limit_check
(id, license_id, user_count_when_email_last_sent, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id
`

func (c dbUserLimitCheckers) GetByLicenseID(ctx context.Context, licenseID string) (*dbUserLimitChecker, error) {
	row := c.db.QueryRowContext(ctx, getByLicenseIDFmtStr, licenseID)

	var limitChecker dbUserLimitChecker
	err := row.Scan(
		&limitChecker.ID,
		&limitChecker.LicenseID,
		&limitChecker.UserCountAlertSentAt,
		&limitChecker.UserCountWhenEmailLastSent,
		&limitChecker.CreatedAt,
		&limitChecker.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}
	return &limitChecker, nil
}

const getByLicenseIDFmtStr = `
SELECT
	id,
	license_id,
	user_count_alert_sent_at,
	user_count_when_email_last_sent,
	created_at,
	updated_at
FROM license_user_limit_check
WHERE license_id = $1
`

func (c dbUserLimitCheckers) Update(ctx context.Context, checkerId string, currentUserCount int) error {
	q := sqlf.Sprintf(
		updateLimitCheckerFmtStr,
		currentUserCount,
		checkerId,
	)

	res, err := c.db.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return errLimitCheckerNotFound
	}
	return nil
}

const updateLimitCheckerFmtStr = `
UPDATE license_user_limit_check
SET updated_at = NOW(),
	user_count_alert_sent_at = NOW(),
	user_count_when_email_last_sent = %s
WHERE id = %s
`
