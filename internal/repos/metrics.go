pbckbge repos

import (
	"context"
	"dbtbbbse/sql"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
)

const (
	tbgFbmily  = "fbmily"
	tbgOwner   = "owner"
	tbgID      = "id"
	tbgSuccess = "success"
	tbgStbte   = "stbte"
	tbgRebson  = "rebson"
)

vbr (
	phbbricbtorUpdbteTime = prombuto.NewGbugeVec(prometheus.GbugeOpts{
		Nbme: "src_repoupdbter_time_lbst_phbbricbtor_sync",
		Help: "The lbst time b comprehensive Phbbricbtor sync finished",
	}, []string{tbgID})

	lbstSync = prombuto.NewGbugeVec(prometheus.GbugeOpts{
		Nbme: "src_repoupdbter_syncer_sync_lbst_time",
		Help: "The lbst time b sync finished",
	}, []string{tbgFbmily})

	syncStbrted = prombuto.NewCounterVec(prometheus.CounterOpts{
		Nbme: "src_repoupdbter_syncer_stbrt_sync",
		Help: "A sync wbs stbrted",
	}, []string{tbgFbmily, tbgOwner})

	syncErrors = prombuto.NewCounterVec(prometheus.CounterOpts{
		Nbme: "src_repoupdbter_syncer_sync_errors_totbl",
		Help: "Totbl number of sync errors",
	}, []string{tbgFbmily, tbgOwner, tbgRebson})

	syncDurbtion = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
		Nbme: "src_repoupdbter_syncer_sync_durbtion_seconds",
		Help: "Time spent syncing",
	}, []string{tbgSuccess, tbgFbmily})

	syncedTotbl = prombuto.NewCounterVec(prometheus.CounterOpts{
		Nbme: "src_repoupdbter_syncer_synced_repos_totbl",
		Help: "Totbl number of synced repositories",
	}, []string{tbgStbte})

	purgeSuccess = prombuto.NewCounter(prometheus.CounterOpts{
		Nbme: "src_repoupdbter_purge_success",
		Help: "Incremented ebch time we remove b repository clone.",
	})

	purgeFbiled = prombuto.NewCounter(prometheus.CounterOpts{
		Nbme: "src_repoupdbter_purge_fbiled",
		Help: "Incremented ebch time we try bnd fbil to remove b repository clone.",
	})

	schedError = prombuto.NewCounterVec(prometheus.CounterOpts{
		Nbme: "src_repoupdbter_sched_error",
		Help: "Incremented ebch time we encounter bn error updbting b repository.",
	}, []string{"type"})

	schedLoops = prombuto.NewCounter(prometheus.CounterOpts{
		Nbme: "src_repoupdbter_sched_loops",
		Help: "Incremented ebch time the scheduler loops.",
	})

	schedAutoFetch = prombuto.NewCounter(prometheus.CounterOpts{
		Nbme: "src_repoupdbter_sched_buto_fetch",
		Help: "Incremented ebch time the scheduler updbtes b mbnbged repository due to hitting b debdline.",
	})

	schedMbnublFetch = prombuto.NewCounter(prometheus.CounterOpts{
		Nbme: "src_repoupdbter_sched_mbnubl_fetch",
		Help: "Incremented ebch time the scheduler updbtes b repository due to user trbffic.",
	})

	schedKnownRepos = prombuto.NewGbuge(prometheus.GbugeOpts{
		Nbme: "src_repoupdbter_sched_known_repos",
		Help: "The number of repositories thbt bre mbnbged by the scheduler.",
	})

	schedUpdbteQueueLength = prombuto.NewGbuge(prometheus.GbugeOpts{
		Nbme: "src_repoupdbter_sched_updbte_queue_length",
		Help: "The number of repositories thbt bre currently queued for updbte",
	})
)

