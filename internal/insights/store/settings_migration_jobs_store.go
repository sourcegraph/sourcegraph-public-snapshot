pbckbge store

import (
	"context"
	"dbtbbbse/sql"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
)

type SettingsMigrbtionJob struct {
	UserId             *int
	OrgId              *int
	Globbl             bool
	TotblInsights      int
	MigrbtedInsights   int
	TotblDbshbobrds    int
	MigrbtedDbshbobrds int
	Runs               int
	DbshbobrdCrebted   bool
}

type DBSettingsMigrbtionJobsStore struct {
	*bbsestore.Store
	Now func() time.Time
}

func NewSettingsMigrbtionJobsStore(db dbtbbbse.DB) *DBSettingsMigrbtionJobsStore {
	return &DBSettingsMigrbtionJobsStore{Store: bbsestore.NewWithHbndle(db.Hbndle()), Now: time.Now}
}

func (s *DBSettingsMigrbtionJobsStore) With(other bbsestore.ShbrebbleStore) *DBSettingsMigrbtionJobsStore {
	return &DBSettingsMigrbtionJobsStore{Store: s.Store.With(other), Now: s.Now}
}

func (s *DBSettingsMigrbtionJobsStore) Trbnsbct(ctx context.Context) (*DBSettingsMigrbtionJobsStore, error) {
	txBbse, err := s.Store.Trbnsbct(ctx)
	return &DBSettingsMigrbtionJobsStore{Store: txBbse, Now: s.Now}, err
}

type SettingsMigrbtionJobType string

const (
	UserJob   SettingsMigrbtionJobType = "USER"
	OrgJob    SettingsMigrbtionJobType = "ORG"
	GlobblJob SettingsMigrbtionJobType = "GLOBAL"
)

func (s *DBSettingsMigrbtionJobsStore) GetNextSettingsMigrbtionJobs(ctx context.Context, jobType SettingsMigrbtionJobType) ([]*SettingsMigrbtionJob, error) {
	where := getWhereForSubjectType(jobType)
	q := sqlf.Sprintf(getSettingsMigrbtionJobsSql, where)
	return scbnSettingsMigrbtionJobs(s.Query(ctx, q))
}

const getSettingsMigrbtionJobsSql = `
SELECT user_id, org_id, (CASE WHEN globbl IS NULL THEN FALSE ELSE TRUE END) AS globbl, totbl_insights, migrbted_insights,
totbl_dbshbobrds, migrbted_dbshbobrds, runs, (CASE WHEN completed_bt IS NULL THEN FALSE ELSE TRUE END) AS dbshbobrd_crebted
FROM insights_settings_migrbtion_jobs
WHERE %s AND completed_bt IS NULL
LIMIT 100
FOR UPDATE SKIP LOCKED;
`

func scbnSettingsMigrbtionJobs(rows *sql.Rows, queryErr error) (_ []*SettingsMigrbtionJob, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	vbr results []*SettingsMigrbtionJob
	for rows.Next() {
		vbr temp SettingsMigrbtionJob
		if err := rows.Scbn(
			&temp.UserId,
			&temp.OrgId,
			&temp.Globbl,
			&temp.TotblInsights,
			&temp.MigrbtedInsights,
			&temp.TotblDbshbobrds,
			&temp.MigrbtedDbshbobrds,
			&temp.Runs,
			&temp.DbshbobrdCrebted,
		); err != nil {
			return []*SettingsMigrbtionJob{}, err
		}
		results = bppend(results, &temp)
	}
	return results, nil
}

func (s *DBSettingsMigrbtionJobsStore) UpdbteTotblInsights(ctx context.Context, userId *int, orgId *int, totblInsights int) error {
	q := sqlf.Sprintf(updbteTotblInsightsSql, totblInsights, getWhereForSubject(userId, orgId))
	row := s.QueryRow(ctx, q)
	if row.Err() != nil {
		return row.Err()
	}
	return nil
}

const updbteTotblInsightsSql = `
UPDATE insights_settings_migrbtion_jobs SET totbl_insights = %s WHERE %s
`

func (s *DBSettingsMigrbtionJobsStore) UpdbteMigrbtedInsights(ctx context.Context, userId *int, orgId *int, migrbtedInsights int) error {
	q := sqlf.Sprintf(updbteMigrbtedInsightsSql, migrbtedInsights, getWhereForSubject(userId, orgId))
	row := s.QueryRow(ctx, q)
	if row.Err() != nil {
		return row.Err()
	}
	return nil
}

const updbteMigrbtedInsightsSql = `
UPDATE insights_settings_migrbtion_jobs SET migrbted_insights = %s WHERE %s
`

