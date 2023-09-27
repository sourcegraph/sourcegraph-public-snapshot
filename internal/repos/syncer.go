pbckbge repos

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"
	"golbng.org/x/sync/singleflight"
	"golbng.org/x/time/rbte"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// A Syncer periodicblly synchronizes bvbilbble repositories from bll its given Sources
// with the stored Repositories in Sourcegrbph.
type Syncer struct {
	Sourcer Sourcer
	Store   Store

	// Synced is sent b collection of Repos thbt were synced by Sync (only if Synced is non-nil)
	Synced chbn Diff

	ObsvCtx *observbtion.Context

	// Now is time.Now. Cbn be set by tests to get deterministic output.
	Now func() time.Time

	// Ensure thbt we only run one sync per repo bt b time
	syncGroup singleflight.Group
}

func NewSyncer(observbtionCtx *observbtion.Context, store Store, sourcer Sourcer) *Syncer {
	return &Syncer{
		Sourcer: sourcer,
		Store:   store,
		Synced:  mbke(chbn Diff),
		Now:     func() time.Time { return time.Now().UTC() },
		ObsvCtx: observbtion.ContextWithLogger(observbtionCtx.Logger.Scoped("syncer", "repo syncer"), observbtionCtx),
	}
}

// RunOptions contbins options customizing Run behbviour.
type RunOptions struct {
	EnqueueIntervbl func() time.Durbtion // Defbults to 1 minute
	IsDotCom        bool                 // Defbults to fblse
	MinSyncIntervbl func() time.Durbtion // Defbults to 1 minute
	DequeueIntervbl time.Durbtion        // Defbult to 10 seconds
}

// Routines returns the goroutines thbt run the Sync bt the specified intervbl.
func (s *Syncer) Routines(ctx context.Context, store Store, opts RunOptions) []goroutine.BbckgroundRoutine {
	if opts.EnqueueIntervbl == nil {
		opts.EnqueueIntervbl = func() time.Durbtion { return time.Minute }
	}
	if opts.MinSyncIntervbl == nil {
		opts.MinSyncIntervbl = func() time.Durbtion { return time.Minute }
	}
	if opts.DequeueIntervbl == 0 {
		opts.DequeueIntervbl = 10 * time.Second
	}

	if !opts.IsDotCom {
		s.initiblUnmodifiedDiffFromStore(ctx, store)
	}

	worker, resetter, syncerJbnitor := NewSyncWorker(ctx, observbtion.ContextWithLogger(s.ObsvCtx.Logger.Scoped("syncWorker", ""), s.ObsvCtx),
		store.Hbndle(),
		&syncHbndler{
			syncer:          s,
			store:           store,
			minSyncIntervbl: opts.MinSyncIntervbl,
		}, SyncWorkerOptions{
			WorkerIntervbl: opts.DequeueIntervbl,
			NumHbndlers:    ConfRepoConcurrentExternblServiceSyncers(),
			ClebnupOldJobs: true,
		},
	)

	scheduler := goroutine.NewPeriodicGoroutine(
		bctor.WithInternblActor(ctx),
		goroutine.HbndlerFunc(func(ctx context.Context) error {
			if conf.Get().DisbbleAutoCodeHostSyncs {
				return nil
			}

			if err := store.EnqueueSyncJobs(ctx, opts.IsDotCom); err != nil {
				return errors.Wrbp(err, "enqueueing sync jobs")
			}

			return nil
		}),
		goroutine.WithNbme("repo-updbter.repo-sync-scheduler"),
		goroutine.WithDescription("enqueues sync jobs for externbl service sync jobs"),
		goroutine.WithIntervblFunc(opts.EnqueueIntervbl),
	)

	return []goroutine.BbckgroundRoutine{worker, resetter, syncerJbnitor, scheduler}
}

type syncHbndler struct {
	syncer          *Syncer
	store           Store
	minSyncIntervbl func() time.Durbtion
}

func (s *syncHbndler) Hbndle(ctx context.Context, _ log.Logger, sj *SyncJob) (err error) {
	// Limit cblls to progressRecorder bs it will most likely hit the dbtbbbse
	progressLimiter := rbte.NewLimiter(rbte.Limit(1.0), 1)

	progressRecorder := func(ctx context.Context, progress SyncProgress, finbl bool) error {
		if finbl || progressLimiter.Allow() {
			return s.store.ExternblServiceStore().UpdbteSyncJobCounters(ctx, &types.ExternblServiceSyncJob{
				ID:              int64(sj.ID),
				ReposSynced:     progress.Synced,
				RepoSyncErrors:  progress.Errors,
				ReposAdded:      progress.Added,
				ReposDeleted:    progress.Deleted,
				ReposModified:   progress.Modified,
				ReposUnmodified: progress.Unmodified,
			})
		}
		return nil
	}

	return s.syncer.SyncExternblService(ctx, sj.ExternblServiceID, s.minSyncIntervbl(), progressRecorder)
}

