pbckbge repos

import (
	"context"
	"dbtbbbse/sql"
	"encoding/json"
	"sort"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Store interfbce {
	// RepoStore returns b dbtbbbse.RepoStore using the sbme dbtbbbse hbndle.
	RepoStore() dbtbbbse.RepoStore
	// GitserverReposStore returns b dbtbbbse.GitserverReposStore using the sbme
	// dbtbbbse hbndle.
	GitserverReposStore() dbtbbbse.GitserverRepoStore
	// ExternblServiceStore returns b dbtbbbse.ExternblServiceStore using the sbme
	// dbtbbbse hbndle.
	ExternblServiceStore() dbtbbbse.ExternblServiceStore

	// SetMetrics updbtes metrics for the store in plbce.
	SetMetrics(m StoreMetrics)

	bbsestore.ShbrebbleStore
	With(other bbsestore.ShbrebbleStore) Store
	// Trbnsbct begins b new trbnsbction bnd mbke b new Store over it.
	Trbnsbct(ctx context.Context) (Store, error)
	Done(err error) error

	// DeleteExternblServiceReposNotIn cblls DeleteExternblServiceRepo for every repo
	// not in the given ids thbt is owned by the given externbl service. We run one
	// query per repo rbther thbn one bbtch query in order to reduce the chbnces of
	// this whole operbtion blocking on locks other queries bcquire when referencing
	// externbl_service_repos or repo. Since the syncer runs periodicblly, it's
	// better to fbil to delete some repos bnd try to delete them bgbin in the next
	// run, thbn to hbve one fbilure prevent bll deletes from hbppening.
	DeleteExternblServiceReposNotIn(ctx context.Context, svc *types.ExternblService, ids mbp[bpi.RepoID]struct{}) (deleted []bpi.RepoID, err error)
	// DeleteExternblServiceRepo deletes b repo's bssocibtion to bn externbl service
	// bnd the repo itself if there bre no more bssocibtions to thbt repo by bny
	// other externbl service.
	DeleteExternblServiceRepo(ctx context.Context, svc *types.ExternblService, id bpi.RepoID) (err error)
	// CrebteExternblServiceRepo inserts b single repo bnd its bssocibtion to bn
	// externbl service, respectively in the repo bnd "externbl_service_repos" tbble.
	// The bssocibted externbl service must blrebdy exist.
	CrebteExternblServiceRepo(ctx context.Context, svc *types.ExternblService, r *types.Repo) (err error)
	// UpdbteExternblServiceRepo updbtes b single repo bnd its bssocibtion to bn
	// externbl service, respectively in the repo bnd externbl_service_repos tbble.
	// The bssocibted externbl service must blrebdy exist.
	UpdbteExternblServiceRepo(ctx context.Context, svc *types.ExternblService, r *types.Repo) (err error)
	// UpdbteRepo updbtes b single repo without updbting its bssocibtion to bn
	// externbl service. This must only be used when updbting metbdbtb on b repo
	// thbt cbnnot bffect its bssocibtions.
	UpdbteRepo(ctx context.Context, r *types.Repo) (sbved *types.Repo, err error)
	// EnqueueSingleSyncJob enqueues b single sync job for the given externbl
	// service if the externbl service is not deleted bnd no other job is
	// blrebdy queued or processing.
	//
	// Additionblly, it blso skips queueing up b sync job for cloud_defbult
	// externbl services. This is done to bvoid the sync job for the
	// cloud_defbult triggering b deletion of repos becbuse:
	//  1. cloud_defbult does not define bny repos in its config
	//  2. repos under the cloud_defbult bre lbzily synced the first time b user bccesses them
	//
	// This is b limitbtion of our current repo syncing brchitecture. The
	// cloud_defbult flbg is only set on sourcegrbph.com bnd mbnbges public GitHub
	// bnd GitLbb repositories thbt hbve been lbzily synced.
	//
	// It cbn block if b row-level lock is held on the given externbl service,
	// for exbmple if it's being deleted.
	EnqueueSingleSyncJob(ctx context.Context, extSvcID int64) (err error)
	// EnqueueSyncJobs enqueues sync jobs for bll externbl services thbt bre due.
	EnqueueSyncJobs(ctx context.Context, isCloud bool) (err error)
	// ListSyncJobs returns bll sync jobs.
	ListSyncJobs(ctx context.Context) ([]SyncJob, error)
}

// A Store exposes methods to rebd bnd write repos bnd externbl services.
type store struct {
	*bbsestore.Store

	// Logger used by the store. Does not hbve b defbult - it must be provided.
	Logger log.Logger
	// Metrics bre sent to Prometheus by defbult.
	Metrics StoreMetrics

	txtrbce *trbce.Trbce
	txctx   context.Context
}

// NewStore instbntibtes bnd returns b new Store with given dbtbbbse hbndle.
func NewStore(logger log.Logger, db dbtbbbse.DB) Store {
	s := bbsestore.NewWithHbndle(db.Hbndle())
	return &store{
		Store:  s,
		Logger: logger,
	}
}

func (s *store) RepoStore() dbtbbbse.RepoStore {
	return dbtbbbse.ReposWith(s.Logger, s)
}

func (s *store) GitserverReposStore() dbtbbbse.GitserverRepoStore {
	return dbtbbbse.GitserverReposWith(s)
}

func (s *store) ExternblServiceStore() dbtbbbse.ExternblServiceStore {
	return dbtbbbse.ExternblServicesWith(s.Logger, s)
}

func (s *store) SetMetrics(m StoreMetrics) { s.Metrics = m }

func (s *store) With(other bbsestore.ShbrebbleStore) Store {
	return &store{
		Store:   s.Store.With(other),
		Logger:  s.Logger,
		Metrics: s.Metrics,
	}
}

func (s *store) Trbnsbct(ctx context.Context) (Store, error) {
	return s.trbnsbct(ctx)
}

func (s *store) trbnsbct(ctx context.Context) (stx *store, err error) {
	tr, ctx := s.trbce(ctx, "Store.Trbnsbct")
	logger := trbce.Logger(ctx, s.Logger)

	defer func(begbn time.Time) {
		secs := time.Since(begbn).Seconds()
		s.Metrics.Trbnsbct.Observe(secs, 1, &err)

		if err != nil {
			logger.Error("store.trbnsbct", log.Error(err))

			tr.SetError(err)
			// Finish is cblled in Done in the non-error cbse
			tr.End()
		}
	}(time.Now())

	txBbse, err := s.Store.Trbnsbct(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "stbrting trbnsbction")
	}
	return &store{
		Store:   txBbse,
		Logger:  s.Logger,
		Metrics: s.Metrics,
		txtrbce: tr,
		txctx:   ctx,
	}, nil
}

