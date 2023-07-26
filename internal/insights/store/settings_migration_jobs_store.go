package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
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

func NewSettingsMigrationJobsStore(db database.DB) *DBSettingsMigrationJobsStore {
	return &DBSettingsMigrationJobsStore{Store: basestore.NewWithHandle(db.Handle()), Now: time.Now}
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
	where := getWhereForSubjectType(jobType)
	q := sqlf.Sprintf(getSettingsMigrationJobsSql, where)
	return scanSettingsMigrationJobs(s.Query(ctx, q))
}

const getSettingsMigrationJobsSql = `
SELECT user_id, org_id, (CASE WHEN global IS NULL THEN FALSE ELSE TRUE END) AS global, total_insights, migrated_insights,
total_dashboards, migrated_dashboards, runs, (CASE WHEN completed_at IS NULL THEN FALSE ELSE TRUE END) AS dashboard_created
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

func (s *DBSettingsMigrationJobsStore) UpdateTotalInsights(ctx context.Context, userId *int, orgId *int, totalInsights int) error {
	q := sqlf.Sprintf(updateTotalInsightsSql, totalInsights, getWhereForSubject(userId, orgId))
	row := s.QueryRow(ctx, q)
	if row.Err() != nil {
		return row.Err()
	}
	return nil
}

const updateTotalInsightsSql = `
UPDATE insights_settings_migration_jobs SET total_insights = %s WHERE %s
`

func (s *DBSettingsMigrationJobsStore) UpdateMigratedInsights(ctx context.Context, userId *int, orgId *int, migratedInsights int) error {
	q := sqlf.Sprintf(updateMigratedInsightsSql, migratedInsights, getWhereForSubject(userId, orgId))
	row := s.QueryRow(ctx, q)
	if row.Err() != nil {
		return row.Err()
	}
	return nil
}

const updateMigratedInsightsSql = `
UPDATE insights_settings_migration_jobs SET migrated_insights = %s WHERE %s
`

func (s *DBSettingsMigrationJobsStore) UpdateTotalDashboards(ctx context.Context, userId *int, orgId *int, totalDashboards int) error {
	q := sqlf.Sprintf(updateTotalDashboardsSql, totalDashboards, getWhereForSubject(userId, orgId))
	row := s.QueryRow(ctx, q)
	if row.Err() != nil {
		return row.Err()
	}
	return nil
}

const updateTotalDashboardsSql = `
UPDATE insights_settings_migration_jobs SET total_dashboards = %s WHERE %s
`

func (s *DBSettingsMigrationJobsStore) UpdateMigratedDashboards(ctx context.Context, userId *int, orgId *int, migratedDashboards int) error {
	q := sqlf.Sprintf(updateMigratedDashboardsSql, migratedDashboards, getWhereForSubject(userId, orgId))
	row := s.QueryRow(ctx, q)
	if row.Err() != nil {
		return row.Err()
	}
	return nil
}

const updateMigratedDashboardsSql = `
UPDATE insights_settings_migration_jobs SET migrated_dashboards = %s WHERE %s
`

func (s *DBSettingsMigrationJobsStore) UpdateRuns(ctx context.Context, userId *int, orgId *int, runs int) error {
	q := sqlf.Sprintf(updateRunsSql, runs, getWhereForSubject(userId, orgId))
	row := s.QueryRow(ctx, q)
	if row.Err() != nil {
		return row.Err()
	}
	return nil
}

const updateRunsSql = `
UPDATE insights_settings_migration_jobs SET runs = %s WHERE %s
`

func (s *DBSettingsMigrationJobsStore) MarkCompleted(ctx context.Context, userId *int, orgId *int) error {
	q := sqlf.Sprintf(markCompletedSql, s.Now(), getWhereForSubject(userId, orgId))
	row := s.QueryRow(ctx, q)
	if row.Err() != nil {
		return row.Err()
	}
	return nil
}

const markCompletedSql = `
UPDATE insights_settings_migration_jobs SET completed_at = %s WHERE %s
`

func (s *DBSettingsMigrationJobsStore) CountSettingsMigrationJobs(ctx context.Context) (int, error) {
	count, _, err := basestore.ScanFirstInt(s.Query(ctx, sqlf.Sprintf(countSettingsMigrationJobsSql)))
	return count, err
}

const countSettingsMigrationJobsSql = `
SELECT COUNT(*) from insights_settings_migration_jobs;
`

func (s *DBSettingsMigrationJobsStore) IsJobTypeComplete(ctx context.Context, jobType SettingsMigrationJobType) (bool, error) {
	where := getWhereForSubjectType(jobType)
	q := sqlf.Sprintf(countIncompleteJobsSql, where)

	count, _, err := basestore.ScanFirstInt(s.Query(ctx, q))
	return count == 0, err
}

const countIncompleteJobsSql = `
SELECT COUNT(*) FROM insights_settings_migration_jobs
WHERE %s AND completed_at IS NULL;
`

func getWhereForSubject(userId *int, orgId *int) *sqlf.Query {
	if userId != nil {
		return sqlf.Sprintf("user_id = %s", *userId)
	} else if orgId != nil {
		return sqlf.Sprintf("org_id = %s", *orgId)
	} else {
		return sqlf.Sprintf("global IS TRUE")
	}
}

func getWhereForSubjectType(jobType SettingsMigrationJobType) *sqlf.Query {
	if jobType == UserJob {
		return sqlf.Sprintf("user_id IS NOT NULL")
	} else if jobType == OrgJob {
		return sqlf.Sprintf("org_id IS NOT NULL")
	} else {
		return sqlf.Sprintf("global IS TRUE")
	}
}

type SettingsMigrationJobsStore interface {
	UpdateTotalInsights(ctx context.Context, userId *int, orgId *int, totalInsights int) error
	UpdateMigratedInsights(ctx context.Context, userId *int, orgId *int, migratedInsights int) error
	UpdateTotalDashboards(ctx context.Context, userId *int, orgId *int, totalDashboards int) error
	UpdateMigratedDashboards(ctx context.Context, userId *int, orgId *int, migratedDashboards int) error
	UpdateRuns(ctx context.Context, userId *int, orgId *int, runs int) error
	MarkCompleted(ctx context.Context, userId *int, orgId *int) error
	CountSettingsMigrationJobs(ctx context.Context) (int, error)
	GetNextSettingsMigrationJobs(ctx context.Context, jobType SettingsMigrationJobType) ([]*SettingsMigrationJob, error)
	IsJobTypeComplete(ctx context.Context, jobType SettingsMigrationJobType) (bool, error)
}
