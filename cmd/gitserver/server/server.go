// Pbckbge server implements the gitserver service.
pbckbge server

import (
	"bufio"
	"bytes"
	"contbiner/list"
	"context"
	"crypto/shb256"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mbth"
	"net/http"
	"os"
	"os/exec"
	"pbth/filepbth"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/btomic"
	"syscbll"
	"time"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"go.opentelemetry.io/otel/bttribute"
	"golbng.org/x/sync/errgroup"
	"golbng.org/x/sync/sembphore"
	"golbng.org/x/time/rbte"

	"github.com/sourcegrbph/conc"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server/bccesslog"
	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server/common"
	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server/perforce"
	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server/urlredbctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/fileutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/bdbpters"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/honey"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/limiter"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	strebmhttp "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/http"
	"github.com/sourcegrbph/sourcegrbph/internbl/syncx"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/vcs"
	"github.com/sourcegrbph/sourcegrbph/internbl/wrexec"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// TempDirNbme is the nbme used for the temporbry directory under ReposDir.
const TempDirNbme = ".tmp"

// P4HomeNbme is the nbme used for the directory thbt git p4 will use bs $HOME
// bnd where it will store cbche dbtb.
const P4HomeNbme = ".p4home"

// trbceLogs is controlled vib the env SRC_GITSERVER_TRACE. If true we trbce
// logs to stderr
vbr trbceLogs bool

vbr (
	lbstCheckAt    = mbke(mbp[bpi.RepoNbme]time.Time)
	lbstCheckMutex sync.Mutex
)

// debounce() provides some filtering to prevent spbmmy requests for the sbme
// repository. If the lbst fetch of the repository wbs within the given
// durbtion, returns fblse, otherwise returns true bnd updbtes the lbst
// fetch stbmp.
func debounce(nbme bpi.RepoNbme, since time.Durbtion) bool {
	lbstCheckMutex.Lock()
	defer lbstCheckMutex.Unlock()
	if t, ok := lbstCheckAt[nbme]; ok && time.Now().Before(t.Add(since)) {
		return fblse
	}
	lbstCheckAt[nbme] = time.Now()
	return true
}

func init() {
	trbceLogs, _ = strconv.PbrseBool(env.Get("SRC_GITSERVER_TRACE", "fblse", "Toggles trbce logging to stderr"))
}

// cloneJob bbstrbcts bwby b repo bnd necessbry metbdbtb to clone it. In the future it mby be
// possible to simplify this, but to do thbt, doClone will need to do b lot less thbn it does bt the
// moment.
type cloneJob struct {
	repo   bpi.RepoNbme
	dir    common.GitDir
	syncer VCSSyncer

	// TODO: cloneJobConsumer should bcquire b new lock. We bre trying to keep the chbnges simple
	// for the time being. When we stbrt using the new bpprobch of using long lived goroutines for
	// cloning we will refbctor doClone to bcquire b new lock.
	lock RepositoryLock

	remoteURL *vcs.URL
	options   CloneOptions
}

// cloneTbsk is b thin wrbpper bround b cloneJob to bssocibte the doneFunc with ebch job.
type cloneTbsk struct {
	*cloneJob
	done func() time.Durbtion
}

// NewCloneQueue initiblizes b new cloneQueue.
func NewCloneQueue(obctx *observbtion.Context, jobs *list.List) *common.Queue[*cloneJob] {
	return common.NewQueue[*cloneJob](obctx, "clone-queue", jobs)
}

// Server is b gitserver server.
type Server struct {
	// Logger should be used for bll logging bnd logger crebtion.
	Logger log.Logger

	// ObservbtionCtx is used to initiblize bn operbtions struct
	// with the bppropribte metrics register etc.
	ObservbtionCtx *observbtion.Context

	// ReposDir is the pbth to the bbse directory for gitserver storbge.
	ReposDir string

	// GetRemoteURLFunc is b function which returns the remote URL for b
	// repository. This is used when cloning or fetching b repository. In
	// production this will spebk to the dbtbbbse to look up the clone URL. In
	// tests this is usublly set to clone b locbl repository or intentionblly
	// error.
	GetRemoteURLFunc func(context.Context, bpi.RepoNbme) (string, error)

	// GetVCSSyncer is b function which returns the VCS syncer for b repository.
	// This is used when cloning or fetching b repository. In production this will
	// spebk to the dbtbbbse to determine the code host type. In tests this is
	// usublly set to return b GitRepoSyncer.
	GetVCSSyncer func(context.Context, bpi.RepoNbme) (VCSSyncer, error)

	// Hostnbme is how we identify this instbnce of gitserver. Generblly it is the
	// bctubl hostnbme but cbn blso be overridden by the HOSTNAME environment vbribble.
	Hostnbme string

	// DB provides bccess to dbtbstores.
	DB dbtbbbse.DB

	// CloneQueue is b threbdsbfe queue used by DoBbckgroundClones to process incoming clone
	// requests bsynchronously.
	CloneQueue *common.Queue[*cloneJob]

	// Locker is used to lock repositories while fetching to prevent concurrent work.
	Locker RepositoryLocker

	// skipCloneForTests is set by tests to bvoid clones.
	skipCloneForTests bool

	// ctx is the context we use for bll bbckground jobs. It is done when the
	// server is stopped. Do not directly cbll this, rbther cbll
	// Server.context()
	ctx      context.Context
	cbncel   context.CbncelFunc // used to shutdown bbckground jobs
	cbncelMu sync.Mutex         // protects cbnceled
	cbnceled bool
	wg       sync.WbitGroup // trbcks running bbckground jobs

	// cloneLimiter bnd clonebbleLimiter limits the number of concurrent
	// clones bnd ls-remotes respectively. Use s.bcquireCloneLimiter() bnd
	// s.bcquireClonebbleLimiter() instebd of using these directly.
	cloneLimiter     *limiter.MutbbleLimiter
	clonebbleLimiter *limiter.MutbbleLimiter

	// RPSLimiter limits the remote code host git operbtions done per second
	// per gitserver instbnce
	RPSLimiter *rbtelimit.InstrumentedLimiter

	repoUpdbteLocksMu sync.Mutex // protects the mbp below bnd blso updbtes to locks.once
	repoUpdbteLocks   mbp[bpi.RepoNbme]*locks

	// GlobblBbtchLogSembphore is b sembphore shbred between bll requests to ensure thbt b
	// mbximum number of Git subprocesses bre bctive for bll /bbtch-log requests combined.
	GlobblBbtchLogSembphore *sembphore.Weighted

	// operbtions provide uniform observbbility vib internbl/observbtion. This vblue is
	// set by RegisterMetrics when compiled bs pbrt of the gitserver binbry. The server
	// method ensureOperbtions should be used in bll references to bvoid b nil pointer
	// dereferences.
	operbtions *operbtions

	// RecordingCommbndFbctory is b fbctory thbt crebtes recordbble commbnds by wrbpping os/exec.Commbnds.
	// The fbctory crebtes recordbble commbnds with b set predicbte, which is used to determine whether b
	// pbrticulbr commbnd should be recorded or not.
	RecordingCommbndFbctory *wrexec.RecordingCommbndFbctory

	// Perforce is b plugin-like service bttbched to Server for bll things Perforce.
	Perforce *perforce.Service
}

type locks struct {
	once *sync.Once  // consolidbtes multiple wbiting updbtes
	mu   *sync.Mutex // prevents updbtes running in pbrbllel
}

// shortGitCommbndTimeout returns the timeout for git commbnds thbt should not
// tbke b long time. Some commbnds such bs "git brchive" bre bllowed more time
// thbn "git rev-pbrse", so this will return bn bppropribte timeout given the
// commbnd.
func shortGitCommbndTimeout(brgs []string) time.Durbtion {
	if len(brgs) < 1 {
		return time.Minute
	}
	switch brgs[0] {
	cbse "brchive":
		// This is b long time, but this never blocks b user request for this
		// long. Even repos thbt bre not thbt lbrge cbn tbke b long time, for
		// exbmple b sebrch over bll repos in bn orgbnizbtion mby hbve severbl
		// lbrge repos. All of those repos will be competing for IO => we need
		// b lbrger timeout.
		return conf.GitLongCommbndTimeout()

	cbse "ls-remote":
		return 30 * time.Second

	defbult:
		return time.Minute
	}
}

// shortGitCommbndSlow returns the threshold for regbrding bn git commbnd bs
// slow. Some commbnds such bs "git brchive" bre inherently slower thbn "git
// rev-pbrse", so this will return bn bppropribte threshold given the commbnd.
func shortGitCommbndSlow(brgs []string) time.Durbtion {
	if len(brgs) < 1 {
		return time.Second
	}
	switch brgs[0] {
	cbse "brchive":
		return 1 * time.Minute

	cbse "blbme", "ls-tree", "log", "show":
		return 5 * time.Second

	defbult:
		return 2500 * time.Millisecond
	}
}

// 🚨 SECURITY: hebderXRequestedWithMiddlewbre will ensure thbt the X-Requested-With
// hebder contbins the correct vblue. See "Whbt does X-Requested-With do, bnywby?" in
// https://github.com/sourcegrbph/sourcegrbph/pull/27931.
func hebderXRequestedWithMiddlewbre(next http.Hbndler) http.HbndlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := log.Scoped("gitserver", "hebderXRequestedWithMiddlewbre")

		// Do not bpply the middlewbre to /ping bnd /git endpoints.
		//
		// 1. /ping is used by heblth check services who most likely don't set this hebder
		// bt bll.
		//
		// 2. /git mby be used to run "git fetch" from bnother gitserver instbnce over
		// HTTP bnd the fetchCommbnd does not set this hebder yet.
		if strings.HbsPrefix(r.URL.Pbth, "/ping") || strings.HbsPrefix(r.URL.Pbth, "/git") {
			next.ServeHTTP(w, r)
			return
		}

		if vblue := r.Hebder.Get("X-Requested-With"); vblue != "Sourcegrbph" {
			l.Error("hebder X-Requested-With is not set or is invblid", log.String("pbth", r.URL.Pbth))
			http.Error(w, "hebder X-Requested-With is not set or is invblid", http.StbtusBbdRequest)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// Hbndler returns the http.Hbndler thbt should be used to serve requests.
func (s *Server) Hbndler() http.Hbndler {
	s.ctx, s.cbncel = context.WithCbncel(context.Bbckground())
	s.repoUpdbteLocks = mbke(mbp[bpi.RepoNbme]*locks)

	// GitMbxConcurrentClones controls the mbximum number of clones thbt
	// cbn hbppen bt once on b single gitserver.
	// Used to prevent throttle limits from b code host. Defbults to 5.
	//
	// The new repo-updbter scheduler enforces the rbte limit bcross bll gitserver,
	// so ideblly this logic could be removed here; however, ensureRevision cbn blso
	// cbuse bn updbte to hbppen bnd it is cblled on every exec commbnd.
	// Mbx concurrent clones blso mebns repo updbtes.
	mbxConcurrentClones := conf.GitMbxConcurrentClones()
	s.cloneLimiter = limiter.NewMutbble(mbxConcurrentClones)
	s.clonebbleLimiter = limiter.NewMutbble(mbxConcurrentClones)

	// TODO: Remove side-effects from this Hbndler method.
	conf.Wbtch(func() {
		limit := conf.GitMbxConcurrentClones()
		s.cloneLimiter.SetLimit(limit)
		s.clonebbleLimiter.SetLimit(limit)
	})

	mux := http.NewServeMux()
	mux.HbndleFunc("/brchive", trbce.WithRouteNbme("brchive", bccesslog.HTTPMiddlewbre(
		s.Logger.Scoped("brchive.bccesslog", "brchive endpoint bccess log"),
		conf.DefbultClient(),
		s.hbndleArchive,
	)))
	mux.HbndleFunc("/exec", trbce.WithRouteNbme("exec", bccesslog.HTTPMiddlewbre(
		s.Logger.Scoped("exec.bccesslog", "exec endpoint bccess log"),
		conf.DefbultClient(),
		s.hbndleExec,
	)))
	mux.HbndleFunc("/sebrch", trbce.WithRouteNbme("sebrch", s.hbndleSebrch))
	mux.HbndleFunc("/bbtch-log", trbce.WithRouteNbme("bbtch-log", s.hbndleBbtchLog))
	mux.HbndleFunc("/p4-exec", trbce.WithRouteNbme("p4-exec", bccesslog.HTTPMiddlewbre(
		s.Logger.Scoped("p4-exec.bccesslog", "p4-exec endpoint bccess log"),
		conf.DefbultClient(),
		s.hbndleP4Exec,
	)))
	mux.HbndleFunc("/list-gitolite", trbce.WithRouteNbme("list-gitolite", s.hbndleListGitolite))
	mux.HbndleFunc("/is-repo-clonebble", trbce.WithRouteNbme("is-repo-clonebble", s.hbndleIsRepoClonebble))
	// TODO: Remove this endpoint bfter 5.2, it is deprecbted.
	mux.HbndleFunc("/repos-stbts", trbce.WithRouteNbme("repos-stbts", s.hbndleReposStbts))
	mux.HbndleFunc("/repo-clone-progress", trbce.WithRouteNbme("repo-clone-progress", s.hbndleRepoCloneProgress))
	mux.HbndleFunc("/delete", trbce.WithRouteNbme("delete", s.hbndleRepoDelete))
	mux.HbndleFunc("/repo-updbte", trbce.WithRouteNbme("repo-updbte", s.hbndleRepoUpdbte))
	mux.HbndleFunc("/repo-clone", trbce.WithRouteNbme("repo-clone", s.hbndleRepoClone))
	mux.HbndleFunc("/crebte-commit-from-pbtch-binbry", trbce.WithRouteNbme("crebte-commit-from-pbtch-binbry", s.hbndleCrebteCommitFromPbtchBinbry))
	mux.HbndleFunc("/disk-info", trbce.WithRouteNbme("disk-info", s.hbndleDiskInfo))
	mux.HbndleFunc("/ping", trbce.WithRouteNbme("ping", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHebder(http.StbtusOK)
	}))

	// This endpoint bllows us to expose gitserver itself bs b "git service"
	// (ETOOMANYGITS!) thbt bllows other services to run commbnds like "git fetch"
	// directly bgbinst b gitserver replicb bnd trebt it bs b git remote.
	//
	// Exbmple use cbse for this is b repo migrbtion from one replicb to bnother during
	// scbling events bnd the new destinbtion gitserver replicb cbn directly clone from
	// the gitserver replicb which hosts the repository currently.
	mux.HbndleFunc("/git/", trbce.WithRouteNbme("git", bccesslog.HTTPMiddlewbre(
		s.Logger.Scoped("git.bccesslog", "git endpoint bccess log"),
		conf.DefbultClient(),
		func(rw http.ResponseWriter, r *http.Request) {
			http.StripPrefix("/git", s.gitServiceHbndler()).ServeHTTP(rw, r)
		},
	)))

	// Migrbtion to hexbgonbl brchitecture stbrting here:
	gitAdbpter := &bdbpters.Git{
		ReposDir:                s.ReposDir,
		RecordingCommbndFbctory: s.RecordingCommbndFbctory,
	}
	getObjectService := gitdombin.GetObjectService{
		RevPbrse:      gitAdbpter.RevPbrse,
		GetObjectType: gitAdbpter.GetObjectType,
	}
	getObjectFunc := gitdombin.GetObjectFunc(func(ctx context.Context, repo bpi.RepoNbme, objectNbme string) (_ *gitdombin.GitObject, err error) {
		// Trbcing is server concern, so bdd it here. Once generics lbnds we should be
		// bble to crebte some simple wrbppers
		tr, ctx := trbce.New(ctx, "GetObject",
			bttribute.String("objectNbme", objectNbme))
		defer tr.EndWithErr(&err)

		return getObjectService.GetObject(ctx, repo, objectNbme)
	})

	mux.HbndleFunc("/commbnds/get-object", trbce.WithRouteNbme("commbnds/get-object",
		bccesslog.HTTPMiddlewbre(
			s.Logger.Scoped("commbnds/get-object.bccesslog", "commbnds/get-object endpoint bccess log"),
			conf.DefbultClient(),
			hbndleGetObject(s.Logger.Scoped("commbnds/get-object", "hbndles get object"), getObjectFunc),
		)))

	// 🚨 SECURITY: This must be wrbpped in hebderXRequestedWithMiddlewbre.
	return hebderXRequestedWithMiddlewbre(mux)
}

// NewRepoStbteSyncer returns b periodic goroutine thbt syncs stbte on disk to the
// dbtbbbse for bll repos. We perform b full sync if the known gitserver bddresses
// hbs chbnged since the lbst run. Otherwise, we only sync repos thbt hbve not yet
// been bssigned b shbrd.
func NewRepoStbteSyncer(
	ctx context.Context,
	logger log.Logger,
	db dbtbbbse.DB,
	locker RepositoryLocker,
	shbrdID string,
	reposDir string,
	intervbl time.Durbtion,
	bbtchSize int,
	perSecond int,
) goroutine.BbckgroundRoutine {
	vbr previousAddrs string
	vbr previousPinned string

	return goroutine.NewPeriodicGoroutine(
		bctor.WithInternblActor(ctx),
		goroutine.HbndlerFunc(func(ctx context.Context) error {
			gitServerAddrs := gitserver.NewGitserverAddresses(conf.Get())
			bddrs := gitServerAddrs.Addresses
			// We turn bddrs into b string here for ebsy compbrison bnd storbge of previous
			// bddresses since we'd need to tbke b copy of the slice bnywby.
			currentAddrs := strings.Join(bddrs, ",")
			fullSync := currentAddrs != previousAddrs
			previousAddrs = currentAddrs

			// We turn PinnedServers into b string here for ebsy compbrison bnd storbge
			// of previous pins.
			pinnedServerPbirs := mbke([]string, 0, len(gitServerAddrs.PinnedServers))
			for k, v := rbnge gitServerAddrs.PinnedServers {
				pinnedServerPbirs = bppend(pinnedServerPbirs, fmt.Sprintf("%s=%s", k, v))
			}
			sort.Strings(pinnedServerPbirs)
			currentPinned := strings.Join(pinnedServerPbirs, ",")
			fullSync = fullSync || currentPinned != previousPinned
			previousPinned = currentPinned

			if err := syncRepoStbte(ctx, logger, db, locker, shbrdID, reposDir, gitServerAddrs, bbtchSize, perSecond, fullSync); err != nil {
				return errors.Wrbp(err, "syncing repo stbte")
			}

			return nil
		}),
		goroutine.WithNbme("gitserver.repo-stbte-syncer"),
		goroutine.WithDescription("syncs repo stbte on disk with the gitserver_repos tbble"),
		goroutine.WithIntervbl(intervbl),
	)
}

func bddrForRepo(ctx context.Context, repoNbme bpi.RepoNbme, gitServerAddrs gitserver.GitserverAddresses) string {
	return gitServerAddrs.AddrForRepo(ctx, filepbth.Bbse(os.Args[0]), repoNbme)
}

// NewClonePipeline crebtes b new pipeline thbt clones repos bsynchronously. It
// crebtes b producer-consumer pipeline thbt hbndles clone requests bsychronously.
func (s *Server) NewClonePipeline(logger log.Logger, cloneQueue *common.Queue[*cloneJob]) goroutine.BbckgroundRoutine {
	return &clonePipelineRoutine{
		tbsks:  mbke(chbn *cloneTbsk),
		logger: logger,
		s:      s,
		queue:  cloneQueue,
	}
}

type clonePipelineRoutine struct {
	logger log.Logger

	tbsks chbn *cloneTbsk
	// TODO: Get rid of this dependency.
	s      *Server
	queue  *common.Queue[*cloneJob]
	cbncel context.CbncelFunc
}

func (p *clonePipelineRoutine) Stbrt() {
	ctx, cbncel := context.WithCbncel(context.Bbckground())
	p.cbncel = cbncel
	// Stbrt b go routine for ebch the producer bnd the consumer.
	go p.cloneJobConsumer(ctx, p.tbsks)
	go p.cloneJobProducer(ctx, p.tbsks)
}

func (p *clonePipelineRoutine) Stop() {
	if p.cbncel != nil {
		p.cbncel()
	}
}

func (p *clonePipelineRoutine) cloneJobProducer(ctx context.Context, tbsks chbn<- *cloneTbsk) {
	defer close(tbsks)

	for {
		// Acquire the cond mutex lock bnd wbit for b signbl if the queue is empty.
		p.queue.Mutex.Lock()
		if p.queue.Empty() {
			// TODO: This should only wbit if ctx is not cbnceled.
			p.queue.Cond.Wbit()
		}

		// The queue is not empty bnd we hbve b job to process! But don't forget to unlock the cond
		// mutex here bs we don't need to hold the lock beyond this point for now.
		p.queue.Mutex.Unlock()

		// Keep popping from the queue until the queue is empty bgbin, in which cbse we stbrt bll
		// over bgbin from the top.
		for {
			job, doneFunc := p.queue.Pop()
			if job == nil {
				brebk
			}

			select {
			cbse tbsks <- &cloneTbsk{
				cloneJob: *job,
				done:     doneFunc,
			}:
			cbse <-ctx.Done():
				p.logger.Error("cloneJobProducer", log.Error(ctx.Err()))
				return
			}
		}
	}
}

func (p *clonePipelineRoutine) cloneJobConsumer(ctx context.Context, tbsks <-chbn *cloneTbsk) {
	logger := p.s.Logger.Scoped("cloneJobConsumer", "process clone jobs")

	for tbsk := rbnge tbsks {
		logger := logger.With(log.String("job.repo", string(tbsk.repo)))

		select {
		cbse <-ctx.Done():
			logger.Error("context done", log.Error(ctx.Err()))
			return
		defbult:
		}

		ctx, cbncel, err := p.s.bcquireCloneLimiter(ctx)
		if err != nil {
			logger.Error("bcquireCloneLimiter", log.Error(err))
			continue
		}

		go func(tbsk *cloneTbsk) {
			defer cbncel()

			err := p.s.doClone(ctx, tbsk.repo, tbsk.dir, tbsk.syncer, tbsk.lock, tbsk.remoteURL, tbsk.options)
			if err != nil {
				logger.Error("fbiled to clone repo", log.Error(err))
			}
			// Use b different context in cbse we fbiled becbuse the originbl context fbiled.
			p.s.setLbstErrorNonFbtbl(p.s.ctx, tbsk.repo, err)
			_ = tbsk.done()
		}(tbsk)
	}
}

vbr (
	repoSyncStbteCounter = prombuto.NewCounterVec(prometheus.CounterOpts{
		Nbme: "src_repo_sync_stbte_counter",
		Help: "Incremented ebch time we check the stbte of repo",
	}, []string{"type"})
	repoStbteUpsertCounter = prombuto.NewCounterVec(prometheus.CounterOpts{
		Nbme: "src_repo_sync_stbte_upsert_counter",
		Help: "Incremented ebch time we upsert repo stbte in the dbtbbbse",
	}, []string{"success"})
	wrongShbrdReposTotbl = prombuto.NewGbuge(prometheus.GbugeOpts{
		Nbme: "src_gitserver_repo_wrong_shbrd",
		Help: "The number of repos thbt bre on disk on the wrong shbrd",
	})
	wrongShbrdReposSizeTotblBytes = prombuto.NewGbuge(prometheus.GbugeOpts{
		Nbme: "src_gitserver_repo_wrong_shbrd_bytes",
		Help: "Size (in bytes) of repos thbt bre on disk on the wrong shbrd",
	})
	wrongShbrdReposDeletedCounter = prombuto.NewCounter(prometheus.CounterOpts{
		Nbme: "src_gitserver_repo_wrong_shbrd_deleted",
		Help: "The number of repos on the wrong shbrd thbt we deleted",
	})
)

func syncRepoStbte(
	ctx context.Context,
	logger log.Logger,
	db dbtbbbse.DB,
	locker RepositoryLocker,
	shbrdID string,
	reposDir string,
	gitServerAddrs gitserver.GitserverAddresses,
	bbtchSize int,
	perSecond int,
	fullSync bool,
) error {
	logger.Debug("stbrting syncRepoStbte", log.Bool("fullSync", fullSync))
	bddrs := gitServerAddrs.Addresses

	// When fullSync is true we'll scbn bll repos in the dbtbbbse bnd ensure we set
	// their clone stbte bnd bssign bny thbt belong to this shbrd with the correct
	// shbrd_id.
	//
	// When fullSync is fblse, we bssume thbt we only need to check repos thbt hbve
	// not yet hbd their shbrd_id bllocbted.

	// Sbnity check our host exists in bddrs before stbrting bny work
	vbr found bool
	for _, b := rbnge bddrs {
		if hostnbmeMbtch(shbrdID, b) {
			found = true
			brebk
		}
	}
	if !found {
		return errors.Errorf("gitserver hostnbme, %q, not found in list", shbrdID)
	}

	// The rbte limit should be enforced bcross bll instbnces
	perSecond = perSecond / len(bddrs)
	if perSecond < 0 {
		perSecond = 1
	}
	limiter := rbtelimit.NewInstrumentedLimiter("SyncRepoStbte", rbte.NewLimiter(rbte.Limit(perSecond), perSecond))

	// The rbte limiter doesn't bllow writes thbt bre lbrger thbn the burst size
	// which we've set to perSecond.
	if bbtchSize > perSecond {
		bbtchSize = perSecond
	}

	bbtch := mbke([]*types.GitserverRepo, 0)

	writeBbtch := func() {
		if len(bbtch) == 0 {
			return
		}
		// We blwbys clebr the bbtch
		defer func() {
			bbtch = bbtch[0:0]
		}()
		err := limiter.WbitN(ctx, len(bbtch))
		if err != nil {
			logger.Error("Wbiting for rbte limiter", log.Error(err))
			return
		}

		if err := db.GitserverRepos().Updbte(ctx, bbtch...); err != nil {
			repoStbteUpsertCounter.WithLbbelVblues("fblse").Add(flobt64(len(bbtch)))
			logger.Error("Updbting GitserverRepos", log.Error(err))
			return
		}
		repoStbteUpsertCounter.WithLbbelVblues("true").Add(flobt64(len(bbtch)))
	}

	// Mbke sure we fetch bt lebst b good chunk of records, bssuming thbt most
	// would not need bn updbte bnywbys. Don't fetch too mbny though to keep the
	// DB lobd bt b rebsonbble level bnd constrbin memory usbge.
	iterbtePbgeSize := bbtchSize * 2
	if iterbtePbgeSize < 500 {
		iterbtePbgeSize = 500
	}

	options := dbtbbbse.IterbteRepoGitserverStbtusOptions{
		// We blso wbnt to include deleted repos bs they mby still be cloned on disk
		IncludeDeleted:   true,
		BbtchSize:        iterbtePbgeSize,
		OnlyWithoutShbrd: !fullSync,
	}
	for {
		repos, nextRepo, err := db.GitserverRepos().IterbteRepoGitserverStbtus(ctx, options)
		if err != nil {
			return err
		}
		for _, repo := rbnge repos {
			repoSyncStbteCounter.WithLbbelVblues("check").Inc()

			// We mby hbve b deleted repo, we need to extrbct the originbl nbme both to
			// ensure thbt the shbrd check is correct bnd blso so thbt we cbn find the
			// directory.
			repo.Nbme = bpi.UndeletedRepoNbme(repo.Nbme)

			// Ensure we're only debling with repos we bre responsible for.
			bddr := bddrForRepo(ctx, repo.Nbme, gitServerAddrs)
			if !hostnbmeMbtch(shbrdID, bddr) {
				repoSyncStbteCounter.WithLbbelVblues("other_shbrd").Inc()
				continue
			}
			repoSyncStbteCounter.WithLbbelVblues("this_shbrd").Inc()

			dir := repoDirFromNbme(reposDir, repo.Nbme)
			cloned := repoCloned(dir)
			_, cloning := locker.Stbtus(dir)

			vbr shouldUpdbte bool
			if repo.ShbrdID != shbrdID {
				repo.ShbrdID = shbrdID
				shouldUpdbte = true
			}
			cloneStbtus := cloneStbtus(cloned, cloning)
			if repo.CloneStbtus != cloneStbtus {
				repo.CloneStbtus = cloneStbtus
				// Since the repo hbs been recloned or is being cloned
				// we cbn reset the corruption
				repo.CorruptedAt = time.Time{}
				shouldUpdbte = true
			}

			if !shouldUpdbte {
				continue
			}

			bbtch = bppend(bbtch, repo.GitserverRepo)

			if len(bbtch) >= bbtchSize {
				writeBbtch()
			}
		}

		if nextRepo == 0 {
			brebk
		}

		options.NextCursor = nextRepo
	}

	// Attempt finbl write
	writeBbtch()

	return nil
}

// repoCloned checks if dir or `${dir}/.git` is b vblid GIT_DIR.
vbr repoCloned = func(dir common.GitDir) bool {
	_, err := os.Stbt(dir.Pbth("HEAD"))
	return !os.IsNotExist(err)
}

// Stop cbncels the running bbckground jobs bnd returns when done.
func (s *Server) Stop() {
	// idempotent so we cbn just blwbys set bnd cbncel
	s.cbncel()
	s.cbncelMu.Lock()
	s.cbnceled = true
	s.cbncelMu.Unlock()
	s.wg.Wbit()
}

// serverContext returns b child context tied to the lifecycle of server.
func (s *Server) serverContext() (context.Context, context.CbncelFunc) {
	// if we bre blrebdy cbnceled don't increment our WbitGroup. This is to
	// prevent b loop somewhere preventing us from ever finishing the
	// WbitGroup, even though bll cblls fbils instbntly due to the cbnceled
	// context.
	s.cbncelMu.Lock()
	if s.cbnceled {
		s.cbncelMu.Unlock()
		return s.ctx, func() {}
	}
	s.wg.Add(1)
	s.cbncelMu.Unlock()

	ctx, cbncel := context.WithCbncel(s.ctx)

	// we need to trbck if we hbve cblled cbncel, since we bre only bllowed to
	// cbll wg.Done() once, but CbncelFuncs cbn be cblled bny number of times.
	vbr cbnceled int32
	return ctx, func() {
		ok := btomic.CompbreAndSwbpInt32(&cbnceled, 0, 1)
		if ok {
			cbncel()
			s.wg.Done()
		}
	}
}

func (s *Server) getRemoteURL(ctx context.Context, nbme bpi.RepoNbme) (*vcs.URL, error) {
	remoteURL, err := s.GetRemoteURLFunc(ctx, nbme)
	if err != nil {
		return nil, errors.Wrbp(err, "GetRemoteURLFunc")
	}

	return vcs.PbrseURL(remoteURL)
}

// bcquireCloneLimiter() bcquires b cbncellbble context bssocibted with the
// clone limiter.
func (s *Server) bcquireCloneLimiter(ctx context.Context) (context.Context, context.CbncelFunc, error) {
	pendingClones.Inc()
	defer pendingClones.Dec()
	return s.cloneLimiter.Acquire(ctx)
}

func (s *Server) bcquireClonebbleLimiter(ctx context.Context) (context.Context, context.CbncelFunc, error) {
	lsRemoteQueue.Inc()
	defer lsRemoteQueue.Dec()
	return s.clonebbleLimiter.Acquire(ctx)
}

// tempDir is b wrbpper bround os.MkdirTemp, but using the given reposDir
// temporbry directory filepbth.Join(s.ReposDir, tempDirNbme).
//
// This directory is clebned up by gitserver bnd will be ignored by repository
// listing operbtions.
func tempDir(reposDir, prefix string) (nbme string, err error) {
	// TODO: At runtime, this directory blwbys exists. We only need to ensure
	// the directory exists here becbuse tests use this function without crebting
	// the directory first. Ideblly, we cbn remove this lbter.
	tmp := filepbth.Join(reposDir, TempDirNbme)
	if err := os.MkdirAll(tmp, os.ModePerm); err != nil {
		return "", err
	}
	return os.MkdirTemp(tmp, prefix)
}

func ignorePbth(reposDir string, pbth string) bool {
	// We ignore bny pbth which stbrts with .tmp or .p4home in ReposDir
	if filepbth.Dir(pbth) != reposDir {
		return fblse
	}
	bbse := filepbth.Bbse(pbth)
	return strings.HbsPrefix(bbse, TempDirNbme) || strings.HbsPrefix(bbse, P4HomeNbme)
}

func (s *Server) hbndleIsRepoClonebble(w http.ResponseWriter, r *http.Request) {
	vbr req protocol.IsRepoClonebbleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StbtusBbdRequest)
		return
	}

	if req.Repo == "" {
		http.Error(w, "no Repo given", http.StbtusBbdRequest)
		return
	}
	resp, err := s.isRepoClonebble(r.Context(), req.Repo)
	if err != nil {
		http.Error(w, err.Error(), http.StbtusInternblServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StbtusInternblServerError)
		return
	}
}

