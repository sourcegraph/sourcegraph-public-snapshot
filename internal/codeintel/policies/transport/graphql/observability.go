package graphql

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	// Configurations
	getConfigurationPolicies      *observation.Operation
	getConfigurationPolicyByID    *observation.Operation
	createConfigurationPolicy     *observation.Operation
	updateConfigurationPolicy     *observation.Operation
	deleteConfigurationPolicyByID *observation.Operation

	// Retention Policy
	getRetentionPolicyOverview *observation.Operation

	// Repository
	getPreviewRepositoryFilter *observation.Operation
	getPreviewGitObjectFilter  *observation.Operation

	// Factory
	getPolicyResolverFactory *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	m := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_policies_transport_graphql",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.policies.transport.graphql.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	return &operations{
		// Configurations
		getConfigurationPolicies:      op("GetConfigurationPolicies"),
		getConfigurationPolicyByID:    op("GetConfigurationPolicyByID"),
		createConfigurationPolicy:     op("CreateConfigurationPolicy"),
		updateConfigurationPolicy:     op("UpdateConfigurationPolicy"),
		deleteConfigurationPolicyByID: op("DeleteConfigurationPolicyByID"),

		// Retention
		getRetentionPolicyOverview: op("GetRetentionPolicyOverview"),

		// Repository
		getPreviewRepositoryFilter: op("PreviewRepositoryFilter"),
		getPreviewGitObjectFilter:  op("PreviewGitObjectFilter"),

		// Factory
		getPolicyResolverFactory: op("GetPolicyResolverFactory"),
	}
}

func observeResolver(
	ctx context.Context,
	err *error,
	operation *observation.Operation,
	threshold time.Duration,
	observationArgs observation.Args,
) (context.Context, observation.TraceLogger, func()) {
	start := time.Now()
	ctx, trace, endObservation := operation.With(ctx, err, observationArgs)

	return ctx, trace, func() {
		duration := time.Since(start)
		endObservation(1, observation.Args{})

		if duration >= threshold {
			// use trace logger which includes all relevant fields
			lowSlowRequest(trace, duration, err)
		}
	}
}

func lowSlowRequest(logger log.Logger, duration time.Duration, err *error) {
	fields := []log.Field{log.Duration("duration", duration)}
	if err != nil && *err != nil {
		fields = append(fields, log.Error(*err))
	}
	logger.Warn("Slow codeintel request", fields...)
}