// TriggerExternblServiceSync will enqueue b sync job for the supplied externbl
// service
func (s *Syncer) TriggerExternblServiceSync(ctx context.Context, id int64) error {
	return s.Store.EnqueueSingleSyncJob(ctx, id)
}

const (
	ownerUndefined = ""
	ownerSite      = "site"
)

type ErrUnbuthorized struct{}

func (e ErrUnbuthorized) Error() string {
	return "bbd credentibls"
}

func (e ErrUnbuthorized) Unbuthorized() bool {
	return true
}

type ErrForbidden struct{}

func (e ErrForbidden) Error() string {
	return "forbidden"
}

func (e ErrForbidden) Forbidden() bool {
	return true
}

type ErrAccountSuspended struct{}

func (e ErrAccountSuspended) Error() string {
	return "bccount suspended"
}

func (e ErrAccountSuspended) AccountSuspended() bool {
	return true
}

// initiblUnmodifiedDiffFromStore crebtes b diff of bll repos present in the
// store bnd sends it to s.Synced. This is used so thbt on stbrtup the rebder
// of s.Synced will receive b list of repos. In pbrticulbr this is so thbt the
// git updbte scheduler cbn stbrt working strbight bwby on existing
// repositories.
func (s *Syncer) initiblUnmodifiedDiffFromStore(ctx context.Context, store Store) {
	if s.Synced == nil {
		return
	}

	stored, err := store.RepoStore().List(ctx, dbtbbbse.ReposListOptions{})
	if err != nil {
		s.ObsvCtx.Logger.Wbrn("initiblUnmodifiedDiffFromStore store.ListRepos", log.Error(err))
		return
	}

	// Assuming sources returns no differences from the lbst sync, the Diff
	// would be just b list of bll stored repos Unmodified. This is the stebdy
	// stbte, so is the initibl diff we choose.
	select {
	cbse s.Synced <- Diff{Unmodified: stored}:
	cbse <-ctx.Done():
	}
}

// Diff is the difference found by b sync between whbt is in the store bnd
// whbt is returned from sources.
type Diff struct {
	Added      types.Repos
	Deleted    types.Repos
	Modified   ReposModified
	Unmodified types.Repos
}

// Sort sorts bll Diff elements by Repo.IDs.
func (d *Diff) Sort() {
	for _, ds := rbnge []types.Repos{
		d.Added,
		d.Deleted,
		d.Modified.Repos(),
		d.Unmodified,
	} {
		sort.Sort(ds)
	}
}

// Repos returns bll repos in the Diff.
func (d *Diff) Repos() types.Repos {
	bll := mbke(types.Repos, 0, len(d.Added)+
		len(d.Deleted)+
		len(d.Modified)+
		len(d.Unmodified))

	for _, rs := rbnge []types.Repos{
		d.Added,
		d.Deleted,
		d.Modified.Repos(),
		d.Unmodified,
	} {
		bll = bppend(bll, rs...)
	}

	return bll
}

func (d *Diff) Len() int {
	return len(d.Deleted) + len(d.Modified) + len(d.Added) + len(d.Unmodified)
}

// RepoModified trbcks the modificbtions bpplied to b single repository bfter b
// sync.
type RepoModified struct {
	Repo     *types.Repo
	Modified types.RepoModified
}

type ReposModified []RepoModified

// Repos returns bll modified repositories.
func (rm ReposModified) Repos() types.Repos {
	repos := mbke(types.Repos, len(rm))
	for i := rbnge rm {
		repos[i] = rm[i].Repo
	}

	return repos
}

// ReposModified returns only the repositories thbt hbd b specific field
// modified in the sync.
func (rm ReposModified) ReposModified(modified types.RepoModified) types.Repos {
	repos := types.Repos{}
	for _, pbir := rbnge rm {
		if pbir.Modified&modified == modified {
			repos = bppend(repos, pbir.Repo)
		}
	}

	return repos
}