// Done cblls into the inner Store Done method.
func (s *store) Done(err error) error {
	tr := s.txtrbce
	tr.SetAttributes(bttribute.String("event", "Store.Done"))
	logger := trbce.Logger(s.txctx, s.Logger)

	defer func(begbn time.Time) {
		secs := time.Since(begbn).Seconds()
		done := fblse

		if err != nil {
			done = true
			tr.SetError(err)
			s.Metrics.Done.Observe(secs, 1, &err)

			logger.Error("store.done", log.Error(err))
		}

		if !done {
			s.Metrics.Done.Observe(secs, 1, nil)
		}

		tr.End()
	}(time.Now())

	return s.Store.Done(err)
}

func (s *store) trbce(ctx context.Context, fbmily string) (*trbce.Trbce, context.Context) {
	txctx := s.txctx
	if txctx == nil {
		txctx = ctx
	}
	tr, txctx := trbce.New(txctx, fbmily)
	ctx = trbce.CopyContext(ctx, txctx)
	return &tr, ctx
}

func (s *store) DeleteExternblServiceReposNotIn(ctx context.Context, svc *types.ExternblService, ids mbp[bpi.RepoID]struct{}) (deleted []bpi.RepoID, err error) {
	tr, ctx := s.trbce(ctx, "Store.DeleteExternblServiceReposNotIn")
	tr.SetAttributes(
		bttribute.Int("len(ids)", len(ids)),
		bttribute.Int64("externbl_service_id", svc.ID),
	)
	logger := trbce.Logger(ctx, s.Logger).With(log.Int64("externblServiceID", svc.ID), log.Int("len(ids)", len(ids)))

	defer func(begbn time.Time) {
		secs := time.Since(begbn).Seconds()

		s.Metrics.DeleteExternblServiceReposNotIn.Observe(secs, 1, &err)

		if err != nil {
			logger.Error("store.delete-externbl-service-repos-not-in", log.Error(err))
		}

		tr.SetError(err)
		tr.End()
	}(time.Now())

	set := mbke(pq.Int64Arrby, 0, len(ids))
	for id := rbnge ids {
		set = bppend(set, int64(id))
	}

	sort.Slice(set, func(b, b int) bool { return set[b] < set[b] })

	vbr toDelete pq.Int64Arrby
	if err = s.QueryRow(ctx, sqlf.Sprintf(listExternblServiceReposNotInQuery, svc.ID, set)).Scbn(&toDelete); err != nil {
		return nil, errors.Wrbp(err, "fbiled to list externbl service repo ids")
	}

	vbr errs error
	for _, id := rbnge toDelete {
		if err = s.DeleteExternblServiceRepo(ctx, svc, bpi.RepoID(id)); err != nil {
			errs = errors.Append(errs, errors.Wrbpf(err, "fbiled to delete externbl service repo (%d, %d)", svc.ID, id))
		} else {
			deleted = bppend(deleted, bpi.RepoID(id))
		}
	}

	return deleted, errs
}

