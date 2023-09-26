package shared

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/hubspot/hubspotutil"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/profiler"
	"github.com/sourcegraph/sourcegraph/internal/pubsub"
	"github.com/sourcegraph/sourcegraph/internal/service"
	"github.com/sourcegraph/sourcegraph/internal/updatecheck"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func Main(ctx context.Context, obctx *observation.Context, ready service.ReadyFunc, config *Config) error {
	profiler.Init()

	shutdownOtel, err := initOpenTelemetry(ctx, obctx.Logger, config.OpenTelemetry)
	if err != nil {
		return errors.Wrap(err, "initOpenTelemetry")
	}
	defer shutdownOtel()

	// Initialize HTTP server
	serverHandler, err := newServerHandler(obctx.Logger, config)
	if err != nil {
		return errors.Errorf("create server handler: %v", err)
	}

	addr := fmt.Sprintf(":%d", config.Port)
	server := httpserver.NewFromAddr(
		addr,
		&http.Server{
			ReadTimeout:  time.Minute,
			WriteTimeout: time.Minute,
			Handler:      serverHandler,
		},
	)

	// Mark health server as ready and go!
	ready()
	obctx.Logger.Info("service ready", log.String("address", addr))

	// Block until done
	goroutine.MonitorBackgroundRoutines(ctx, server)
	return nil
}

var meter = otel.GetMeterProvider().Meter("pings/shared")

func newServerHandler(logger log.Logger, config *Config) (http.Handler, error) {
	r := mux.NewRouter()

	r.Path("/").Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://docs.sourcegraph.com/admin/pings", http.StatusFound)
	})

	r.Path("/-/version").Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(version.Version()))
	})

	pubsubClient, err := pubsub.NewTopicClient(config.PubSub.ProjectID, config.PubSub.TopicID)
	if err != nil {
		return nil, errors.Errorf("create Pub/Sub client: %v", err)
	}
	r.Path("/-/healthz").Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secret := strings.TrimPrefix(strings.ToLower(r.Header.Get("Authorization")), "bearer ")
		if subtle.ConstantTimeCompare([]byte(secret), []byte(config.DiagnosticsSecret)) == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.URL.Query().Get("full-suite") == "" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
			return
		}

		// NOTE: Only mark as failed and respond with a non-200 status code if a critical
		// component fails, otherwise the service would be marked as unhealthy and stop
		// serving requests (in Cloud Run).
		failed := false
		status := make(map[string]string)
		if err := pubsubClient.Ping(r.Context()); err != nil {
			failed = true
			status["pubsubClient"] = err.Error()
			logger.Error("failed to ping Pub/Sub client", log.Error(err))
		} else {
			status["pubsubClient"] = "OK"
		}

		if hubspotutil.HasAPIKey() {
			if err := hubspotutil.Client().Ping(r.Context(), 30*time.Second); err != nil {
				status["hubspotClient"] = err.Error()
				logger.Error("failed to ping HubSpot client", log.Error(err))
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

	requestCounter, err := meter.Int64Counter(
		"pings.request_count",
		metric.WithDescription("number of requests to the update check handler"),
	)
	if err != nil {
		return nil, errors.Errorf("create request counter: %v", err)
	}
	requestHasUpdateCounter, err := meter.Int64Counter(
		"pings.request_has_update_count",
		metric.WithDescription("number of requests to the update check handler where an update is available"),
	)
	if err != nil {
		return nil, errors.Errorf("create request has update counter: %v", err)
	}
	errorCounter, err := meter.Int64Counter(
		"pings.error_count",
		metric.WithDescription("number of errors that occur while publishing server pings"),
	)
	if err != nil {
		return nil, errors.Errorf("create request counter: %v", err)
	}
	errorCounter.Add(context.Background(), 0) // Add a zero value to ensure the metric is visible to scrapers.
	meter := &updatecheck.Meter{
		RequestCounter:          requestCounter,
		RequestHasUpdateCounter: requestHasUpdateCounter,
		ErrorCounter:            errorCounter,
	}
	r.Path("/updates").
		Methods(http.MethodGet, http.MethodPost).
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			updatecheck.Handle(logger, pubsubClient, meter, w, r)
		})
	return r, nil
}