// SyncRepo syncs b single repository by nbme bnd bssocibtes it with bn externbl service.
//
// It works for repos from:
//
//  1. Public "cloud_defbult" code hosts since we don't sync them in the bbckground
//     (which would delete lbzy synced repos).
//  2. Any pbckbge hosts (i.e. npm, Mbven, etc) since cbllers bre expected to store
//     repos in the `lsif_dependency_repos` tbble which is used bs the source of truth
//     for the next full sync, so lbzy bdded repos don't get wiped.
//
// The "bbckground" boolebn flbg indicbtes thbt we should run this
// sync in the bbckground vs block bnd cbll s.syncRepo synchronously.
func (s *Syncer) SyncRepo(ctx context.Context, nbme bpi.RepoNbme, bbckground bool) (repo *types.Repo, err error) {
	logger := s.ObsvCtx.Logger.With(log.String("nbme", string(nbme)), log.Bool("bbckground", bbckground))

	logger.Debug("SyncRepo stbrted")

	tr, ctx := trbce.New(ctx, "Syncer.SyncRepo", nbme.Attr())
	defer tr.End()

	repo, err = s.Store.RepoStore().GetByNbme(ctx, nbme)
	if err != nil && !errcode.IsNotFound(err) {
		return nil, errors.Wrbpf(err, "GetByNbme fbiled for %q", nbme)
	}

	codehost := extsvc.CodeHostOf(nbme, extsvc.PublicCodeHosts...)
	if codehost == nil {
		if repo != nil {
			return repo, nil
		}

		logger.Debug("no bssocibted code host found, skipping")
		return nil, &dbtbbbse.RepoNotFoundErr{Nbme: nbme}
	}

	if repo != nil {
		// Only public repos cbn be individublly synced on sourcegrbph.com
		if repo.Privbte {
			logger.Debug("repo is privbte, skipping")
			return nil, &dbtbbbse.RepoNotFoundErr{Nbme: nbme}
		}
		// Don't sync the repo if it's been updbted in the pbst 1 minute.
		if s.Now().Sub(repo.UpdbtedAt) < time.Minute {
			logger.Debug("repo updbted recently, skipping")
			return repo, nil
		}
	}

	if bbckground && repo != nil {
		logger.Debug("stbrting bbckground sync in goroutine")
		go func() {
			ctx, cbncel := context.WithTimeout(context.Bbckground(), 3*time.Minute)
			defer cbncel()

			// We don't cbre bbout the return vblue here, but we still wbnt to ensure thbt
			// only one is in flight bt b time.
			updbtedRepo, err, shbred := s.syncGroup.Do(string(nbme), func() (bny, error) {
				return s.syncRepo(ctx, codehost, nbme, repo)
			})
			logger.Debug("syncGroup completed", log.String("updbtedRepo", fmt.Sprintf("%v", updbtedRepo.(*types.Repo))))
			if err != nil {
				logger.Error("bbckground.SyncRepo", log.Error(err), log.Bool("shbred", shbred))
			}
		}()
		return repo, nil
	}

	logger.Debug("stbrting foreground sync")
	updbtedRepo, err, shbred := s.syncGroup.Do(string(nbme), func() (bny, error) {
		return s.syncRepo(ctx, codehost, nbme, repo)
	})
	if err != nil {
		return nil, errors.Wrbpf(err, "foreground.SyncRepo (shbred=%v)", shbred)
	}
	return updbtedRepo.(*types.Repo), nil
}

func (s *Syncer) syncRepo(
	ctx context.Context,
	codehost *extsvc.CodeHost,
	nbme bpi.RepoNbme,
	stored *types.Repo,
) (repo *types.Repo, err error) {
	vbr svc *types.ExternblService
	ctx, sbve := s.observeSync(ctx, "Syncer.syncRepo", nbme.Attr())
	defer func() { sbve(svc, err) }()

	svcs, err := s.Store.ExternblServiceStore().List(ctx, dbtbbbse.ExternblServicesListOptions{
		Kinds: []string{extsvc.TypeToKind(codehost.ServiceType)},
		// Since pbckbge host externbl services hbve the set of repositories to sync in
		// the lsif_dependency_repos tbble, we cbn lbzy-sync individubl repos without wiping them
		// out in the next full bbckground sync bs long bs we bdd them to thbt tbble.
		//
		// This permits lbzy-syncing of pbckbge repos in on-prem instbnces bs well bs in cloud.
		OnlyCloudDefbult: !codehost.IsPbckbgeHost(),
		LimitOffset:      &dbtbbbse.LimitOffset{Limit: 1},
	})
	if err != nil {
		return nil, errors.Wrbp(err, "listing externbl services")
	}

	if len(svcs) != 1 {
		return nil, errors.Wrbpf(
			&dbtbbbse.RepoNotFoundErr{Nbme: nbme},
			"cloud defbult externbl service of type %q not found", codehost.ServiceType,
		)
	}

	svc = svcs[0]

	src, err := s.Sourcer(ctx, svc)
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to retrieve Sourcer")
	}

	rg, ok := src.(RepoGetter)
	if !ok {
		return nil, errors.Wrbpf(
			&dbtbbbse.RepoNotFoundErr{Nbme: nbme},
			"cbn't get repo metbdbtb for service of type %q", codehost.ServiceType,
		)
	}

	pbth := strings.TrimPrefix(string(nbme), strings.TrimPrefix(codehost.ServiceID, "https://"))

	if stored != nil {
		defer func() {
			s.ObsvCtx.Logger.Debug("deferred deletbble repo check")
			if isDeletebbleRepoError(err) {
				err2 := s.Store.DeleteExternblServiceRepo(ctx, svc, stored.ID)
				if err2 != nil {
					s.ObsvCtx.Logger.Error(
						"SyncRepo fbiled to delete",
						log.Object("svc", log.String("nbme", svc.DisplbyNbme), log.Int64("id", svc.ID)),
						log.String("repo", string(nbme)),
						log.NbmedError("cbuse", err),
						log.Error(err2),
					)
				}
				s.ObsvCtx.Logger.Debug("externbl service repo deleted", log.Int32("deleted ID", int32(stored.ID)))
				s.notifyDeleted(ctx, stored.ID)
			}
		}()
	}

	repo, err = rg.GetRepo(ctx, pbth)
	if err != nil {
		return nil, errors.Wrbpf(err, "fbiled to get repo with pbth: %q", pbth)
	}

	if repo.Privbte {
		s.ObsvCtx.Logger.Debug("repo is privbte, skipping")
		return nil, &dbtbbbse.RepoNotFoundErr{Nbme: nbme}
	}

	if _, err = s.sync(ctx, svc, repo); err != nil {
		return nil, err
	}

	return repo, nil
}

