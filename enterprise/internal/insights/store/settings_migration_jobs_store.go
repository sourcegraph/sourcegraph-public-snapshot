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

func scanSettingsMigrationJob(rows *sql.Rows, queryErr error) (_ []*SettingsMigrationJob, err error) {
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
		fmt.Println("Adding job for user", *args.UserId)
	} else if args.OrgId != nil {
		q = sqlf.Sprintf(insertOrgSettingsMigrationJobsSql, *args.OrgId, args.TotalItems)
		fmt.Println("Adding job for org", *args.OrgId)
	} else {
		q = sqlf.Sprintf(insertGlobalSettingsMigrationJobsSql, args.TotalItems)
		fmt.Println("Adding job for global")
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

type SettingsMigrationJobsStore interface {
	CreateSettingsMigrationJob(ctx context.Context, args CreateSettingsMigrationJobArgs) error
	CountSettingsMigrationJobs(ctx context.Context) (int, error)
}
