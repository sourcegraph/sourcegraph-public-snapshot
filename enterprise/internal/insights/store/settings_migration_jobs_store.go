package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type SettingsMigrationJob struct {
	UserId             *int
	OrgId              *int
	TotalItems         int
	MigratedItems      int
	DashboardCreatedAt time.Time
	Runs               int
}

type DBSettingsMigrationJobsStore struct {
	*basestore.Store
	Now func() time.Time
}

func NewSettingsMigrationJobsStore(db dbutil.DB) *DBSettingsMigrationJobsStore {
	return &DBSettingsMigrationJobsStore{Store: basestore.NewWithDB(db, sql.TxOptions{}), Now: time.Now}
}

func (s *DBSettingsMigrationJobsStore) Handle() *basestore.TransactableHandle {
	return s.Store.Handle()
}

func (s *DBSettingsMigrationJobsStore) With(other basestore.ShareableStore) *DBSettingsMigrationJobsStore {
	return &DBSettingsMigrationJobsStore{Store: s.Store.With(other), Now: s.Now}
}

func (s *DBSettingsMigrationJobsStore) Transact(ctx context.Context) (*DBSettingsMigrationJobsStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &DBSettingsMigrationJobsStore{Store: txBase, Now: s.Now}, err
}

type SettingsMigrationJobType string

const (
	UserJob   SettingsMigrationJobType = "USER"
	OrgJob    SettingsMigrationJobType = "ORG"
	GlobalJob SettingsMigrationJobType = "GLOBAL"
)

func (s *DBSettingsMigrationJobsStore) GetNextSettingsMigrationJobs(ctx context.Context, jobType SettingsMigrationJobType) ([]*SettingsMigrationJob, error) {
	var q *sqlf.Query
	if jobType == UserJob {
		q = sqlf.Sprintf("WHERE user_id > 0 AND org_id == 0 AND global IS FALSE")
	} else if jobType == OrgJob {
		q = sqlf.Sprintf("WHERE user_id == 0 AND org_id > 0 AND global IS FALSE")
	} else {
		q = sqlf.Sprintf("WHERE org_id == 0 AND user_id == 0 AND global IS TRUE")
	}

	return scanSettingsMigrationJobs(s.Query(ctx, sqlf.Sprintf(getSettingsMigrationJobsSql, q)))
}

const getSettingsMigrationJobsSql = `
-- source: enterprise/internal/insights/store/settings_migration_jobs.go:GetSettingsMigrationJob
SELECT * FROM settings_migration_jobs
WHERE %s AND (total_items > migrated_items OR dashboard_created_at IS NULL)
LIMIT 100
FOR UPDATE SKIP LOCKED;
`

func scanSettingsMigrationJobs(rows *sql.Rows, queryErr error) (_ []*SettingsMigrationJob, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var results []*SettingsMigrationJob
	for rows.Next() {
		var temp SettingsMigrationJob
		if err := rows.Scan(
			&temp.UserId,
			&temp.OrgId,
			&temp.TotalItems,
			&temp.MigratedItems,
			&temp.DashboardCreatedAt,
			&temp.Runs,
		); err != nil {
			return []*SettingsMigrationJob{}, err
		}
		results = append(results, &temp)
	}
	return results, nil
}

type CreateSettingsMigrationJobArgs struct {
	UserId     *int32
	OrgId      *int32
	TotalItems int
}

func (s *DBSettingsMigrationJobsStore) CreateSettingsMigrationJob(ctx context.Context, args CreateSettingsMigrationJobArgs) error {
	var q *sqlf.Query
	if args.UserId != nil {
		q = sqlf.Sprintf(insertUserSettingsMigrationJobsSql, *args.UserId, args.TotalItems)
	} else if args.OrgId != nil {
		q = sqlf.Sprintf(insertOrgSettingsMigrationJobsSql, *args.OrgId, args.TotalItems)
	} else {
		q = sqlf.Sprintf(insertGlobalSettingsMigrationJobsSql, args.TotalItems)
	}
	row := s.QueryRow(ctx, q)
	if row.Err() != nil {
		return row.Err()
	}

	return nil
}

const insertUserSettingsMigrationJobsSql = `
-- source: enterprise/internal/insights/store/settings_migration_jobs_store.go:CreateSettingsMigrationJob
INSERT INTO settings_migration_jobs (user_id, total_items) VALUES (%s, %s)
ON CONFLICT DO NOTHING;
`

const insertOrgSettingsMigrationJobsSql = `
-- source: enterprise/internal/insights/store/settings_migration_jobs.go:CreateSettingsMigrationJob
INSERT INTO settings_migration_jobs (org_id, total_items) VALUES (%s, %s)
ON CONFLICT DO NOTHING;
`

const insertGlobalSettingsMigrationJobsSql = `
-- source: enterprise/internal/insights/store/settings_migration_jobs.go:CreateSettingsMigrationJob
INSERT INTO settings_migration_jobs (global, total_items) VALUES (true, %s)
ON CONFLICT DO NOTHING;
`

func (s *DBSettingsMigrationJobsStore) CountSettingsMigrationJobs(ctx context.Context) (int, error) {
	count, _, err := basestore.ScanFirstInt(s.Query(ctx, sqlf.Sprintf(countSettingsMigrationJobsSql)))
	return count, err
}

const countSettingsMigrationJobsSql = `
-- source: enterprise/internal/insights/store/settings_migration_jobs.go:CreateSettingsMigrationJob
SELECT COUNT(*) from settings_migration_jobs;
`

func (s *DBSettingsMigrationJobsStore) IsJobTypeComplete(ctx context.Context, jobType SettingsMigrationJobType) (bool, error) {
	var q *sqlf.Query
	if jobType == UserJob {
		q = sqlf.Sprintf("WHERE user_id > 0 AND org_id == 0 AND global IS FALSE")
	} else if jobType == OrgJob {
		q = sqlf.Sprintf("WHERE user_id == 0 AND org_id > 0 AND global IS FALSE")
	} else {
		q = sqlf.Sprintf("WHERE org_id == 0 AND user_id == 0 AND global IS TRUE")
	}

	count, _, err := basestore.ScanFirstInt(s.Query(ctx, sqlf.Sprintf(countIncompleteJobsSql, q)))
	return count == 0, err
}

const countIncompleteJobsSql = `
-- source: enterprise/internal/insights/store/settings_migration_jobs.go:IsJobTypeComplete
SELECT COUNT(*) FROM settings_migration_jobs
WHERE %s AND (total_items > migrated_items OR dashboard_created_at IS NULL);
`

type SettingsMigrationJobsStore interface {
	CreateSettingsMigrationJob(ctx context.Context, args CreateSettingsMigrationJobArgs) error
	CountSettingsMigrationJobs(ctx context.Context) (int, error)
	GetNextSettingsMigrationJobs(ctx context.Context, jobType SettingsMigrationJobType) ([]*SettingsMigrationJob, error)
	IsJobTypeComplete(ctx context.Context, jobType SettingsMigrationJobType) (bool, error)
}
