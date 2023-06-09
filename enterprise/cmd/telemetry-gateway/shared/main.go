package shared

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/telemetry-gateway/internal/events"
	events2 "github.com/sourcegraph/sourcegraph/enterprise/cmd/telemetry-gateway/shared/events"
	sgtrace "github.com/sourcegraph/sourcegraph/internal/trace"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/instrumentation"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/requestclient"
	"github.com/sourcegraph/sourcegraph/internal/service"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func Main(ctx context.Context, obctx *observation.Context, ready service.ReadyFunc, config *Config) error {
	// Enable tracing, at this point tracing wouldn't have been enabled yet because
	// we run Cody Gateway without conf which means Sourcegraph tracing is not enabled.
	shutdownTracing, err := maybeEnableTracing(ctx,
		obctx.Logger.Scoped("tracing", "tracing configuration"),
		config.Trace)
	if err != nil {
		return errors.Wrap(err, "maybeEnableTracing")
	}
	defer shutdownTracing()

	// var eventLogger events.Logger
	// if config.BigQuery.ProjectID != "" {
	// 	eventLogger, err = events.NewBigQueryLogger(config.BigQuery.ProjectID, config.BigQuery.Dataset, config.BigQuery.Table)
	// 	if err != nil {
	// 		return errors.Wrap(err, "create BigQuery event logger")
	// 	}
	//
	// 	// If a buffer is configured, wrap in events.BufferedLogger
	// 	if config.BigQuery.EventBufferSize > 0 {
	// 		eventLogger = events.NewBufferedLogger(obctx.Logger, eventLogger, config.BigQuery.EventBufferSize)
	// 	}
	// } else {
	// 	eventLogger = events.NewStdoutLogger(obctx.Logger)
	//
	// 	// Useful for testing event logging in a way that has latency that is
	// 	// somewhat similar to BigQuery.
	// 	if os.Getenv("CODY_GATEWAY_BUFFERED_LAGGY_EVENT_LOGGING_FUN_TIMES_MODE") == "true" {
	// 		eventLogger = events.NewBufferedLogger(
	// 			obctx.Logger,
	// 			events.NewDelayedLogger(eventLogger),
	// 			config.BigQuery.EventBufferSize)
	// 	}
	// }
	//
	// // Supported actor/auth sources
	// sources := actor.Sources{
	// 	anonymous.NewSource(true, config.ActorConcurrencyLimit),
	// }
	//
	// authr := &auth.Authenticator{
	// 	Logger:      obctx.Logger.Scoped("auth", "authentication middleware"),
	// 	EventLogger: eventLogger,
	// 	Sources:     sources,
	// }

	topicConfig := events.TopicConfig{
		ProjectName: config.PubSub.ProjectName,
		TopicName:   config.PubSub.TopicName,
	}

	// Set up our handler chain,
	handler := NewHandler(topicConfig)

	// Instrumentation layers
	handler = requestLogger(obctx.Logger.Scoped("requests", "HTTP requests"), handler)
	var otelhttpOpts []otelhttp.Option
	if !config.InsecureDev {
		// Outside of dev, we're probably running as a standalone service, so treat
		// incoming spans as links
		otelhttpOpts = append(otelhttpOpts, otelhttp.WithPublicEndpoint())
	}
	handler = instrumentation.HTTPMiddleware("telemetry-gateway", handler, otelhttpOpts...)

	// Collect request client for downstream handlers. Outside of dev, we always set up
	// Cloudflare in from of Cody Gateway. This comes first.
	hasCloudflare := !config.InsecureDev
	handler = requestclient.ExternalHTTPMiddleware(handler, hasCloudflare)

	// Initialize our server
	server := httpserver.NewFromAddr(config.Address, &http.Server{
		ReadTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Handler:      handler,
	})
	// Mark health server as ready and go!
	ready()
	obctx.Logger.Info("service ready", log.String("address", config.Address))

	// Collect background routines
	backgroundRoutines := []goroutine.BackgroundRoutine{server}
	// Block until done
	goroutine.MonitorBackgroundRoutines(ctx, backgroundRoutines...)

	return nil
}

func NewHandler(config events.TopicConfig) http.Handler {
	r := mux.NewRouter()

	// V1 service routes
	v1 := r.PathPrefix("/v1").Subrouter()

	v1.Path("/events").Methods(http.MethodPost).Handler(eventsHandler(config))

	return r
}

func requestLogger(logger log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only requestclient is available at the point, actor middleware is later
		rc := requestclient.FromContext(r.Context())

		sgtrace.Logger(r.Context(), logger).Debug("Request",
			log.String("method", r.Method),
			log.String("path", r.URL.Path),
			log.String("requestclient.ip", rc.IP),
			log.String("requestclient.forwardedFor", rc.ForwardedFor))

		next.ServeHTTP(w, r)
	})
}

func eventsHandler(config events.TopicConfig) http.Handler {
	sender := events.FakeEventSender{}

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var proxyReq events2.TelemetryGatewayProxyRequest
		if err := json.NewDecoder(request.Body).Decode(&proxyReq); err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		// err := events.SendEvents(request.Context(), proxyReq, events.TopicConfig{})
		// if err != nil {
		// 	http.Error(writer, err.Error(), http.StatusInternalServerError) // need to improve this
		// }

		err := sender.SendEvents(request.Context(), proxyReq)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		// Use proxyReq...
		_, _ = writer.Write([]byte("jobs done"))
	})
}