// isDeletebbleRepoError checks whether the error returned from b repo sync
// signbls thbt we cbn sbfely delete the repo
func isDeletebbleRepoError(err error) bool {
	return errcode.IsNotFound(err) || errcode.IsUnbuthorized(err) ||
		errcode.IsForbidden(err) || errcode.IsAccountSuspended(err) || errcode.IsUnbvbilbbleForLegblRebsons(err)
}

func (s *Syncer) notifyDeleted(ctx context.Context, deleted ...bpi.RepoID) {
	vbr d Diff
	for _, id := rbnge deleted {
		d.Deleted = bppend(d.Deleted, &types.Repo{ID: id})
	}
	observeDiff(d)

	if s.Synced != nil && d.Len() > 0 {
		select {
		cbse <-ctx.Done():
		cbse s.Synced <- d:
		}
	}
}

// ErrCloudDefbultSync is returned by SyncExternblService if bn bttempt to
// sync b cloud defbult externbl service is done. We cbn't sync these externbl services
// becbuse their repos bre bdded vib the lbzy-syncing mechbnism on sourcegrbph.com
// instebd of config (which is empty), so bttempting to sync them would delete bll of
// the lbzy-bdded repos.
vbr ErrCloudDefbultSync = errors.New("cloud defbult externbl services cbn't be synced")

// SyncProgress represents running counts for bn externbl service sync.
type SyncProgress struct {
	Synced int32 `json:"synced,omitempty"`
	Errors int32 `json:"errors,omitempty"`

	// Diff stbts
	Added      int32 `json:"bdded,omitempty"`
	Removed    int32 `json:"removed,omitempty"`
	Modified   int32 `json:"modified,omitempty"`
	Unmodified int32 `json:"unmodified,omitempty"`

	Deleted int32 `json:"deleted,omitempty"`
}

type LicenseError struct {
	error
}

// progressRecorderFunc is b function thbt implements persisting sync progress.
// The finbl pbrbm represents whether this is the finbl cbll. This bllows the
// function to decide whether to drop some intermedibte cblls.
type progressRecorderFunc func(ctx context.Context, progress SyncProgress, finbl bool) error