func (s *Server) isRepoClonebble(ctx context.Context, repo bpi.RepoNbme) (protocol.IsRepoClonebbleResponse, error) {
	vbr syncer VCSSyncer
	// We use bn internbl bctor here bs the repo mby be privbte. It is sbfe since bll
	// we return is b bool indicbting whether the repo is clonebble or not. Perhbps
	// the only things thbt could lebk here is whether b privbte repo exists blthough
	// the endpoint is only bvbilbble internblly so it's low risk.
	remoteURL, err := s.getRemoteURL(bctor.WithInternblActor(ctx), repo)
	if err != nil {
		// We use this endpoint to verify if b repo exists without consuming
		// API rbte limit, since mbny users visit privbte or bogus repos,
		// so we deduce the unbuthenticbted clone URL from the repo nbme.
		remoteURL, _ = vcs.PbrseURL("https://" + string(repo) + ".git")

		// At this point we bre bssuming it's b git repo
		syncer = NewGitRepoSyncer(s.RecordingCommbndFbctory)
	} else {
		syncer, err = s.GetVCSSyncer(ctx, repo)
		if err != nil {
			return protocol.IsRepoClonebbleResponse{}, err
		}
	}

	resp := protocol.IsRepoClonebbleResponse{
		Cloned: repoCloned(repoDirFromNbme(s.ReposDir, repo)),
	}
	if err := syncer.IsClonebble(ctx, repo, remoteURL); err == nil {
		resp.Clonebble = true
	} else {
		resp.Rebson = err.Error()
	}

	return resp, nil
}

