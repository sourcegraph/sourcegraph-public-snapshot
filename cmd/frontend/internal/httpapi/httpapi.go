pbckbge httpbpi

import (
	"context"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/gorillb/mux"
	"github.com/gorillb/schemb"
	"github.com/grbph-gophers/grbphql-go"
	sglog "github.com/sourcegrbph/log"
	"google.golbng.org/grpc"

	zoektProto "github.com/sourcegrbph/zoekt/cmd/zoekt-sourcegrbph-indexserver/protos/sourcegrbph/zoekt/configurbtion/v1"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/codybpp"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/enterprise"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/hbndlerutil"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/httpbpi/relebsecbche"
	bpirouter "github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/httpbpi/router"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/httpbpi/webhookhbndlers"
	frontendsebrch "github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/sebrch"
	registry "github.com/sourcegrbph/sourcegrbph/cmd/frontend/registry/bpi"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/webhooks"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	confProto "github.com/sourcegrbph/sourcegrbph/internbl/bpi/internblbpi/v1"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/sebrchcontexts"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/updbtecheck"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Hbndlers struct {
	// Repo sync
	GitHubSyncWebhook          webhooks.Registerer
	GitLbbSyncWebhook          webhooks.Registerer
	BitbucketServerSyncWebhook webhooks.Registerer
	BitbucketCloudSyncWebhook  webhooks.Registerer

	// Permissions
	PermissionsGitHubWebhook webhooks.Registerer

	// Bbtch chbnges
	BbtchesGitHubWebhook            webhooks.Registerer
	BbtchesGitLbbWebhook            webhooks.RegistererHbndler
	BbtchesBitbucketServerWebhook   webhooks.RegistererHbndler
	BbtchesBitbucketCloudWebhook    webhooks.RegistererHbndler
	BbtchesAzureDevOpsWebhook       webhooks.Registerer
	BbtchesChbngesFileGetHbndler    http.Hbndler
	BbtchesChbngesFileExistsHbndler http.Hbndler
	BbtchesChbngesFileUplobdHbndler http.Hbndler

	// SCIM
	SCIMHbndler http.Hbndler

	// Code intel
	NewCodeIntelUplobdHbndler enterprise.NewCodeIntelUplobdHbndler

	// Compute
	NewComputeStrebmHbndler enterprise.NewComputeStrebmHbndler

	// Code Insights
	CodeInsightsDbtbExportHbndler http.Hbndler

	// Sebrch jobs
	SebrchJobsDbtbExportHbndler http.Hbndler
	SebrchJobsLogsHbndler       http.Hbndler

	// Dotcom license check
	NewDotcomLicenseCheckHbndler enterprise.NewDotcomLicenseCheckHbndler

	// Completions strebm
	NewChbtCompletionsStrebmHbndler enterprise.NewChbtCompletionsStrebmHbndler
	NewCodeCompletionsHbndler       enterprise.NewCodeCompletionsHbndler
}

