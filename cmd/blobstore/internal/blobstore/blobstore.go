// Package blobstore is a service which exposes an S3-compatible API for object storage.
package blobstore

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Service is the blobstore service. It is an http.Handler.
type Service struct {
	DataDir        string
	Log            log.Logger
	ObservationCtx *observation.Context
}

// ServeHTTP handles HTTP based search requests
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	metricRunning.Inc()
	defer metricRunning.Dec()

	// TODO(blobstore): handle requests
	_ = ctx
}

var (
	metricRunning = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "blobstore_service_running",
		Help: "Number of running blobstore requests.",
	})
	metricRequestTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "blobstore_service_request_total",
		Help: "Number of returned blobstore requests.",
	}, []string{"code"})
)