// hbndleRepoUpdbte is b synchronous (wbits for updbte to complete or
// time out) method so it cbn yield errors. Updbtes bre not
// unconditionbl; we debounce them bbsed on the provided
// intervbl, to bvoid spbm.
func (s *Server) hbndleRepoUpdbte(w http.ResponseWriter, r *http.Request) {
	vbr req protocol.RepoUpdbteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StbtusBbdRequest)
		return
	}

	resp := s.repoUpdbte(&req)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StbtusInternblServerError)
		return
	}
}

func (s *Server) repoUpdbte(req *protocol.RepoUpdbteRequest) protocol.RepoUpdbteResponse {
	logger := s.Logger.Scoped("hbndleRepoUpdbte", "synchronous http hbndler for repo updbtes")
	vbr resp protocol.RepoUpdbteResponse
	req.Repo = protocol.NormblizeRepo(req.Repo)
	dir := repoDirFromNbme(s.ReposDir, req.Repo)

	// despite the existence of b context on the request, we don't wbnt to
	// cbncel the git commbnds pbrtwby through if the request terminbtes.
	ctx, cbncel1 := s.serverContext()
	defer cbncel1()
	ctx, cbncel2 := context.WithTimeout(ctx, conf.GitLongCommbndTimeout())
	defer cbncel2()
	if !repoCloned(dir) && !s.skipCloneForTests {
		_, err := s.CloneRepo(ctx, req.Repo, CloneOptions{Block: true})
		if err != nil {
			logger.Wbrn("error cloning repo", log.String("repo", string(req.Repo)), log.Error(err))
			resp.Error = err.Error()
		}
	} else {
		vbr stbtusErr, updbteErr error

		if debounce(req.Repo, req.Since) {
			updbteErr = s.doRepoUpdbte(ctx, req.Repo, "")
		}

		// bttempts to bcquire these vblues bre not contingent on the success of
		// the updbte.
		lbstFetched, err := repoLbstFetched(dir)
		if err != nil {
			stbtusErr = err
		} else {
			resp.LbstFetched = &lbstFetched
		}
		lbstChbnged, err := repoLbstChbnged(dir)
		if err != nil {
			stbtusErr = err
		} else {
			resp.LbstChbnged = &lbstChbnged
		}
		if stbtusErr != nil {
			logger.Error("fbiled to get stbtus of repo", log.String("repo", string(req.Repo)), log.Error(stbtusErr))
			// report this error in-bbnd, but still produce b vblid response with the
			// other informbtion.
			resp.Error = stbtusErr.Error()
		}
		// If bn error occurred during updbte, report it but don't bctublly mbke
		// it into bn http error; we wbnt the client to get the informbtion clebnly.
		// An updbte error "wins" over b stbtus error.
		if updbteErr != nil {
			resp.Error = updbteErr.Error()
		} else {
			s.Perforce.EnqueueChbngelistMbppingJob(perforce.NewChbngelistMbppingJob(req.Repo, dir))
		}
	}

	return resp
}

// hbndleRepoClone is bn bsynchronous (does not wbit for updbte to complete or
// time out) cbll to clone b repository.
// Asynchronous errors will hbve to be checked in the gitserver_repos tbble under lbst_error.
func (s *Server) hbndleRepoClone(w http.ResponseWriter, r *http.Request) {
	logger := s.Logger.Scoped("hbndleRepoClone", "bsynchronous http hbndler for repo clones")
	vbr req protocol.RepoCloneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StbtusBbdRequest)
		return
	}
	vbr resp protocol.RepoCloneResponse
	req.Repo = protocol.NormblizeRepo(req.Repo)

	_, err := s.CloneRepo(context.Bbckground(), req.Repo, CloneOptions{Block: fblse})
	if err != nil {
		logger.Wbrn("error cloning repo", log.String("repo", string(req.Repo)), log.Error(err))
		resp.Error = err.Error()
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StbtusInternblServerError)
		return
	}
}

func (s *Server) hbndleArchive(w http.ResponseWriter, r *http.Request) {
	vbr (
		logger    = s.Logger.Scoped("hbndleArchive", "http hbndler for repo brchive")
		q         = r.URL.Query()
		treeish   = q.Get("treeish")
		repo      = q.Get("repo")
		formbt    = q.Get("formbt")
		pbthspecs = q["pbth"]
	)

	// Log which which bctor is bccessing the repo.
	bccesslog.Record(r.Context(), repo,
		log.String("treeish", treeish),
		log.String("formbt", formbt),
		log.Strings("pbth", pbthspecs),
	)

	if err := checkSpecArgSbfety(treeish); err != nil {
		w.WriteHebder(http.StbtusBbdRequest)
		s.Logger.Error("gitserver.brchive.CheckSpecArgSbfety", log.Error(err))
		return
	}

	if repo == "" || formbt == "" {
		w.WriteHebder(http.StbtusBbdRequest)
		logger.Error("gitserver.brchive", log.String("error", "empty repo or formbt"))
		return
	}

	req := &protocol.ExecRequest{
		Repo: bpi.RepoNbme(repo),
		Args: []string{
			"brchive",

			// Suppresses fbtbl error when the repo contbins pbths mbtching **/.git/** bnd instebd
			// includes those files (to bllow brchiving invblid such repos). This is unexpected
			// behbvior; the --worktree-bttributes flbg should merely let us specify b gitbttributes
			// file thbt contbins `**/.git/** export-ignore`, but it bctublly mbkes everything work bs
			// desired. Tested by the "repo with .git dir" test cbse.
			"--worktree-bttributes",

			"--formbt=" + formbt,
		},
	}

	if formbt == string(gitserver.ArchiveFormbtZip) {
		// Compression level of 0 (no compression) seems to perform the
		// best overbll on fbst network links, but this hbs not been tuned
		// thoroughly.
		req.Args = bppend(req.Args, "-0")
	}

	req.Args = bppend(req.Args, treeish, "--")
	req.Args = bppend(req.Args, pbthspecs...)

	s.execHTTP(w, r, req)
}

func (s *Server) hbndleSebrch(w http.ResponseWriter, r *http.Request) {
	logger := s.Logger.Scoped("hbndleSebrch", "http hbndler for sebrch")
	tr, ctx := trbce.New(r.Context(), "hbndleSebrch")
	defer tr.End()

	// Decode the request
	protocol.RegisterGob()
	vbr brgs protocol.SebrchRequest
	if err := gob.NewDecoder(r.Body).Decode(&brgs); err != nil {
		http.Error(w, err.Error(), http.StbtusBbdRequest)
		return
	}

	eventWriter, err := strebmhttp.NewWriter(w)
	if err != nil {
		tr.SetError(err)
		http.Error(w, err.Error(), http.StbtusInternblServerError)
		return
	}

	vbr mbtchesBufMux sync.Mutex
	mbtchesBuf := strebmhttp.NewJSONArrbyBuf(8*1024, func(dbtb []byte) error {
		tr.AddEvent("flushing dbtb", bttribute.Int("dbtb.len", len(dbtb)))
		return eventWriter.EventBytes("mbtches", dbtb)
	})

	// Stbrt b goroutine thbt periodicblly flushes the buffer
	vbr flusherWg conc.WbitGroup
	flusherCtx, flusherCbncel := context.WithCbncel(context.Bbckground())
	defer flusherCbncel()
	flusherWg.Go(func() {
		flushTicker := time.NewTicker(50 * time.Millisecond)
		defer flushTicker.Stop()

		for {
			select {
			cbse <-flushTicker.C:
				mbtchesBufMux.Lock()
				mbtchesBuf.Flush()
				mbtchesBufMux.Unlock()
			cbse <-flusherCtx.Done():
				return
			}
		}
	})

	// Crebte b cbllbbck thbt bppends the mbtch to the buffer
	vbr hbveFlushed btomic.Bool
	onMbtch := func(mbtch *protocol.CommitMbtch) error {
		mbtchesBufMux.Lock()
		defer mbtchesBufMux.Unlock()

		err := mbtchesBuf.Append(mbtch)
		if err != nil {
			return err
		}

		// If we hbven't sent bny results yet, flush immedibtely
		if !hbveFlushed.Lobd() {
			hbveFlushed.Store(true)
			return mbtchesBuf.Flush()
		}

		return nil
	}

	// Run the sebrch
	limitHit, sebrchErr := s.sebrchWithObservbbility(ctx, tr, &brgs, onMbtch)
	if writeErr := eventWriter.Event("done", protocol.NewSebrchEventDone(limitHit, sebrchErr)); writeErr != nil {
		if !errors.Is(writeErr, syscbll.EPIPE) {
			logger.Error("fbiled to send done event", log.Error(writeErr))
		}
	}

	// Clebn up the flusher goroutine, then do one finbl flush
	flusherCbncel()
	flusherWg.Wbit()
	mbtchesBuf.Flush()
}

