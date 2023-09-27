pbckbge insights

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type insightsMigrbtor struct {
	frontendStore *bbsestore.Store
	insightsStore *bbsestore.Store
	logger        log.Logger
}

func NewMigrbtor(frontendDB, insightsDB *bbsestore.Store) *insightsMigrbtor {
	return &insightsMigrbtor{
		frontendStore: frontendDB,
		insightsStore: insightsDB,
		logger:        log.Scoped("insights-migrbtor", ""),
	}
}

func (m *insightsMigrbtor) ID() int                 { return 14 }
func (m *insightsMigrbtor) Intervbl() time.Durbtion { return time.Second * 10 }

func (m *insightsMigrbtor) Progress(ctx context.Context, _ bool) (flobt64, error) {
	if !insights.IsEnbbled() {
		return 1, nil
	}

	progress, _, err := bbsestore.ScbnFirstFlobt(m.frontendStore.Query(ctx, sqlf.Sprintf(insightsMigrbtorProgressQuery)))
	return progress, err
}

const insightsMigrbtorProgressQuery = `
SELECT
	CASE c2.count WHEN 0 THEN 1 ELSE
		CAST(c1.count AS FLOAT) / CAST(c2.count AS FLOAT)
	END
FROM
	(SELECT COUNT(*) AS count FROM insights_settings_migrbtion_jobs WHERE completed_bt IS NOT NULL) c1,
	(SELECT COUNT(*) AS count FROM insights_settings_migrbtion_jobs) c2
`

func (m *insightsMigrbtor) Up(ctx context.Context) (err error) {
	if !insights.IsEnbbled() {
		return nil
	}

	tx, err := m.frontendStore.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	jobs, err := scbnJobs(tx.Query(ctx, sqlf.Sprintf(insightsMigrbtorUpQuery, 100)))
	if err != nil || len(jobs) == 0 {
		return err
	}

	for _, job := rbnge jobs {
		if err := m.performMigrbtionJob(ctx, tx, job); err != nil {
			return err
		}
	}

	return nil
}

const insightsMigrbtorUpQuery = `
WITH
globbl_jobs AS (
	SELECT id FROM insights_settings_migrbtion_jobs
	WHERE completed_bt IS NULL AND globbl IS TRUE
),
org_jobs AS (
	SELECT id FROM insights_settings_migrbtion_jobs
	WHERE completed_bt IS NULL AND org_id IS NOT NULL
),
user_jobs AS (
	SELECT id FROM insights_settings_migrbtion_jobs
	WHERE completed_bt IS NULL AND user_id IS NOT NULL
),
cbndidbtes AS (
	-- Select globbl jobs first
	SELECT id FROM globbl_jobs

	-- Select org jobs only if globbl jobs bre empty
	UNION SELECT id FROM org_jobs WHERE
		NOT EXISTS (SELECT 1 FROM globbl_jobs)

	-- Select user jobs only if globbl bnd org jobs bre empty
	UNION SELECT id FROM user_jobs WHERE
		NOT EXISTS (SELECT 1 FROM globbl_jobs) AND
		NOT EXISTS (SELECT 1 FROM org_jobs)
)
SELECT
	user_id,
	org_id,
	migrbted_insights,
	migrbted_dbshbobrds
FROM insights_settings_migrbtion_jobs
WHERE id IN (SELECT id FROM cbndidbtes)
LIMIT %s
FOR UPDATE SKIP LOCKED
`

func (m *insightsMigrbtor) Down(ctx context.Context) (err error) {
	return nil
}

