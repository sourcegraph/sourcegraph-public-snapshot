// Package blobstore is a service which exposes an S3-compatible API for object storage.
package blobstore

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Service is the blobstore service. It is an http.Handler.
type Service struct {
	DataDir        string
	Log            log.Logger
	ObservationCtx *observation.Context
}

// ServeHTTP handles HTTP based search requests
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metricRunning.Inc()
	defer metricRunning.Dec()

	err := s.serve(w, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.Log.Error("serving request", sglog.Error(err))
		fmt.Fprintf(w, "error: %v", err)
		return
	}
}

func (s *Service) serve(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	path := strings.FieldsFunc(r.URL.Path, func(r rune) bool { return r == '/' })
	switch r.Method {
	case "PUT":
		if len(path) == 1 { // PUT /<bucket>
			if r.ContentLength != 0 {
				return errors.Newf("expected CreateBucket request to have content length 0: %s %s", r.Method, r.URL)
			}
			if err := s.createBucket(ctx, path[0]); err != nil {
				return err
			}
			w.WriteHeader(http.StatusOK)
			return nil
		}
		return errors.Newf("unexpected request(0): %s %s", r.Method, r.URL)
	case "GET":
		return errors.Newf("unexpected request(1): %s %s", r.Method, r.URL)
	default:
		return errors.Newf("unexpected request(2): %s %s", r.Method, r.URL)
	}
}

func (s *Service) createBucket(ctx context.Context, name string) error {
	fmt.Println("TODO(blobstore): create bucket", name)
	return nil
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