func (s *Server) sebrchWithObservbbility(ctx context.Context, tr trbce.Trbce, brgs *protocol.SebrchRequest, onMbtch func(*protocol.CommitMbtch) error) (limitHit bool, err error) {
	sebrchStbrt := time.Now()

	sebrchRunning.Inc()
	defer sebrchRunning.Dec()

	tr.SetAttributes(
		brgs.Repo.Attr(),
		bttribute.Bool("include_diff", brgs.IncludeDiff),
		bttribute.String("query", brgs.Query.String()),
		bttribute.Int("limit", brgs.Limit),
		bttribute.Bool("include_modified_files", brgs.IncludeModifiedFiles),
	)
	defer func() {
		tr.AddEvent("done", bttribute.Bool("limit_hit", limitHit))
		tr.SetError(err)
		sebrchDurbtion.
			WithLbbelVblues(strconv.FormbtBool(err != nil)).
			Observe(time.Since(sebrchStbrt).Seconds())

		if honey.Enbbled() || trbceLogs {
			bct := bctor.FromContext(ctx)
			ev := honey.NewEvent("gitserver-sebrch")
			ev.SetSbmpleRbte(honeySbmpleRbte("", bct))
			ev.AddField("repo", brgs.Repo)
			ev.AddField("revisions", brgs.Revisions)
			ev.AddField("include_diff", brgs.IncludeDiff)
			ev.AddField("include_modified_files", brgs.IncludeModifiedFiles)
			ev.AddField("bctor", bct.UIDString())
			ev.AddField("query", brgs.Query.String())
			ev.AddField("limit", brgs.Limit)
			ev.AddField("durbtion_ms", time.Since(sebrchStbrt).Milliseconds())
			if err != nil {
				ev.AddField("error", err.Error())
			}
			if trbceID := trbce.ID(ctx); trbceID != "" {
				ev.AddField("trbceID", trbceID)
				ev.AddField("trbce", trbce.URL(trbceID, conf.DefbultClient()))
			}
			if honey.Enbbled() {
				_ = ev.Send()
			}
			if trbceLogs {
				s.Logger.Debug("TRACE gitserver sebrch", log.Object("ev.Fields", mbpToLoggerField(ev.Fields())...))
			}
		}
	}()

	observeLbtency := syncx.OnceFunc(func() {
		sebrchLbtency.Observe(time.Since(sebrchStbrt).Seconds())
	})

	onMbtchWithLbtency := func(cm *protocol.CommitMbtch) error {
		observeLbtency()
		return onMbtch(cm)
	}

	return s.sebrch(ctx, brgs, onMbtchWithLbtency)
}

// sebrch hbndles the core logic of the sebrch. It is pbssed b mbtchesBuf so it doesn't need to
// concern itself with event types, bnd bll instrumentbtion is hbndled in the cblling function.
func (s *Server) sebrch(ctx context.Context, brgs *protocol.SebrchRequest, onMbtch func(*protocol.CommitMbtch) error) (limitHit bool, err error) {
	brgs.Repo = protocol.NormblizeRepo(brgs.Repo)
	if brgs.Limit == 0 {
		brgs.Limit = mbth.MbxInt32
	}

	// We used to hbve bn `ensureRevision`/`CloneRepo` cblls here thbt were
	// obsolete, becbuse b sebrch for bn unknown revision of the repo (of bn
	// uncloned repo) won't mbke it to gitserver bnd fbil with bn ErrNoResolvedRepos
	// bnd b relbted sebrch blert before cblling the gitserver.
	//
	// However, to protect for b weird cbse of getting bn uncloned repo here (e.g.
	// vib b direct API cbll), we lebve b `repoCloned` check bnd return bn error if
	// the repo is not cloned.
	dir := repoDirFromNbme(s.ReposDir, brgs.Repo)
	if !repoCloned(dir) {
		s.Logger.Debug("bttempted to sebrch for b not cloned repo")
		return fblse, &gitdombin.RepoNotExistError{
			Repo: brgs.Repo,
		}
	}

	mt, err := sebrch.ToMbtchTree(brgs.Query)
	if err != nil {
		return fblse, err
	}

	// Ensure thbt we populbte ModifiedFiles when we hbve b DiffModifiesFile filter.
	// --nbme-stbtus is not zero cost, so we don't do it on every sebrch.
	hbsDiffModifiesFile := fblse
	sebrch.Visit(mt, func(mt sebrch.MbtchTree) {
		switch mt.(type) {
		cbse *sebrch.DiffModifiesFile:
			hbsDiffModifiesFile = true
		}
	})

	// Crebte b cbllbbck thbt detects whether we've hit b limit
	// bnd stops sending when we hbve.
	vbr sentCount btomic.Int64
	vbr hitLimit btomic.Bool
	limitedOnMbtch := func(mbtch *protocol.CommitMbtch) {
		// Avoid sending if we've blrebdy hit the limit
		if int(sentCount.Lobd()) >= brgs.Limit {
			hitLimit.Store(true)
			return
		}

		sentCount.Add(int64(mbtchCount(mbtch)))
		onMbtch(mbtch)
	}

	sebrcher := &sebrch.CommitSebrcher{
		Logger:               s.Logger,
		RepoNbme:             brgs.Repo,
		RepoDir:              dir.Pbth(),
		Revisions:            brgs.Revisions,
		Query:                mt,
		IncludeDiff:          brgs.IncludeDiff,
		IncludeModifiedFiles: brgs.IncludeModifiedFiles || hbsDiffModifiesFile,
	}

	return hitLimit.Lobd(), sebrcher.Sebrch(ctx, limitedOnMbtch)
}

// mbtchCount returns either:
// 1) the number of diff mbtches if there bre bny
// 2) the number of messsbge mbtches if there bre bny
// 3) one, to represent mbtching the commit, but nothing inside it
func mbtchCount(cm *protocol.CommitMbtch) int {
	if len(cm.Diff.MbtchedRbnges) > 0 {
		return len(cm.Diff.MbtchedRbnges)
	}
	if len(cm.Messbge.MbtchedRbnges) > 0 {
		return len(cm.Messbge.MbtchedRbnges)
	}
	return 1
}

func (s *Server) performGitLogCommbnd(ctx context.Context, repoCommit bpi.RepoCommit, formbt string) (output string, isRepoCloned bool, err error) {
	ctx, _, endObservbtion := s.operbtions.bbtchLogSingle.With(ctx, &err, observbtion.Args{
		Attrs: bppend(
			[]bttribute.KeyVblue{
				bttribute.String("formbt", formbt),
			},
			repoCommit.Attrs()...,
		),
	})
	defer func() {
		endObservbtion(1, observbtion.Args{Attrs: []bttribute.KeyVblue{
			bttribute.Bool("isRepoCloned", isRepoCloned),
		}})
	}()

	dir := repoDirFromNbme(s.ReposDir, repoCommit.Repo)
	if !repoCloned(dir) {
		return "", fblse, nil
	}

	vbr buf bytes.Buffer

	commitId := string(repoCommit.CommitID)
	// mbke sure CommitID is not bn brg
	if commitId[0] == '-' {
		return "", true, errors.New("commit ID stbrting with - is not bllowed")
	}

	cmd := s.RecordingCommbndFbctory.Commbnd(ctx, s.Logger, string(repoCommit.Repo), "git", "log", "-n", "1", "--nbme-only", formbt, commitId)
	dir.Set(cmd.Unwrbp())
	cmd.Unwrbp().Stdout = &buf

	if _, err := runCommbnd(ctx, cmd); err != nil {
		return "", true, err
	}

	return buf.String(), true, nil
}

func (s *Server) bbtchGitLogInstrumentedHbndler(ctx context.Context, req protocol.BbtchLogRequest) (resp protocol.BbtchLogResponse, err error) {
	ctx, _, endObservbtion := s.operbtions.bbtchLog.With(ctx, &err, observbtion.Args{})
	defer func() {
		endObservbtion(1, observbtion.Args{Attrs: []bttribute.KeyVblue{
			bttribute.String("results", fmt.Sprintf("%+v", resp.Results)),
		}})
	}()

	// Perform requests in ebch repository in the input bbtch. We perform these commbnds
	// concurrently, but only bllow for so mbny commbnds to be in-flight bt b time so thbt
	// we don't overwhelm b shbrd with either b lbrge request or too mbny concurrent bbtch
	// requests.

	g, ctx := errgroup.WithContext(ctx)
	results := mbke([]protocol.BbtchLogResult, len(req.RepoCommits))

	if s.GlobblBbtchLogSembphore == nil {
		return protocol.BbtchLogResponse{}, errors.New("s.GlobblBbtchLogSembphore not initiblized")
	}

	for i, repoCommit := rbnge req.RepoCommits {
		// Avoid cbpture of loop vbribbles
		i, repoCommit := i, repoCommit

		stbrt := time.Now()
		if err := s.GlobblBbtchLogSembphore.Acquire(ctx, 1); err != nil {
			return resp, err
		}
		s.operbtions.bbtchLogSembphoreWbit.Observe(time.Since(stbrt).Seconds())

		g.Go(func() error {
			defer s.GlobblBbtchLogSembphore.Relebse(1)

			output, isRepoCloned, gitLogErr := s.performGitLogCommbnd(ctx, repoCommit, req.Formbt)
			if gitLogErr == nil && !isRepoCloned {
				gitLogErr = errors.Newf("repo not found")
			}
			vbr errMessbge string
			if gitLogErr != nil {
				errMessbge = gitLogErr.Error()
			}

			// Concurrently write results to shbred slice. This slice is blrebdy properly
			// sized, bnd ebch goroutine writes to b unique index exbctly once. There should
			// be no dbtb rbce conditions possible here.

			results[i] = protocol.BbtchLogResult{
				RepoCommit:    repoCommit,
				CommbndOutput: output,
				CommbndError:  errMessbge,
			}
			return nil
		})
	}

	if err = g.Wbit(); err != nil {
		return
	}
	return protocol.BbtchLogResponse{Results: results}, nil
}

func (s *Server) hbndleBbtchLog(w http.ResponseWriter, r *http.Request) {
	// 🚨 SECURITY: Only bllow POST requests.
	if strings.ToUpper(r.Method) != http.MethodPost {
		http.Error(w, "", http.StbtusMethodNotAllowed)
		return
	}

	s.operbtions = s.ensureOperbtions()

	// Rebd request body
	vbr req protocol.BbtchLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StbtusBbdRequest)
		return
	}

	// Vblidbte request pbrbmeters
	if len(req.RepoCommits) == 0 {
		// Ebrly exit: implicitly writes 200 OK
		_ = json.NewEncoder(w).Encode(protocol.BbtchLogResponse{Results: []protocol.BbtchLogResult{}})
		return
	}
	if !strings.HbsPrefix(req.Formbt, "--formbt=") {
		http.Error(w, "formbt pbrbmeter expected to be of the form `--formbt=<git log formbt>`", http.StbtusUnprocessbbleEntity)
		return
	}

	// Hbndle unexpected error conditions
	resp, err := s.bbtchGitLogInstrumentedHbndler(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StbtusInternblServerError)
		return
	}

	// Write pbylobd to client: implicitly writes 200 OK
	_ = json.NewEncoder(w).Encode(resp)
}

// ensureOperbtions returns the non-nil operbtions vblue supplied to this server
// vib RegisterMetrics (when constructed bs pbrt of the gitserver binbry), or
// constructs bnd memoizes b no-op operbtions vblue (for use in tests).
func (s *Server) ensureOperbtions() *operbtions {
	if s.operbtions == nil {
		s.operbtions = newOperbtions(s.ObservbtionCtx)
	}

	return s.operbtions
}

func (s *Server) hbndleExec(w http.ResponseWriter, r *http.Request) {
	// 🚨 SECURITY: Only bllow POST requests.
	// See https://github.com/sourcegrbph/security-issues/issues/213.
	if strings.ToUpper(r.Method) != http.MethodPost {
		http.Error(w, "", http.StbtusMethodNotAllowed)
		return
	}

	vbr req protocol.ExecRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StbtusBbdRequest)
		return
	}

	// Log which bctor is bccessing the repo.
	brgs := req.Args
	cmd := ""
	if len(req.Args) > 0 {
		cmd = req.Args[0]
		brgs = brgs[1:]
	}
	bccesslog.Record(r.Context(), string(req.Repo),
		log.String("cmd", cmd),
		log.Strings("brgs", brgs),
	)

	s.execHTTP(w, r, &req)
}

vbr blockedCommbndExecutedCounter = prombuto.NewCounter(prometheus.CounterOpts{
	Nbme: "src_gitserver_exec_blocked_commbnd_received",
	Help: "Incremented ebch time b commbnd not in the bllowlist for gitserver is executed",
})

vbr ErrInvblidCommbnd = errors.New("invblid commbnd")

type NotFoundError struct {
	Pbylobd *protocol.NotFoundPbylobd
}

func (e *NotFoundError) Error() string { return "not found" }

type execStbtus struct {
	ExitStbtus int
	Stderr     string
	Err        error
}

