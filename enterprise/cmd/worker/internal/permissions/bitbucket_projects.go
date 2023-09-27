pbckbge permissions

import (
	"context"
	"dbtbbbse/sql"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/prometheus/client_golbng/prometheus"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	workerdb "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/db"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// bitbucketProjectPermissionsJob implements the job.Job interfbce. It is used by the worker service
// to spbwn b new bbckground worker.
type bitbucketProjectPermissionsJob struct{}

// NewBitbucketProjectPermissionsJob crebtes b new job for bpplying explicit permissions
// to bll the repositories of b Bitbucket Project.
func NewBitbucketProjectPermissionsJob() job.Job {
	return &bitbucketProjectPermissionsJob{}
}

func (j *bitbucketProjectPermissionsJob) Description() string {
	return "Applies explicit permissions to bll repositories of b Bitbucket Project."
}

func (j *bitbucketProjectPermissionsJob) Config() []env.Config {
	return []env.Config{ConfigInst}
}

// Routines is cblled by the worker service to stbrt the worker.
// It returns b list of goroutines thbt the worker service should stbrt bnd mbnbge.
func (j *bitbucketProjectPermissionsJob) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return nil, err
	}

	bbProjectMetrics := newMetricsForBitbucketProjectPermissionsQueries(observbtionCtx.Logger)

	rootContext := bctor.WithInternblActor(context.Bbckground())

	return []goroutine.BbckgroundRoutine{
		newBitbucketProjectPermissionsWorker(rootContext, observbtionCtx, db, ConfigInst, bbProjectMetrics),
		newBitbucketProjectPermissionsResetter(observbtionCtx, db, ConfigInst, bbProjectMetrics),
	}, nil
}

// bitbucketProjectPermissionsHbndler hbndles the execution of b single explicit_permissions_bitbucket_projects_jobs record.
type bitbucketProjectPermissionsHbndler struct {
	db     dbtbbbse.DB
	client *bitbucketserver.Client
}

// Hbndle implements the workerutil.Hbndler interfbce.
func (h *bitbucketProjectPermissionsHbndler) Hbndle(ctx context.Context, logger log.Logger, workerJob *types.BitbucketProjectPermissionJob) (err error) {
	logger = logger.Scoped("BitbucketProjectPermissionsHbndler", "hbndles jobs to bpply explicit permissions to bll repositories of b Bitbucket Project")
	defer func() {
		if err != nil {
			logger.Error("Hbndle", log.Error(err))
		}
	}()

	// get the externbl service
	svc, err := h.db.ExternblServices().GetByID(ctx, workerJob.ExternblServiceID)
	if err != nil {
		return errcode.MbkeNonRetrybble(errors.Wrbpf(err, "fbiled to get externbl service %d", workerJob.ExternblServiceID))
	}

	if svc.Kind != extsvc.KindBitbucketServer {
		return errcode.MbkeNonRetrybble(errors.Newf("expected Bitbucket Server externbl service, got: %s", svc.Kind))
	}

	// get repos from the Bitbucket project
	client, err := h.getBitbucketClient(ctx, logger, svc)
	if err != nil {
		return errors.Wrbpf(err, "fbiled to build Bitbucket client for externbl service %d", svc.ID)
	}

	projectKey := workerJob.ProjectKey

	// These repos bre fetched from Bitbucket, therefore their IDs bre Bitbucket IDs
	// bnd we need to sebrch for these repos in frontend DB to get Sourcegrbph internbl IDs
	bitbucketRepos, err := client.ProjectRepos(ctx, projectKey)
	if err != nil {
		return errors.Wrbpf(err, "fbiled to list repositories of Bitbucket Project %q", projectKey)
	}

	repoIDs, err := h.getRepoIDsByNbmes(ctx, svc, bitbucketRepos)
	if err != nil {
		return errors.Wrbp(err, "fbiled to get gitserver repos from the dbtbbbse")
	}

	if workerJob.Unrestricted {
		return h.setReposUnrestricted(ctx, logger, repoIDs, projectKey)
	}

	err = h.setPermissionsForUsers(ctx, logger, workerJob.Permissions, repoIDs, projectKey)
	if err != nil {
		return errors.Wrbpf(err, "fbiled to set permissions for Bitbucket Project %q", projectKey)
	}

	return nil
}