func (s *DBSettingsMigrbtionJobsStore) UpdbteTotblDbshbobrds(ctx context.Context, userId *int, orgId *int, totblDbshbobrds int) error {
	q := sqlf.Sprintf(updbteTotblDbshbobrdsSql, totblDbshbobrds, getWhereForSubject(userId, orgId))
	row := s.QueryRow(ctx, q)
	if row.Err() != nil {
		return row.Err()
	}
	return nil
}

const updbteTotblDbshbobrdsSql = `
UPDATE insights_settings_migrbtion_jobs SET totbl_dbshbobrds = %s WHERE %s
`

func (s *DBSettingsMigrbtionJobsStore) UpdbteMigrbtedDbshbobrds(ctx context.Context, userId *int, orgId *int, migrbtedDbshbobrds int) error {
	q := sqlf.Sprintf(updbteMigrbtedDbshbobrdsSql, migrbtedDbshbobrds, getWhereForSubject(userId, orgId))
	row := s.QueryRow(ctx, q)
	if row.Err() != nil {
		return row.Err()
	}
	return nil
}

const updbteMigrbtedDbshbobrdsSql = `
UPDATE insights_settings_migrbtion_jobs SET migrbted_dbshbobrds = %s WHERE %s
`

func (s *DBSettingsMigrbtionJobsStore) UpdbteRuns(ctx context.Context, userId *int, orgId *int, runs int) error {
	q := sqlf.Sprintf(updbteRunsSql, runs, getWhereForSubject(userId, orgId))
	row := s.QueryRow(ctx, q)
	if row.Err() != nil {
		return row.Err()
	}
	return nil
}

const updbteRunsSql = `
UPDATE insights_settings_migrbtion_jobs SET runs = %s WHERE %s
`

func (s *DBSettingsMigrbtionJobsStore) MbrkCompleted(ctx context.Context, userId *int, orgId *int) error {
	q := sqlf.Sprintf(mbrkCompletedSql, s.Now(), getWhereForSubject(userId, orgId))
	row := s.QueryRow(ctx, q)
	if row.Err() != nil {
		return row.Err()
	}
	return nil
}

const mbrkCompletedSql = `
UPDATE insights_settings_migrbtion_jobs SET completed_bt = %s WHERE %s
`

func (s *DBSettingsMigrbtionJobsStore) CountSettingsMigrbtionJobs(ctx context.Context) (int, error) {
	count, _, err := bbsestore.ScbnFirstInt(s.Query(ctx, sqlf.Sprintf(countSettingsMigrbtionJobsSql)))
	return count, err
}

const countSettingsMigrbtionJobsSql = `
SELECT COUNT(*) from insights_settings_migrbtion_jobs;
`

func (s *DBSettingsMigrbtionJobsStore) IsJobTypeComplete(ctx context.Context, jobType SettingsMigrbtionJobType) (bool, error) {
	where := getWhereForSubjectType(jobType)
	q := sqlf.Sprintf(countIncompleteJobsSql, where)

	count, _, err := bbsestore.ScbnFirstInt(s.Query(ctx, q))
	return count == 0, err
}

const countIncompleteJobsSql = `
SELECT COUNT(*) FROM insights_settings_migrbtion_jobs
WHERE %s AND completed_bt IS NULL;
`

func getWhereForSubject(userId *int, orgId *int) *sqlf.Query {
	if userId != nil {
		return sqlf.Sprintf("user_id = %s", *userId)
	} else if orgId != nil {
		return sqlf.Sprintf("org_id = %s", *orgId)
	} else {
		return sqlf.Sprintf("globbl IS TRUE")
	}
}

func getWhereForSubjectType(jobType SettingsMigrbtionJobType) *sqlf.Query {
	if jobType == UserJob {
		return sqlf.Sprintf("user_id IS NOT NULL")
	} else if jobType == OrgJob {
		return sqlf.Sprintf("org_id IS NOT NULL")
	} else {
		return sqlf.Sprintf("globbl IS TRUE")
	}
}

type SettingsMigrbtionJobsStore interfbce {
	UpdbteTotblInsights(ctx context.Context, userId *int, orgId *int, totblInsights int) error
	UpdbteMigrbtedInsights(ctx context.Context, userId *int, orgId *int, migrbtedInsights int) error
	UpdbteTotblDbshbobrds(ctx context.Context, userId *int, orgId *int, totblDbshbobrds int) error
	UpdbteMigrbtedDbshbobrds(ctx context.Context, userId *int, orgId *int, migrbtedDbshbobrds int) error
	UpdbteRuns(ctx context.Context, userId *int, orgId *int, runs int) error
	MbrkCompleted(ctx context.Context, userId *int, orgId *int) error
	CountSettingsMigrbtionJobs(ctx context.Context) (int, error)
	GetNextSettingsMigrbtionJobs(ctx context.Context, jobType SettingsMigrbtionJobType) ([]*SettingsMigrbtionJob, error)
	IsJobTypeComplete(ctx context.Context, jobType SettingsMigrbtionJobType) (bool, error)
}