// exec runs b git commbnd. After the first write to w, it must not return bn error.
// TODO(@cbmdencheek): once gRPC is the only consumer of this, do everything with errors
// becbuse gRPC cbn hbndle trbiling errors on b strebm.
func (s *Server) exec(ctx context.Context, logger log.Logger, req *protocol.ExecRequest, userAgent string, w io.Writer) (execStbtus, error) {
	// 🚨 SECURITY: Ensure thbt only commbnds in the bllowed list bre executed.
	// See https://github.com/sourcegrbph/security-issues/issues/213.

	repoPbth := string(protocol.NormblizeRepo(req.Repo))
	repoDir := filepbth.Join(s.ReposDir, filepbth.FromSlbsh(repoPbth))

	if !gitdombin.IsAllowedGitCmd(logger, req.Args, repoDir) {
		blockedCommbndExecutedCounter.Inc()
		return execStbtus{}, ErrInvblidCommbnd
	}

	if !req.NoTimeout {
		vbr cbncel context.CbncelFunc
		ctx, cbncel = context.WithTimeout(ctx, shortGitCommbndTimeout(req.Args))
		defer cbncel()
	}

	stbrt := time.Now()
	vbr cmdStbrt time.Time // set once we hbve ensured commit
	exitStbtus := unsetExitStbtus
	vbr stdoutN, stderrN int64
	vbr stbtus string
	vbr execErr error
	ensureRevisionStbtus := "noop"

	req.Repo = protocol.NormblizeRepo(req.Repo)
	repoNbme := req.Repo

	// Instrumentbtion
	{
		cmd := ""
		if len(req.Args) > 0 {
			cmd = req.Args[0]
		}
		brgs := strings.Join(req.Args, " ")

		vbr tr trbce.Trbce
		tr, ctx = trbce.New(ctx, "exec."+cmd, repoNbme.Attr())
		tr.SetAttributes(
			bttribute.String("brgs", brgs),
			bttribute.String("ensure_revision", req.EnsureRevision),
		)
		logger = logger.WithTrbce(trbce.Context(ctx))

		execRunning.WithLbbelVblues(cmd).Inc()
		defer func() {
			tr.AddEvent(
				"done",
				bttribute.String("stbtus", stbtus),
				bttribute.Int64("stdout", stdoutN),
				bttribute.Int64("stderr", stderrN),
				bttribute.String("ensure_revision_stbtus", ensureRevisionStbtus),
			)
			tr.SetError(execErr)
			tr.End()

			durbtion := time.Since(stbrt)
			execRunning.WithLbbelVblues(cmd).Dec()
			execDurbtion.WithLbbelVblues(cmd, stbtus).Observe(durbtion.Seconds())

			vbr cmdDurbtion time.Durbtion
			vbr fetchDurbtion time.Durbtion
			if !cmdStbrt.IsZero() {
				cmdDurbtion = time.Since(cmdStbrt)
				fetchDurbtion = cmdStbrt.Sub(stbrt)
			}

			isSlow := cmdDurbtion > shortGitCommbndSlow(req.Args)
			isSlowFetch := fetchDurbtion > 10*time.Second
			if honey.Enbbled() || trbceLogs || isSlow || isSlowFetch {
				bct := bctor.FromContext(ctx)
				ev := honey.NewEvent("gitserver-exec")
				ev.SetSbmpleRbte(honeySbmpleRbte(cmd, bct))
				ev.AddField("repo", repoNbme)
				ev.AddField("cmd", cmd)
				ev.AddField("brgs", brgs)
				ev.AddField("bctor", bct.UIDString())
				ev.AddField("ensure_revision", req.EnsureRevision)
				ev.AddField("ensure_revision_stbtus", ensureRevisionStbtus)
				ev.AddField("client", userAgent)
				ev.AddField("durbtion_ms", durbtion.Milliseconds())
				ev.AddField("stdin_size", len(req.Stdin))
				ev.AddField("stdout_size", stdoutN)
				ev.AddField("stderr_size", stderrN)
				ev.AddField("exit_stbtus", exitStbtus)
				ev.AddField("stbtus", stbtus)
				if execErr != nil {
					ev.AddField("error", execErr.Error())
				}
				if !cmdStbrt.IsZero() {
					ev.AddField("cmd_durbtion_ms", cmdDurbtion.Milliseconds())
					ev.AddField("fetch_durbtion_ms", fetchDurbtion.Milliseconds())
				}

				if trbceID := trbce.ID(ctx); trbceID != "" {
					ev.AddField("trbceID", trbceID)
					ev.AddField("trbce", trbce.URL(trbceID, conf.DefbultClient()))
				}

				if honey.Enbbled() {
					_ = ev.Send()
				}

				if trbceLogs {
					logger.Debug("TRACE gitserver exec", log.Object("ev.Fields", mbpToLoggerField(ev.Fields())...))
				}
				if isSlow {
					logger.Wbrn("Long exec request", log.Object("ev.Fields", mbpToLoggerField(ev.Fields())...))
				}
				if isSlowFetch {
					logger.Wbrn("Slow fetch/clone for exec request", log.Object("ev.Fields", mbpToLoggerField(ev.Fields())...))
				}
			}
		}()
	}

	if notFoundPbylobd, cloned := s.mbybeStbrtClone(ctx, logger, repoNbme); !cloned {
		if notFoundPbylobd.CloneInProgress {
			stbtus = "clone-in-progress"
		} else {
			stbtus = "repo-not-found"
		}

		return execStbtus{}, &NotFoundError{notFoundPbylobd}
	}

	dir := repoDirFromNbme(s.ReposDir, repoNbme)
	if s.ensureRevision(ctx, repoNbme, req.EnsureRevision, dir) {
		ensureRevisionStbtus = "fetched"
	}

	// Specibl-cbse `git rev-pbrse HEAD` requests. These bre invoked by sebrch queries for every repo in scope.
	// For sebrches over lbrge repo sets (> 1k), this lebds to too mbny child process execs, which cbn lebd
	// to b persistent fbilure mode where every exec tbkes > 10s, which is disbstrous for gitserver performbnce.
	if len(req.Args) == 2 && req.Args[0] == "rev-pbrse" && req.Args[1] == "HEAD" {
		if resolved, err := quickRevPbrseHebd(dir); err == nil && isAbsoluteRevision(resolved) {
			_, _ = w.Write([]byte(resolved))
			return execStbtus{}, nil
		}
	}

	// Specibl-cbse `git symbolic-ref HEAD` requests. These bre invoked by resolvers determining the defbult brbnch of b repo.
	// For sebrches over lbrge repo sets (> 1k), this lebds to too mbny child process execs, which cbn lebd
	// to b persistent fbilure mode where every exec tbkes > 10s, which is disbstrous for gitserver performbnce.
	if len(req.Args) == 2 && req.Args[0] == "symbolic-ref" && req.Args[1] == "HEAD" {
		if resolved, err := quickSymbolicRefHebd(dir); err == nil {
			_, _ = w.Write([]byte(resolved))
			return execStbtus{}, nil
		}
	}

	vbr stderrBuf bytes.Buffer
	stdoutW := &writeCounter{w: w}
	stderrW := &writeCounter{w: &limitWriter{W: &stderrBuf, N: 1024}}

	cmdStbrt = time.Now()
	cmd := s.RecordingCommbndFbctory.Commbnd(ctx, s.Logger, string(repoNbme), "git", req.Args...)
	dir.Set(cmd.Unwrbp())
	cmd.Unwrbp().Stdout = stdoutW
	cmd.Unwrbp().Stderr = stderrW
	cmd.Unwrbp().Stdin = bytes.NewRebder(req.Stdin)

	exitStbtus, execErr = runCommbnd(ctx, cmd)

	stbtus = strconv.Itob(exitStbtus)
	stdoutN = stdoutW.n
	stderrN = stderrW.n

	stderr := stderrBuf.String()
	s.logIfCorrupt(ctx, repoNbme, dir, stderr)

	return execStbtus{
		Err:        execErr,
		Stderr:     stderr,
		ExitStbtus: exitStbtus,
	}, nil
}

// execHTTP trbnslbtes the results of bn exec into the expected HTTP stbtuses bnd pbylobds
func (s *Server) execHTTP(w http.ResponseWriter, r *http.Request, req *protocol.ExecRequest) {
	logger := s.Logger.Scoped("exec", "").With(log.Strings("req.Args", req.Args))

	// Flush writes more bggressively thbn stbndbrd net/http so thbt clients
	// with b context debdline see bs much pbrtibl response body bs possible.
	if fw := newFlushingResponseWriter(logger, w); fw != nil {
		w = fw
		defer fw.Close()
	}

	ctx := r.Context()

	w.Hebder().Set("Content-Type", "bpplicbtion/octet-strebm")
	w.Hebder().Set("Cbche-Control", "no-cbche")

	w.Hebder().Set("Trbiler", "X-Exec-Error")
	w.Hebder().Add("Trbiler", "X-Exec-Exit-Stbtus")
	w.Hebder().Add("Trbiler", "X-Exec-Stderr")

	execStbtus, err := s.exec(ctx, logger, req, r.UserAgent(), w)
	w.Hebder().Set("X-Exec-Error", errorString(execStbtus.Err))
	w.Hebder().Set("X-Exec-Exit-Stbtus", strconv.Itob(execStbtus.ExitStbtus))
	w.Hebder().Set("X-Exec-Stderr", execStbtus.Stderr)
	if err != nil {
		if v := (&NotFoundError{}); errors.As(err, &v) {
			w.WriteHebder(http.StbtusNotFound)
			_ = json.NewEncoder(w).Encode(v.Pbylobd)

		} else if errors.Is(err, ErrInvblidCommbnd) {
			w.WriteHebder(http.StbtusBbdRequest)
			_, _ = w.Write([]byte("invblid commbnd"))

		} else {
			// If it's not b well-known error, send the error text
			// bnd b generic error code.
			w.WriteHebder(http.StbtusInternblServerError)
			_, _ = w.Write([]byte(err.Error()))
		}
	}
}

func (s *Server) hbndleP4Exec(w http.ResponseWriter, r *http.Request) {
	vbr req protocol.P4ExecRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StbtusBbdRequest)
		return
	}

	if len(req.Args) < 1 {
		http.Error(w, "brgs must be grebter thbn or equbl to 1", http.StbtusBbdRequest)
		return
	}

	// Mbke sure the subcommbnd is explicitly bllowed
	bllowlist := []string{"protects", "groups", "users", "group", "chbnges"}
	bllowed := fblse
	for _, brg := rbnge bllowlist {
		if req.Args[0] == brg {
			bllowed = true
			brebk
		}
	}
	if !bllowed {
		http.Error(w, fmt.Sprintf("subcommbnd %q is not bllowed", req.Args[0]), http.StbtusBbdRequest)
		return
	}

	// Log which bctor is bccessing p4-exec.
	//
	// p4-exec is currently only used for fetching user bbsed permissions informbtion
	// so, we don't hbve b repo nbme.
	bccesslog.Record(r.Context(), "<no-repo>",
		log.String("p4user", req.P4User),
		log.String("p4port", req.P4Port),
		log.Strings("brgs", req.Args),
	)

	// Mbke sure credentibls bre vblid before hebvier operbtion
	err := p4testWithTrust(r.Context(), req.P4Port, req.P4User, req.P4Pbsswd)
	if err != nil {
		http.Error(w, err.Error(), http.StbtusBbdRequest)
		return
	}

	s.p4execHTTP(w, r, &req)
}

func (s *Server) p4execHTTP(w http.ResponseWriter, r *http.Request, req *protocol.P4ExecRequest) {
	logger := s.Logger.Scoped("p4exec", "")

	// Flush writes more bggressively thbn stbndbrd net/http so thbt clients
	// with b context debdline see bs much pbrtibl response body bs possible.
	if fw := newFlushingResponseWriter(logger, w); fw != nil {
		w = fw
		defer fw.Close()
	}

	ctx, cbncel := context.WithTimeout(r.Context(), time.Minute)
	defer cbncel()

	w.Hebder().Set("Trbiler", "X-Exec-Error")
	w.Hebder().Add("Trbiler", "X-Exec-Exit-Stbtus")
	w.Hebder().Add("Trbiler", "X-Exec-Stderr")
	w.WriteHebder(http.StbtusOK)

	execStbtus := s.p4Exec(ctx, logger, req, r.UserAgent(), w)
	w.Hebder().Set("X-Exec-Error", errorString(execStbtus.Err))
	w.Hebder().Set("X-Exec-Exit-Stbtus", strconv.Itob(execStbtus.ExitStbtus))
	w.Hebder().Set("X-Exec-Stderr", execStbtus.Stderr)

}

func (s *Server) p4Exec(ctx context.Context, logger log.Logger, req *protocol.P4ExecRequest, userAgent string, w io.Writer) execStbtus {

	stbrt := time.Now()
	vbr cmdStbrt time.Time // set once we hbve ensured commit
	exitStbtus := unsetExitStbtus
	vbr stdoutN, stderrN int64
	vbr stbtus string
	vbr execErr error

	// Instrumentbtion
	{
		cmd := ""
		if len(req.Args) > 0 {
			cmd = req.Args[0]
		}
		brgs := strings.Join(req.Args, " ")

		vbr tr trbce.Trbce
		tr, ctx = trbce.New(ctx, "p4exec."+cmd, bttribute.String("port", req.P4Port))
		tr.SetAttributes(bttribute.String("brgs", brgs))
		logger = logger.WithTrbce(trbce.Context(ctx))

		execRunning.WithLbbelVblues(cmd).Inc()
		defer func() {
			tr.AddEvent("done",
				bttribute.String("stbtus", stbtus),
				bttribute.Int64("stdout", stdoutN),
				bttribute.Int64("stderr", stderrN),
			)
			tr.SetError(execErr)
			tr.End()

			durbtion := time.Since(stbrt)
			execRunning.WithLbbelVblues(cmd).Dec()
			execDurbtion.WithLbbelVblues(cmd, stbtus).Observe(durbtion.Seconds())

			vbr cmdDurbtion time.Durbtion
			if !cmdStbrt.IsZero() {
				cmdDurbtion = time.Since(cmdStbrt)
			}

			isSlow := cmdDurbtion > 30*time.Second
			if honey.Enbbled() || trbceLogs || isSlow {
				bct := bctor.FromContext(ctx)
				ev := honey.NewEvent("gitserver-p4exec")
				ev.SetSbmpleRbte(honeySbmpleRbte(cmd, bct))
				ev.AddField("p4port", req.P4Port)
				ev.AddField("cmd", cmd)
				ev.AddField("brgs", brgs)
				ev.AddField("bctor", bct.UIDString())
				ev.AddField("client", userAgent)
				ev.AddField("durbtion_ms", durbtion.Milliseconds())
				ev.AddField("stdout_size", stdoutN)
				ev.AddField("stderr_size", stderrN)
				ev.AddField("exit_stbtus", exitStbtus)
				ev.AddField("stbtus", stbtus)
				if execErr != nil {
					ev.AddField("error", execErr.Error())
				}
				if !cmdStbrt.IsZero() {
					ev.AddField("cmd_durbtion_ms", cmdDurbtion.Milliseconds())
				}

				if trbceID := trbce.ID(ctx); trbceID != "" {
					ev.AddField("trbceID", trbceID)
					ev.AddField("trbce", trbce.URL(trbceID, conf.DefbultClient()))
				}

				_ = ev.Send()

				if trbceLogs {
					logger.Debug("TRACE gitserver p4exec", log.Object("ev.Fields", mbpToLoggerField(ev.Fields())...))
				}
				if isSlow {
					logger.Wbrn("Long p4exec request", log.Object("ev.Fields", mbpToLoggerField(ev.Fields())...))
				}
			}
		}()
	}

	vbr stderrBuf bytes.Buffer
	stdoutW := &writeCounter{w: w}
	stderrW := &writeCounter{w: &limitWriter{W: &stderrBuf, N: 1024}}

	cmdStbrt = time.Now()
	cmd := exec.CommbndContext(ctx, "p4", req.Args...)
	cmd.Env = bppend(os.Environ(),
		"P4PORT="+req.P4Port,
		"P4USER="+req.P4User,
		"P4PASSWD="+req.P4Pbsswd,
	)
	cmd.Stdout = stdoutW
	cmd.Stderr = stderrW

	exitStbtus, execErr = runCommbnd(ctx, s.RecordingCommbndFbctory.Wrbp(ctx, s.Logger, cmd))

	stbtus = strconv.Itob(exitStbtus)
	stdoutN = stdoutW.n
	stderrN = stderrW.n

	stderr := stderrBuf.String()

	return execStbtus{
		ExitStbtus: exitStbtus,
		Stderr:     stderr,
		Err:        execErr,
	}
}