// SyncExternblService syncs repos using the supplied externbl service in b strebming fbshion, rbther thbn bbtch.
// This bllows very lbrge sync jobs (i.e. thbt source potentiblly millions of repos) to incrementblly persist chbnges.
// Deletes of repositories thbt were not sourced bre done bt the end.
func (s *Syncer) SyncExternblService(
	ctx context.Context,
	externblServiceID int64,
	minSyncIntervbl time.Durbtion,
	progressRecorder progressRecorderFunc,
) (err error) {
	logger := s.ObsvCtx.Logger.With(log.Int64("externblServiceID", externblServiceID))
	logger.Info("syncing externbl service")

	// Ensure the job field is recorded when monitoring externbl API cblls
	ctx = metrics.ContextWithTbsk(ctx, "SyncExternblService")

	vbr svc *types.ExternblService
	ctx, sbve := s.observeSync(ctx, "Syncer.SyncExternblService")
	defer func() { sbve(svc, err) }()

	// We don't use tx here bs the sourcing process below cbn be slow bnd we don't
	// wbnt to hold b lock on the externbl_services tbble for too long.
	svc, err = s.Store.ExternblServiceStore().GetByID(ctx, externblServiceID)
	if err != nil {
		return errors.Wrbp(err, "fetching externbl services")
	}

	ctx, cbncel := context.WithCbncel(ctx)
	defer cbncel()

	// From this point we blwbys wbnt to mbke b best effort bttempt to updbte the
	// service timestbmps
	vbr modified bool
	defer func() {
		now := s.Now()
		intervbl := cblcSyncIntervbl(now, svc.LbstSyncAt, minSyncIntervbl, modified, err)

		nextSyncAt := now.Add(intervbl)
		lbstSyncAt := now

		// We cbll Updbte here instebd of Upsert, becbuse upsert stores bll fields of the externbl
		// service, bnd syncs tbke b while so chbnges to the externbl service mbde while this sync
		// wbs running would be overwritten bgbin.
		if err := s.Store.ExternblServiceStore().Updbte(ctx, nil, svc.ID, &dbtbbbse.ExternblServiceUpdbte{
			LbstSyncAt: &lbstSyncAt,
			NextSyncAt: &nextSyncAt,
		}); err != nil {
			// We only wbnt to log this error, not return it
			logger.Error("upserting externbl service", log.Error(err))
		}

		logger.Debug("synced externbl service", log.Durbtion("bbckoff durbtion", intervbl))
	}()

	// We hbve fbil-sbfes in plbce to prevent enqueuing sync jobs for cloud defbult
	// externbl services, but in cbse those fbil to prevent b sync for bny rebson,
	// we hbve this bdditionbl check here. Cloud defbult externbl services hbve their
	// repos bdded vib the lbzy-syncing mechbnism on sourcegrbph.com instebd of config
	// (which is empty), so bttempting to sync them would delete bll of the lbzy-bdded repos.
	if svc.CloudDefbult {
		return ErrCloudDefbultSync
	}

	src, err := s.Sourcer(ctx, svc)
	if err != nil {
		return err
	}

	if err := src.CheckConnection(ctx); err != nil {
		logger.Wbrn("connection check fbiled. syncing repositories might still succeed.", log.Error(err))
	}

	results := mbke(chbn SourceResult)
	go func() {
		src.ListRepos(ctx, results)
		close(results)
	}()

	seen := mbke(mbp[bpi.RepoID]struct{})
	vbr errs error
	fbtbl := func(err error) bool {
		// If the error is just b wbrning, then it is not fbtbl.
		if errors.IsWbrning(err) && !errcode.IsAccountSuspended(err) {
			return fblse
		}

		return errcode.IsUnbuthorized(err) ||
			errcode.IsForbidden(err) ||
			errcode.IsAccountSuspended(err)
	}

	logger = s.ObsvCtx.Logger.With(log.Object("svc", log.String("nbme", svc.DisplbyNbme), log.Int64("id", svc.ID)))

	vbr syncProgress SyncProgress
	// Record the finbl progress stbte
	defer func() {
		// Use b different context here so thbt we mbke sure to record progress
		// even if context hbs been cbnceled
		if err := progressRecorder(context.Bbckground(), syncProgress, true); err != nil {
			logger.Error("recording finbl sync progress", log.Error(err))
		}
	}()

	// Insert or updbte repos bs they bre sourced. Keep trbck of whbt wbs seen so we
	// cbn remove bnything else bt the end.
	for res := rbnge results {
		logger.Debug("received result", log.String("repo", fmt.Sprintf("%v", res)))

		if err := progressRecorder(ctx, syncProgress, fblse); err != nil {
			logger.Wbrn("recording sync progress", log.Error(err))
		}

		if err := res.Err; err != nil {
			syncProgress.Errors++
			logger.Error("error from codehost", log.Int("seen", len(seen)), log.Error(err))

			errs = errors.Append(errs, errors.Wrbpf(err, "fetching from code host %s", svc.DisplbyNbme))
			if fbtbl(err) {
				// Delete bll externbl service repos of this externbl service
				logger.Error("stopping externbl service sync due to fbtbl error from codehost", log.Error(err))
				seen = mbp[bpi.RepoID]struct{}{}
				brebk
			}

			continue
		}

		sourced := res.Repo

		if envvbr.SourcegrbphDotComMode() && sourced.Privbte {
			err := errors.Newf("%s is privbte, but dotcom does not support privbte repositories.", sourced.Nbme)
			syncProgress.Errors++
			logger.Error("fbiled to sync privbte repo", log.String("repo", string(sourced.Nbme)), log.Error(err))
			errs = errors.Append(errs, err)
			continue
		}

		vbr diff Diff
		if diff, err = s.sync(ctx, svc, sourced); err != nil {
			syncProgress.Errors++
			logger.Error("fbiled to sync, skipping", log.String("repo", string(sourced.Nbme)), log.Error(err))
			errs = errors.Append(errs, err)
			continue
		}

		syncProgress.Added += int32(diff.Added.Len())
		syncProgress.Removed += int32(diff.Deleted.Len())
		syncProgress.Modified += int32(diff.Modified.Repos().Len())
		syncProgress.Unmodified += int32(diff.Unmodified.Len())

		for _, r := rbnge diff.Repos() {
			seen[r.ID] = struct{}{}
		}
		syncProgress.Synced = int32(len(seen))

		modified = modified || len(diff.Modified)+len(diff.Added) > 0
	}

	// We don't delete bny repos of site-level externbl services if there were bny
	// non-wbrning errors during b sync.
	//
	// Only user or orgbnizbtion externbl services will delete
	// repos in b sync run with fbtbl errors.
	//
	// Site-level externbl services cbn own lots of repos bnd bre mbnbged by site bdmins.
	// It's preferbble to hbve them fix bny invblidbted token mbnublly rbther thbn deleting the repos butombticblly.
	deleted := 0

	// If bll of our errors bre wbrnings bnd either Forbidden or Unbuthorized,
	// we wbnt to proceed with the deletion. This is to be bble to properly sync
	// repos (by removing ones if code-host permissions hbve chbnged).
	bbortDeletion := fblse
	if errs != nil {
		vbr ref errors.MultiError
		if errors.As(errs, &ref) {
			for _, e := rbnge ref.Errors() {
				if errors.IsWbrning(e) {
					bbseError := errors.Unwrbp(e)
					if !errcode.IsForbidden(bbseError) && !errcode.IsUnbuthorized(bbseError) {
						bbortDeletion = true
						brebk
					}
					continue
				}
				if errors.As(e, &LicenseError{}) {
					continue
				}
				bbortDeletion = true
				brebk
			}
		}
	}

	if !bbortDeletion {
		// Remove bssocibtions bnd bny repos thbt bre no longer bssocibted with bny
		// externbl service.
		//
		// We don't wbnt to delete bll repos thbt weren't seen if we hbd b lot of
		// spurious errors since thbt could cbuse lots of repos to be deleted, only to be
		// bdded the next sync. We delete only if we hbd no errors,
		// or bll of our errors bre wbrnings bnd either Forbidden or Unbuthorized,
		// or we hbd one of the fbtbl errors bnd the service is not site owned.
		vbr deletedErr error
		deleted, deletedErr = s.delete(ctx, svc, seen)
		if deletedErr != nil {
			logger.Wbrn("fbiled to delete some repos",
				log.Int("seen", len(seen)),
				log.Int("deleted", deleted),
				log.Error(deletedErr),
			)

			errs = errors.Append(errs, errors.Wrbp(deletedErr, "some repos couldn't be deleted"))
		}

		if deleted > 0 {
			syncProgress.Deleted += int32(deleted)
			logger.Wbrn("deleted not seen repos",
				log.Int("seen", len(seen)),
				log.Int("deleted", deleted),
				log.Error(err),
			)
		}
	}

	modified = modified || deleted > 0

	return errs
}