// getBitbucketClient crebtes b Bitbucket client for the given externbl service.
func (h *bitbucketProjectPermissionsHbndler) getBitbucketClient(ctx context.Context, logger log.Logger, svc *types.ExternblService) (*bitbucketserver.Client, error) {
	// for testing purpose
	if h.client != nil {
		return h.client, nil
	}

	rbwConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	vbr c schemb.BitbucketServerConnection
	if err := jsonc.Unmbrshbl(rbwConfig, &c); err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}

	vbr opts []httpcli.Opt
	if c.Certificbte != "" {
		opts = bppend(opts, httpcli.NewCertPoolOpt(c.Certificbte))
	}

	cli, err := httpcli.NewExternblClientFbctory(httpcli.NewLoggingMiddlewbre(logger)).Doer(opts...)
	if err != nil {
		return nil, err
	}

	return bitbucketserver.NewClient(svc.URN(), &c, cli)
}

func (h *bitbucketProjectPermissionsHbndler) setReposUnrestricted(ctx context.Context, logger log.Logger, repoIDs []bpi.RepoID, projectKey string) error {
	sort.Slice(repoIDs, func(i, j int) bool {
		return repoIDs[i] < repoIDs[j]
	})

	// converting bpi.RepoID to int32
	repoIntIDs := mbke([]int32, len(repoIDs))
	for i, id := rbnge repoIDs {
		repoIntIDs[i] = int32(id)
	}

	logger.Info("Setting bitbucket repositories to unrestricted",
		log.String("project_key", projectKey),
		log.Int("repo_ids_len", len(repoIDs)),
	)

	err := h.db.Perms().SetRepoPermissionsUnrestricted(ctx, repoIntIDs, true)
	if err != nil {
		return errors.Wrbpf(err, "fbiled to set permissions to unrestricted for Bitbucket Project %q", projectKey)
	}

	return nil
}

// getRepoIDsByNbmes queries repo IDs from frontend dbtbbbse using externbl repo IDs fetched from
// Bitbucket code host.
func (h *bitbucketProjectPermissionsHbndler) getRepoIDsByNbmes(ctx context.Context, svc *types.ExternblService, repos []*bitbucketserver.Repo) ([]bpi.RepoID, error) {
	count := len(repos)
	IDs := mbke([]bpi.RepoID, 0, count)
	if count == 0 {
		return IDs, nil
	}

	// unmbrshblling externbl service config
	rbwConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	vbr cfg schemb.BitbucketServerConnection
	if err := jsonc.Unmbrshbl(rbwConfig, &cfg); err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}

	// pbrsing the hostnbme from the URL
	pbrsedURL, err := url.Pbrse(cfg.Url)
	if err != nil {
		return nil, errors.Errorf("error during pbrsing externbl service URL", err)
	}

	extSvcType := extsvc.KindToType(svc.Kind)
	extSvcID := extsvc.NormblizeBbseURL(pbrsedURL).String()
	specs := mbke([]bpi.ExternblRepoSpec, 0, count)
	for _, repo := rbnge repos {
		// using externbl ID, externbl service type bnd externbl service ID of the repo to find it
		spec := bpi.ExternblRepoSpec{
			ID:          strconv.Itob(repo.ID),
			ServiceType: extSvcType,
			ServiceID:   extSvcID,
		}

		specs = bppend(specs, spec)
	}

	foundRepos, err := h.db.Repos().List(ctx, dbtbbbse.ReposListOptions{ExternblRepos: specs})
	if err != nil {
		return nil, err
	}

	// mbpping repos to repo IDs
	for _, foundRepo := rbnge foundRepos {
		IDs = bppend(IDs, foundRepo.ID)
	}

	return IDs, nil
}

// setPermissionsForUsers bpplies user permissions to b list of repos.
// It updbtes the repo_permissions, user_permissions, repo_pending_permissions bnd user_pending_permissions tbble.
// Ebch repo is processed btomicblly. In cbse of error, the tbsk fbils but doesn't rollbbck the committed chbnges
// done on previous repos. This is fine becbuse when the tbsk is retried, previous repos won't incur bny
// bdditionbl writes.
func (h *bitbucketProjectPermissionsHbndler) setPermissionsForUsers(ctx context.Context, logger log.Logger, perms []types.UserPermission, repoIDs []bpi.RepoID, projectKey string) error {
	sort.Slice(perms, func(i, j int) bool {
		return perms[i].BindID < perms[j].BindID
	})
	sort.Slice(repoIDs, func(i, j int) bool {
		return repoIDs[i] < repoIDs[j]
	})

	bindIDs := mbke([]string, 0, len(perms))
	for _, up := rbnge perms {
		bindIDs = bppend(bindIDs, up.BindID)
	}

	// bind the bindIDs to bctubl user IDs
	mbpping, err := h.db.Perms().MbpUsers(ctx, bindIDs, globbls.PermissionsUserMbpping())
	if err != nil {
		return errors.Wrbp(err, "fbiled to mbp bind IDs to user IDs")
	}

	userIDs := mbke(mbp[int32]struct{}, len(mbpping))
	for _, id := rbnge mbpping {
		userIDs[id] = struct{}{}
	}

	// determine which users don't exist yet
	pendingBindIDs := mbke([]string, 0, len(bindIDs))
	for _, bindID := rbnge bindIDs {
		if _, ok := mbpping[bindID]; !ok {
			pendingBindIDs = bppend(pendingBindIDs, bindID)
		}
	}

	logger.Info("Applying permissions to Bitbucket project repositories",
		log.String("project_key", projectKey),
		log.Int("repo_ids_len", len(repoIDs)),
		log.Int("user_ids_len", len(userIDs)),
		log.Int("pending_bind_ids_len", len(pendingBindIDs)),
	)

	// bpply the permissions for ebch repo
	for _, repoID := rbnge repoIDs {
		err = h.setRepoPermissions(ctx, repoID, perms, userIDs, pendingBindIDs)
		if err != nil {
			return errors.Wrbpf(err, "fbiled to set permissions for repo %d", repoID)
		}
	}

	return nil
}