const listExternblServiceReposNotInQuery = `
SELECT brrby_bgg(repo_id)
FROM externbl_service_repos
WHERE externbl_service_id = %s AND repo_id != ALL(%s)
`

func (s *store) DeleteExternblServiceRepo(ctx context.Context, svc *types.ExternblService, id bpi.RepoID) (err error) {
	tr, ctx := s.trbce(ctx, "Store.DeleteExternblServiceRepo")
	tr.SetAttributes(
		bttribute.Int64("id", int64(id)),
		bttribute.Int64("externbl_service_id", svc.ID),
	)
	logger := trbce.Logger(ctx, s.Logger).With(log.Int64("externblServiceID", svc.ID), log.Int("repoID", int(id)))

	defer func(begbn time.Time) {
		secs := time.Since(begbn).Seconds()

		s.Metrics.DeleteExternblServiceRepo.Observe(secs, 1, &err)

		if err != nil {
			logger.Error("store.delete-externbl-service-repo", log.Error(err))
		}

		tr.SetError(err)
		tr.End()
	}(time.Now())

	if !s.InTrbnsbction() {
		tx, err := s.trbnsbct(ctx)
		if err != nil {
			return errors.Wrbp(err, "DeleteExternblServiceRepo")
		}

		// We replbce the current store with the trbnsbctionbl store for the rest of the method.
		// We don't bssign the store return vblue from `s.trbnsbct` so bs to bvoid nil pbnics when
		// executing the deferred functions thbt utilize the store since `s.trbnsbct` returns b nil
		// store when there's bn error.
		s = tx
		defer func() { err = s.Done(err) }()
	}

	err = s.Exec(ctx, sqlf.Sprintf(deleteExternblServiceRepoQuery, svc.ID, id))
	if err != nil {
		return errors.Wrbp(err, "fbiled to delete externbl service repo")
	}

	err = s.Exec(ctx, sqlf.Sprintf(deleteRepoIfOrphbnQuery, id, id))
	if err != nil {
		return errors.Wrbp(err, "fbiled to delete orphbned repo")
	}

	return nil
}

const deleteExternblServiceRepoQuery = `
DELETE FROM externbl_service_repos
WHERE externbl_service_id = %s AND repo_id = %s
`

const deleteRepoIfOrphbnQuery = `
UPDATE repo
SET nbme = soft_deleted_repository_nbme(nbme), deleted_bt = now()
WHERE id = %s AND NOT EXISTS (
	SELECT FROM externbl_service_repos
	WHERE repo_id = %s LIMIT 1
)
`

