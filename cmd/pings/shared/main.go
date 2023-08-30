package shared

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/hubspot/hubspotutil"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/pubsub"
	"github.com/sourcegraph/sourcegraph/internal/service"
	"github.com/sourcegraph/sourcegraph/internal/updatecheck"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func Main(ctx context.Context, obctx *observation.Context, ready service.ReadyFunc, config *Config) error {
	// Initialize our server
	serverHandler, err := newServerHandler(obctx.Logger, config)
	if err != nil {
		return errors.Errorf("create server handler: %v", err)
	}
	server := httpserver.NewFromAddr(
		config.Address,
		&http.Server{
			ReadTimeout:  75 * time.Second,
			WriteTimeout: 10 * time.Minute,
			Handler:      serverHandler,
		},
	)

	// Mark health server as ready and go!
	ready()
	obctx.Logger.Info("service ready", log.String("address", config.Address))

	// Block until done
	goroutine.MonitorBackgroundRoutines(ctx, server)
	return nil
}

func newServerHandler(logger log.Logger, config *Config) (http.Handler, error) {
	r := mux.NewRouter()

	r.Path("/-/version").Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(version.Version()))
	})

	pubsubClient, err := pubsub.NewTopicClient(config.PubSub.ProjectID, config.PubSub.TopicID)
	if err != nil {
		return nil, errors.Errorf("create Pub/Sub client: %v", err)
	}
	r.Path("/-/healthz").Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// NOTE: Only mark as failed and respond with a non-200 status code if a critical
		// component fails, otherwise the service would be marked as unhealthy and stop
		// serving requests (in Cloud Run).
		failed := false
		status := make(map[string]string)
		if err := pubsubClient.Ping(context.Background()); err != nil {
			failed = true
			status["pubsubClient"] = err.Error()
		} else {
			status["pubsubClient"] = "OK"
		}

		if hubspotutil.HasAPIKey() {
			if err := hubspotutil.Client().Ping(30 * time.Second); err != nil {
				status["hubspotClient"] = err.Error()
			} else {
				status["hubspotClient"] = "OK"
			}
		} else {
			status["hubspotClient"] = "Not configured"
		}

		if failed {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		err := json.NewEncoder(w).Encode(status)
		if err != nil {
			logger.Error("failed to encode health check status", log.Error(err))
		}
		return
	})
	r.Path("/updates").
		Methods(http.MethodGet, http.MethodPost).
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			updatecheck.HandlePingRequest(logger, pubsubClient, w, r)
		})
	return r, nil
}