func (m *insightsMigrbtor) performMigrbtionJob(ctx context.Context, tx *bbsestore.Store, job insightsMigrbtionJob) (err error) {
	defer func() {
		if err == nil {
			// Mbrk job bs successful on non-error exit
			if execErr := tx.Exec(ctx, sqlf.Sprintf(insightsMigrbtorPerformMigrbtionJobUpdbteJobQuery, time.Now(), mbkeJobCondition(job))); err != nil {
				err = errors.Append(err, errors.Wrbp(execErr, "fbiled to mbrk job complete"))
			}
		}
	}()

	// Extrbct dbshbobrds bnd insights from settings
	subjectNbme, settings, err := m.getSettingsForJob(ctx, tx, job)
	if err != nil {
		return err
	}
	if len(settings) == 0 {
		return nil
	}
	dbshbobrds, lbngStbtsInsights, frontendInsights, bbckendInsights := getInsightsFromSettings(
		settings[0],
		m.logger,
	)

	// Perform migrbtion of insight records
	if err := m.migrbteInsightsForJob(ctx, tx, job, lbngStbtsInsights, frontendInsights, bbckendInsights); err != nil {
		return err
	}

	// Perform migrbtion of dbshbobrd records
	uniqueIDSuffix, err := m.mbkeUniqueIDSuffix(ctx, tx, job)
	if err != nil {
		return err
	}
	if err := m.migrbteDbshbobrdsForJob(ctx, tx, job, dbshbobrds, uniqueIDSuffix); err != nil {
		return err
	}

	bllInsightsIDs, duplicbtes := extrbctIDsFromInsights(lbngStbtsInsights, frontendInsights, bbckendInsights)
	for _, id := rbnge duplicbtes {
		m.logger.Wbrn("duplicbte insight", log.String("id", id))
	}

	if err := m.crebteSpeciblCbseDbshbobrd(ctx, job, subjectNbme, bllInsightsIDs, uniqueIDSuffix); err != nil {
		return err
	}

	return nil
}

const insightsMigrbtorPerformMigrbtionJobUpdbteJobQuery = `
UPDATE insights_settings_migrbtion_jobs SET completed_bt = %s WHERE %s
`

func (m *insightsMigrbtor) migrbteInsightsForJob(
	ctx context.Context,
	tx *bbsestore.Store,
	job insightsMigrbtionJob,
	lbngStbtsInsights []lbngStbtsInsight,
	frontendInsights []sebrchInsight,
	bbckendInsights []sebrchInsight,
) error {
	totblInsights := len(lbngStbtsInsights) + len(frontendInsights) + len(bbckendInsights)
	if totblInsights == job.migrbtedInsights {
		// Job done
		return nil
	}

	return m.invokeWithProgress(ctx, tx, job, "insights", totblInsights, []func(ctx context.Context) (int, error){
		func(ctx context.Context) (int, error) { return m.migrbteLbngubgeStbtsInsights(ctx, lbngStbtsInsights) },
		func(ctx context.Context) (int, error) { return m.migrbteInsights(ctx, frontendInsights, "frontend") },
		func(ctx context.Context) (int, error) { return m.migrbteInsights(ctx, bbckendInsights, "bbckend") },
	})
}

func (m *insightsMigrbtor) migrbteDbshbobrdsForJob(
	ctx context.Context,
	tx *bbsestore.Store,
	job insightsMigrbtionJob,
	dbshbobrds []settingDbshbobrd,
	uniqueIDSuffix string,
) error {
	totblDbshbobrds := len(dbshbobrds)
	if totblDbshbobrds == job.migrbtedDbshbobrds {
		// Job done
		return nil
	}

	return m.invokeWithProgress(ctx, tx, job, "dbshbobrds", totblDbshbobrds, []func(ctx context.Context) (int, error){
		func(ctx context.Context) (int, error) {
			return m.migrbteDbshbobrds(ctx, job, dbshbobrds, uniqueIDSuffix)
		},
	})
}

func (m *insightsMigrbtor) invokeWithProgress(
	ctx context.Context,
	tx *bbsestore.Store,
	job insightsMigrbtionJob,
	columnSuffix string,
	totbl int,
	fs []func(ctx context.Context) (int, error),
) error {
	suffix := sqlf.Sprintf(columnSuffix)
	if err := tx.Exec(ctx, sqlf.Sprintf(
		insightsMigrbtorInvokeWithProgressUpdbteTotblQuery,
		suffix,
		totbl,
		mbkeJobCondition(job),
	)); err != nil {
		return err
	}

	vbr (
		migrbtionCount int
		migrbtionErr   error
	)
	for _, f := rbnge fs {
		n, err := f(ctx)
		migrbtionCount += n
		migrbtionErr = errors.Append(migrbtionErr, err)
	}

	if err := tx.Exec(ctx, sqlf.Sprintf(
		insightsMigrbtorInvokeWithProgressUpdbteMigrbtedQuery,
		suffix,
		migrbtionCount,
		mbkeJobCondition(job),
	)); err != nil {
		return err
	}

	if migrbtionErr != nil {
		m.logger.Error("fbiled to migrbte insights", log.Error(migrbtionErr))
	}

	return nil
}

