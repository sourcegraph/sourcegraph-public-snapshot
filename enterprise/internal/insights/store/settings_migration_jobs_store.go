package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type SettingsMigrationJob struct {
	UserId             *int
	OrgId              *int
	Global             bool
	TotalInsights      int
	MigratedInsights   int
	TotalDashboards    int
	MigratedDashboards int
	Runs               int
	DashboardCreated   bool
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
	var where *sqlf.Query
	if jobType == UserJob {
		where = sqlf.Sprintf("user_id IS NOT NULL AND org_id IS NULL AND global IS FALSE")
	} else if jobType == OrgJob {
		where = sqlf.Sprintf("user_id IS NULL AND org_id IS NOT NULL AND global IS FALSE")
	} else {
		where = sqlf.Sprintf("org_id IS NULL AND user_id IS NULL AND global IS TRUE")
	}

	q := sqlf.Sprintf(getSettingsMigrationJobsSql, where)
	fmt.Println(q)

	return scanSettingsMigrationJobs(s.Query(ctx, q))
}

const getSettingsMigrationJobsSql = `
-- source: enterprise/internal/insights/store/settings_migration_jobs.go:GetSettingsMigrationJob
SELECT user_id, org_id, global, total_insights, migrated_insights, total_dashboards, migrated_dashboards, runs,
(CASE WHEN virtual_dashboard_created_at IS NULL THEN FALSE ELSE TRUE END) AS virtual_dashboard_created
FROM settings_migration_jobs
WHERE %s AND (total_insights > migrated_insights OR total_dashboards > migrated_dashboards OR virtual_dashboard_created_at IS NULL)
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
			&temp.Global,
			&temp.TotalInsights,
			&temp.MigratedInsights,
			&temp.TotalDashboards,
			&temp.MigratedDashboards,
			&temp.Runs,
			&temp.DashboardCreated,
		); err != nil {
			return []*SettingsMigrationJob{}, err
		}
		results = append(results, &temp)
	}
	return results, nil
}

type CreateSettingsMigrationJobArgs struct {
	UserId *int32
	OrgId  *int32
}

func (s *DBSettingsMigrationJobsStore) CreateSettingsMigrationJob(ctx context.Context, args CreateSettingsMigrationJobArgs) error {
	var q *sqlf.Query
	if args.UserId != nil {
		q = sqlf.Sprintf(insertUserSettingsMigrationJobsSql, *args.UserId)
	} else if args.OrgId != nil {
		q = sqlf.Sprintf(insertOrgSettingsMigrationJobsSql, *args.OrgId)
	} else {
		q = sqlf.Sprintf(insertGlobalSettingsMigrationJobsSql)
	}
	row := s.QueryRow(ctx, q)
	if row.Err() != nil {
		return row.Err()
	}

	return nil
}

const insertUserSettingsMigrationJobsSql = `
-- source: enterprise/internal/insights/store/settings_migration_jobs_store.go:CreateSettingsMigrationJob
INSERT INTO settings_migration_jobs (user_id) VALUES (%s)
ON CONFLICT DO NOTHING;
`

const insertOrgSettingsMigrationJobsSql = `
-- source: enterprise/internal/insights/store/settings_migration_jobs.go:CreateSettingsMigrationJob
INSERT INTO settings_migration_jobs (org_id) VALUES (%s)
ON CONFLICT DO NOTHING;
`

const insertGlobalSettingsMigrationJobsSql = `
-- source: enterprise/internal/insights/store/settings_migration_jobs.go:CreateSettingsMigrationJob
INSERT INTO settings_migration_jobs (global) VALUES (true)
ON CONFLICT DO NOTHING;
`

type UpdateSettingsMigrationJobArgs struct {
	UserId             *int
	OrgId              *int
	TotalInsights      int
	MigratedInsights   int
	TotalDashboards    int
	MigratedDashboards int
	Runs               int
}

func (s *DBSettingsMigrationJobsStore) UpdateSettingsMigrationJob(ctx context.Context, args UpdateSettingsMigrationJobArgs) error {
	var where *sqlf.Query
	if args.UserId != nil {
		where = sqlf.Sprintf("user_id = %s", args.UserId)
	} else if args.OrgId != nil {
		where = sqlf.Sprintf("org_id = %s", args.OrgId)
	} else {
		where = sqlf.Sprintf("global IS TRUE")
	}

	q := sqlf.Sprintf(
		updateSettingsMigrationJobSql,
		args.TotalInsights,
		args.MigratedInsights,
		args.TotalDashboards,
		args.MigratedDashboards,
		args.Runs,
		where,
	)

	row := s.QueryRow(ctx, q)
	if row.Err() != nil {
		return row.Err()
	}

	return nil
}

const updateSettingsMigrationJobSql = `
-- source: enterprise/internal/insights/store/settings_migration_jobs.go:CreateSettingsMigrationJob
UPDATE settings_migration_jobs
SET total_insights = %s, migrated_insights = %s, total_dashboards = %s, migrated_dashboards = %s, runs = %s
WHERE %s
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
	var where *sqlf.Query
	if jobType == UserJob {
		where = sqlf.Sprintf("user_id IS NOT NULL AND org_id IS NULL AND global IS FALSE")
	} else if jobType == OrgJob {
		where = sqlf.Sprintf("user_id IS NULL AND org_id IS NOT NULL AND global IS FALSE")
	} else {
		where = sqlf.Sprintf("org_id IS NULL AND user_id IS NULL AND global IS TRUE")
	}

	q := sqlf.Sprintf(countIncompleteJobsSql, where)
	fmt.Println(q)

	count, _, err := basestore.ScanFirstInt(s.Query(ctx, q))
	return count == 0, err
}

const countIncompleteJobsSql = `
-- source: enterprise/internal/insights/store/settings_migration_jobs.go:IsJobTypeComplete
SELECT COUNT(*) FROM settings_migration_jobs
WHERE %s AND (total_insights > migrated_insights OR total_dashboards > migrated_dashboards OR virtual_dashboard_created_at IS NULL);
`

type SettingsMigrationJobsStore interface {
	CreateSettingsMigrationJob(ctx context.Context, args CreateSettingsMigrationJobArgs) error
	UpdateSettingsMigrationJob(ctx context.Context, args UpdateSettingsMigrationJobArgs) error
	CountSettingsMigrationJobs(ctx context.Context) (int, error)
	GetNextSettingsMigrationJobs(ctx context.Context, jobType SettingsMigrationJobType) ([]*SettingsMigrationJob, error)
	IsJobTypeComplete(ctx context.Context, jobType SettingsMigrationJobType) (bool, error)
}