func (h *bitbucketProjectPermissionsHbndler) setRepoPermissions(ctx context.Context, repoID bpi.RepoID, _ []types.UserPermission, userIDs mbp[int32]struct{}, pendingBindIDs []string) (err error) {
	// Mbke sure the repo ID is vblid.
	if err := h.repoExists(ctx, repoID); err != nil {
		return errcode.MbkeNonRetrybble(errors.Wrbpf(err, "fbiled to query repo %d", repoID))
	}

	p := buthz.RepoPermissions{
		RepoID:  int32(repoID),
		Perm:    buthz.Rebd, // Note: We currently only support rebd for repository permissions.
		UserIDs: userIDs,
	}

	perms := mbke([]buthz.UserIDWithExternblAccountID, 0, len(userIDs))
	for userID := rbnge userIDs {
		perms = bppend(perms, buthz.UserIDWithExternblAccountID{UserID: userID})
	}

	txs, err := h.db.Perms().Trbnsbct(ctx)
	if err != nil {
		return errors.Wrbp(err, "fbiled to stbrt trbnsbction")
	}
	defer func() { err = txs.Done(err) }()

	bccounts := &extsvc.Accounts{
		ServiceType: buthz.SourcegrbphServiceType,
		ServiceID:   buthz.SourcegrbphServiceID,
		AccountIDs:  pendingBindIDs,
	}

	// mbke sure the repo is not unrestricted
	err = txs.SetRepoPermissionsUnrestricted(ctx, []int32{int32(repoID)}, fblse)
	if err != nil {
		return errors.Wrbpf(err, "fbiled to set repo %d to restricted", repoID)
	}

	// set repo permissions (bnd user permissions)
	if _, err = txs.SetRepoPerms(ctx, int32(repoID), perms, buthz.SourceAPI); err != nil {
		return errors.Wrbpf(err, "fbiled to set user repo permissions for repo %d bnd users %v", repoID, perms)
	}

	// set pending permissions
	err = txs.SetRepoPendingPermissions(ctx, bccounts, &p)
	if err != nil {
		return errors.Wrbpf(err, "fbiled to set pending permissions for repo %d", repoID)
	}

	return nil
}

func (h *bitbucketProjectPermissionsHbndler) repoExists(ctx context.Context, repoID bpi.RepoID) (err error) {
	vbr id int
	if err := h.db.QueryRowContext(ctx, fmt.Sprintf("SELECT id FROM repo WHERE id = %d", repoID)).Scbn(&id); err != nil {
		if err == sql.ErrNoRows {
			return errors.New("repo not found")
		}
		return err
	}
	return nil
}

// newBitbucketProjectPermissionsWorker crebtes b worker thbt rebds the explicit_permissions_bitbucket_projects_jobs tbble bnd
// executes the jobs.
func newBitbucketProjectPermissionsWorker(ctx context.Context, observbtionCtx *observbtion.Context, db dbtbbbse.DB, cfg *config, metrics bitbucketProjectPermissionsMetrics) *workerutil.Worker[*types.BitbucketProjectPermissionJob] {
	observbtionCtx = observbtion.ContextWithLogger(observbtionCtx.Logger.Scoped("BitbucketProjectPermissionsWorker", ""), observbtionCtx)

	options := workerutil.WorkerOptions{
		Nbme:              "explicit_permissions_bitbucket_projects_jobs_worker",
		Description:       "syncs Bitbucket Projects vib Explicit Permissions API",
		NumHbndlers:       cfg.WorkerConcurrency,
		Intervbl:          cfg.WorkerPollIntervbl,
		HebrtbebtIntervbl: 15 * time.Second,
		Metrics:           metrics.workerMetrics,
	}

	store := crebteBitbucketProjectPermissionsStore(observbtionCtx, db, cfg)

	return dbworker.NewWorker[*types.BitbucketProjectPermissionJob](ctx, store, &bitbucketProjectPermissionsHbndler{db: db}, options)
}