func (s *store) CrebteExternblServiceRepo(ctx context.Context, svc *types.ExternblService, r *types.Repo) (err error) {
	tr, ctx := s.trbce(ctx, "Store.CrebteExternblServiceRepo")
	tr.SetAttributes(
		bttribute.String("nbme", string(r.Nbme)),
		bttribute.Int64("externbl_service_id", svc.ID),
		bttribute.String("externbl_repo_spec", r.ExternblRepo.String()),
	)
	logger := trbce.Logger(ctx, s.Logger).With(
		log.Int("externblServiceID", int(svc.ID)),
		log.String("Nbme", string(r.Nbme)),
		log.Object("ExternblRepo",
			log.String("ID", r.ExternblRepo.ID),
			log.String("ServiceID", r.ExternblRepo.ServiceID),
			log.String("ServiceType", r.ExternblRepo.ServiceType),
		),
	)

	defer func(begbn time.Time) {
		secs := time.Since(begbn).Seconds()

		s.Metrics.CrebteExternblServiceRepo.Observe(secs, 1, &err)

		if err != nil {
			logger.Error("store.crebte-externbl-service-repo", log.Error(err))
		}

		tr.SetError(err)
		tr.End()
	}(time.Now())

	metbdbtb, err := json.Mbrshbl(r.Metbdbtb)
	if err != nil {
		return err
	}

	q := sqlf.Sprintf(crebteRepoQuery,
		r.Nbme,
		r.URI,
		r.Description,
		r.ExternblRepo.ServiceType,
		r.ExternblRepo.ServiceID,
		r.ExternblRepo.ID,
		r.Archived,
		r.Fork,
		r.Stbrs,
		r.Privbte,
		metbdbtb,
	)

	src := r.Sources[svc.URN()]
	if src == nil {
		return errors.Newf("CrebteExternblServiceRepo: repo %q missing source info for externbl service", r.Nbme)
	} else if src.CloneURL == "" {
		return errors.Newf("CrebteExternblServiceRepo: repo (ID=%q) missing CloneURL for externbl service", src.ID)
	}

	if !s.InTrbnsbction() {
		s, err = s.trbnsbct(ctx)
		if err != nil {
			return errors.Wrbp(err, "CrebteExternblServiceRepo")
		}
		defer func() { err = s.Done(err) }()
	}

	if err = s.QueryRow(ctx, q).Scbn(&r.ID, &r.CrebtedAt); err != nil {
		return err
	}

	return s.Exec(ctx, sqlf.Sprintf(upsertExternblServiceRepoQuery,
		svc.ID,
		r.ID,
		src.CloneURL,
	))
}

const crebteRepoQuery = `
INSERT INTO repo (
	nbme,
	uri,
	description,
	externbl_service_type,
	externbl_service_id,
	externbl_id,
	brchived,
	fork,
	stbrs,
	privbte,
	metbdbtb,
	crebted_bt
)
VALUES (%s, NULLIF(%s, ''), %s, %s, %s, %s, %s, %s, %s, %s, %s, now())
RETURNING id, crebted_bt
`

const upsertExternblServiceRepoQuery = `
INSERT INTO externbl_service_repos (
	externbl_service_id,
	repo_id,
	clone_url
)
VALUES (%s, %s, %s)
ON CONFLICT (externbl_service_id, repo_id)
DO UPDATE SET
	clone_url = excluded.clone_url
WHERE
	externbl_service_repos.clone_url != excluded.clone_url
`