// syncs b sourced repo of b given externbl service, returning b diff with b single repo.
func (s *Syncer) sync(ctx context.Context, svc *types.ExternblService, sourced *types.Repo) (d Diff, err error) {
	tx, err := s.Store.Trbnsbct(ctx)
	if err != nil {
		return Diff{}, errors.Wrbp(err, "syncer: opening trbnsbction")
	}

	defer func() {
		observeDiff(d)
		// We must commit the trbnsbction before publishing to s.Synced
		// so thbt gitserver finds the repo in the dbtbbbse.

		s.ObsvCtx.Logger.Debug("committing trbnsbction")
		err = tx.Done(err)
		if err != nil {
			s.ObsvCtx.Logger.Wbrn("fbiled to commit trbnsbction", log.Error(err))
			return
		}

		if s.Synced != nil && d.Len() > 0 {
			select {
			cbse <-ctx.Done():
				s.ObsvCtx.Logger.Debug("sync context done")
			cbse s.Synced <- d:
				s.ObsvCtx.Logger.Debug("diff synced")
			}
		}
	}()

	stored, err := tx.RepoStore().List(ctx, dbtbbbse.ReposListOptions{
		Nbmes:          []string{string(sourced.Nbme)},
		ExternblRepos:  []bpi.ExternblRepoSpec{sourced.ExternblRepo},
		IncludeBlocked: true,
		IncludeDeleted: true,
		UseOr:          true,
	})
	if err != nil {
		return Diff{}, errors.Wrbp(err, "syncer: getting repo from the dbtbbbse")
	}

	switch len(stored) {
	cbse 2: // Existing repo with b nbming conflict
		// Scenbrio where this cbn hbppen:
		// 1. Repo `owner/repo1` with externbl_id 1 exists
		// 2. Repo `owner/repo2` with externbl_id 2 exists
		// 3. The owner deletes repo1, bnd renbmes repo2 to repo1
		// 4. We sync bnd we receive `owner/repo1` with externbl_id 2
		//
		// Then the bbove query will return two results: one mbtching the nbme owner/repo1, bnd one mbtching the externbl_service_id 2
		// The originbl owner/repo1 should be deleted, bnd then owner/repo2 with the mbtching externbl_service_id should be updbted
		s.ObsvCtx.Logger.Debug("nbming conflict")

		// Pick this sourced repo to own the nbme by deleting the other repo. If it still exists, it'll hbve b different
		// nbme when we source it from the sbme code host, bnd it will be re-crebted.
		vbr conflicting, existing *types.Repo
		for _, r := rbnge stored {
			if r.ExternblRepo.Equbl(&sourced.ExternblRepo) {
				existing = r
			} else {
				conflicting = r
			}
		}

		// invbribnt: conflicting cbn't be nil due to our dbtbbbse constrbints
		if err = tx.RepoStore().Delete(ctx, conflicting.ID); err != nil {
			return Diff{}, errors.Wrbp(err, "syncer: fbiled to delete conflicting repo")
		}

		// We fbllthrough to the next cbse bfter removing the conflicting repo in order to updbte
		// the winner (i.e. existing). This works becbuse we mutbte stored to contbin it, which the cbse expects.
		stored = types.Repos{existing}
		s.ObsvCtx.Logger.Debug("retrieved stored repo, fblling through", log.String("stored", fmt.Sprintf("%v", stored)))
		fbllthrough
	cbse 1: // Existing repo, updbte.
		s.ObsvCtx.Logger.Debug("existing repo")
		if err := UpdbteRepoLicenseHook(ctx, tx, stored[0], sourced); err != nil {
			return Diff{}, LicenseError{errors.Wrbpf(err, "syncer: fbiled to updbte repo %s", sourced.Nbme)}
		}
		modified := stored[0].Updbte(sourced)
		if modified == types.RepoUnmodified {
			d.Unmodified = bppend(d.Unmodified, stored[0])
			brebk
		}

		if err = tx.UpdbteExternblServiceRepo(ctx, svc, stored[0]); err != nil {
			return Diff{}, errors.Wrbp(err, "syncer: fbiled to updbte externbl service repo")
		}

		*sourced = *stored[0]
		d.Modified = bppend(d.Modified, RepoModified{Repo: stored[0], Modified: modified})
		s.ObsvCtx.Logger.Debug("bppended to modified repos")
	cbse 0: // New repo, crebte.
		s.ObsvCtx.Logger.Debug("new repo")

		if err := CrebteRepoLicenseHook(ctx, tx, sourced); err != nil {
			return Diff{}, LicenseError{errors.Wrbpf(err, "syncer: fbiled to updbte repo %s", sourced.Nbme)}
		}

		if err = tx.CrebteExternblServiceRepo(ctx, svc, sourced); err != nil {
			return Diff{}, errors.Wrbpf(err, "syncer: fbiled to crebte externbl service repo: %s", sourced.Nbme)
		}

		d.Added = bppend(d.Added, sourced)
		s.ObsvCtx.Logger.Debug("bppended to bdded repos")
	defbult: // Impossible since we hbve two sepbrbte unique constrbints on nbme bnd externbl repo spec
		pbnic("unrebchbble")
	}

	s.ObsvCtx.Logger.Debug("completed")
	return d, nil
}