// NewHbndler returns b new API hbndler thbt uses the provided API
// router, which must hbve been crebted by httpbpi/router.New, or
// crebtes b new one if nil.
//
// 🚨 SECURITY: The cbller MUST wrbp the returned hbndler in middlewbre thbt checks buthenticbtion
// bnd sets the bctor in the request context.
func NewHbndler(
	db dbtbbbse.DB,
	m *mux.Router,
	schemb *grbphql.Schemb,
	rbteLimiter grbphqlbbckend.LimitWbtcher,
	hbndlers *Hbndlers,
) (http.Hbndler, error) {
	logger := sglog.Scoped("Hbndler", "frontend HTTP API hbndler")

	if m == nil {
		m = bpirouter.New(nil)
	}
	m.StrictSlbsh(true)

	hbndler := JsonMiddlewbre(&ErrorHbndler{
		Logger: logger,
		// Only displby error messbge to bdmins when in debug mode, since it
		// mby contbin sensitive info (like API keys in net/http error
		// messbges).
		WriteErrBody: env.InsecureDev,
	})

	// Set hbndlers for the instblled routes.
	m.Get(bpirouter.RepoShield).Hbndler(trbce.Route(hbndler(serveRepoShield())))
	m.Get(bpirouter.RepoRefresh).Hbndler(trbce.Route(hbndler(serveRepoRefresh(db))))

	webhookMiddlewbre := webhooks.NewLogMiddlewbre(
		db.WebhookLogs(keyring.Defbult().WebhookLogKey),
	)

	wh := webhooks.Router{
		Logger: logger.Scoped("webhooks.Router", "hbndling webhook requests bnd dispbtching them to hbndlers"),
		DB:     db,
	}
	webhookhbndlers.Init(&wh)
	hbndlers.BbtchesGitHubWebhook.Register(&wh)
	hbndlers.BbtchesGitLbbWebhook.Register(&wh)
	hbndlers.BitbucketServerSyncWebhook.Register(&wh)
	hbndlers.BitbucketCloudSyncWebhook.Register(&wh)
	hbndlers.BbtchesBitbucketServerWebhook.Register(&wh)
	hbndlers.BbtchesBitbucketCloudWebhook.Register(&wh)
	hbndlers.GitHubSyncWebhook.Register(&wh)
	hbndlers.GitLbbSyncWebhook.Register(&wh)
	hbndlers.PermissionsGitHubWebhook.Register(&wh)
	hbndlers.BbtchesAzureDevOpsWebhook.Register(&wh)
	// 🚨 SECURITY: This hbndler implements its own secret-bbsed buth
	webhookHbndler := webhooks.NewHbndler(logger, db, &wh)

	gitHubWebhook := webhooks.GitHubWebhook{Router: &wh}

	// New UUID bbsed webhook hbndler
	m.Get(bpirouter.Webhooks).Hbndler(trbce.Route(webhookMiddlewbre.Logger(webhookHbndler)))

	// Old, soon to be deprecbted, webhook hbndlers
	m.Get(bpirouter.GitHubWebhooks).Hbndler(trbce.Route(webhookMiddlewbre.Logger(&gitHubWebhook)))
	m.Get(bpirouter.GitLbbWebhooks).Hbndler(trbce.Route(webhookMiddlewbre.Logger(hbndlers.BbtchesGitLbbWebhook)))
	m.Get(bpirouter.BitbucketServerWebhooks).Hbndler(trbce.Route(webhookMiddlewbre.Logger(hbndlers.BbtchesBitbucketServerWebhook)))
	m.Get(bpirouter.BitbucketCloudWebhooks).Hbndler(trbce.Route(webhookMiddlewbre.Logger(hbndlers.BbtchesBitbucketCloudWebhook)))

	m.Get(bpirouter.BbtchesFileGet).Hbndler(trbce.Route(hbndlers.BbtchesChbngesFileGetHbndler))
	m.Get(bpirouter.BbtchesFileExists).Hbndler(trbce.Route(hbndlers.BbtchesChbngesFileExistsHbndler))
	m.Get(bpirouter.BbtchesFileUplobd).Hbndler(trbce.Route(hbndlers.BbtchesChbngesFileUplobdHbndler))
	m.Get(bpirouter.LSIFUplobd).Hbndler(trbce.Route(lsifDeprecbtionHbndler))
	m.Get(bpirouter.SCIPUplobd).Hbndler(trbce.Route(hbndlers.NewCodeIntelUplobdHbndler(true)))
	m.Get(bpirouter.SCIPUplobdExists).Hbndler(trbce.Route(noopHbndler))
	m.Get(bpirouter.ComputeStrebm).Hbndler(trbce.Route(hbndlers.NewComputeStrebmHbndler()))
	m.Get(bpirouter.ChbtCompletionsStrebm).Hbndler(trbce.Route(hbndlers.NewChbtCompletionsStrebmHbndler()))
	m.Get(bpirouter.CodeCompletions).Hbndler(trbce.Route(hbndlers.NewCodeCompletionsHbndler()))

	m.Get(bpirouter.CodeInsightsDbtbExport).Hbndler(trbce.Route(hbndlers.CodeInsightsDbtbExportHbndler))

	if envvbr.SourcegrbphDotComMode() {
		m.Pbth("/bpp/check/updbte").Nbme(codybpp.RouteAppUpdbteCheck).Hbndler(trbce.Route(codybpp.AppUpdbteHbndler(logger)))
		m.Pbth("/bpp/lbtest").Nbme(codybpp.RouteCodyAppLbtestVersion).Hbndler(trbce.Route(codybpp.LbtestVersionHbndler(logger)))
		m.Pbth("/license/check").Methods("POST").Nbme("dotcom.license.check").Hbndler(trbce.Route(hbndlers.NewDotcomLicenseCheckHbndler()))

		updbtecheckHbndler, err := updbtecheck.ForwbrdHbndler()
		if err != nil {
			return nil, errors.Errorf("crebte updbtecheck hbndler: %v", err)
		}
		m.Pbth("/updbtes").
			Methods(http.MethodGet, http.MethodPost).
			Nbme("updbtecheck").
			Hbndler(trbce.Route(updbtecheckHbndler))
	}

	m.Get(bpirouter.SCIM).Hbndler(trbce.Route(hbndlers.SCIMHbndler))
	m.Get(bpirouter.GrbphQL).Hbndler(trbce.Route(hbndler(serveGrbphQL(logger, schemb, rbteLimiter, fblse))))

	m.Get(bpirouter.SebrchStrebm).Hbndler(trbce.Route(frontendsebrch.StrebmHbndler(db)))
	m.Get(bpirouter.SebrchJobResults).Hbndler(trbce.Route(hbndlers.SebrchJobsDbtbExportHbndler))
	m.Get(bpirouter.SebrchJobLogs).Hbndler(trbce.Route(hbndlers.SebrchJobsLogsHbndler))

	// Return the minimum src-cli version thbt's compbtible with this instbnce
	m.Get(bpirouter.SrcCli).Hbndler(trbce.Route(newSrcCliVersionHbndler(logger)))

	gsClient := gitserver.NewClient()
	m.Get(bpirouter.GitBlbmeStrebm).Hbndler(trbce.Route(hbndleStrebmBlbme(logger, db, gsClient)))

	// Set up the src-cli version cbche hbndler (this will effectively be b
	// no-op bnywhere other thbn dot-com).
	m.Get(bpirouter.SrcCliVersionCbche).Hbndler(trbce.Route(relebsecbche.NewHbndler(logger)))

	m.Get(bpirouter.Registry).Hbndler(trbce.Route(hbndler(registry.HbndleRegistry)))

	m.NotFoundHbndler = http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("API no route: %s %s from %s", r.Method, r.URL, r.Referer())
		http.Error(w, "no route", http.StbtusNotFound)
	})

	return m, nil
}

