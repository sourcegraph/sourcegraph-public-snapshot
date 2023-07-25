package userlimitchecker

import (
	"context"
	"time"

	"github.com/jackc/pgconn"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type LicenseUserLimitCheckStore interface {
	basestore.ShareableStore
	Done(error) error

	CreateUserLimitChecker(ctx context.Context, licenseId string) error
	UpdateUserLimitChecker(ctx context.Context, licenseId string) error
	GetAlertSentAt(ctx context.Context, licenseId string) error
	GetCountWhenAlertLastSent(ctx context.Context, licenseId string) error
	DeleteUserLimitChecker(ctx context.Context, licenseId string) error
}

type licenseUserLimitCheckStore struct {
	*basestore.Store
}

var ErrCheckerAlreadyExists = errors.New("user limit checker already exists for this license")
var columns = []*sqlf.Query{
	sqlf.Sprintf("license_id"),
	sqlf.Sprintf("user_count_alert_sent_at"),
	sqlf.Sprintf("user_count_when_email_last_sent"),
}

func (checker *licenseUserLimitCheckStore) CreateUserLimitChecker(ctx context.Context, licenseId string, userCount int) error {
	q := sqlf.Sprintf(
		createUserLimitCheckerFmtStr,
		sqlf.Join(columns, ","),
		licenseId,
		userCount,
	)

	if _, err := checker.Handle().ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...); err != nil {
		var e *pgconn.PgError
		if errors.As(err, &e) {
			switch e.ConstraintName {
			case "license_id":
				return ErrCheckerAlreadyExists
			}
		}
		return err
	}
	return nil
}

const createUserLimitCheckerFmtStr = `
INSERT INTO license_user_limit_check
(%s)
VALUES (%s, %s, %s, %s, %s)
`

// TODO: Write these functions
func (checker *licenseUserLimitCheckStore) UpdateUserLimitChecker(ctx context.Context, licenseId string) error {
	return nil
}
func (checker *licenseUserLimitCheckStore) GetAlertSentAt(ctx context.Context, licenseId string) (time.Time, error) {
	q := sqlf.Sprintf(
		"SELECT user_count_alert_sent_at FROM license_user_limit_check WHERE id = %s",
		licenseId,
	)

	var userCountAlertSentAt time.Time
	err := checker.Handle().QueryRowContext(
		ctx,
		q.Query(sqlf.PostgresBindVar),
		q.Args()...,
	).Scan(&dbutil.NullTime{Time: &userCountAlertSentAt})
	if err != nil {
		return time.Time{}, errors.Wrap(err, "could not get userCountAlertSentAt")
	}
	return userCountAlertSentAt, nil
}

func (checker *licenseUserLimitCheckStore) GetCountWhenAlertLastSent(ctx context.Context, licenseId string) (int, error) {
	q := sqlf.Sprintf(
		"SELECT user_count_when_email_last_sent FROM license_user_limit_check WHERE id = %s",
		licenseId,
	)

	var userCountWhenEmailLastSent int
	err := checker.Handle().QueryRowContext(
		ctx,
		q.Query(sqlf.PostgresBindVar),
		q.Args()...,
	).Scan(&userCountWhenEmailLastSent)
	if err != nil {
		return 0, errors.Wrap(err, "could not get userCountAlertSentAt")
	}
	return userCountWhenEmailLastSent, nil
}

func (checker *licenseUserLimitCheckStore) DeleteUserLimitChecker(ctx context.Context, licenseId string) error {
	return nil
}