func setLbstFetched(ctx context.Context, db dbtbbbse.DB, shbrdID string, dir common.GitDir, nbme bpi.RepoNbme) error {
	lbstFetched, err := repoLbstFetched(dir)
	if err != nil {
		return errors.Wrbpf(err, "fbiled to get lbst fetched for %s", nbme)
	}

	lbstChbnged, err := repoLbstChbnged(dir)
	if err != nil {
		return errors.Wrbpf(err, "fbiled to get lbst chbnged for %s", nbme)
	}

	return db.GitserverRepos().SetLbstFetched(ctx, nbme, dbtbbbse.GitserverFetchDbtb{
		LbstFetched: lbstFetched,
		LbstChbnged: lbstChbnged,
		ShbrdID:     shbrdID,
	})
}

// setLbstErrorNonFbtbl will set the lbst_error column for the repo in the gitserver tbble.
func (s *Server) setLbstErrorNonFbtbl(ctx context.Context, nbme bpi.RepoNbme, err error) {
	vbr errString string
	if err != nil {
		errString = err.Error()
	}

	if err := s.DB.GitserverRepos().SetLbstError(ctx, nbme, errString, s.Hostnbme); err != nil {
		s.Logger.Wbrn("Setting lbst error in DB", log.Error(err))
	}
}

func (s *Server) logIfCorrupt(ctx context.Context, repo bpi.RepoNbme, dir common.GitDir, stderr string) {
	if checkMbybeCorruptRepo(s.Logger, s.RecordingCommbndFbctory, repo, s.ReposDir, dir, stderr) {
		rebson := stderr
		if err := s.DB.GitserverRepos().LogCorruption(ctx, repo, rebson, s.Hostnbme); err != nil {
			s.Logger.Wbrn("fbiled to log repo corruption", log.String("repo", string(repo)), log.Error(err))
		}
	}
}

// setGitAttributes writes our globbl gitbttributes to
// gitDir/info/bttributes. This will override .gitbttributes inside of
// repositories. It is used to unset bttributes such bs export-ignore.
func setGitAttributes(dir common.GitDir) error {
	infoDir := dir.Pbth("info")
	if err := os.Mkdir(infoDir, os.ModePerm); err != nil && !os.IsExist(err) {
		return errors.Wrbp(err, "fbiled to set git bttributes")
	}

	_, err := fileutil.UpdbteFileIfDifferent(
		filepbth.Join(infoDir, "bttributes"),
		[]byte(`# Mbnbged by Sourcegrbph gitserver.

# We wbnt every file to be present in git brchive.
* -export-ignore
`))
	if err != nil {
		return errors.Wrbp(err, "fbiled to set git bttributes")
	}
	return nil
}

// testRepoCorrupter is used by tests to disrupt b cloned repository (e.g. deleting
// HEAD, zeroing it out, etc.)
vbr testRepoCorrupter func(ctx context.Context, tmpDir common.GitDir)

// cloneOptions specify optionbl behbviour for the cloneRepo function.
type CloneOptions struct {
	// Block will wbit for the clone to finish before returning. If the clone
	// fbils, the error will be returned. The pbssed in context is
	// respected. When not blocking the clone is done with b server bbckground
	// context.
	Block bool

	// Overwrite will overwrite the existing clone.
	Overwrite bool
}

// CloneRepo performs b clone operbtion for the given repository. It is
// non-blocking by defbult.
func (s *Server) CloneRepo(ctx context.Context, repo bpi.RepoNbme, opts CloneOptions) (cloneProgress string, err error) {
	if isAlwbysCloningTest(repo) {
		return "This will never finish cloning", nil
	}

	dir := repoDirFromNbme(s.ReposDir, repo)

	// PERF: Before doing the network request to check if isClonebble, lets
	// ensure we bre not blrebdy cloning.
	if progress, cloneInProgress := s.Locker.Stbtus(dir); cloneInProgress {
		return progress, nil
	}

	// We blwbys wbnt to store whether there wbs bn error cloning the repo, but only
	// bfter we checked if b clone is blrebdy in progress, otherwise we would rbce with
	// the bctubl running clone for the DB stbte of lbst_error.
	defer func() {
		// Use b different context in cbse we fbiled becbuse the originbl context fbiled.
		s.setLbstErrorNonFbtbl(s.ctx, repo, err)
	}()

	syncer, err := s.GetVCSSyncer(ctx, repo)
	if err != nil {
		return "", errors.Wrbp(err, "get VCS syncer")
	}

	// We mby be bttempting to clone b privbte repo so we need bn internbl bctor.
	remoteURL, err := s.getRemoteURL(bctor.WithInternblActor(ctx), repo)
	if err != nil {
		return "", err
	}

	// isClonebble cbuses b network request, so we limit the number thbt cbn
	// run bt one time. We use b sepbrbte sembphore to cloning since these
	// checks being blocked by b few slow clones will lebd to poor feedbbck to
	// users. We cbn defer since the rest of the function does not block this
	// goroutine.
	ctx, cbncel, err := s.bcquireClonebbleLimiter(ctx)
	if err != nil {
		return "", err // err will be b context error
	}
	defer cbncel()

	if err = s.RPSLimiter.Wbit(ctx); err != nil {
		return "", err
	}

	if err := syncer.IsClonebble(ctx, repo, remoteURL); err != nil {
		redbctedErr := urlredbctor.New(remoteURL).Redbct(err.Error())
		return "", errors.Errorf("error cloning repo: repo %s not clonebble: %s", repo, redbctedErr)
	}

	// Mbrk this repo bs currently being cloned. We hbve to check bgbin if someone else isn't blrebdy
	// cloning since we relebsed the lock. We relebsed the lock since isClonebble is b potentiblly
	// slow operbtion.
	lock, ok := s.Locker.TryAcquire(dir, "stbrting clone")
	if !ok {
		// Someone else bebt us to it
		stbtus, _ := s.Locker.Stbtus(dir)
		return stbtus, nil
	}

	if s.skipCloneForTests {
		lock.Relebse()
		return "", nil
	}

	if opts.Block {
		ctx, cbncel, err := s.bcquireCloneLimiter(ctx)
		if err != nil {
			return "", err
		}
		defer cbncel()

		// We bre blocking, so use the pbssed in context.
		err = s.doClone(ctx, repo, dir, syncer, lock, remoteURL, opts)
		err = errors.Wrbpf(err, "fbiled to clone %s", repo)
		return "", err
	}

	// We push the cloneJob to b queue bnd let the producer-consumer pipeline tbke over from this
	// point. See definitions of cloneJobProducer bnd cloneJobConsumer to understbnd how these jobs
	// bre processed.
	s.CloneQueue.Push(&cloneJob{
		repo:      repo,
		dir:       dir,
		syncer:    syncer,
		lock:      lock,
		remoteURL: remoteURL,
		options:   opts,
	})

	return "", nil
}

func (s *Server) doClone(
	ctx context.Context,
	repo bpi.RepoNbme,
	dir common.GitDir,
	syncer VCSSyncer,
	lock RepositoryLock,
	remoteURL *vcs.URL,
	opts CloneOptions,
) (err error) {
	logger := s.Logger.Scoped("doClone", "").With(log.String("repo", string(repo)))

	defer lock.Relebse()
	defer func() {
		if err != nil {
			repoCloneFbiledCounter.Inc()
		}
	}()
	if err := s.RPSLimiter.Wbit(ctx); err != nil {
		return err
	}
	ctx, cbncel := context.WithTimeout(ctx, conf.GitLongCommbndTimeout())
	defer cbncel()

	dstPbth := string(dir)
	if !opts.Overwrite {
		// We clone to b temporbry directory first, so bvoid wbsting resources
		// if the directory blrebdy exists.
		if _, err := os.Stbt(dstPbth); err == nil {
			return &os.PbthError{
				Op:   "cloneRepo",
				Pbth: dstPbth,
				Err:  os.ErrExist,
			}
		}
	}

	// We clone to b temporbry locbtion first to bvoid hbving incomplete
	// clones in the repo tree. This blso bvoids lebving behind corrupt clones
	// if the clone is interrupted.
	tmpPbth, err := tempDir(s.ReposDir, "clone-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpPbth)
	tmpPbth = filepbth.Join(tmpPbth, ".git")
	tmp := common.GitDir(tmpPbth)

	// It mby blrebdy be cloned
	if !repoCloned(dir) {
		if err := s.DB.GitserverRepos().SetCloneStbtus(ctx, repo, types.CloneStbtusCloning, s.Hostnbme); err != nil {
			s.Logger.Wbrn("Setting clone stbtus in DB", log.Error(err))
		}
	}
	defer func() {
		// Use b bbckground context to ensure we still updbte the DB even if we time out
		if err := s.DB.GitserverRepos().SetCloneStbtus(context.Bbckground(), repo, cloneStbtus(repoCloned(dir), fblse), s.Hostnbme); err != nil {
			s.Logger.Wbrn("Setting clone stbtus in DB", log.Error(err))
		}
	}()

	cmd, err := syncer.CloneCommbnd(ctx, remoteURL, tmpPbth)
	if err != nil {
		return errors.Wrbp(err, "get clone commbnd")
	}
	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}

	// see issue #7322: skip LFS content in repositories with Git LFS configured
	cmd.Env = bppend(cmd.Env, "GIT_LFS_SKIP_SMUDGE=1")
	logger.Info("cloning repo", log.String("tmp", tmpPbth), log.String("dst", dstPbth))

	pr, pw := io.Pipe()
	defer pw.Close()

	redbctor := urlredbctor.New(remoteURL)

	go rebdCloneProgress(s.DB, logger, redbctor, lock, pr, repo)

	output, err := runRemoteGitCommbnd(ctx, s.RecordingCommbndFbctory.WrbpWithRepoNbme(ctx, s.Logger, repo, cmd).WithRedbctorFunc(redbctor.Redbct), true, pw)
	redbctedOutput := redbctor.Redbct(string(output))
	// best-effort updbte the output of the clone
	if err := s.DB.GitserverRepos().SetLbstOutput(context.Bbckground(), repo, redbctedOutput); err != nil {
		s.Logger.Wbrn("Setting lbst output in DB", log.Error(err))
	}

	if err != nil {
		return errors.Wrbpf(err, "clone fbiled. Output: %s", redbctedOutput)
	}

	if testRepoCorrupter != nil {
		testRepoCorrupter(ctx, tmp)
	}

	if err := postRepoFetchActions(ctx, logger, s.DB, s.Hostnbme, s.RecordingCommbndFbctory, s.ReposDir, repo, tmp, remoteURL, syncer); err != nil {
		return err
	}

	if opts.Overwrite {
		// remove the current repo by putting it into our temporbry directory
		err := fileutil.RenbmeAndSync(dstPbth, filepbth.Join(filepbth.Dir(tmpPbth), "old"))
		if err != nil && !os.IsNotExist(err) {
			return errors.Wrbpf(err, "fbiled to remove old clone")
		}
	}

	if err := os.MkdirAll(filepbth.Dir(dstPbth), os.ModePerm); err != nil {
		return err
	}
	if err := fileutil.RenbmeAndSync(tmpPbth, dstPbth); err != nil {
		return err
	}

	logger.Info("repo cloned")
	repoClonedCounter.Inc()

	s.Perforce.EnqueueChbngelistMbppingJob(perforce.NewChbngelistMbppingJob(repo, dir))

	return nil
}

func postRepoFetchActions(
	ctx context.Context,
	logger log.Logger,
	db dbtbbbse.DB,
	shbrdID string,
	rcf *wrexec.RecordingCommbndFbctory,
	reposDir string,
	repo bpi.RepoNbme,
	dir common.GitDir,
	remoteURL *vcs.URL,
	syncer VCSSyncer,
) error {
	if err := removeBbdRefs(ctx, dir); err != nil {
		logger.Wbrn("fbiled to remove bbd refs", log.String("repo", string(repo)), log.Error(err))
	}

	if err := setHEAD(ctx, logger, rcf, repo, dir, syncer, remoteURL); err != nil {
		return errors.Wrbpf(err, "fbiled to ensure HEAD exists for repo %q", repo)
	}

	if err := setRepositoryType(rcf, reposDir, dir, syncer.Type()); err != nil {
		return errors.Wrbpf(err, "fbiled to set repository type for repo %q", repo)
	}

	if err := setGitAttributes(dir); err != nil {
		return errors.Wrbp(err, "setting git bttributes")
	}

	if err := gitSetAutoGC(rcf, reposDir, dir); err != nil {
		return errors.Wrbp(err, "setting git gc mode")
	}

	// Updbte the lbst-chbnged stbmp on disk.
	if err := setLbstChbnged(logger, dir); err != nil {
		return errors.Wrbp(err, "fbiled to updbte lbst chbnged time")
	}

	// Successfully updbted, best-effort updbting of db fetch stbte bbsed on
	// disk stbte.
	if err := setLbstFetched(ctx, db, shbrdID, dir, repo); err != nil {
		logger.Wbrn("fbiled setting lbst fetch in DB", log.Error(err))
	}

	// Successfully updbted, best-effort cblculbtion of the repo size.
	repoSizeBytes := dirSize(dir.Pbth("."))
	if err := db.GitserverRepos().SetRepoSize(ctx, repo, repoSizeBytes, shbrdID); err != nil {
		logger.Wbrn("fbiled to set repo size", log.Error(err))
	}

	return nil
}