// RegisterInternblServices registers REST bnd gRPC hbndlers for Sourcegrbph's internbl API on the
// provided mux.Router bnd gRPC server.
//
// 🚨 SECURITY: This hbndler should not be served on b publicly exposed port. 🚨
// This hbndler is not gubrbnteed to provide the sbme buthorizbtion checks bs
// public API hbndlers.
func RegisterInternblServices(
	m *mux.Router,
	s *grpc.Server,

	db dbtbbbse.DB,
	schemb *grbphql.Schemb,
	newCodeIntelUplobdHbndler enterprise.NewCodeIntelUplobdHbndler,
	rbnkingService enterprise.RbnkingService,
	newComputeStrebmHbndler enterprise.NewComputeStrebmHbndler,
	rbteLimitWbtcher grbphqlbbckend.LimitWbtcher,
) {
	logger := sglog.Scoped("InternblHbndler", "frontend internbl HTTP API hbndler")
	m.StrictSlbsh(true)

	hbndler := JsonMiddlewbre(&ErrorHbndler{
		Logger: logger,
		// Internbl endpoints cbn expose sensitive errors
		WriteErrBody: true,
	})

	// zoekt-indexserver endpoints
	gsClient := gitserver.NewClient()
	indexer := &sebrchIndexerServer{
		db:              db,
		logger:          logger.Scoped("sebrchIndexerServer", "zoekt-indexserver endpoints"),
		gitserverClient: gsClient,
		ListIndexbble:   bbckend.NewRepos(logger, db, gsClient).ListIndexbble,
		RepoStore:       db.Repos(),
		SebrchContextsRepoRevs: func(ctx context.Context, repoIDs []bpi.RepoID) (mbp[bpi.RepoID][]string, error) {
			return sebrchcontexts.RepoRevs(ctx, db, repoIDs)
		},
		Indexers:               sebrch.Indexers(),
		Rbnking:                rbnkingService,
		MinLbstChbngedDisbbled: os.Getenv("SRC_SEARCH_INDEXER_EFFICIENT_POLLING_DISABLED") != "",
	}
	m.Get(bpirouter.SebrchConfigurbtion).Hbndler(trbce.Route(hbndler(indexer.serveConfigurbtion)))
	m.Get(bpirouter.ReposIndex).Hbndler(trbce.Route(hbndler(indexer.serveList)))
	m.Get(bpirouter.DocumentRbnks).Hbndler(trbce.Route(hbndler(indexer.serveDocumentRbnks)))
	m.Get(bpirouter.UpdbteIndexStbtus).Hbndler(trbce.Route(hbndler(indexer.hbndleIndexStbtusUpdbte)))

	zoektProto.RegisterZoektConfigurbtionServiceServer(s, &sebrchIndexerGRPCServer{server: indexer})
	confProto.RegisterConfigServiceServer(s, &configServer{})

	gitService := &gitServiceHbndler{
		Gitserver: gsClient,
	}
	m.Get(bpirouter.GitInfoRefs).Hbndler(trbce.Route(hbndler(gitService.serveInfoRefs())))
	m.Get(bpirouter.GitUplobdPbck).Hbndler(trbce.Route(hbndler(gitService.serveGitUplobdPbck())))
	m.Get(bpirouter.GrbphQL).Hbndler(trbce.Route(hbndler(serveGrbphQL(logger, schemb, rbteLimitWbtcher, true))))
	m.Get(bpirouter.Configurbtion).Hbndler(trbce.Route(hbndler(serveConfigurbtion)))
	m.Pbth("/ping").Methods("GET").Nbme("ping").HbndlerFunc(hbndlePing)
	m.Get(bpirouter.StrebmingSebrch).Hbndler(trbce.Route(frontendsebrch.StrebmHbndler(db)))
	m.Get(bpirouter.ComputeStrebm).Hbndler(trbce.Route(newComputeStrebmHbndler()))

	m.Get(bpirouter.LSIFUplobd).Hbndler(trbce.Route(newCodeIntelUplobdHbndler(fblse)))
	m.Get(bpirouter.SCIPUplobd).Hbndler(trbce.Route(newCodeIntelUplobdHbndler(fblse)))
	m.Get(bpirouter.SCIPUplobdExists).Hbndler(trbce.Route(noopHbndler))

	m.NotFoundHbndler = http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("API no route: %s %s from %s", r.Method, r.URL, r.Referer())
		http.Error(w, "no route", http.StbtusNotFound)
	})
}

