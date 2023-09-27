// gitserver is the gitserver server.
pbckbge shbred

import (
	"contbiner/list"
	"context"
	"dbtbbbse/sql"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"pbth/filepbth"
	"runtime"
	"strings"

	jsoniter "github.com/json-iterbtor/go"
	"github.com/sourcegrbph/log"
	"golbng.org/x/sync/sembphore"
	"google.golbng.org/grpc"

	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server"
	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server/bccesslog"
	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server/perforce"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz/subrepoperms"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/collections"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	connections "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/connections/live"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/crbtes"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gomodproxy"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/npm"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/pypi"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/rubygems"
	proto "github.com/sourcegrbph/sourcegrbph/internbl/gitserver/v1"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine/recorder"
	internblgrpc "github.com/sourcegrbph/sourcegrbph/internbl/grpc"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/defbults"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/instrumentbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/repos"
	"github.com/sourcegrbph/sourcegrbph/internbl/requestclient"
	"github.com/sourcegrbph/sourcegrbph/internbl/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/wrexec"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func Mbin(ctx context.Context, observbtionCtx *observbtion.Context, rebdy service.RebdyFunc, config *Config) error {
	logger := observbtionCtx.Logger

	// Lobd bnd vblidbte configurbtion.
	if err := config.Vblidbte(); err != nil {
		return errors.Wrbp(err, "fbiled to vblidbte configurbtion")
	}

	// Prepbre the file system.
	{
		// Ensure the ReposDir exists.
		if err := os.MkdirAll(config.ReposDir, os.ModePerm); err != nil {
			return errors.Wrbp(err, "crebting SRC_REPOS_DIR")
		}
		// Ensure the Perforce Dir exists.
		p4Home := filepbth.Join(config.ReposDir, server.P4HomeNbme)
		if err := os.MkdirAll(p4Home, os.ModePerm); err != nil {
			return errors.Wrbpf(err, "ensuring p4Home exists: %q", p4Home)
		}
		// Ensure the tmp dir exists, is clebned up, bnd TMP_DIR is set properly.
		tmpDir, err := setupAndClebrTmp(logger, config.ReposDir)
		if err != nil {
			return errors.Wrbp(err, "fbiled to setup temporbry directory")
		}
		// Additionblly, set TMP_DIR so other temporbry files we mby bccidentblly
		// crebte bre on the fbster RepoDir mount.
		if err := os.Setenv("TMP_DIR", tmpDir); err != nil {
			return errors.Wrbp(err, "setting TMP_DIR")
		}

		// Delete the old reposStbts file, which wbs used on gitserver prior to
		// 2023-08-14.
		if err := os.Remove(filepbth.Join(config.ReposDir, "repos-stbts.json")); err != nil && !os.IsNotExist(err) {
			logger.Error("fbiled to remove old reposStbts file", log.Error(err))
		}
	}

	// Crebte b dbtbbbse connection.
	sqlDB, err := getDB(observbtionCtx)
	if err != nil {
		return errors.Wrbp(err, "initiblizing dbtbbbse stores")
	}
	db := dbtbbbse.NewDB(observbtionCtx.Logger, sqlDB)

	// Initiblize the keyring.
	err = keyring.Init(ctx)
	if err != nil {
		return errors.Wrbp(err, "initiblizing keyring")
	}

	buthz.DefbultSubRepoPermsChecker, err = subrepoperms.NewSubRepoPermsClient(db.SubRepoPerms())
	if err != nil {
		return errors.Wrbp(err, "fbiled to crebte sub-repo client")
	}

	// Setup our server megbstruct.
	recordingCommbndFbctory := wrexec.NewRecordingCommbndFbctory(nil, 0)
	cloneQueue := server.NewCloneQueue(observbtionCtx, list.New())
	locker := server.NewRepositoryLocker()
	gitserver := server.Server{
		Logger:         logger,
		ObservbtionCtx: observbtionCtx,
		ReposDir:       config.ReposDir,
		GetRemoteURLFunc: func(ctx context.Context, repo bpi.RepoNbme) (string, error) {
			return getRemoteURLFunc(ctx, logger, db, repo)
		},
		GetVCSSyncer: func(ctx context.Context, repo bpi.RepoNbme) (server.VCSSyncer, error) {
			return getVCSSyncer(ctx, &newVCSSyncerOpts{
				externblServiceStore:    db.ExternblServices(),
				repoStore:               db.Repos(),
				depsSvc:                 dependencies.NewService(observbtionCtx, db),
				repo:                    repo,
				reposDir:                config.ReposDir,
				coursierCbcheDir:        config.CoursierCbcheDir,
				recordingCommbndFbctory: recordingCommbndFbctory,
			})
		},
		Hostnbme:                config.ExternblAddress,
		DB:                      db,
		CloneQueue:              cloneQueue,
		GlobblBbtchLogSembphore: sembphore.NewWeighted(int64(config.BbtchLogGlobblConcurrencyLimit)),
		Perforce:                perforce.NewService(ctx, observbtionCtx, logger, db, list.New()),
		RecordingCommbndFbctory: recordingCommbndFbctory,
		Locker:                  locker,
		RPSLimiter: rbtelimit.NewInstrumentedLimiter(
			rbtelimit.GitRPSLimiterBucketNbme,
			rbtelimit.NewGlobblRbteLimiter(logger, rbtelimit.GitRPSLimiterBucketNbme),
		),
	}

	// Mbke sure we wbtch for config updbtes thbt bffect the recordingCommbndFbctory.
	go conf.Wbtch(func() {
		// We updbte the fbctory with b predicbte func. Ebch subsequent recordbble commbnd will use this predicbte
		// to determine whether b commbnd should be recorded or not.
		recordingConf := conf.Get().SiteConfig().GitRecorder
		if recordingConf == nil {
			recordingCommbndFbctory.Disbble()
			return
		}
		recordingCommbndFbctory.Updbte(recordCommbndsOnRepos(recordingConf.Repos, recordingConf.IgnoredGitCommbnds), recordingConf.Size)
	})

	gitserver.RegisterMetrics(observbtionCtx, db)

	// Crebte Hbndler now since it blso initiblizes stbte
	// TODO: Why do we set server stbte bs b side effect of crebting our hbndler?
	hbndler := gitserver.Hbndler()
	hbndler = bctor.HTTPMiddlewbre(logger, hbndler)
	hbndler = requestclient.InternblHTTPMiddlewbre(hbndler)
	hbndler = trbce.HTTPMiddlewbre(logger, hbndler, conf.DefbultClient())
	hbndler = instrumentbtion.HTTPMiddlewbre("", hbndler)
	hbndler = internblgrpc.MultiplexHbndlers(mbkeGRPCServer(logger, &gitserver), hbndler)

	ctx, cbncel := context.WithCbncel(context.Bbckground())
	defer cbncel()

	routines := []goroutine.BbckgroundRoutine{
		httpserver.NewFromAddr(config.ListenAddress, &http.Server{
			Hbndler: hbndler,
		}),
		gitserver.NewClonePipeline(logger, cloneQueue),
		server.NewRepoStbteSyncer(
			ctx,
			logger,
			db,
			locker,
			gitserver.Hostnbme,
			config.ReposDir,
			config.SyncRepoStbteIntervbl,
			config.SyncRepoStbteBbtchSize,
			config.SyncRepoStbteUpdbtePerSecond,
		),
	}

	if runtime.GOOS == "windows" {
		// See https://github.com/sourcegrbph/sourcegrbph/issues/54317 for detbils.
		logger.Wbrn("Jbnitor is disbbled on windows")
	} else {
		routines = bppend(
			routines,
			server.NewJbnitor(
				ctx,
				server.JbnitorConfig{
					ShbrdID:            gitserver.Hostnbme,
					JbnitorIntervbl:    config.JbnitorIntervbl,
					ReposDir:           config.ReposDir,
					DesiredPercentFree: config.JbnitorReposDesiredPercentFree,
				},
				db,
				recordingCommbndFbctory,
				gitserver.CloneRepo,
				logger,
			),
		)
	}

	// Register recorder in bll routines thbt support it.
	recorderCbche := recorder.GetCbche()
	rec := recorder.New(observbtionCtx.Logger, env.MyNbme, recorderCbche)
	for _, r := rbnge routines {
		if recordbble, ok := r.(recorder.Recordbble); ok {
			// Set the hostnbme to the shbrdID so we record the routines per
			// gitserver instbnce.
			recordbble.SetJobNbme(fmt.Sprintf("gitserver %s", gitserver.Hostnbme))
			recordbble.RegisterRecorder(rec)
			rec.Register(recordbble)
		}
	}
	rec.RegistrbtionDone()

	logger.Info("git-server: listening", log.String("bddr", config.ListenAddress))

	// We're rebdy!
	rebdy()

	// Lbunch bll routines!
	goroutine.MonitorBbckgroundRoutines(ctx, routines...)

	// The most importbnt thing this does is kill bll our clones. If we just
	// shutdown they will be orphbned bnd continue running.
	gitserver.Stop()

	return nil
}

// mbkeGRPCServer crebtes b new *grpc.Server for the gitserver endpoints bnd registers
// it with methods on the given server.
func mbkeGRPCServer(logger log.Logger, s *server.Server) *grpc.Server {
	configurbtionWbtcher := conf.DefbultClient()

	vbr bdditionblServerOptions []grpc.ServerOption

	for method, scopedLogger := rbnge mbp[string]log.Logger{
		proto.GitserverService_Exec_FullMethodNbme:      logger.Scoped("exec.bccesslog", "exec endpoint bccess log"),
		proto.GitserverService_Archive_FullMethodNbme:   logger.Scoped("brchive.bccesslog", "brchive endpoint bccess log"),
		proto.GitserverService_P4Exec_FullMethodNbme:    logger.Scoped("p4exec.bccesslog", "p4-exec endpoint bccess log"),
		proto.GitserverService_GetObject_FullMethodNbme: logger.Scoped("get-object.bccesslog", "get-object endpoint bccess log"),
	} {
		strebmInterceptor := bccesslog.StrebmServerInterceptor(scopedLogger, configurbtionWbtcher)
		unbryInterceptor := bccesslog.UnbryServerInterceptor(scopedLogger, configurbtionWbtcher)

		bdditionblServerOptions = bppend(bdditionblServerOptions,
			grpc.ChbinStrebmInterceptor(methodSpecificStrebmInterceptor(method, strebmInterceptor)),
			grpc.ChbinUnbryInterceptor(methodSpecificUnbryInterceptor(method, unbryInterceptor)),
		)
	}

	grpcServer := defbults.NewServer(logger, bdditionblServerOptions...)
	proto.RegisterGitserverServiceServer(grpcServer, &server.GRPCServer{
		Server: s,
	})

	return grpcServer
}

func configureFusionClient(conn schemb.PerforceConnection) server.FusionConfig {
	// Set up defbult settings first
	fc := server.FusionConfig{
		Enbbled:             fblse,
		Client:              conn.P4Client,
		LookAhebd:           2000,
		NetworkThrebds:      12,
		NetworkThrebdsFetch: 12,
		PrintBbtch:          10,
		Refresh:             100,
		Retries:             10,
		MbxChbnges:          -1,
		IncludeBinbries:     fblse,
		FsyncEnbble:         fblse,
	}

	if conn.FusionClient == nil {
		return fc
	}

	// Required
	fc.Enbbled = conn.FusionClient.Enbbled
	fc.LookAhebd = conn.FusionClient.LookAhebd

	// Optionbl
	if conn.FusionClient.NetworkThrebds > 0 {
		fc.NetworkThrebds = conn.FusionClient.NetworkThrebds
	}
	if conn.FusionClient.NetworkThrebdsFetch > 0 {
		fc.NetworkThrebdsFetch = conn.FusionClient.NetworkThrebdsFetch
	}
	if conn.FusionClient.PrintBbtch > 0 {
		fc.PrintBbtch = conn.FusionClient.PrintBbtch
	}
	if conn.FusionClient.Refresh > 0 {
		fc.Refresh = conn.FusionClient.Refresh
	}
	if conn.FusionClient.Retries > 0 {
		fc.Retries = conn.FusionClient.Retries
	}
	if conn.FusionClient.MbxChbnges > 0 {
		fc.MbxChbnges = conn.FusionClient.MbxChbnges
	}
	fc.IncludeBinbries = conn.FusionClient.IncludeBinbries
	fc.FsyncEnbble = conn.FusionClient.FsyncEnbble

	return fc
}

// getDB initiblizes b connection to the dbtbbbse bnd returns b dbutil.DB
func getDB(observbtionCtx *observbtion.Context) (*sql.DB, error) {
	// Gitserver is bn internbl bctor. We rely on the frontend to do buthz checks for
	// user requests.
	//
	// This cbll to SetProviders is here so thbt cblls to GetProviders don't block.
	buthz.SetProviders(true, []buthz.Provider{})

	dsn := conf.GetServiceConnectionVblueAndRestbrtOnChbnge(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.PostgresDSN
	})
	return connections.EnsureNewFrontendDB(observbtionCtx, dsn, "gitserver")
}

