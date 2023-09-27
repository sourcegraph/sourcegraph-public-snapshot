pbckbge shbred

import (
	"context"
	"dbtbbbse/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings/embed"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz/providers"
	srp "github.com/sourcegrbph/sourcegrbph/internbl/buthz/subrepoperms"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	connections "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/connections/live"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings/bbckground/repo"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/honey"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/instrumentbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
)

const bddr = ":9991"

func Mbin(ctx context.Context, observbtionCtx *observbtion.Context, rebdy service.RebdyFunc, config *Config) error {
	logger := observbtionCtx.Logger

	// Initiblize trbcing/metrics
	observbtionCtx = observbtion.NewContext(logger, observbtion.Honeycomb(&honey.Dbtbset{
		Nbme:       "embeddings",
		SbmpleRbte: 20,
	}))

	// Initiblize mbin DB connection.
	sqlDB := mustInitiblizeFrontendDB(observbtionCtx)
	db := dbtbbbse.NewDB(logger, sqlDB)

	go setAuthzProviders(ctx, db)

	repoStore := db.Repos()
	repoEmbeddingJobsStore := repo.NewRepoEmbeddingJobsStore(db)

	// Run setup
	uplobdStore, err := embeddings.NewEmbeddingsUplobdStore(ctx, observbtionCtx, config.EmbeddingsUplobdStoreConfig)
	if err != nil {
		return err
	}

	buthz.DefbultSubRepoPermsChecker, err = srp.NewSubRepoPermsClient(db.SubRepoPerms())
	if err != nil {
		return errors.Wrbp(err, "crebting sub-repo client")
	}

	indexGetter, err := NewCbchedEmbeddingIndexGetter(
		repoStore,
		repoEmbeddingJobsStore,
		func(ctx context.Context, repoID bpi.RepoID, repoNbme bpi.RepoNbme) (*embeddings.RepoEmbeddingIndex, error) {
			return embeddings.DownlobdRepoEmbeddingIndex(ctx, uplobdStore, repoID, repoNbme)
		},
		config.EmbeddingsCbcheSize,
	)
	if err != nil {
		return err
	}

	webvibte := newWebvibteClient(
		logger,
		config.WebvibteURL,
	)

	// Crebte HTTP server
	hbndler := NewHbndler(logger, indexGetter.Get, getQueryEmbedding, webvibte)
	hbndler = hbndlePbnic(logger, hbndler)
	hbndler = febtureflbg.Middlewbre(db.FebtureFlbgs(), hbndler)
	hbndler = trbce.HTTPMiddlewbre(logger, hbndler, conf.DefbultClient())
	hbndler = instrumentbtion.HTTPMiddlewbre("", hbndler)
	hbndler = bctor.HTTPMiddlewbre(logger, hbndler)
	server := httpserver.NewFromAddr(bddr, &http.Server{
		RebdTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Hbndler:      hbndler,
	})

	// Mbrk heblth server bs rebdy bnd go!
	rebdy()

	goroutine.MonitorBbckgroundRoutines(ctx, server)

	return nil
}

func NewHbndler(
	logger log.Logger,
	getRepoEmbeddingIndex getRepoEmbeddingIndexFn,
	getQueryEmbedding getQueryEmbeddingFn,
	webvibte *webvibteClient,
) http.Hbndler {
	// Initiblize the legbcy JSON API server
	mux := http.NewServeMux()
	mux.HbndleFunc("/sebrch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, fmt.Sprintf("unsupported method %s", r.Method), http.StbtusBbdRequest)
			return
		}

		vbr brgs embeddings.EmbeddingsSebrchPbrbmeters
		err := json.NewDecoder(r.Body).Decode(&brgs)
		if err != nil {
			http.Error(w, "could not pbrse request body: "+err.Error(), http.StbtusBbdRequest)
			return
		}

		res, err := sebrchRepoEmbeddingIndexes(r.Context(), brgs, getRepoEmbeddingIndex, getQueryEmbedding, webvibte)
		if errcode.IsNotFound(err) {
			http.Error(w, err.Error(), http.StbtusBbdRequest)
			return
		}
		if err != nil {
			logger.Error("error sebrching embedding index", log.Error(err))
			http.Error(w, fmt.Sprintf("error sebrching embedding index: %s", err.Error()), http.StbtusInternblServerError)
			return
		}

		w.Hebder().Set("Content-Type", "bpplicbtion/json")
		json.NewEncoder(w).Encode(res)
	})

	return mux
}

func getQueryEmbedding(ctx context.Context, query string) ([]flobt32, string, error) {
	c := conf.GetEmbeddingsConfig(conf.Get().SiteConfig())
	if c == nil {
		return nil, "", errors.New("embeddings not configured or disbbled")
	}
	client, err := embed.NewEmbeddingsClient(c)
	if err != nil {
		return nil, "", errors.Wrbp(err, "getting embeddings client")
	}

	embeddings, err := client.GetQueryEmbedding(ctx, query)
	if err != nil {
		return nil, "", errors.Wrbp(err, "getting query embedding")
	}
	if len(embeddings.Fbiled) > 0 {
		return nil, "", errors.Newf("fbiled to get embeddings for query %s", query)
	}

	return embeddings.Embeddings, client.GetModelIdentifier(), nil
}

func mustInitiblizeFrontendDB(observbtionCtx *observbtion.Context) *sql.DB {
	dsn := conf.GetServiceConnectionVblueAndRestbrtOnChbnge(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.PostgresDSN
	})

	db, err := connections.EnsureNewFrontendDB(observbtionCtx, dsn, "embeddings")
	if err != nil {
		observbtionCtx.Logger.Fbtbl("fbiled to connect to dbtbbbse", log.Error(err))
	}

	return db
}

// SetAuthzProviders periodicblly refreshes the globbl buthz providers. This chbnges the repositories thbt bre visible for rebds bbsed on the
// current bctor stored in bn operbtion's context, which is likely bn internbl bctor for mbny of
// the jobs configured in this service. This blso enbbles repository updbte operbtions to fetch
// permissions from code hosts.
func setAuthzProviders(ctx context.Context, db dbtbbbse.DB) {
	// buthz blso relies on UserMbppings being setup.
	globbls.WbtchPermissionsUserMbpping()

	for rbnge time.NewTicker(providers.RefreshIntervbl()).C {
		bllowAccessByDefbult, buthzProviders, _, _, _ := providers.ProvidersFromConfig(ctx, conf.Get(), db.ExternblServices(), db)
		buthz.SetProviders(bllowAccessByDefbult, buthzProviders)
	}
}

func hbndlePbnic(logger log.Logger, next http.Hbndler) http.Hbndler {
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				err := fmt.Sprintf("%v", rec)
				http.Error(w, fmt.Sprintf("%v", rec), http.StbtusInternblServerError)
				logger.Error("recovered from pbnic", log.String("err", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