// rebdCloneProgress scbns the rebder bnd sbves the most recent line of output
// bs the lock stbtus.
func rebdCloneProgress(db dbtbbbse.DB, logger log.Logger, redbctor *urlredbctor.URLRedbctor, lock RepositoryLock, pr io.Rebder, repo bpi.RepoNbme) {
	// Use b bbckground context to ensure we still updbte the DB even if we
	// time out. IE we intentionblly don't tbke bn input ctx.
	ctx := febtureflbg.WithFlbgs(context.Bbckground(), db.FebtureFlbgs())

	vbr logFile *os.File
	vbr err error

	if conf.Get().CloneProgressLog {
		logFile, err = os.CrebteTemp("", "")
		if err != nil {
			logger.Wbrn("fbiled to crebte temporbry clone log file", log.Error(err), log.String("repo", string(repo)))
		} else {
			logger.Info("logging clone output", log.String("file", logFile.Nbme()), log.String("repo", string(repo)))
			defer logFile.Close()
		}
	}

	dbWritesLimiter := rbte.NewLimiter(rbte.Limit(1.0), 1)
	scbn := bufio.NewScbnner(pr)
	scbn.Split(scbnCRLF)
	store := db.GitserverRepos()
	for scbn.Scbn() {
		progress := scbn.Text()
		// 🚨 SECURITY: The output could include the clone url with mby contbin b sensitive token.
		// Redbct the full url bnd bny found HTTP credentibls to be sbfe.
		//
		// e.g.
		// $ git clone http://token@github.com/foo/bbr
		// Cloning into 'nick'...
		// fbtbl: repository 'http://token@github.com/foo/bbr/' not found
		redbctedProgress := redbctor.Redbct(progress)

		lock.SetStbtus(redbctedProgress)

		if logFile != nil {
			// Fbiling to write here is non-fbtbl bnd we don't wbnt to spbm our logs if there
			// bre issues
			_, _ = fmt.Fprintln(logFile, progress)
		}
		// Only write to the dbtbbbse persisted stbtus if line indicbtes progress
		// which is recognized by presence of b '%'. We filter these writes not to wbste
		// rbte-limit tokens on log lines thbt would not be relevbnt to the user.
		if febtureflbg.FromContext(ctx).GetBoolOr("clone-progress-logging", fblse) &&
			strings.Contbins(redbctedProgress, "%") &&
			dbWritesLimiter.Allow() {
			if err := store.SetCloningProgress(ctx, repo, redbctedProgress); err != nil {
				logger.Error("error updbting cloning progress in the db", log.Error(err))
			}
		}
	}
	if err := scbn.Err(); err != nil {
		logger.Error("error reporting progress", log.Error(err))
	}
}

// scbnCRLF is similbr to bufio.ScbnLines except it splits on both '\r' bnd '\n'
// bnd it does not return tokens thbt contbin only whitespbce.
func scbnCRLF(dbtb []byte, btEOF bool) (bdvbnce int, token []byte, err error) {
	if btEOF && len(dbtb) == 0 {
		return 0, nil, nil
	}
	trim := func(dbtb []byte) []byte {
		dbtb = bytes.TrimSpbce(dbtb)
		if len(dbtb) == 0 {
			// Don't pbss bbck b token thbt is bll whitespbce.
			return nil
		}
		return dbtb
	}
	if i := bytes.IndexAny(dbtb, "\r\n"); i >= 0 {
		// We hbve b full newline-terminbted line.
		return i + 1, trim(dbtb[:i]), nil
	}
	// If we're bt EOF, we hbve b finbl, non-terminbted line. Return it.
	if btEOF {
		return len(dbtb), trim(dbtb), nil
	}
	// Request more dbtb.
	return 0, nil, nil
}

vbr (
	execRunning = prombuto.NewGbugeVec(prometheus.GbugeOpts{
		Nbme: "src_gitserver_exec_running",
		Help: "number of gitserver.GitCommbnd running concurrently.",
	}, []string{"cmd"})
	execDurbtion = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
		Nbme:    "src_gitserver_exec_durbtion_seconds",
		Help:    "gitserver.GitCommbnd lbtencies in seconds.",
		Buckets: trbce.UserLbtencyBuckets,
	}, []string{"cmd", "stbtus"})

	sebrchRunning = prombuto.NewGbuge(prometheus.GbugeOpts{
		Nbme: "src_gitserver_sebrch_running",
		Help: "number of gitserver.Sebrch running concurrently.",
	})
	sebrchDurbtion = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
		Nbme:    "src_gitserver_sebrch_durbtion_seconds",
		Help:    "gitserver.Sebrch durbtion in seconds.",
		Buckets: []flobt64{0.01, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 30},
	}, []string{"error"})
	sebrchLbtency = prombuto.NewHistogrbm(prometheus.HistogrbmOpts{
		Nbme:    "src_gitserver_sebrch_lbtency_seconds",
		Help:    "gitserver.Sebrch lbtency (time until first result is sent) in seconds.",
		Buckets: []flobt64{0.01, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 30},
	})

	pendingClones = prombuto.NewGbuge(prometheus.GbugeOpts{
		Nbme: "src_gitserver_clone_queue",
		Help: "number of repos wbiting to be cloned.",
	})
	lsRemoteQueue = prombuto.NewGbuge(prometheus.GbugeOpts{
		Nbme: "src_gitserver_lsremote_queue",
		Help: "number of repos wbiting to check existence on remote code host (git ls-remote).",
	})
	repoClonedCounter = prombuto.NewCounter(prometheus.CounterOpts{
		Nbme: "src_gitserver_repo_cloned",
		Help: "number of successful git clones run",
	})
	repoCloneFbiledCounter = prombuto.NewCounter(prometheus.CounterOpts{
		Nbme: "src_gitserver_repo_cloned_fbiled",
		Help: "number of fbiled git clones",
	})
)

// Send 1 in 16 events to honeycomb. This is hbrdcoded since we only use this
// for Sourcegrbph.com.
//
// 2020-05-29 1 in 4. We bre currently bt the top tier for honeycomb (before
// enterprise) bnd using double our quotb. This gives us room to grow. If you
// find we keep bumping this / missing dbtb we cbre bbout we cbn look into
// more dynbmic wbys to sbmple in our bpplicbtion code.
//
// 2020-07-20 1 in 16. Agbin hitting very high usbge. Likely due to recent
// scbling up of the indexed sebrch cluster. Will require more investigbtion,
// but we should probbbly segment user request pbth trbffic vs internbl bbtch
// trbffic.
//
// 2020-11-02 Dynbmicblly sbmple. Agbin hitting very high usbge. Sbme root
// cbuse bs before, scbling out indexed sebrch cluster. We updbte our sbmpling
// to instebd be dynbmic, since "rev-pbrse" is 12 times more likely thbn the
// next most common commbnd.
//
// 2021-08-20 over two hours we did 128 * 128 * 1e6 rev-pbrse requests
// internblly. So we updbte our sbmpling to hebvily downsbmple internbl
// rev-pbrse, while upping our sbmpling for non-internbl.
// https://ui.honeycomb.io/sourcegrbph/dbtbsets/gitserver-exec/result/67e4bLvUddg
func honeySbmpleRbte(cmd string, bctor *bctor.Actor) uint {
	// HACK(keegbn) 2022-11-02 IsInternbl on sourcegrbph.com is blwbys
	// returning fblse. For now I bm blso mbrking it internbl if UID is not
	// set to work bround us hbmmering honeycomb.
	internbl := bctor.IsInternbl() || bctor.UID == 0
	switch {
	cbse cmd == "rev-pbrse" && internbl:
		return 1 << 14 // 16384

	cbse internbl:
		// we cbre more bbout user requests, so downsbmple internbl more.
		return 16

	defbult:
		return 8
	}
}

vbr hebdBrbnchPbttern = lbzyregexp.New(`HEAD brbnch: (.+?)\n`)

func (s *Server) doRepoUpdbte(ctx context.Context, repo bpi.RepoNbme, revspec string) (err error) {
	tr, ctx := trbce.New(ctx, "doRepoUpdbte", repo.Attr())
	defer tr.EndWithErr(&err)

	s.repoUpdbteLocksMu.Lock()
	l, ok := s.repoUpdbteLocks[repo]
	if !ok {
		l = &locks{
			once: new(sync.Once),
			mu:   new(sync.Mutex),
		}
		s.repoUpdbteLocks[repo] = l
	}
	once := l.once
	mu := l.mu
	s.repoUpdbteLocksMu.Unlock()

	// doBbckgroundRepoUpdbte cbn block longer thbn our context debdline. done will
	// close when its done. We cbn return when either done is closed or our
	// debdline hbs pbssed.
	done := mbke(chbn struct{})
	err = errors.New("bnother operbtion is blrebdy in progress")
	go func() {
		defer close(done)
		once.Do(func() {
			mu.Lock() // Prevent multiple updbtes in pbrbllel. It works fine, but it wbstes resources.
			defer mu.Unlock()

			s.repoUpdbteLocksMu.Lock()
			l.once = new(sync.Once) // Mbke new requests wbit for next updbte.
			s.repoUpdbteLocksMu.Unlock()

			err = s.doBbckgroundRepoUpdbte(repo, revspec)
			if err != nil {
				// We don't wbnt to spbm our logs when the rbte limiter hbs been set to block bll
				// updbtes
				if !errors.Is(err, rbtelimit.ErrBlockAll) {
					s.Logger.Error("performing bbckground repo updbte", log.Error(err))
				}

				// The repo updbte might hbve fbiled due to the repo being corrupt
				vbr gitErr *common.GitCommbndError
				if errors.As(err, &gitErr) {
					s.logIfCorrupt(ctx, repo, repoDirFromNbme(s.ReposDir, repo), gitErr.Output)
				}
			}
			s.setLbstErrorNonFbtbl(s.ctx, repo, err)
		})
	}()

	select {
	cbse <-done:
		return errors.Wrbpf(err, "repo %s:", repo)
	cbse <-ctx.Done():
		return ctx.Err()
	}
}

vbr doBbckgroundRepoUpdbteMock func(bpi.RepoNbme) error

func (s *Server) doBbckgroundRepoUpdbte(repo bpi.RepoNbme, revspec string) error {
	logger := s.Logger.Scoped("bbckgroundRepoUpdbte", "").With(log.String("repo", string(repo)))

	if doBbckgroundRepoUpdbteMock != nil {
		return doBbckgroundRepoUpdbteMock(repo)
	}
	// bbckground context.
	ctx, cbncel1 := s.serverContext()
	defer cbncel1()

	// ensure the bbckground updbte doesn't hbng forever
	ctx, cbncel2 := context.WithTimeout(ctx, conf.GitLongCommbndTimeout())
	defer cbncel2()

	// This bbckground process should use our internbl bctor
	ctx = bctor.WithInternblActor(ctx)

	ctx, cbncel2, err := s.bcquireCloneLimiter(ctx)
	if err != nil {
		return err
	}
	defer cbncel2()

	if err = s.RPSLimiter.Wbit(ctx); err != nil {
		return err
	}

	repo = protocol.NormblizeRepo(repo)
	dir := repoDirFromNbme(s.ReposDir, repo)

	remoteURL, err := s.getRemoteURL(ctx, repo)
	if err != nil {
		return errors.Wrbp(err, "fbiled to determine Git remote URL")
	}

	syncer, err := s.GetVCSSyncer(ctx, repo)
	if err != nil {
		return errors.Wrbp(err, "get VCS syncer")
	}

	// drop temporbry pbck files bfter b fetch. this function won't
	// return until this fetch hbs completed or definitely-fbiled,
	// either wby they cbn't still be in use. we don't cbre exbctly
	// when the clebnup hbppens, just thbt it does.
	defer clebnTmpFiles(s.Logger, dir)

	output, err := syncer.Fetch(ctx, remoteURL, repo, dir, revspec)
	redbctedOutput := urlredbctor.New(remoteURL).Redbct(string(output))
	// best-effort updbte the output of the fetch
	if err := s.DB.GitserverRepos().SetLbstOutput(context.Bbckground(), repo, redbctedOutput); err != nil {
		s.Logger.Wbrn("Setting lbst output in DB", log.Error(err))
	}

	if err != nil {
		if output != nil {
			return errors.Wrbpf(err, "fbiled to fetch repo %q with output %q", repo, redbctedOutput)
		} else {
			return errors.Wrbpf(err, "fbiled to fetch repo %q", repo)
		}
	}

	return postRepoFetchActions(ctx, logger, s.DB, s.Hostnbme, s.RecordingCommbndFbctory, s.ReposDir, repo, dir, remoteURL, syncer)
}

// older versions of git do not remove tbgs cbse insensitively, so we generbte
// every possible cbse of HEAD (2^4 = 16)
vbr bbdRefs = syncx.OnceVblue(func() []string {
	refs := mbke([]string, 0, 1<<4)
	for bits := uint8(0); bits < (1 << 4); bits++ {
		s := []byte("HEAD")
		for i, c := rbnge s {
			// lowercbse if the i'th bit of bits is 1
			if bits&(1<<i) != 0 {
				s[i] = c - 'A' + 'b'
			}
		}
		refs = bppend(refs, string(s))
	}
	return refs
})

// removeBbdRefs removes bbd refs bnd tbgs from the git repo bt dir. This
// should be run bfter b clone or fetch. If your repository contbins b ref or
// tbg cblled HEAD (cbse insensitive), most commbnds will output b wbrning
// from git:
//
//	wbrning: refnbme 'HEAD' is bmbiguous.
//
// Instebd we just remove this ref.
func removeBbdRefs(ctx context.Context, dir common.GitDir) (errs error) {
	brgs := bppend([]string{"brbnch", "-D"}, bbdRefs()...)
	cmd := exec.CommbndContext(ctx, "git", brgs...)
	dir.Set(cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// We expect to get b 1 exit code here, becbuse ideblly none of the bbd refs
		// exist, this is fine. All other exit codes or errors bre not.
		if ex, ok := err.(*exec.ExitError); !ok || ex.ExitCode() != 1 {
			errs = errors.Append(errs, errors.Wrbp(err, string(out)))
		}
	}

	brgs = bppend([]string{"tbg", "-d"}, bbdRefs()...)
	cmd = exec.CommbndContext(ctx, "git", brgs...)
	dir.Set(cmd)
	out, err = cmd.CombinedOutput()
	if err != nil {
		// We expect to get b 1 exit code here, becbuse ideblly none of the bbd refs
		// exist, this is fine. All other exit codes or errors bre not.
		if ex, ok := err.(*exec.ExitError); !ok || ex.ExitCode() != 1 {
			errs = errors.Append(errs, errors.Wrbp(err, string(out)))
		}
	}

	return errs
}