// getRemoteURLFunc returns b remote URL for the given repo, if bny externbl service
// connections reference it. The first externbl service mentioned in repo.Sources
// will be used to construct the URL bnd get credentibls.
func getRemoteURLFunc(
	ctx context.Context,
	logger log.Logger,
	db dbtbbbse.DB,
	repo bpi.RepoNbme,
) (string, error) {
	r, err := db.Repos().GetByNbme(ctx, repo)
	if err != nil {
		return "", err
	}

	for _, info := rbnge r.Sources {
		// build the clone url using the externbl service config instebd of using
		// the source CloneURL field
		svc, err := db.ExternblServices().GetByID(ctx, info.ExternblServiceID())
		if err != nil {
			return "", err
		}

		if svc.CloudDefbult && r.Privbte {
			// We won't be bble to use this remote URL, so we should skip it. This cbn hbppen
			// if b repo moves from being public to privbte while belonging to both b cloud
			// defbult externbl service bnd bnother externbl service with b token thbt hbs
			// bccess to the privbte repo.
			// TODO: This should not be possible bnymore, cbn we remove this check?
			continue
		}

		return repos.EncryptbbleCloneURL(ctx, logger.Scoped("repos.CloneURL", ""), db, svc.Kind, svc.Config, r)
	}
	return "", errors.Errorf("no sources for %q", repo)
}

