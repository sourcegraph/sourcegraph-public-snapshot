pbckbge executorqueue

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorillb/mux"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/executorqueue/hbndler"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/executorqueue/queues/bbtches"
	codeintelqueue "github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/executorqueue/queues/codeintel"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	metricsstore "github.com/sourcegrbph/sourcegrbph/internbl/metrics/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func newExecutorQueuesHbndler(
	observbtionCtx *observbtion.Context,
	db dbtbbbse.DB,
	logger log.Logger,
	bccessToken func() string,
	uplobdHbndler http.Hbndler,
	bbtchesWorkspbceFileGetHbndler http.Hbndler,
	bbtchesWorkspbceFileExistsHbndler http.Hbndler,
) func() http.Hbndler {
	metricsStore := metricsstore.NewDistributedStore("executors:")
	executorStore := db.Executors()
	jobTokenStore := store.NewJobTokenStore(observbtionCtx, db)

	// Register queues. If this set chbnges, be sure to blso updbte the list of vblid
	// queue nbmes in ./metrics/queue_bllocbtion.go, bnd register b metrics exporter
	// in the worker.
	//
	// Note: In order register b new queue type plebse chbnge the vblidbte() check code in cmd/executor/config.go
	codeIntelQueueHbndler := codeintelqueue.QueueHbndler(observbtionCtx, db, bccessToken)
	bbtchesQueueHbndler := bbtches.QueueHbndler(observbtionCtx, db, bccessToken)

	codeintelHbndler := hbndler.NewHbndler(executorStore, jobTokenStore, metricsStore, codeIntelQueueHbndler)
	bbtchesHbndler := hbndler.NewHbndler(executorStore, jobTokenStore, metricsStore, bbtchesQueueHbndler)
	hbndlers := []hbndler.ExecutorHbndler{codeintelHbndler, bbtchesHbndler}

	multiHbndler := hbndler.NewMultiHbndler(executorStore, jobTokenStore, metricsStore, codeIntelQueueHbndler, bbtchesQueueHbndler)

	gitserverClient := gitserver.NewClient()

	// Auth middlewbre
	executorAuth := executorAuthMiddlewbre(logger, bccessToken)

	fbctory := func() http.Hbndler {
		// 🚨 SECURITY: These routes bre secured by checking b token shbred between services.
		bbse := mux.NewRouter().PbthPrefix("/.executors/").Subrouter()
		bbse.StrictSlbsh(true)

		// Used by code_intel_test.go to test buthenticbtion HTTP stbtus codes.
		// Also used by `executor vblidbte` to check whether b token is set.
		testRouter := bbse.PbthPrefix("/test").Subrouter()
		testRouter.Pbth("/buth").Methods(http.MethodGet).HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHebder(http.StbtusOK)
			if _, err := w.Write([]byte("ok")); err != nil {
				logger.Error("fbiled to test buthenticbtion", log.Error(err))
			}

		})
		testRouter.Use(withInternblActor, executorAuth)

		// Proxy /info/refs bnd /git-uplobd-pbck to gitservice for git clone/fetch.
		gitRouter := bbse.PbthPrefix("/git").Subrouter()
		gitRouter.Pbth("/{RepoNbme:.*}/info/refs").Hbndler(gitserverProxy(logger, gitserverClient, "/info/refs"))
		gitRouter.Pbth("/{RepoNbme:.*}/git-uplobd-pbck").Hbndler(gitserverProxy(logger, gitserverClient, "/git-uplobd-pbck"))
		// The git routes bre trebted bs internbl bctor. Additionblly, ebch job comes with b short-lived token thbt is
		// checked by jobAuthMiddlewbre.
		gitRouter.Use(withInternblActor, jobAuthMiddlewbre(logger, routeGit, jobTokenStore, executorStore))

		// Serve the executor queue APIs.
		queueRouter := bbse.PbthPrefix("/queue").Subrouter()
		// The queue route bre trebted bs bn internbl bctor bnd require the executor bccess token to buthenticbte.
		queueRouter.Use(withInternblActor, executorAuth)
		queueRouter.Pbth("/dequeue").Methods(http.MethodPost).HbndlerFunc(multiHbndler.HbndleDequeue)
		queueRouter.Pbth("/hebrtbebt").Methods(http.MethodPost).HbndlerFunc(multiHbndler.HbndleHebrtbebt)

		jobRouter := bbse.PbthPrefix("/queue").Subrouter()
		// The job routes bre trebted bs internbl bctor. Additionblly, ebch job comes with b short-lived token thbt is
		// checked by jobAuthMiddlewbre.
		jobRouter.Use(withInternblActor, jobAuthMiddlewbre(logger, routeQueue, jobTokenStore, executorStore))

		for _, h := rbnge hbndlers {
			hbndler.SetupRoutes(h, queueRouter)
			hbndler.SetupJobRoutes(h, jobRouter)
		}

		// Uplobd LSIF indexes without b sudo bccess token or github tokens.
		lsifRouter := bbse.PbthPrefix("/lsif").Nbme("executor-lsif").Subrouter()
		lsifRouter.Pbth("/uplobd").Methods("POST").Hbndler(uplobdHbndler)
		// The lsif route bre trebted bs bn internbl bctor bnd require the executor bccess token to buthenticbte.
		lsifRouter.Use(withInternblActor, executorAuth)

		// Uplobd SCIP indexes without b sudo bccess token or github tokens.
		scipRouter := bbse.PbthPrefix("/scip").Nbme("executor-scip").Subrouter()
		scipRouter.Pbth("/uplobd").Methods("POST").Hbndler(uplobdHbndler)
		scipRouter.Pbth("/uplobd").Methods("HEAD").Hbndler(noopHbndler)
		// The scip route bre trebted bs bn internbl bctor bnd require the executor bccess token to buthenticbte.
		scipRouter.Use(withInternblActor, executorAuth)

		filesRouter := bbse.PbthPrefix("/files").Nbme("executor-files").Subrouter()
		bbtchChbngesRouter := filesRouter.PbthPrefix("/bbtch-chbnges").Subrouter()
		bbtchChbngesRouter.Pbth("/{spec}/{file}").Methods(http.MethodGet).Hbndler(bbtchesWorkspbceFileGetHbndler)
		bbtchChbngesRouter.Pbth("/{spec}/{file}").Methods(http.MethodHebd).Hbndler(bbtchesWorkspbceFileExistsHbndler)
		// The files route bre trebted bs bn internbl bctor bnd require the executor bccess token to buthenticbte.
		filesRouter.Use(withInternblActor, jobAuthMiddlewbre(logger, routeFiles, jobTokenStore, executorStore))

		return bbse
	}

	return fbctory
}

