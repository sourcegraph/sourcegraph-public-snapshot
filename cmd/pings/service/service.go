package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/hubspot/hubspotutil"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/pubsub"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/background"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime/contract"
)

// Service is the pings service.
type Service struct{}

var _ runtime.Service[Config] = (*Service)(nil)

func (Service) Name() string    { return "pings" }
func (Service) Version() string { return version.Version() }

func (Service) Initialize(ctx context.Context, logger log.Logger, contract runtime.ServiceContract, config Config) (background.Routine, error) {
	pubsubClient, err := pubsub.NewTopicClient(config.PubSub.ProjectID, config.PubSub.TopicID)
	if err != nil {
		return nil, errors.Errorf("create Pub/Sub client: %v", err)
	}

	// Initialize HTTP server
	handler := mux.NewRouter()
	if err := registerServerHandlers(logger, handler, pubsubClient); err != nil {
		return nil, errors.Errorf("register server handlers: %v", err)
	}

	// Register MSP diagnostics ('/-/version', '/-/healthz', etc)
	contract.Diagnostics.RegisterDiagnosticsHandlers(
		&muxRegisterer{handler: handler},
		&serviceState{
			logger:       logger.Scoped("state"),
			pubsubClient: pubsubClient,
		})

	return background.LIFOStopRoutine{
		httpserver.NewFromAddr(
			fmt.Sprintf(":%d", contract.Port),
			&http.Server{
				ReadTimeout:  time.Minute,
				WriteTimeout: time.Minute,
				Handler:      handler,
			},
		),
		background.CallbackRoutine{
			StopFunc: pubsubClient.Stop,
		},
	}, nil
}

type serviceState struct {
	logger       log.Logger
	pubsubClient pubsub.TopicClient
}

var _ contract.ServiceState = (*serviceState)(nil)

func (s *serviceState) Healthy(ctx context.Context, query url.Values) error {
	if query.Get("full-suite") == "" {
		return nil
	}

	// NOTE: Only mark as failed and respond with a non-200 status code if a critical
	// component fails, otherwise the service would be marked as unhealthy and stop
	// serving requests (in Cloud Run).
	failed := false
	status := make(map[string]string)
	if err := s.pubsubClient.Ping(ctx); err != nil {
		failed = true
		status["pubsubClient"] = err.Error()
		s.logger.Error("failed to ping Pub/Sub client", log.Error(err))
	} else {
		status["pubsubClient"] = "OK"
	}

	if hubspotutil.HasAPIKey() {
		if err := hubspotutil.Client().Ping(ctx, 30*time.Second); err != nil {
			status["hubspotClient"] = err.Error()
			s.logger.Error("failed to ping HubSpot client", log.Error(err))
		} else {
			status["hubspotClient"] = "OK"
		}
	} else {
		status["hubspotClient"] = "Not configured"
	}

	if failed {
		msg, err := json.Marshal(status)
		if err != nil {
			s.logger.Error("failed to encode health check status", log.Error(err))
			return errors.Wrap(err,
				"some health check components failed, but failed to render status details")
		}
		return errors.New(string(msg))
	}

	return nil
}

type muxRegisterer struct{ handler *mux.Router }

func (m *muxRegisterer) Handle(pattern string, handler http.Handler) {
	_ = m.handler.Handle(pattern, handler)
}