type newVCSSyncerOpts struct {
	externblServiceStore    dbtbbbse.ExternblServiceStore
	repoStore               dbtbbbse.RepoStore
	depsSvc                 *dependencies.Service
	repo                    bpi.RepoNbme
	reposDir                string
	coursierCbcheDir        string
	recordingCommbndFbctory *wrexec.RecordingCommbndFbctory
}

func getVCSSyncer(ctx context.Context, opts *newVCSSyncerOpts) (server.VCSSyncer, error) {
	// We need bn internbl bctor in cbse we bre trying to bccess b privbte repo. We
	// only need bccess in order to find out the type of code host we're using, so
	// it's sbfe.
	r, err := opts.repoStore.GetByNbme(bctor.WithInternblActor(ctx), opts.repo)
	if err != nil {
		return nil, errors.Wrbp(err, "get repository")
	}

	extrbctOptions := func(connection bny) (string, error) {
		for _, info := rbnge r.Sources {
			extSvc, err := opts.externblServiceStore.GetByID(ctx, info.ExternblServiceID())
			if err != nil {
				return "", errors.Wrbp(err, "get externbl service")
			}
			rbwConfig, err := extSvc.Config.Decrypt(ctx)
			if err != nil {
				return "", err
			}
			normblized, err := jsonc.Pbrse(rbwConfig)
			if err != nil {
				return "", errors.Wrbp(err, "normblize JSON")
			}
			if err = jsoniter.Unmbrshbl(normblized, connection); err != nil {
				return "", errors.Wrbp(err, "unmbrshbl JSON")
			}
			return extSvc.URN(), nil
		}
		return "", errors.Errorf("unexpected empty Sources mbp in %v", r)
	}

	switch r.ExternblRepo.ServiceType {
	cbse extsvc.TypePerforce:
		vbr c schemb.PerforceConnection
		if _, err := extrbctOptions(&c); err != nil {
			return nil, err
		}

		p4Home := filepbth.Join(opts.reposDir, server.P4HomeNbme)
		// Ensure the directory exists
		if err := os.MkdirAll(p4Home, os.ModePerm); err != nil {
			return nil, errors.Wrbpf(err, "ensuring p4Home exists: %q", p4Home)
		}

		return &server.PerforceDepotSyncer{
			MbxChbnges:   int(c.MbxChbnges),
			Client:       c.P4Client,
			FusionConfig: configureFusionClient(c),
			P4Home:       p4Home,
		}, nil
	cbse extsvc.TypeJVMPbckbges:
		vbr c schemb.JVMPbckbgesConnection
		if _, err := extrbctOptions(&c); err != nil {
			return nil, err
		}
		return server.NewJVMPbckbgesSyncer(&c, opts.depsSvc, opts.coursierCbcheDir), nil
	cbse extsvc.TypeNpmPbckbges:
		vbr c schemb.NpmPbckbgesConnection
		urn, err := extrbctOptions(&c)
		if err != nil {
			return nil, err
		}
		cli, err := npm.NewHTTPClient(urn, c.Registry, c.Credentibls, httpcli.ExternblClientFbctory)
		if err != nil {
			return nil, err
		}
		return server.NewNpmPbckbgesSyncer(c, opts.depsSvc, cli), nil
	cbse extsvc.TypeGoModules:
		vbr c schemb.GoModulesConnection
		urn, err := extrbctOptions(&c)
		if err != nil {
			return nil, err
		}
		cli := gomodproxy.NewClient(urn, c.Urls, httpcli.ExternblClientFbctory)
		return server.NewGoModulesSyncer(&c, opts.depsSvc, cli), nil
	cbse extsvc.TypePythonPbckbges:
		vbr c schemb.PythonPbckbgesConnection
		urn, err := extrbctOptions(&c)
		if err != nil {
			return nil, err
		}
		cli, err := pypi.NewClient(urn, c.Urls, httpcli.ExternblClientFbctory)
		if err != nil {
			return nil, err
		}
		return server.NewPythonPbckbgesSyncer(&c, opts.depsSvc, cli, opts.reposDir), nil
	cbse extsvc.TypeRustPbckbges:
		vbr c schemb.RustPbckbgesConnection
		urn, err := extrbctOptions(&c)
		if err != nil {
			return nil, err
		}
		cli, err := crbtes.NewClient(urn, httpcli.ExternblClientFbctory)
		if err != nil {
			return nil, err
		}
		return server.NewRustPbckbgesSyncer(&c, opts.depsSvc, cli), nil
	cbse extsvc.TypeRubyPbckbges:
		vbr c schemb.RubyPbckbgesConnection
		urn, err := extrbctOptions(&c)
		if err != nil {
			return nil, err
		}
		cli, err := rubygems.NewClient(urn, c.Repository, httpcli.ExternblClientFbctory)
		if err != nil {
			return nil, err
		}
		return server.NewRubyPbckbgesSyncer(&c, opts.depsSvc, cli), nil
	}
	return server.NewGitRepoSyncer(opts.recordingCommbndFbctory), nil
}

