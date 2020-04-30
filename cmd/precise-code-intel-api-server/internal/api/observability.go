package api

import (
	"context"

	"github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	bundles "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observability"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// CodeIntelAPIMetrics encapsulates the Prometheus metrics of a Reader.
type CodeIntelAPIMetrics struct {
	FindClosestDumps *metrics.OperationMetrics
	Definitions      *metrics.OperationMetrics
	References       *metrics.OperationMetrics
	Hover            *metrics.OperationMetrics
}

// MustRegister registers all metrics in CodeIntelAPIMetrics in the given
// prometheus.Registerer. It panics in case of failure.
func (rm CodeIntelAPIMetrics) MustRegister(r prometheus.Registerer) {
	for _, om := range []*metrics.OperationMetrics{
		rm.FindClosestDumps,
		rm.Definitions,
		rm.References,
		rm.Hover,
	} {
		om.MustRegister(prometheus.DefaultRegisterer)
	}
}

// NewCodeIntelAPIMetrics returns CodeIntelAPIMetrics that need to be registered in a Prometheus registry.
func NewCodeIntelAPIMetrics(subsystem string) CodeIntelAPIMetrics {
	return CodeIntelAPIMetrics{
		FindClosestDumps: metrics.NewOperationMetrics(subsystem, "code_intel_api", "find_closest_dumps", metrics.WithCountHelp("The total number of dumps returned")),
		Definitions:      metrics.NewOperationMetrics(subsystem, "code_intel_api", "definitions", metrics.WithCountHelp("The total number of definitions returned")),
		References:       metrics.NewOperationMetrics(subsystem, "code_intel_api", "references", metrics.WithCountHelp("The total number of references returned")),
		Hover:            metrics.NewOperationMetrics(subsystem, "code_intel_api", "hover"),
	}
}

// An ObservedCodeIntelAPI wraps another CodeIntelAPI with error logging, Prometheus metrics, and tracing.
type ObservedCodeIntelAPI struct {
	codeIntelAPI CodeIntelAPI
	logger       logging.ErrorLogger
	metrics      CodeIntelAPIMetrics
	tracer       trace.Tracer
}

var _ CodeIntelAPI = &ObservedCodeIntelAPI{}

// NewObservedCodeIntelAPI wraps the given CodeIntelAPI with error logging, Prometheus metrics, and tracing.
func NewObservedCodeIntelAPI(codeIntelAPI CodeIntelAPI, logger logging.ErrorLogger, metrics CodeIntelAPIMetrics, tracer trace.Tracer) CodeIntelAPI {
	return &ObservedCodeIntelAPI{
		codeIntelAPI: codeIntelAPI,
		logger:       logger,
		metrics:      metrics,
		tracer:       tracer,
	}
}

// FindClosestDumps calls into the inner CodeIntelAPI and registers the observed results.
func (api *ObservedCodeIntelAPI) FindClosestDumps(ctx context.Context, repositoryID int, commit, file string) (dumps []db.Dump, err error) {
	ctx, endTrace := api.prepTrace(ctx, &err, api.metrics.FindClosestDumps, "CodeIntelAPI.FindClosestDumps", "codeIntelAPI.find-closest-dumps")
	defer func() { endTrace(float64(len(dumps))) }()

	return api.codeIntelAPI.FindClosestDumps(ctx, repositoryID, commit, file)
}

// Definitions calls into the inner CodeIntelAPI and registers the observed results.
func (api *ObservedCodeIntelAPI) Definitions(ctx context.Context, file string, line, character, uploadID int) (definitions []ResolvedLocation, err error) {
	ctx, endTrace := api.prepTrace(ctx, &err, api.metrics.Definitions, "CodeIntelAPI.Definitions", "codeIntelAPI.definitions")
	defer func() { endTrace(float64(len(definitions))) }()

	return api.codeIntelAPI.Definitions(ctx, file, line, character, uploadID)
}

// References calls into the inner CodeIntelAPI and registers the observed results.
func (api *ObservedCodeIntelAPI) References(ctx context.Context, repositoryID int, commit string, limit int, cursor Cursor) (references []ResolvedLocation, _ Cursor, _ bool, err error) {
	ctx, endTrace := api.prepTrace(ctx, &err, api.metrics.References, "CodeIntelAPI.References", "codeIntelAPI.references")
	defer func() { endTrace(float64(len(references))) }()

	return api.codeIntelAPI.References(ctx, repositoryID, commit, limit, cursor)
}

// Hover calls into the inner CodeIntelAPI and registers the observed results.
func (api *ObservedCodeIntelAPI) Hover(ctx context.Context, file string, line, character, uploadID int) (_ string, _ bundles.Range, _ bool, err error) {
	ctx, endTrace := api.prepTrace(ctx, &err, api.metrics.Hover, "CodeIntelAPI.Hover", "codeIntelAPI.hover")
	defer endTrace(1)

	return api.codeIntelAPI.Hover(ctx, file, line, character, uploadID)
}

func (api *ObservedCodeIntelAPI) prepTrace(
	ctx context.Context,
	err *error,
	metrics *metrics.OperationMetrics,
	traceName string,
	logName string,
	preFields ...log.Field,
) (context.Context, observability.EndTraceFn) {
	return observability.PrepTrace(ctx, api.logger, metrics, api.tracer, err, traceName, logName, preFields...)
}