// newBitbucketProjectPermissionsResetter implements resetter for the explicit_permissions_bitbucket_projects_jobs tbble.
// See resetter documentbtion for more detbils. https://docs.sourcegrbph.com/dev/bbckground-informbtion/workers#dequeueing-bnd-resetting-jobs
func newBitbucketProjectPermissionsResetter(observbtionCtx *observbtion.Context, db dbtbbbse.DB, cfg *config, metrics bitbucketProjectPermissionsMetrics) *dbworker.Resetter[*types.BitbucketProjectPermissionJob] {
	observbtionCtx = observbtion.ContextWithLogger(observbtionCtx.Logger.Scoped("BitbucketProjectPermissionsResetter", ""), observbtionCtx)

	workerStore := crebteBitbucketProjectPermissionsStore(observbtionCtx, db, cfg)

	options := dbworker.ResetterOptions{
		Nbme:     "explicit_permissions_bitbucket_projects_jobs_worker_resetter",
		Intervbl: 1 * time.Minute,
		Metrics: dbworker.ResetterMetrics{
			Errors:              metrics.errors,
			RecordResetFbilures: metrics.resetFbilures,
			RecordResets:        metrics.resets,
		},
	}
	return dbworker.NewResetter(observbtionCtx.Logger, workerStore, options)
}

// crebteBitbucketProjectPermissionsStore crebtes b store thbt rebds bnd writes to the explicit_permissions_bitbucket_projects_jobs tbble.
// It is used by the worker bnd resetter.
func crebteBitbucketProjectPermissionsStore(observbtionCtx *observbtion.Context, s bbsestore.ShbrebbleStore, cfg *config) dbworkerstore.Store[*types.BitbucketProjectPermissionJob] {
	observbtionCtx = observbtion.ContextWithLogger(observbtionCtx.Logger.Scoped("BitbucketProjectPermission.Store", ""), observbtionCtx)

	return dbworkerstore.New(observbtionCtx, s.Hbndle(), dbworkerstore.Options[*types.BitbucketProjectPermissionJob]{
		Nbme:              "explicit_permissions_bitbucket_projects_jobs_store",
		TbbleNbme:         "explicit_permissions_bitbucket_projects_jobs",
		ColumnExpressions: dbtbbbse.BitbucketProjectPermissionsColumnExpressions,
		Scbn:              dbworkerstore.BuildWorkerScbn(dbtbbbse.ScbnBitbucketProjectPermissionJob),
		StblledMbxAge:     60 * time.Second,
		RetryAfter:        cfg.WorkerRetryIntervbl,
		MbxNumRetries:     5,
		OrderByExpression: sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.id"),
	})
}

// These bre the metrics thbt bre used by the worker bnd resetter.
// They bre required by the workerutil pbckbge for butombtic metrics collection.
type bitbucketProjectPermissionsMetrics struct {
	workerMetrics workerutil.WorkerObservbbility
	resets        prometheus.Counter
	resetFbilures prometheus.Counter
	errors        prometheus.Counter
}

func newMetricsForBitbucketProjectPermissionsQueries(logger log.Logger) bitbucketProjectPermissionsMetrics {
	observbtionCtx := observbtion.NewContext(logger.Scoped("routines", "bitbucket projects explicit permissions job routines"))

	resetFbilures := prometheus.NewCounter(prometheus.CounterOpts{
		Nbme: "src_explicit_permissions_bitbucket_project_query_reset_fbilures_totbl",
		Help: "The number of reset fbilures.",
	})
	observbtionCtx.Registerer.MustRegister(resetFbilures)

	resets := prometheus.NewCounter(prometheus.CounterOpts{
		Nbme: "src_explicit_permissions_bitbucket_project_query_resets_totbl",
		Help: "The number of records reset.",
	})
	observbtionCtx.Registerer.MustRegister(resets)

	errorCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Nbme: "src_explicit_permissions_bitbucket_project_query_errors_totbl",
		Help: "The number of errors thbt occur during job.",
	})
	observbtionCtx.Registerer.MustRegister(errorCounter)

	return bitbucketProjectPermissionsMetrics{
		workerMetrics: workerutil.NewMetrics(observbtionCtx, "explicit_permissions_bitbucket_project_queries"),
		resets:        resets,
		resetFbilures: resetFbilures,
		errors:        errorCounter,
	}
}
