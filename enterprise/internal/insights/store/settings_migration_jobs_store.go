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
	where := getWhereForSubjectType(ctx, jobType)
	q := sqlf.Sprintf(getSettingsMigrationJobsSql, where)
	//fmt.Println(q)

	return scanSettingsMigrationJobs(s.Query(ctx, q))
}

const getSettingsMigrationJobsSql = `
-- source: enterprise/internal/insights/store/settings_migration_jobs.go:GetSettingsMigrationJob
SELECT user_id, org_id, global, total_insights, migrated_insights, total_dashboards, migrated_dashboards, runs,
(CASE WHEN completed_at IS NULL THEN FALSE ELSE TRUE END) AS dashboard_created
FROM insights_settings_migration_jobs
WHERE %s AND completed_at IS NULL
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
INSERT INTO insights_settings_migration_jobs (user_id) VALUES (%s)
ON CONFLICT DO NOTHING;
`

const insertOrgSettingsMigrationJobsSql = `
-- source: enterprise/internal/insights/store/settings_migration_jobs.go:CreateSettingsMigrationJob
INSERT INTO insights_settings_migration_jobs (org_id) VALUES (%s)
ON CONFLICT DO NOTHING;
`

const insertGlobalSettingsMigrationJobsSql = `
-- source: enterprise/internal/insights/store/settings_migration_jobs.go:CreateSettingsMigrationJob
INSERT INTO insights_settings_migration_jobs (global) VALUES (true)
ON CONFLICT DO NOTHING;
`

func (s *DBSettingsMigrationJobsStore) UpdateTotalInsights(ctx context.Context, userId *int, orgId *int, totalInsights int) error {
	q := sqlf.Sprintf(updateTotalInsightsSql, totalInsights, getWhereForSubject(ctx, userId, orgId))
	row := s.QueryRow(ctx, q)
	if row.Err() != nil {
		return row.Err()
	}
	return nil
}

const updateTotalInsightsSql = `
-- source: enterprise/internal/insights/store/settings_migration_jobs.go:CreateSettingsMigrationJob
UPDATE insights_settings_migration_jobs SET total_insights = %s WHERE %s
`

func (s *DBSettingsMigrationJobsStore) UpdateMigratedInsights(ctx context.Context, userId *int, orgId *int, migratedInsights int) error {
	q := sqlf.Sprintf(updateMigratedInsightsSql, migratedInsights, getWhereForSubject(ctx, userId, orgId))
	row := s.QueryRow(ctx, q)
	if row.Err() != nil {
		return row.Err()
	}
	return nil
}

const updateMigratedInsightsSql = `
-- source: enterprise/internal/insights/store/settings_migration_jobs.go:CreateSettingsMigrationJob
UPDATE insights_settings_migration_jobs SET migrated_insights = %s WHERE %s
`

func (s *DBSettingsMigrationJobsStore) UpdateTotalDashboards(ctx context.Context, userId *int, orgId *int, totalDashboards int) error {
	q := sqlf.Sprintf(updateTotalDashboardsSql, totalDashboards, getWhereForSubject(ctx, userId, orgId))
	row := s.QueryRow(ctx, q)
	if row.Err() != nil {
		return row.Err()
	}
	return nil
}

const updateTotalDashboardsSql = `
-- source: enterprise/internal/insights/store/settings_migration_jobs.go:CreateSettingsMigrationJob
UPDATE insights_settings_migration_jobs SET total_dashboards = %s WHERE %s
`

func (s *DBSettingsMigrationJobsStore) UpdateMigratedDashboards(ctx context.Context, userId *int, orgId *int, migratedDashboards int) error {
	q := sqlf.Sprintf(updateMigratedDashboardsSql, migratedDashboards, getWhereForSubject(ctx, userId, orgId))
	row := s.QueryRow(ctx, q)
	if row.Err() != nil {
		return row.Err()
	}
	return nil
}

const updateMigratedDashboardsSql = `
-- source: enterprise/internal/insights/store/settings_migration_jobs.go:CreateSettingsMigrationJob
UPDATE insights_settings_migration_jobs SET migrated_dashboards = %s WHERE %s
`

func (s *DBSettingsMigrationJobsStore) CountSettingsMigrationJobs(ctx context.Context) (int, error) {
	count, _, err := basestore.ScanFirstInt(s.Query(ctx, sqlf.Sprintf(countSettingsMigrationJobsSql)))
	return count, err
}

const countSettingsMigrationJobsSql = `
-- source: enterprise/internal/insights/store/settings_migration_jobs.go:CreateSettingsMigrationJob
SELECT COUNT(*) from insights_settings_migration_jobs;
`

func (s *DBSettingsMigrationJobsStore) IsJobTypeComplete(ctx context.Context, jobType SettingsMigrationJobType) (bool, error) {
	where := getWhereForSubjectType(ctx, jobType)
	q := sqlf.Sprintf(countIncompleteJobsSql, where)

	count, _, err := basestore.ScanFirstInt(s.Query(ctx, q))
	return count == 0, err
}

const countIncompleteJobsSql = `
-- source: enterprise/internal/insights/store/settings_migration_jobs.go:IsJobTypeComplete
SELECT COUNT(*) FROM insights_settings_migration_jobs
WHERE %s AND (total_insights > migrated_insights OR total_dashboards > migrated_dashboards OR completed_at IS NULL);
`

func getWhereForSubject(ctx context.Context, userId *int, orgId *int) *sqlf.Query {
	if userId != nil {
		return sqlf.Sprintf("user_id = %s", *userId)
	} else if orgId != nil {
		return sqlf.Sprintf("org_id = %s", *orgId)
	} else {
		return sqlf.Sprintf("global IS TRUE")
	}
}

func getWhereForSubjectType(ctx context.Context, jobType SettingsMigrationJobType) *sqlf.Query {
	if jobType == UserJob {
		return sqlf.Sprintf("user_id IS NOT NULL")
	} else if jobType == OrgJob {
		return sqlf.Sprintf("org_id IS NOT NULL")
	} else {
		return sqlf.Sprintf("global IS TRUE")
	}
}

type SettingsMigrationJobsStore interface {
	CreateSettingsMigrationJob(ctx context.Context, args CreateSettingsMigrationJobArgs) error
	UpdateTotalInsights(ctx context.Context, userId *int, orgId *int, totalInsights int) error
	UpdateMigratedInsights(ctx context.Context, userId *int, orgId *int, migratedInsights int) error
	UpdateTotalDashboards(ctx context.Context, userId *int, orgId *int, totalDashboards int) error
	UpdateMigratedDashboards(ctx context.Context, userId *int, orgId *int, migratedDashboards int) error
	CountSettingsMigrationJobs(ctx context.Context) (int, error)
	GetNextSettingsMigrationJobs(ctx context.Context, jobType SettingsMigrationJobType) ([]*SettingsMigrationJob, error)
	IsJobTypeComplete(ctx context.Context, jobType SettingsMigrationJobType) (bool, error)
}