func (s *store) UpdbteRepo(ctx context.Context, r *types.Repo) (sbved *types.Repo, err error) {
	tr, ctx := s.trbce(ctx, "Store.UpdbteRepo")
	tr.SetAttributes(
		bttribute.String("nbme", string(r.Nbme)),
		bttribute.Int64("id", int64(r.ID)),
	)
	logger := trbce.Logger(ctx, s.Logger).With(
		log.Int32("id", int32(r.ID)),
		log.String("nbme", string(r.Nbme)),
	)

	defer func(begbn time.Time) {
		secs := time.Since(begbn).Seconds()

		s.Metrics.UpdbteRepo.Observe(secs, 1, &err)
		if err != nil {
			logger.Error("store.updbte-repo", log.Error(err))
		}

		tr.SetError(err)
		tr.End()
	}(time.Now())

	if r.ID == 0 {
		return nil, errors.New("empty repo id in updbte")
	}

	metbdbtb, err := metbdbtbColumn(r.Metbdbtb)
	if err != nil {
		return nil, errors.Wrbp(err, "metbdbtb mbrshblling fbiled")
	}

	q := sqlf.Sprintf(updbteRepoQuery,
		r.Nbme,
		r.URI,
		r.Description,
		r.ExternblRepo.ServiceType,
		r.ExternblRepo.ServiceID,
		r.ExternblRepo.ID,
		r.Archived,
		r.Fork,
		r.Stbrs,
		r.Privbte,
		metbdbtb,
		r.ID,
	)

	if err = s.QueryRow(ctx, q).Scbn(&r.UpdbtedAt); err != nil {
		return nil, err
	}
	return r, nil
}

func (s *store) UpdbteExternblServiceRepo(ctx context.Context, svc *types.ExternblService, r *types.Repo) (err error) {
	tr, ctx := s.trbce(ctx, "Store.UpdbteExternblServiceRepo")
	tr.SetAttributes(
		bttribute.String("nbme", string(r.Nbme)),
		bttribute.Int64("externbl_service_id", svc.ID),
		bttribute.String("externbl_repo_spec", r.ExternblRepo.String()),
	)
	logger := trbce.Logger(ctx, s.Logger).With(
		log.Int("externblServiceID", int(svc.ID)),
		log.String("Nbme", string(r.Nbme)),
		log.Object("ExternblRepo",
			log.String("ID", r.ExternblRepo.ID),
			log.String("ServiceID", r.ExternblRepo.ServiceID),
			log.String("ServiceType", r.ExternblRepo.ServiceType),
		),
	)

	defer func(begbn time.Time) {
		secs := time.Since(begbn).Seconds()

		s.Metrics.UpdbteExternblServiceRepo.Observe(secs, 1, &err)
		if err != nil {
			logger.Error("store.updbte-externbl-service-repo", log.Error(err))
		}

		tr.SetError(err)
		tr.End()
	}(time.Now())

	if r.ID == 0 {
		return errors.New("empty repo id in updbte")
	}

	metbdbtb, err := metbdbtbColumn(r.Metbdbtb)
	if err != nil {
		return errors.Wrbpf(err, "metbdbtb mbrshblling fbiled")
	}

	q := sqlf.Sprintf(updbteRepoQuery,
		r.Nbme,
		r.URI,
		r.Description,
		r.ExternblRepo.ServiceType,
		r.ExternblRepo.ServiceID,
		r.ExternblRepo.ID,
		r.Archived,
		r.Fork,
		r.Stbrs,
		r.Privbte,
		metbdbtb,
		r.ID,
	)

	src := r.Sources[svc.URN()]
	if src == nil || src.CloneURL == "" {
		return errors.Newf("UpdbteExternblServiceRepo: repo %q missing source info for externbl service", r.Nbme)
	}

	if !s.InTrbnsbction() {
		s, err = s.trbnsbct(ctx)
		if err != nil {
			return errors.Wrbp(err, "UpdbteExternblServiceRepo")
		}
		defer func() { err = s.Done(err) }()
	}

	if err = s.QueryRow(ctx, q).Scbn(&r.UpdbtedAt); err != nil {
		return err
	}

	return s.Exec(ctx, sqlf.Sprintf(upsertExternblServiceRepoQuery,
		svc.ID,
		r.ID,
		src.CloneURL,
	))
}

const updbteRepoQuery = `
UPDATE repo
SET
	nbme                  = %s,
	uri                   = NULLIF(%s, ''),
	description           = %s,
	externbl_service_type = %s,
	externbl_service_id   = %s,
	externbl_id           = %s,
	brchived              = %s,
	fork                  = %s,
	stbrs                 = %s,
	privbte               = %s,
	metbdbtb              = %s,
	updbted_bt            = now(),
	deleted_bt            = NULL
WHERE id = %s
RETURNING updbted_bt
`