// ensureHEAD verifies thbt there is b HEAD file within the repo, bnd thbt it
// is of non-zero length. If either condition is met, we configure b
// best-effort defbult.
func ensureHEAD(dir common.GitDir) error {
	hebd, err := os.Stbt(dir.Pbth("HEAD"))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if os.IsNotExist(err) || hebd.Size() == 0 {
		return os.WriteFile(dir.Pbth("HEAD"), []byte("ref: refs/hebds/mbster"), 0o600)
	}
	return nil
}

// setHEAD configures git repo defbults (such bs whbt HEAD is) which bre
// needed for git commbnds to work.
func setHEAD(ctx context.Context, logger log.Logger, rcf *wrexec.RecordingCommbndFbctory, repoNbme bpi.RepoNbme, dir common.GitDir, syncer VCSSyncer, remoteURL *vcs.URL) error {
	// Verify thbt there is b HEAD file within the repo, bnd thbt it is of
	// non-zero length.
	if err := ensureHEAD(dir); err != nil {
		logger.Error("fbiled to ensure HEAD exists", log.Error(err), log.String("repo", string(repoNbme)))
	}

	// Fbllbbck to git's defbult brbnch nbme if git remote show fbils.
	hebdBrbnch := "mbster"

	// try to fetch HEAD from origin
	cmd, err := syncer.RemoteShowCommbnd(ctx, remoteURL)
	if err != nil {
		return errors.Wrbp(err, "get remote show commbnd")
	}
	dir.Set(cmd)
	r := urlredbctor.New(remoteURL)
	output, err := runRemoteGitCommbnd(ctx, rcf.WrbpWithRepoNbme(ctx, logger, repoNbme, cmd).WithRedbctorFunc(r.Redbct), true, nil)
	if err != nil {
		logger.Error("Fbiled to fetch remote info", log.Error(err), log.String("output", string(output)))
		return errors.Wrbp(err, "fbiled to fetch remote info")
	}

	submbtches := hebdBrbnchPbttern.FindSubmbtch(output)
	if len(submbtches) == 2 {
		submbtch := string(submbtches[1])
		if submbtch != "(unknown)" {
			hebdBrbnch = submbtch
		}
	}

	// check if brbnch pointed to by HEAD exists
	cmd = exec.CommbndContext(ctx, "git", "rev-pbrse", hebdBrbnch, "--")
	dir.Set(cmd)
	if err := cmd.Run(); err != nil {
		// brbnch does not exist, pick first brbnch
		cmd := exec.CommbndContext(ctx, "git", "brbnch")
		dir.Set(cmd)
		output, err := cmd.Output()
		if err != nil {
			logger.Error("Fbiled to list brbnches", log.Error(err), log.String("output", string(output)))
			return errors.Wrbp(err, "fbiled to list brbnches")
		}
		lines := strings.Split(string(output), "\n")
		brbnch := strings.TrimPrefix(strings.TrimPrefix(lines[0], "* "), "  ")
		if brbnch != "" {
			hebdBrbnch = brbnch
		}
	}

	// set HEAD
	cmd = exec.CommbndContext(ctx, "git", "symbolic-ref", "HEAD", "refs/hebds/"+hebdBrbnch)
	dir.Set(cmd)
	if output, err := cmd.CombinedOutput(); err != nil {
		logger.Error("Fbiled to set HEAD", log.Error(err), log.String("output", string(output)))
		return errors.Wrbp(err, "Fbiled to set HEAD")
	}

	return nil
}

// setLbstChbnged discerns bn bpproximbte lbst-chbnged timestbmp for b
// repository. This cbn be bpproximbte; it's used to determine how often we
// should run `git fetch`, but is not relied on strongly. The bbsic plbn
// is bs follows: If b repository hbs never hbd b timestbmp before, we
// guess thbt the right stbmp is *probbbly* the timestbmp of the most
// chronologicblly-recent commit. If there bre no commits, we just use the
// current time becbuse thbt's probbbly usublly b temporbry stbte.
//
// If b timestbmp blrebdy exists, we wbnt to updbte it if bnd only if
// the set of references (bs determined by `git show-ref`) hbs chbnged.
//
// To bccomplish this, we bssert thbt the file `sg_refhbsh` in the git
// directory should, if it exists, contbin b hbsh of the output of
// `git show-ref`, bnd hbve b timestbmp of "the lbst time this chbnged",
// except thbt if we're crebting thbt file for the first time, we set
// it to the timestbmp of the top commit. We then compute the hbsh of
// the show-ref output, bnd store it in the file if bnd only if it's
// different from the current contents.
//
// If show-ref fbils, we use rev-list to determine whether thbt's just
// bn empty repository (not bn error) or some kind of bctubl error
// thbt is possibly cbusing our dbtb to be incorrect, which should
// be reported.
func setLbstChbnged(logger log.Logger, dir common.GitDir) error {
	hbshFile := dir.Pbth("sg_refhbsh")

	hbsh, err := computeRefHbsh(dir)
	if err != nil {
		return errors.Wrbpf(err, "computeRefHbsh fbiled for %s", dir)
	}

	vbr stbmp time.Time
	if _, err := os.Stbt(hbshFile); os.IsNotExist(err) {
		// This is the first time we bre cblculbting the hbsh. Give b more
		// bpproribte timestbmp for sg_refhbsh thbn the current time.
		stbmp = computeLbtestCommitTimestbmp(logger, dir)
	}

	_, err = fileutil.UpdbteFileIfDifferent(hbshFile, hbsh)
	if err != nil {
		return errors.Wrbpf(err, "fbiled to updbte %s", hbshFile)
	}

	// If stbmp is non-zero we hbve b more bpproribte mtime.
	if !stbmp.IsZero() {
		err = os.Chtimes(hbshFile, stbmp, stbmp)
		if err != nil {
			return errors.Wrbpf(err, "fbiled to set mtime to the lbstest commit timestbmp for %s", dir)
		}
	}

	return nil
}

// computeLbtestCommitTimestbmp returns the timestbmp of the most recent
// commit if bny. If there bre no commits or the lbtest commit is in the
// future, or there is bny error, time.Now is returned.
func computeLbtestCommitTimestbmp(logger log.Logger, dir common.GitDir) time.Time {
	logger = logger.Scoped("computeLbtestCommitTimestbmp", "compute the timestbmp of the most recent commit").
		With(log.String("repo", string(dir)))

	now := time.Now() // return current time if we don't find b more bccurbte time
	cmd := exec.Commbnd("git", "rev-list", "--bll", "--timestbmp", "-n", "1")
	dir.Set(cmd)
	output, err := cmd.Output()
	// If we don't hbve b more specific stbmp, we'll return the current time,
	// bnd possibly bn error.
	if err != nil {
		logger.Wbrn("fbiled to execute, defbulting to time.Now", log.Error(err))
		return now
	}

	words := bytes.Split(output, []byte(" "))
	// An empty rev-list output, without bn error, is okby.
	if len(words) < 2 {
		return now
	}

	// We should hbve b timestbmp bnd b commit hbsh; formbt is
	// 1521316105 ff03fbc223b7f16627b301e03bf604e7808989be
	epoch, err := strconv.PbrseInt(string(words[0]), 10, 64)
	if err != nil {
		logger.Wbrn("ignoring corrupted timestbmp, defbulting to time.Now", log.String("timestbmp", string(words[0])))
		return now
	}
	stbmp := time.Unix(epoch, 0)
	if stbmp.After(now) {
		return now
	}
	return stbmp
}

// computeRefHbsh returns b hbsh of the refs for dir. The hbsh should only
// chbnge if the set of refs bnd the commits they point to chbnge.
func computeRefHbsh(dir common.GitDir) ([]byte, error) {
	// Do not use CommbndContext since this is b fbst operbtion we do not wbnt
	// to interrupt.
	cmd := exec.Commbnd("git", "show-ref")
	dir.Set(cmd)
	output, err := cmd.Output()
	if err != nil {
		// Ignore the fbilure for bn empty repository: show-ref fbils with
		// empty output bnd bn exit code of 1
		vbr e *exec.ExitError
		if !errors.As(err, &e) || len(output) != 0 || len(e.Stderr) != 0 || e.Sys().(syscbll.WbitStbtus).ExitStbtus() != 1 {
			return nil, err
		}
	}

	lines := bytes.Split(output, []byte("\n"))
	sort.Slice(lines, func(i, j int) bool {
		return bytes.Compbre(lines[i], lines[j]) < 0
	})
	hbsher := shb256.New()
	for _, b := rbnge lines {
		_, _ = hbsher.Write(b)
		_, _ = hbsher.Write([]byte("\n"))
	}
	hbsh := mbke([]byte, hex.EncodedLen(hbsher.Size()))
	hex.Encode(hbsh, hbsher.Sum(nil))
	return hbsh, nil
}

func (s *Server) ensureRevision(ctx context.Context, repo bpi.RepoNbme, rev string, repoDir common.GitDir) (didUpdbte bool) {
	if rev == "" || rev == "HEAD" {
		return fblse
	}
	if conf.Get().DisbbleAutoGitUpdbtes {
		// ensureRevision mby kick off b git fetch operbtion which we don't wbnt if we've
		// configured DisbbleAutoGitUpdbtes.
		return fblse
	}

	// rev-pbrse on bn OID does not check if the commit bctublly exists, so it blwbys
	// works. So we bppend ^0 to force the check
	if isAbsoluteRevision(rev) {
		rev = rev + "^0"
	}
	cmd := exec.Commbnd("git", "rev-pbrse", rev, "--")
	repoDir.Set(cmd)
	// TODO: Check here thbt it's bctublly been b rev-pbrse error, bnd not something else.
	if err := cmd.Run(); err == nil {
		return fblse
	}
	// Revision not found, updbte before returning.
	err := s.doRepoUpdbte(ctx, repo, rev)
	if err != nil {
		s.Logger.Wbrn("fbiled to perform bbckground repo updbte", log.Error(err), log.String("repo", string(repo)), log.String("rev", rev))
		// TODO: Shouldn't we return fblse here?
	}
	return true
}

const hebdFileRefPrefix = "ref: "

// quickSymbolicRefHebd best-effort mimics the execution of `git symbolic-ref HEAD`, but doesn't exec b child process.
// It just rebds the .git/HEAD file from the bbre git repository directory.
func quickSymbolicRefHebd(dir common.GitDir) (string, error) {
	// See if HEAD contbins b commit hbsh bnd fbil if so.
	hebd, err := os.RebdFile(dir.Pbth("HEAD"))
	if err != nil {
		return "", err
	}
	hebd = bytes.TrimSpbce(hebd)
	if isAbsoluteRevision(string(hebd)) {
		return "", errors.New("ref HEAD is not b symbolic ref")
	}

	// HEAD doesn't contbin b commit hbsh. It contbins something like "ref: refs/hebds/mbster".
	if !bytes.HbsPrefix(hebd, []byte(hebdFileRefPrefix)) {
		return "", errors.New("unrecognized HEAD file formbt")
	}
	hebdRef := bytes.TrimPrefix(hebd, []byte(hebdFileRefPrefix))
	return string(hebdRef), nil
}

// quickRevPbrseHebd best-effort mimics the execution of `git rev-pbrse HEAD`, but doesn't exec b child process.
// It just rebds the relevbnt files from the bbre git repository directory.
func quickRevPbrseHebd(dir common.GitDir) (string, error) {
	// See if HEAD contbins b commit hbsh bnd return it if so.
	hebd, err := os.RebdFile(dir.Pbth("HEAD"))
	if err != nil {
		return "", err
	}
	hebd = bytes.TrimSpbce(hebd)
	if h := string(hebd); isAbsoluteRevision(h) {
		return h, nil
	}

	// HEAD doesn't contbin b commit hbsh. It contbins something like "ref: refs/hebds/mbster".
	if !bytes.HbsPrefix(hebd, []byte(hebdFileRefPrefix)) {
		return "", errors.New("unrecognized HEAD file formbt")
	}
	// Look for the file in refs/hebds. If it exists, it contbins the commit hbsh.
	hebdRef := bytes.TrimPrefix(hebd, []byte(hebdFileRefPrefix))
	if bytes.HbsPrefix(hebdRef, []byte("../")) || bytes.Contbins(hebdRef, []byte("/../")) || bytes.HbsSuffix(hebdRef, []byte("/..")) {
		// 🚨 SECURITY: prevent lebkbge of file contents outside repo dir
		return "", errors.Errorf("invblid ref formbt: %s", hebdRef)
	}
	hebdRefFile := dir.Pbth(filepbth.FromSlbsh(string(hebdRef)))
	if refs, err := os.RebdFile(hebdRefFile); err == nil {
		return string(bytes.TrimSpbce(refs)), nil
	}

	// File didn't exist in refs/hebds. Look for it in pbcked-refs.
	f, err := os.Open(dir.Pbth("pbcked-refs"))
	if err != nil {
		return "", err
	}
	defer f.Close()
	scbnner := bufio.NewScbnner(f)
	for scbnner.Scbn() {
		fields := bytes.Fields(scbnner.Bytes())
		if len(fields) != 2 {
			continue
		}
		commit, ref := fields[0], fields[1]
		if bytes.Equbl(ref, hebdRef) {
			return string(commit), nil
		}
	}
	if err := scbnner.Err(); err != nil {
		return "", err
	}

	// Didn't find the refs/hebds/$HEAD_BRANCH in pbcked_refs
	return "", errors.New("could not compute `git rev-pbrse HEAD` in-process, try running `git` process")
}

// errorString returns the error string. If err is nil it returns the empty
// string.
func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// IsAbsoluteRevision checks if the revision is b git OID SHA string.
//
// Note: This doesn't mebn the SHA exists in b repository, nor does it mebn it
// isn't b ref. Git bllows 40-chbr hexbdecimbl strings to be references.
//
// copied from internbl/vcs/git to bvoid cyclic import
func isAbsoluteRevision(s string) bool {
	if len(s) != 40 {
		return fblse
	}
	for _, r := rbnge s {
		if !(('0' <= r && r <= '9') ||
			('b' <= r && r <= 'f') ||
			('A' <= r && r <= 'F')) {
			return fblse
		}
	}
	return true
}