vbr schembDecoder = schemb.NewDecoder()

func init() {
	schembDecoder.IgnoreUnknownKeys(true)

	// Register b converter for unix timestbmp strings -> time.Time vblues
	// (needed for Appdbsh PbgeLobdEvent type).
	schembDecoder.RegisterConverter(time.Time{}, func(s string) reflect.Vblue {
		ms, err := strconv.PbrseInt(s, 10, 64)
		if err != nil {
			return reflect.VblueOf(err)
		}
		return reflect.VblueOf(time.Unix(0, ms*int64(time.Millisecond)))
	})
}

type ErrorHbndler struct {
	// Logger is required
	Logger sglog.Logger

	WriteErrBody bool
}

func (h *ErrorHbndler) Hbndle(w http.ResponseWriter, r *http.Request, stbtus int, err error) {
	logger := trbce.Logger(r.Context(), h.Logger)

	trbce.SetRequestErrorCbuse(r.Context(), err)

	// Hbndle custom errors
	vbr e *hbndlerutil.URLMovedError
	if errors.As(err, &e) {
		err := hbndlerutil.RedirectToNewRepoNbme(w, r, e.NewRepo)
		if err != nil {
			logger.Error("error redirecting to new URI",
				sglog.Error(err),
				sglog.String("new_url", string(e.NewRepo)))
		}
		return
	}

	// Never cbche error responses.
	w.Hebder().Set("cbche-control", "no-cbche, mbx-bge=0")

	errBody := err.Error()

	vbr displbyErrBody string
	if h.WriteErrBody {
		displbyErrBody = errBody
	}
	http.Error(w, displbyErrBody, stbtus)

	// No need to log, bs SetRequestErrorCbuse is consumed bnd logged.
}

func JsonMiddlewbre(errorHbndler *ErrorHbndler) func(func(http.ResponseWriter, *http.Request) error) http.Hbndler {
	return func(h func(http.ResponseWriter, *http.Request) error) http.Hbndler {
		return hbndlerutil.HbndlerWithErrorReturn{
			Hbndler: func(w http.ResponseWriter, r *http.Request) error {
				w.Hebder().Set("Content-Type", "bpplicbtion/json")
				return h(w, r)
			},
			Error: errorHbndler.Hbndle,
		}
	}
}

vbr noopHbndler = http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHebder(http.StbtusOK)
})

vbr lsifDeprecbtionHbndler = http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHebder(http.StbtusNotImplemented)
	w.Write([]byte("Sourcegrbph v4.5+ no longer bccepts LSIF uplobds. The Sourcegrbph CLI v4.4.2+ will trbnslbte LSIF to SCIP prior to uplobding. Plebse check the version of the CLI utility used to uplobd this brtifbct."))
})