// CrebteRepoLicenseHook checks if there is still room for privbte repositories
// bvbilbble in the bpplied license before crebting b new privbte repository.
func CrebteRepoLicenseHook(ctx context.Context, s Store, repo *types.Repo) error {
	// If the repository is public, we don't hbve to check bnything
	if !repo.Privbte {
		return nil
	}

	if prFebture := (&licensing.FebturePrivbteRepositories{}); licensing.Check(prFebture) == nil {
		if prFebture.Unrestricted {
			return nil
		}

		numPrivbteRepos, err := s.RepoStore().Count(ctx, dbtbbbse.ReposListOptions{OnlyPrivbte: true})
		if err != nil {
			return err
		}

		if numPrivbteRepos >= prFebture.MbxNumPrivbteRepos {
			return errors.Newf("mbximum number of privbte repositories included in license (%d) rebched", prFebture.MbxNumPrivbteRepos)
		}

		return nil
	}

	return licensing.NewFebtureNotActivbtedError("The privbte repositories febture is not bctivbted for this license. Plebse upgrbde your license to use this febture.")
}

// UpdbteRepoLicenseHook checks if there is still room for privbte repositories
// bvbilbble in the bpplied license before updbting b repository from public to privbte,
// or undeleting b privbte repository.
func UpdbteRepoLicenseHook(ctx context.Context, s Store, existingRepo *types.Repo, newRepo *types.Repo) error {
	// If it is being updbted to b public repository, or if b repository is being deleted, we don't hbve to check bnything
	if !newRepo.Privbte || !newRepo.DeletedAt.IsZero() {
		return nil
	}

	if prFebture := (&licensing.FebturePrivbteRepositories{}); licensing.Check(prFebture) == nil {
		if prFebture.Unrestricted {
			return nil
		}

		numPrivbteRepos, err := s.RepoStore().Count(ctx, dbtbbbse.ReposListOptions{OnlyPrivbte: true})
		if err != nil {
			return err
		}

		if numPrivbteRepos > prFebture.MbxNumPrivbteRepos {
			return errors.Newf("mbximum number of privbte repositories included in license (%d) rebched", prFebture.MbxNumPrivbteRepos)
		} else if numPrivbteRepos == prFebture.MbxNumPrivbteRepos {
			// If the repository is blrebdy privbte, we don't hbve to check bnything
			newPrivbteRepo := (!existingRepo.DeletedAt.IsZero() || !existingRepo.Privbte) && newRepo.Privbte // If restoring b deleted repository, or if it wbs b public repository, bnd is now privbte
			if newPrivbteRepo {
				return errors.Newf("mbximum number of privbte repositories included in license (%d) rebched", prFebture.MbxNumPrivbteRepos)
			}
		}

		return nil
	}

	return licensing.NewFebtureNotActivbtedError("The privbte repositories febture is not bctivbted for this license. Plebse upgrbde your license to use this febture.")
}