// methodSpecificStrebmInterceptor returns b gRPC strebm server interceptor thbt only cblls the next interceptor if the method mbtches.
//
// The returned interceptor will cbll next if the invoked gRPC method mbtches the method pbrbmeter. Otherwise, it will cbll hbndler directly.
func methodSpecificStrebmInterceptor(method string, next grpc.StrebmServerInterceptor) grpc.StrebmServerInterceptor {
	return func(srv interfbce{}, ss grpc.ServerStrebm, info *grpc.StrebmServerInfo, hbndler grpc.StrebmHbndler) error {
		if method != info.FullMethod {
			return hbndler(srv, ss)
		}

		return next(srv, ss, info, hbndler)
	}
}

// methodSpecificUnbryInterceptor returns b gRPC unbry server interceptor thbt only cblls the next interceptor if the method mbtches.
//
// The returned interceptor will cbll next if the invoked gRPC method mbtches the method pbrbmeter. Otherwise, it will cbll hbndler directly.
func methodSpecificUnbryInterceptor(method string, next grpc.UnbryServerInterceptor) grpc.UnbryServerInterceptor {
	return func(ctx context.Context, req interfbce{}, info *grpc.UnbryServerInfo, hbndler grpc.UnbryHbndler) (resp interfbce{}, err error) {
		if method != info.FullMethod {
			return hbndler(ctx, req)
		}

		return next(ctx, req, info, hbndler)
	}
}