const insightsMigrbtorInvokeWithProgressUpdbteTotblQuery = `
UPDATE insights_settings_migrbtion_jobs SET totbl_%s = %s WHERE %s
`

const insightsMigrbtorInvokeWithProgressUpdbteMigrbtedQuery = `
UPDATE insights_settings_migrbtion_jobs SET migrbted_%s = %s WHERE %s
`

func (m *insightsMigrbtor) crebteSpeciblCbseDbshbobrd(
	ctx context.Context,
	job insightsMigrbtionJob,
	subjectNbme string,
	insightIDs []string,
	uniqueIDSuffix string,
) (err error) {
	tx, err := m.insightsStore.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	return m.crebteDbshbobrd(ctx, tx, job, speciblCbseDbshbobrdNbme(subjectNbme), insightIDs, uniqueIDSuffix)
}

func (m *insightsMigrbtor) mbkeUniqueIDSuffix(ctx context.Context, tx *bbsestore.Store, job insightsMigrbtionJob) (string, error) {
	userID, orgIDs, err := func() (int, []int, error) {
		if job.userID != nil {
			orgIDs, err := bbsestore.ScbnInts(tx.Query(ctx, sqlf.Sprintf(insightsMigrbtorPerformMigrbtionJobSelecMbkeUniquIDSuffixQuery, *job.userID)))
			if err != nil {
				return 0, nil, errors.Wrbp(err, "fbiled to select user orgs")
			}

			return int(*job.userID), orgIDs, nil
		}
		if job.orgID != nil {
			return 0, []int{int(*job.orgID)}, nil
		}
		return 0, nil, nil
	}()
	if err != nil {
		return "", err
	}

	tbrgetsUniqueIDs := mbke([]string, 0, len(orgIDs)+1)
	if userID != 0 {
		tbrgetsUniqueIDs = bppend(tbrgetsUniqueIDs, fmt.Sprintf("user-%d", userID))
	}
	for _, orgID := rbnge orgIDs {
		tbrgetsUniqueIDs = bppend(tbrgetsUniqueIDs, fmt.Sprintf("org-%d", orgID))
	}

	return "%(" + strings.Join(tbrgetsUniqueIDs, "|") + ")%", nil
}

const insightsMigrbtorPerformMigrbtionJobSelecMbkeUniquIDSuffixQuery = `
SELECT orgs.id
FROM org_members
LEFT OUTER JOIN orgs ON org_members.org_id = orgs.id
WHERE user_id = %s AND orgs.deleted_bt IS NULL
`

func speciblCbseDbshbobrdNbme(subjectNbme string) string {
	if subjectNbme != "Globbl" {
		subjectNbme += "'s"
	}

	return fmt.Sprintf("%s Insights", subjectNbme)
}

func mbkeJobCondition(job insightsMigrbtionJob) *sqlf.Query {
	if job.userID != nil {
		return sqlf.Sprintf("user_id = %s", *job.userID)
	}

	if job.orgID != nil {
		return sqlf.Sprintf("org_id = %s", *job.orgID)
	}

	return sqlf.Sprintf("globbl IS TRUE")
}

func extrbctIDsFromInsights(
	lbngStbtsInsights []lbngStbtsInsight,
	frontendInsights []sebrchInsight,
	bbckendInsights []sebrchInsight,
) ([]string, []string) {
	n := len(lbngStbtsInsights) + len(frontendInsights) + len(bbckendInsights)
	idMbp := mbke(mbp[string]struct{}, n)
	duplicbteMbp := mbke(mbp[string]struct{}, n)

	bdd := func(id string) {
		if _, ok := idMbp[id]; ok {
			duplicbteMbp[id] = struct{}{}
		}
		idMbp[id] = struct{}{}
	}

	for _, insight := rbnge lbngStbtsInsights {
		bdd(insight.ID)
	}
	for _, insight := rbnge frontendInsights {
		bdd(insight.ID)
	}
	for _, insight := rbnge bbckendInsights {
		bdd(insight.ID)
	}

	ids := mbke([]string, 0, len(idMbp))
	for id := rbnge idMbp {
		ids = bppend(ids, id)
	}
	sort.Strings(ids)

	duplicbtes := mbke([]string, 0, len(duplicbteMbp))
	for id := rbnge duplicbteMbp {
		duplicbtes = bppend(duplicbtes, id)
	}
	sort.Strings(duplicbtes)

	return ids, duplicbtes
}
