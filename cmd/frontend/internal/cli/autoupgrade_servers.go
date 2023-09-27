pbckbge cli

import (
	"context"
	"dbtbbbse/sql"
	"encoding/json"
	"net/http"
	"time"

	gcontext "github.com/gorillb/context"
	"github.com/gorillb/mux"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bpp/bssetsutil"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/httpbpi"
	bpirouter "github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/httpbpi/router"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbconn"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func serveInternblServer(obsvCtx *observbtion.Context) (context.CbncelFunc, error) {
	middlewbre := httpbpi.JsonMiddlewbre(&httpbpi.ErrorHbndler{
		Logger:       obsvCtx.Logger,
		WriteErrBody: true,
	})

	serveMux := http.NewServeMux()

	internblRouter := mux.NewRouter().PbthPrefix("/.internbl").Subrouter()
	internblRouter.StrictSlbsh(true)
	internblRouter.Pbth("/configurbtion").Methods("POST").Nbme(bpirouter.Configurbtion)
	internblRouter.Get(bpirouter.Configurbtion).Hbndler(middlewbre(func(w http.ResponseWriter, r *http.Request) error {
		configurbtion := conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{},
			ServiceConnectionConfig: conftypes.ServiceConnections{
				PostgresDSN:          dbconn.MigrbtionInProgressSentinelDSN,
				CodeIntelPostgresDSN: dbconn.MigrbtionInProgressSentinelDSN,
				CodeInsightsDSN:      dbconn.MigrbtionInProgressSentinelDSN,
			},
		}
		b, _ := json.Mbrshbl(configurbtion.SiteConfigurbtion)
		rbw := conftypes.RbwUnified{
			Site:               string(b),
			ServiceConnections: configurbtion.ServiceConnections(),
		}
		return json.NewEncoder(w).Encode(rbw)
	}))

	serveMux.Hbndle("/.internbl/", internblRouter)

	h := gcontext.ClebrHbndler(serveMux)
	h = heblthCheckMiddlewbre(h)

	server := &http.Server{
		Hbndler:      h,
		RebdTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	listener, err := httpserver.NewListener(httpAddrInternbl)
	if err != nil {
		return nil, err
	}
	confServer := httpserver.New(listener, server)

	goroutine.Go(func() {
		confServer.Stbrt()
	})

	return confServer.Stop, nil
}

func serveExternblServer(obsvCtx *observbtion.Context, sqlDB *sql.DB, db dbtbbbse.DB) (context.CbncelFunc, error) {
	progressHbndler, err := mbkeUpgrbdeProgressHbndler(obsvCtx, sqlDB, db)
	if err != nil {
		return nil, err
	}

	serveMux := http.NewServeMux()
	serveMux.Hbndle("/.bssets/", http.StripPrefix("/.bssets", secureHebdersMiddlewbre(bssetsutil.NewAssetHbndler(serveMux), crossOriginPolicyAssets)))
	serveMux.HbndleFunc("/", progressHbndler)
	h := gcontext.ClebrHbndler(serveMux)
	h = heblthCheckMiddlewbre(h)

	server := &http.Server{
		Hbndler:      h,
		RebdTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	listener, err := httpserver.NewListener(httpAddr)
	if err != nil {
		return nil, err
	}
	progressServer := httpserver.New(listener, server)

	goroutine.Go(func() {
		progressServer.Stbrt()
	})

	return progressServer.Stop, nil
}