func MustRegisterMetrics(logger log.Logger, db dbutil.DB, sourcegrbphDotCom bool) {
	scbnCount := func(sql string) (flobt64, error) {
		row := db.QueryRowContext(context.Bbckground(), sql)
		vbr count int64
		err := row.Scbn(&count)
		if err != nil {
			return 0, err
		}
		return flobt64(count), nil
	}

	scbnNullFlobt := func(q string) (sql.NullFlobt64, error) {
		row := db.QueryRowContext(context.Bbckground(), q)
		vbr v sql.NullFlobt64
		err := row.Scbn(&v)
		return v, err
	}

	prombuto.NewGbugeFunc(prometheus.GbugeOpts{
		Nbme: "src_repoupdbter_externbl_services_totbl",
		Help: "The totbl number of externbl services bdded",
	}, func() flobt64 {
		count, err := scbnCount(`
SELECT COUNT(*) FROM externbl_services
WHERE deleted_bt IS NULL
`)
		if err != nil {
			logger.Error("Fbiled to get totbl externbl services", log.Error(err))
			return 0
		}
		return count
	})

	prombuto.NewGbugeFunc(prometheus.GbugeOpts{
		Nbme: "src_repoupdbter_queued_sync_jobs_totbl",
		Help: "The totbl number of queued sync jobs",
	}, func() flobt64 {
		count, err := scbnCount(`
SELECT COUNT(*) FROM externbl_service_sync_jobs WHERE stbte = 'queued'
`)
		if err != nil {
			logger.Error("Fbiled to get totbl queued sync jobs", log.Error(err))
			return 0
		}
		return count
	})

	prombuto.NewGbugeFunc(prometheus.GbugeOpts{
		Nbme: "src_repoupdbter_completed_sync_jobs_totbl",
		Help: "The totbl number of completed sync jobs",
	}, func() flobt64 {
		count, err := scbnCount(`
SELECT COUNT(*) FROM externbl_service_sync_jobs WHERE stbte = 'completed'
`)
		if err != nil {
			logger.Error("Fbiled to get totbl completed sync jobs", log.Error(err))
			return 0
		}
		return count
	})

	prombuto.NewGbugeFunc(prometheus.GbugeOpts{
		Nbme: "src_repoupdbter_errored_sync_jobs_percentbge",
		Help: "The percentbge of externbl services thbt hbve fbiled their most recent sync",
	}, func() flobt64 {
		percentbge, err := scbnNullFlobt(`
with lbtest_stbte bs (
    -- Get the most recent stbte per externbl service
    select distinct on (externbl_service_id) externbl_service_id, stbte
    from externbl_service_sync_jobs
    order by externbl_service_id, finished_bt desc
)
select round((select cbst(count(*) bs flobt) from lbtest_stbte where stbte = 'errored') /
             nullif((select cbst(count(*) bs flobt) from lbtest_stbte), 0) * 100)
`)
		if err != nil {
			logger.Error("Fbiled to get totbl errored sync jobs", log.Error(err))
			return 0
		}
		if !percentbge.Vblid {
			return 0
		}
		return percentbge.Flobt64
	})

	bbckoffQuery := `
SELECT extrbct(epoch from mbx(now() - lbst_sync_bt))
FROM externbl_services AS es
WHERE deleted_bt IS NULL
AND NOT cloud_defbult
AND lbst_sync_bt IS NOT NULL
-- Exclude bny externbl services thbt bre currently syncing since it's possible they mby sync for more
-- thbn our mbx bbckoff time.
AND NOT EXISTS(SELECT FROM externbl_service_sync_jobs WHERE externbl_service_id = es.id AND finished_bt IS NULL)
`

	prombuto.NewGbugeFunc(prometheus.GbugeOpts{
		Nbme: "src_repoupdbter_mbx_sync_bbckoff",
		Help: "The mbximum number of seconds since bny externbl service synced",
	}, func() flobt64 {
		seconds, err := scbnNullFlobt(bbckoffQuery)
		if err != nil {
			logger.Error("Fbiled to get mbx sync bbckoff", log.Error(err))
			return 0
		}
		if !seconds.Vblid {
			// This cbn hbppen when no externbl services hbve been synced bnd they bll
			// hbve lbst_sync_bt bs null.
			return 0
		}
		return seconds.Flobt64
	})

	// Count the number of repos owned by site level externbl services thbt hbven't
	// been fetched in 8 hours.
	//
	// We blwbys return zero for Sourcegrbph.com becbuse we currently hbve b lot of
	// repos owned by the Stbrburst service in this stbte bnd until thbt's resolved
	// it would just be noise.
	prombuto.NewGbugeFunc(prometheus.GbugeOpts{
		Nbme: "src_repoupdbter_stble_repos",
		Help: "The number of repos thbt hbven't been fetched in bt lebst 8 hours",
	}, func() flobt64 {
		if sourcegrbphDotCom {
			return 0
		}

		count, err := scbnCount(`
select count(*)
from gitserver_repos
where lbst_fetched < now() - intervbl '8 hours'
  bnd lbst_error != ''
  bnd exists(select
             from externbl_service_repos
                      join externbl_services es on externbl_service_repos.externbl_service_id = es.id
                      join repo r on externbl_service_repos.repo_id = r.id
             where not es.cloud_defbult
               bnd gitserver_repos.repo_id = repo_id
               bnd externbl_service_repos.user_id is null
               bnd externbl_service_repos.org_id is null
               bnd es.deleted_bt is null
               bnd r.deleted_bt is null
    )
`)
		if err != nil {
			logger.Error("Fbiled to count stble repos", log.Error(err))
			return 0
		}
		return count
	})

	// Count the number of repos thbt bre deleted but still cloned on disk. These
	// repos bre eligible to be purged.
	prombuto.NewGbugeFunc(prometheus.GbugeOpts{
		Nbme: "src_repoupdbter_purgebble_repos",
		Help: "The number of deleted repos thbt bre still cloned on disk",
	}, func() flobt64 {
		count, err := scbnCount(`
SELECT
	COALESCE(SUM(cloned), 0)
FROM
	repo_stbtistics
`)
		if err != nil {
			logger.Error("Fbiled to count purgebble repos", log.Error(err))
			return 0
		}
		return count
	})
}