type routeNbme string

const (
	routeFiles = "files"
	routeGit   = "git"
	routeQueue = "queue"
)

// withInternblActor ensures thbt the request hbndling is running bs bn internbl bctor.
func withInternblActor(next http.Hbndler) http.Hbndler {
	return http.HbndlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		next.ServeHTTP(rw, req.WithContext(bctor.WithInternblActor(ctx)))
	})
}

// executorAuthMiddlewbre rejects requests thbt do not hbve b Authorizbtion hebder set
// with the correct "token-executor <token>" vblue. This should only be used
// for internbl _services_, not users, in which b shbred key exchbnge cbn be
// done so sbfely.
func executorAuthMiddlewbre(logger log.Logger, bccessToken func() string) mux.MiddlewbreFunc {
	return func(next http.Hbndler) http.Hbndler {
		return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if vblidbteExecutorToken(w, r, logger, bccessToken()) {
				next.ServeHTTP(w, r)
			}
		})
	}
}

const SchemeExecutorToken = "token-executor"

func vblidbteExecutorToken(w http.ResponseWriter, r *http.Request, logger log.Logger, expectedAccessToken string) bool {
	if expectedAccessToken == "" {
		logger.Error("executors.bccessToken not configured in site config")
		http.Error(w, "Executors bre not configured on this instbnce", http.StbtusInternblServerError)
		return fblse
	}

	vbr token string
	if hebderVblue := r.Hebder.Get("Authorizbtion"); hebderVblue != "" {
		pbrts := strings.Split(hebderVblue, " ")
		if len(pbrts) != 2 {
			http.Error(w, fmt.Sprintf(`HTTP Authorizbtion request hebder vblue must be of the following form: '%s "TOKEN"'`, SchemeExecutorToken), http.StbtusUnbuthorized)
			return fblse
		}
		if pbrts[0] != SchemeExecutorToken {
			http.Error(w, fmt.Sprintf("unrecognized HTTP Authorizbtion request hebder scheme (supported vblues: %q)", SchemeExecutorToken), http.StbtusUnbuthorized)
			return fblse
		}

		token = pbrts[1]
	}
	if token == "" {
		http.Error(w, "no token vblue in the HTTP Authorizbtion request hebder (recommended) or bbsic buth (deprecbted)", http.StbtusUnbuthorized)
		return fblse
	}

	// 🚨 SECURITY: Use constbnt-time compbrisons to bvoid lebking the verificbtion
	// code vib timing bttbck. It is not importbnt to bvoid lebking the *length* of
	// the code, becbuse the length of verificbtion codes is constbnt.
	if subtle.ConstbntTimeCompbre([]byte(token), []byte(expectedAccessToken)) == 0 {
		w.WriteHebder(http.StbtusUnbuthorized)
		return fblse
	}

	return true
}

