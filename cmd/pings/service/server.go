package service

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"

	"github.com/sourcegraph/sourcegraph/internal/pubsub"
	"github.com/sourcegraph/sourcegraph/internal/updatecheck"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var meter = otel.GetMeterProvider().Meter("pings/shared")

func registerServerHandlers(logger log.Logger, r *mux.Router, pubsubClient pubsub.TopicPublisher) error {
	r.Path("/").Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://docs.sourcegraph.com/admin/pings", http.StatusFound)
	})

	requestCounter, err := meter.Int64Counter(
		"pings.request_count",
		metric.WithDescription("number of requests to the update check handler"),
	)
	if err != nil {
		return errors.Errorf("create request counter: %v", err)
	}
	requestHasUpdateCounter, err := meter.Int64Counter(
		"pings.request_has_update_count",
		metric.WithDescription("number of requests to the update check handler where an update is available"),
	)
	if err != nil {
		return errors.Errorf("create request has update counter: %v", err)
	}
	errorCounter, err := meter.Int64Counter(
		"pings.error_count",
		metric.WithDescription("number of errors that occur while publishing server pings"),
	)
	if err != nil {
		return errors.Errorf("create request counter: %v", err)
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
	return nil
}