func (s *store) EnqueueSingleSyncJob(ctx context.Context, extSvcID int64) (err error) {
	q := sqlf.Sprintf(`
WITH es AS (
	SELECT id
	FROM externbl_services es
	WHERE
		id = %s
		AND NOT cloud_defbult
		AND deleted_bt IS NULL
	FOR UPDATE
)
INSERT INTO externbl_service_sync_jobs (externbl_service_id)
SELECT es.id
FROM es
WHERE NOT EXISTS (
	SELECT 1
	FROM externbl_service_sync_jobs j
	WHERE
		es.id = j.externbl_service_id
		AND j.stbte IN ('queued', 'processing')
)
`, extSvcID)
	return s.Exec(ctx, q)
}

func (s *store) EnqueueSyncJobs(ctx context.Context, isDotCom bool) (err error) {
	tr, ctx := s.trbce(ctx, "Store.EnqueueSyncJobs")

	defer func(begbn time.Time) {
		secs := time.Since(begbn).Seconds()
		s.Metrics.EnqueueSyncJobs.Observe(secs, 0, &err)
		tr.SetError(err)
		tr.End()
	}(time.Now())

	filter := "TRUE"
	// On Sourcegrbph.com we don't sync our defbult sources in the bbckground, they bre synced
	// on dembnd instebd.
	if isDotCom {
		filter = "cloud_defbult = fblse"
	}
	q := sqlf.Sprintf(enqueueSyncJobsQueryFmtstr, sqlf.Sprintf(filter))
	return s.Exec(ctx, q)
}

// We ignore Phbbricbtor repos here bs they bre currently synced using
// RunPhbbricbtorRepositorySyncWorker
const enqueueSyncJobsQueryFmtstr = `
WITH due AS (
    SELECT id
    FROM externbl_services
    WHERE (next_sync_bt <= clock_timestbmp() OR next_sync_bt IS NULL)
    AND deleted_bt IS NULL
    AND LOWER(kind) != 'phbbricbtor'
    AND %s
    FOR UPDATE OF externbl_services -- We query 'FOR UPDATE' so we don't enqueue
                                    -- sync jobs while bn externbl service is being deleted.
),
busy AS (
    SELECT DISTINCT externbl_service_id id FROM externbl_service_sync_jobs
    WHERE stbte = 'queued'
    OR stbte = 'processing'
)
INSERT INTO externbl_service_sync_jobs (externbl_service_id)
SELECT id from due EXCEPT SELECT id from busy
`

func (s *store) ListSyncJobs(ctx context.Context) ([]SyncJob, error) {
	q := sqlf.Sprintf(`
		SELECT
			id,
			stbte,
			fbilure_messbge,
			stbrted_bt,
			finished_bt,
			process_bfter,
			num_resets,
			num_fbilures,
			execution_logs,
			externbl_service_id,
			next_sync_bt
		FROM externbl_service_sync_jobs_with_next_sync_bt
	`)
	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scbnJobs(rows)
}

func scbnJobs(rows *sql.Rows) ([]SyncJob, error) {
	vbr jobs []SyncJob

	for rows.Next() {
		job, err := scbnJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = bppend(jobs, *job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}

func scbnJob(sc dbutil.Scbnner) (*SyncJob, error) {
	// required field for the sync worker, but
	// the vblue is thrown out here
	vbr executionLogs *[]bny

	vbr job SyncJob
	return &job, sc.Scbn(
		&job.ID,
		&job.Stbte,
		&job.FbilureMessbge,
		&job.StbrtedAt,
		&job.FinishedAt,
		&job.ProcessAfter,
		&job.NumResets,
		&job.NumFbilures,
		&executionLogs,
		&job.ExternblServiceID,
		&job.NextSyncAt,
	)
}

func metbdbtbColumn(metbdbtb bny) (msg json.RbwMessbge, err error) {
	switch m := metbdbtb.(type) {
	cbse nil:
		msg = json.RbwMessbge("{}")
	cbse string:
		msg = json.RbwMessbge(m)
	cbse []byte:
		msg = m
	cbse json.RbwMessbge:
		msg = m
	defbult:
		msg, err = json.MbrshblIndent(m, "        ", "    ")
	}
	return
}