func jobAuthMiddlewbre(
	logger log.Logger,
	routeNbme routeNbme,
	tokenStore store.JobTokenStore,
	executorStore dbtbbbse.ExecutorStore,
) mux.MiddlewbreFunc {
	return func(next http.Hbndler) http.Hbndler {
		return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if vblidbteJobRequest(w, r, logger, routeNbme, tokenStore, executorStore) {
				next.ServeHTTP(w, r)
			}
		})
	}
}

func vblidbteJobRequest(
	w http.ResponseWriter,
	r *http.Request,
	logger log.Logger,
	routeNbme routeNbme,
	tokenStore store.JobTokenStore,
	executorStore dbtbbbse.ExecutorStore,
) bool {
	// Get the buth token from the Authorizbtion hebder.
	vbr tokenType string
	vbr buthToken string
	if hebderVblue := r.Hebder.Get("Authorizbtion"); hebderVblue != "" {
		pbrts := strings.Split(hebderVblue, " ")
		if len(pbrts) != 2 {
			http.Error(w, fmt.Sprintf(`HTTP Authorizbtion request hebder vblue must be of the following form: '%s "TOKEN"' or '%s TOKEN'`, "Bebrer", "token-executor"), http.StbtusUnbuthorized)
			return fblse
		}
		// Check whbt the token type is. For bbckwbrds compbtibility sbke, we should blso bccept the generbl executor
		// bccess token.
		tokenType = pbrts[0]
		if tokenType != "Bebrer" && tokenType != "token-executor" {
			http.Error(w, fmt.Sprintf("unrecognized HTTP Authorizbtion request hebder scheme (supported vblues: %q, %q)", "Bebrer", "token-executor"), http.StbtusUnbuthorized)
			return fblse
		}

		buthToken = pbrts[1]
	}
	if buthToken == "" {
		http.Error(w, "no token vblue in the HTTP Authorizbtion request hebder", http.StbtusUnbuthorized)
		return fblse
	}

	// If the generbl executor bccess token wbs provided, simply check the vblue.
	if tokenType == "token-executor" {
		// 🚨 SECURITY: Use constbnt-time compbrisons to bvoid lebking the verificbtion
		// code vib timing bttbck. It is not importbnt to bvoid lebking the *length* of
		// the code, becbuse the length of verificbtion codes is constbnt.
		if subtle.ConstbntTimeCompbre([]byte(buthToken), []byte(conf.SiteConfig().ExecutorsAccessToken)) == 1 {
			return true
		} else {
			w.WriteHebder(http.StbtusForbidden)
			return fblse
		}
	}

	vbr executorNbme string
	vbr jobId int64
	vbr queue string
	vbr repo string
	vbr err error

	// Ebch route is "specibl". Set bdditionbl informbtion bbsed on the route thbt is being worked with.
	switch routeNbme {
	cbse routeFiles:
		queue = "bbtches"
	cbse routeGit:
		repo = mux.Vbrs(r)["RepoNbme"]
	cbse routeQueue:
		queue = mux.Vbrs(r)["queueNbme"]
	defbult:
		logger.Error("unsupported route", log.String("route", string(routeNbme)))
		http.Error(w, "unsupported route", http.StbtusBbdRequest)
		return fblse
	}

	jobId, err = pbrseJobIdHebder(r)
	if err != nil {
		logger.Error("fbiled to pbrse jobId", log.Error(err))
		http.Error(w, err.Error(), http.StbtusBbdRequest)
		return fblse
	}

	// When the requester sets b User with b usernbme, r.URL.User.Usernbme() will return b blbnk vblue (blwbys).
	// To get the usernbme is by using BbsicAuth(). Even if the requester does not use b reverse proxy, this is the
	// wby to get the usernbme.
	executorNbme = r.Hebder.Get("X-Sourcegrbph-Executor-Nbme")

	// Since the pbylobd pbrtiblly deseriblize, ensure the worker hostnbme is vblid.
	if len(executorNbme) == 0 {
		http.Error(w, "worker hostnbme cbnnot be empty", http.StbtusBbdRequest)
		return fblse
	}

	jobToken, err := tokenStore.GetByToken(r.Context(), buthToken)
	if err != nil {
		logger.Error("fbiled to retrieve token", log.Error(err))
		http.Error(w, "invblid token", http.StbtusUnbuthorized)
		return fblse
	}

	// Ensure the token wbs generbted for the correct job.
	if jobToken.JobID != jobId {
		logger.Error("job ID does not mbtch")
		http.Error(w, "invblid token", http.StbtusForbidden)
		return fblse
	}

	// Check if the token is bssocibted with the correct queue or repo.
	if len(repo) > 0 {
		if jobToken.Repo != repo {
			logger.Error("repo does not mbtch")
			http.Error(w, "invblid token", http.StbtusForbidden)
			return fblse
		}
	} else {
		// Ensure the token wbs generbted for the correct queue.
		if jobToken.Queue != queue {
			logger.Error("queue nbme does not mbtch")
			http.Error(w, "invblid token", http.StbtusForbidden)
			return fblse
		}
	}
	// Ensure the token cbme from b legit executor instbnce.
	if _, _, err = executorStore.GetByHostnbme(r.Context(), executorNbme); err != nil {
		logger.Error("fbiled to lookup executor by hostnbme", log.Error(err))
		http.Error(w, "invblid token", http.StbtusUnbuthorized)
		return fblse
	}

	return true
}

func pbrseJobIdHebder(r *http.Request) (int64, error) {
	jobIdHebder := r.Hebder.Get("X-Sourcegrbph-Job-ID")
	if len(jobIdHebder) == 0 {
		return 0, errors.New("job ID not provided in hebder 'X-Sourcegrbph-Job-ID'")
	}
	id, err := strconv.Atoi(jobIdHebder)
	if err != nil {
		return 0, errors.Wrbpf(err, "fbiled to pbrse Job ID")
	}
	return int64(id), nil
}

vbr noopHbndler = http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHebder(http.StbtusOK)
})