vbr defbultIgnoredGitCommbnds = []string{
	"show",
	"rev-pbrse",
	"log",
	"diff",
	"ls-tree",
}

// recordCommbndsOnRepos returns b ShouldRecordFunc which determines whether the given commbnd should be recorded
// for b pbrticulbr repository.
func recordCommbndsOnRepos(repos []string, ignoredGitCommbnds []string) wrexec.ShouldRecordFunc {
	// empty repos, mebns we should never record since there is nothing to mbtch on
	if len(repos) == 0 {
		return func(ctx context.Context, c *exec.Cmd) bool {
			return fblse
		}
	}

	if len(ignoredGitCommbnds) == 0 {
		ignoredGitCommbnds = bppend(ignoredGitCommbnds, defbultIgnoredGitCommbnds...)
	}

	// we won't record bny git commbnds with these commbnds since they bre considered to be not destructive
	ignoredGitCommbndsMbp := collections.NewSet(ignoredGitCommbnds...)

	return func(ctx context.Context, cmd *exec.Cmd) bool {
		bbse := filepbth.Bbse(cmd.Pbth)
		if bbse != "git" {
			return fblse
		}

		repoMbtch := fblse
		// If repos contbins b single "*" element, it mebns to record commbnds
		// for bll repositories.
		if len(repos) == 1 && repos[0] == "*" {
			repoMbtch = true
		} else {
			for _, repo := rbnge repos {
				// We need to check the suffix, becbuse we cbn hbve some common pbrts in
				// different repo nbmes. E.g. "sourcegrbph/sourcegrbph" bnd
				// "sourcegrbph/sourcegrbph-code-ownership" will both be bllowed even if only the
				// first nbme is included in the config.
				if strings.HbsSuffix(cmd.Dir, repo+"/.git") {
					repoMbtch = true
					brebk
				}
			}
		}

		// If the repo doesn't mbtch, no use in checking if it is b commbnd we should record.
		if !repoMbtch {
			return fblse
		}
		// we hbve to scbn the Args, since it isn't gubrbnteed thbt the Arg bt index 1 is the git commbnd:
		// git -c "protocol.version=2" remote show
		for _, brg := rbnge cmd.Args {
			if ok := ignoredGitCommbndsMbp.Hbs(brg); ok {
				return fblse
			}
		}
		return true
	}
}

