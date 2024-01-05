package cli

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	gcontext "github.com/gorilla/context"
	"github.com/gorilla/mux"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/assetsutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/schema"
)

func serveInternalServer(obsvCtx *observation.Context) (context.CancelFunc, error) {
	middleware := httpapi.JsonMiddleware(&httpapi.ErrorHandler{
		Logger:       obsvCtx.Logger,
		WriteErrBody: true,
	})

	serveMux := http.NewServeMux()

	internalRouter := mux.NewRouter().PathPrefix("/.internal").Subrouter()
	internalRouter.StrictSlash(true)
	internalRouter.Path("/configuration").Methods("POST").
		Handler(middleware(func(w http.ResponseWriter, r *http.Request) error {
			configuration := conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{},
				ServiceConnectionConfig: conftypes.ServiceConnections{
					PostgresDSN:          dbconn.MigrationInProgressSentinelDSN,
					CodeIntelPostgresDSN: dbconn.MigrationInProgressSentinelDSN,
					CodeInsightsDSN:      dbconn.MigrationInProgressSentinelDSN,
				},
			}
			b, _ := json.Marshal(configuration.SiteConfiguration)
			raw := conftypes.RawUnified{
				Site:               string(b),
				ServiceConnections: configuration.ServiceConnections(),
			}
			return json.NewEncoder(w).Encode(raw)
		}))

	serveMux.Handle("/.internal/", internalRouter)

	h := gcontext.ClearHandler(serveMux)
	h = healthCheckMiddleware(h)

	server := &http.Server{
		Handler:      h,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	listener, err := httpserver.NewListener(httpAddrInternal)
	if err != nil {
		return nil, err
	}
	confServer := httpserver.New(listener, server)

	goroutine.Go(func() {
		confServer.Start()
	})

	return confServer.Stop, nil
}

func serveExternalServer(obsvCtx *observation.Context, sqlDB *sql.DB, db database.DB) (context.CancelFunc, error) {
	progressHandler, err := makeUpgradeProgressHandler(obsvCtx, sqlDB, db)
	if err != nil {
		return nil, err
	}

	serveMux := http.NewServeMux()
	serveMux.Handle("/.assets/", http.StripPrefix("/.assets", secureHeadersMiddleware(assetsutil.NewAssetHandler(serveMux), crossOriginPolicyAssets)))
	serveMux.HandleFunc("/", progressHandler)
	h := gcontext.ClearHandler(serveMux)
	h = healthCheckMiddleware(h)

	server := &http.Server{
		Handler:      h,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	listener, err := httpserver.NewListener(httpAddr)
	if err != nil {
		return nil, err
	}
	progressServer := httpserver.New(listener, server)

	goroutine.Go(func() {
		progressServer.Start()
	})

	return progressServer.Stop, nil
}
