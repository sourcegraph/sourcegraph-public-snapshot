package api

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	bundles "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// ErrorLogger captures the method required for logging an error.
type ErrorLogger interface {
	Error(msg string, ctx ...interface{})
}

// OperationMetrics contains three common metrics for any operation.
type OperationMetrics struct {
	Duration *prometheus.HistogramVec // How long did it take?
	Count    *prometheus.CounterVec   // How many things were processed?
	Errors   *prometheus.CounterVec   // How many errors occurred?
}

// Observe registers an observation of a single operation.
func (m *OperationMetrics) Observe(secs, count float64, err error, lvals ...string) {
	if m == nil {
		return
	}

	m.Duration.WithLabelValues(lvals...).Observe(secs)
	m.Count.WithLabelValues(lvals...).Add(count)
	if err != nil {
		m.Errors.WithLabelValues(lvals...).Add(1)
	}
}

// MustRegister registers all metrics in OperationMetrics in the given prometheus.Registerer.
// It panics in case of failure.
func (m *OperationMetrics) MustRegister(r prometheus.Registerer) {
	r.MustRegister(m.Duration)
	r.MustRegister(m.Count)
	r.MustRegister(m.Errors)
}

// CodeIntelAPIMetrics encapsulates the Prometheus metrics of a CodeIntelAPI.
type CodeIntelAPIMetrics struct {
	FindClosestDumps *OperationMetrics
	Definitions      *OperationMetrics
	References       *OperationMetrics
	Hover            *OperationMetrics
}

// NewCodeIntelAPIMetrics returns CodeIntelAPIMetrics that need to be registered in a Prometheus registry.
func NewCodeIntelAPIMetrics() CodeIntelAPIMetrics {
	return CodeIntelAPIMetrics{
		FindClosestDumps: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-api-server",
				Name:      "code_intel_api_find_closest_dumps_duration_seconds",
				Help:      "Time spent performing find closest dumps queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-api-server",
				Name:      "code_intel_api_find_closest_dumps_total",
				Help:      "Total number of find closest dumps queries",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-api-server",
				Name:      "code_intel_api_find_closest_dumps_errors_total",
				Help:      "Total number of errors when performing find closest dumps queries",
			}, []string{}),
		},
		Definitions: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-api-server",
				Name:      "code_intel_api_definitions_duration_seconds",
				Help:      "Time spent performing definitions queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-api-server",
				Name:      "code_intel_api_definitions_total",
				Help:      "Total number of definitions queries",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-api-server",
				Name:      "code_intel_api_definitions_errors_total",
				Help:      "Total number of errors when performing definitions queries",
			}, []string{}),
		},
		References: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-api-server",
				Name:      "code_intel_api_references_duration_seconds",
				Help:      "Time spent performing references queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-api-server",
				Name:      "code_intel_api_references_total",
				Help:      "Total number of references queries",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-api-server",
				Name:      "code_intel_api_references_errors_total",
				Help:      "Total number of errors when performing references queries",
			}, []string{}),
		},
		Hover: &OperationMetrics{
			Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-api-server",
				Name:      "code_intel_api_hover_duration_seconds",
				Help:      "Time spent performing hover queries",
			}, []string{}),
			Count: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-api-server",
				Name:      "code_intel_api_hover_total",
				Help:      "Total number of hover queries",
			}, []string{}),
			Errors: prometheus.NewCounterVec(prometheus.CounterOpts{
				Namespace: "src",
				Subsystem: "precise-code-intel-api-server",
				Name:      "code_intel_api_hover_errors_total",
				Help:      "Total number of errors when performing hover queries",
			}, []string{}),
		},
	}
}

// An ObservedCodeIntelAPI wraps another CodeIntelAPI with error logging, Prometheus metrics, and tracing.
type ObservedCodeIntelAPI struct {
	codeIntelAPI CodeIntelAPI
	logger       ErrorLogger
	metrics      CodeIntelAPIMetrics
	tracer       trace.Tracer
}

var _ CodeIntelAPI = &ObservedCodeIntelAPI{}

// NewObservedCodeIntelAPI wraps the given CodeIntelAPI with error logging, Prometheus metrics, and tracing.
func NewObservedCodeIntelAPI(codeIntelAPI CodeIntelAPI, logger ErrorLogger, metrics CodeIntelAPIMetrics, tracer trace.Tracer) CodeIntelAPI {
	return &ObservedCodeIntelAPI{
		codeIntelAPI: codeIntelAPI,
		logger:       logger,
		metrics:      metrics,
		tracer:       tracer,
	}
}

func (api *ObservedCodeIntelAPI) FindClosestDumps(ctx context.Context, repositoryID int, commit, file string) (_ []db.Dump, err error) {
	tr, ctx := api.tracer.New(ctx, "CodeIntelAPI.FindClosestDumps", "")
	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		api.metrics.FindClosestDumps.Observe(secs, 1, err)
		log(api.logger, "code-intel-api.find-closest-dumps", err)
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	return api.codeIntelAPI.FindClosestDumps(ctx, repositoryID, commit, file)
}

func (api *ObservedCodeIntelAPI) Definitions(ctx context.Context, file string, line, character, uploadID int) (_ []ResolvedLocation, err error) {
	tr, ctx := api.tracer.New(ctx, "CodeIntelAPI.Definitions", "")
	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		api.metrics.Definitions.Observe(secs, 1, err)
		log(api.logger, "code-intel-api.definitions", err)
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	return api.codeIntelAPI.Definitions(ctx, file, line, character, uploadID)
}

func (api *ObservedCodeIntelAPI) References(ctx context.Context, repositoryID int, commit string, limit int, cursor Cursor) (_ []ResolvedLocation, _ Cursor, _ bool, err error) {
	tr, ctx := api.tracer.New(ctx, "CodeIntelAPI.References", "")
	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		api.metrics.References.Observe(secs, 1, err)
		log(api.logger, "code-intel-api.references", err)
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	return api.codeIntelAPI.References(ctx, repositoryID, commit, limit, cursor)
}

func (api *ObservedCodeIntelAPI) Hover(ctx context.Context, file string, line, character, uploadID int) (_ string, _ bundles.Range, _ bool, err error) {
	tr, ctx := api.tracer.New(ctx, "CodeIntelAPI.Hover", "")
	defer func(began time.Time) {
		secs := time.Since(began).Seconds()
		api.metrics.Hover.Observe(secs, 1, err)
		log(api.logger, "code-intel-api.hover", err)
		tr.SetError(err)
		tr.Finish()
	}(time.Now())

	return api.codeIntelAPI.Hover(ctx, file, line, character, uploadID)
}

func log(lg ErrorLogger, msg string, err error, ctx ...interface{}) {
	if err == nil {
		return
	}

	lg.Error(msg, append(append(make([]interface{}, 0, len(ctx)+2), "error", err), ctx...)...)
}