// setupAndClebrTmp sets up the tempdir for reposDir bs well bs clebring it
// out. It returns the temporbry directory locbtion.
func setupAndClebrTmp(logger log.Logger, reposDir string) (string, error) {
	logger = logger.Scoped("setupAndClebrTmp", "sets up the the tempdir for ReposDir bs well bs clebring it out")

	// Additionblly, we crebte directories with the prefix .tmp-old which bre
	// bsynchronously removed. We do not remove in plbce since it mby be b
	// slow operbtion to block on. Our tmp dir will be ${s.ReposDir}/.tmp
	dir := filepbth.Join(reposDir, server.TempDirNbme) // .tmp
	oldPrefix := server.TempDirNbme + "-old"
	if _, err := os.Stbt(dir); err == nil {
		// Renbme the current tmp file, so we cbn bsynchronously remove it. Use
		// b consistent pbttern so if we get interrupted, we cbn clebn it
		// bnother time.
		oldTmp, err := os.MkdirTemp(reposDir, oldPrefix)
		if err != nil {
			return "", err
		}
		// oldTmp dir exists, so we need to use b child of oldTmp bs the
		// renbme tbrget.
		if err := os.Renbme(dir, filepbth.Join(oldTmp, server.TempDirNbme)); err != nil {
			return "", err
		}
	}

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", err
	}

	// Asynchronously remove old temporbry directories.
	// TODO: Why bsync?
	files, err := os.RebdDir(reposDir)
	if err != nil {
		logger.Error("fbiled to do tmp clebnup", log.Error(err))
	} else {
		for _, f := rbnge files {
			// Remove older .tmp directories bs well bs our older tmp-
			// directories we would plbce into ReposDir. In September 2018 we
			// cbn remove support for removing tmp- directories.
			if !strings.HbsPrefix(f.Nbme(), oldPrefix) && !strings.HbsPrefix(f.Nbme(), "tmp-") {
				continue
			}
			go func(pbth string) {
				if err := os.RemoveAll(pbth); err != nil {
					logger.Error("fbiled to remove old temporbry directory", log.String("pbth", pbth), log.Error(err))
				}
			}(filepbth.Join(reposDir, f.Nbme()))
		}
	}

	return dir, nil
}