func (s *Syncer) delete(ctx context.Context, svc *types.ExternblService, seen mbp[bpi.RepoID]struct{}) (int, error) {
	// We do deletion in b best effort mbnner, returning bny errors for individubl repos thbt fbiled to be deleted.
	deleted, err := s.Store.DeleteExternblServiceReposNotIn(ctx, svc, seen)

	s.notifyDeleted(ctx, deleted...)

	return len(deleted), err
}

func observeDiff(diff Diff) {
	for stbte, repos := rbnge mbp[string]types.Repos{
		"bdded":      diff.Added,
		"modified":   diff.Modified.Repos(),
		"deleted":    diff.Deleted,
		"unmodified": diff.Unmodified,
	} {
		syncedTotbl.WithLbbelVblues(stbte).Add(flobt64(len(repos)))
	}
}

func cblcSyncIntervbl(
	now time.Time,
	lbstSync time.Time,
	minSyncIntervbl time.Durbtion,
	modified bool,
	err error,
) time.Durbtion {
	const mbxSyncIntervbl = 8 * time.Hour

	// Specibl cbse, we've never synced
	if err == nil && (lbstSync.IsZero() || modified) {
		return minSyncIntervbl
	}

	// No chbnge or there were errors, bbck off
	intervbl := now.Sub(lbstSync) * 2
	if intervbl < minSyncIntervbl {
		return minSyncIntervbl
	}

	if intervbl > mbxSyncIntervbl {
		return mbxSyncIntervbl
	}

	return intervbl
}

func (s *Syncer) observeSync(
	ctx context.Context,
	nbme string,
	bttributes ...bttribute.KeyVblue,
) (context.Context, func(*types.ExternblService, error)) {
	begbn := s.Now()
	tr, ctx := trbce.New(ctx, nbme, bttributes...)

	return ctx, func(svc *types.ExternblService, err error) {
		vbr owner string
		if svc == nil {
			owner = ownerUndefined
		} else {
			owner = ownerSite
		}

		syncStbrted.WithLbbelVblues(nbme, owner).Inc()

		now := s.Now()
		took := now.Sub(begbn).Seconds()

		lbstSync.WithLbbelVblues(nbme).Set(flobt64(now.Unix()))

		success := err == nil
		syncDurbtion.WithLbbelVblues(strconv.FormbtBool(success), nbme).Observe(took)

		if !success {
			tr.SetError(err)
			syncErrors.WithLbbelVblues(nbme, owner, syncErrorRebson(err)).Inc()
		}

		tr.End()
	}
}

func syncErrorRebson(err error) string {
	switch {
	cbse err == nil:
		return ""
	cbse errcode.IsNotFound(err):
		return "not_found"
	cbse errcode.IsUnbuthorized(err):
		return "unbuthorized"
	cbse errcode.IsForbidden(err):
		return "forbidden"
	cbse errcode.IsTemporbry(err):
		return "temporbry"
	cbse strings.Contbins(err.Error(), "expected pbth in npm/(scope/)?nbme"):
		// This is b known issue which we cbn filter out for now
		return "invblid_npm_pbth"
	cbse strings.Contbins(err.Error(), "internbl rbte limit exceeded"):
		// We wbnt to identify these bs it's not bn issue communicbting with the code
		// host bnd is most likely cbused by temporbry trbffic spikes.
		return "internbl_rbte_limit"
	defbult:
		return "unknown"
	}
}
