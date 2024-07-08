package server

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/redact"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/sourcegraph/sourcegraph/cmd/telemetry-gateway/internal/events"
	sgtrace "github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	telemetrygatewayv1 "github.com/sourcegraph/sourcegraph/lib/telemetrygateway/v1"
)

func handlePublishEvents(
	ctx context.Context,
	logger log.Logger,
	payloadMetrics *recordEventsRequestPayloadMetrics,
	publisher *events.Publisher,
	events []*telemetrygatewayv1.Event,
) *telemetrygatewayv1.RecordEventsResponse {
	start := time.Now()
	var tr sgtrace.Trace
	tr, ctx = sgtrace.New(ctx, "handlePublishEvents",
		attribute.Int("events", len(events)))
	defer tr.End()

	logger = sgtrace.Logger(ctx, logger)

	// Send off our events
	results := publisher.Publish(ctx, events)

	// Aggregate failure details
	summary := summarizePublishEventsResults(results, summarizePublishEventsResultsOpts{
		onlyReportRetriableAsFailed: publisher.IsSourcegraphInstance(),
	})

	// Record the result on the trace and metrics
	resultAttribute := attribute.String("result", summary.result)
	sourceAttribute := attribute.String("source", publisher.GetSourceName())
	tr.SetAttributes(resultAttribute, sourceAttribute)
	payloadMetrics.length.Record(ctx, int64(len(events)),
		metric.WithAttributes(resultAttribute, sourceAttribute))
	payloadMetrics.processedEvents.Add(ctx, int64(len(summary.succeededEvents)),
		metric.WithAttributes(attribute.Bool("succeeded", true), resultAttribute, sourceAttribute))
	payloadMetrics.processedEvents.Add(ctx, int64(len(summary.failedEvents)),
		metric.WithAttributes(attribute.Bool("succeeded", false), resultAttribute, sourceAttribute))

	// Generate a log message for convenience
	duration := time.Since(start)
	summaryFields := []log.Field{
		log.String("result", summary.result),
		log.Int("submitted", len(events)),
		log.Int("succeeded", len(summary.succeededEvents)),
		log.Int("failed", len(summary.failedEvents)),
		log.Duration("duration", duration),
	}
	if len(summary.failedEvents) > 0 {
		tr.SetError(errors.New(summary.message)) // mark span as failed
		tr.SetAttributes(attribute.Int("failed", len(summary.failedEvents)))

		// Dangerously slow for a single batch, we're at risk of hitting other
		// types of timeouts. Add an error for reporting purposes. We may need
		// to increase default concurrencies or batch sizes.
		if duration > 10*time.Second {
			summaryFields = append(summaryFields,
				log.NamedError("publishDurationError", errors.New("slow publish")))
		}

		logger.Error(summary.message, append(summaryFields, summary.errorFields...)...)
	} else {
		logger.Info(summary.message, summaryFields...)
	}

	return &telemetrygatewayv1.RecordEventsResponse{
		SucceededEvents: summary.succeededEvents,
	}
}

type publishEventsSummary struct {
	// message is a human-readable summary summarizing the result
	message string
	// result is a low-cardinality indicator of the result category
	result string

	errorFields     []log.Field
	succeededEvents []string
	failedEvents    []events.PublishEventResult
}

type summarizePublishEventsResultsOpts struct {
	onlyReportRetriableAsFailed bool
}

func summarizePublishEventsResults(results []events.PublishEventResult, opts summarizePublishEventsResultsOpts) publishEventsSummary {
	var (
		errFields = make([]log.Field, 0)
		succeeded = make([]string, 0, len(results))
		failed    = make([]events.PublishEventResult, 0)
	)

	// We aggregate all errors on a single log entry to get accurate
	// representations of issues in Sentry, while not generating thousands of
	// log entries at the same time. Because this means we only get higher-level
	// logger context, we must annotate the errors with some hidden details to
	// preserve Sentry grouping while adding context for diagnostics.
	for i, result := range results {
		if result.PublishError != nil {
			if result.Retryable {
				// Let the client know that this event failed to submit, so they
				// can retry it.
				failed = append(failed, result)
			} else if opts.onlyReportRetriableAsFailed {
				// Sourcegraph instances will continue to retry unretriable
				// failures - for clients where we should only provide retriable
				// failures, we want to PRETEND that non-retriable issues were
				// successful, so that clients don't retry endlessly.
				//
				// We will still generate a log entry for these errors in all
				// cases, so this discard is okay.
				succeeded = append(succeeded, result.EventID)
			} else {
				// Let the client know that this event failed to submit.
				failed = append(failed, result)
			}

			// Don't record an error for context canceled errors, since it's
			// generally not very interesting - the client probably decided to
			// cancel or time out. Instead, just record the error string, so
			// that it doesn't go to Sentry.
			if errors.IsContextCanceled(result.PublishError) {
				errFields = append(errFields,
					log.String(fmt.Sprintf("error.%d", i), result.PublishError.Error()))
				continue
			}

			// Construct details to annotate the error with in Sentry reports
			// without affecting the error itself (which is important for
			// grouping within Sentry)
			errFields = append(errFields, log.NamedError(fmt.Sprintf("error.%d", i),
				errors.WithSafeDetails(result.PublishError,
					"feature:%[1]q action:%[2]q id:%[3]q %[4]s", // mimic format of result.EventSource
					redact.Safe(result.EventFeature),
					redact.Safe(result.EventAction),
					redact.Safe(result.EventID),
					redact.Safe(result.EventSource),
				),
			))
		} else {
			succeeded = append(succeeded, result.EventID)
		}
	}

	var message, category string
	switch {
	case len(failed) == len(results):
		message = "all events in batch failed to submit"
		category = "complete_failure"
	case len(failed) > 0 && len(failed) < len(results):
		message = "some events in batch failed to submit"
		category = "partial_failure"
	case len(failed) == 0:
		message = "all events in batch submitted successfully"
		category = "success"
	}

	return publishEventsSummary{
		message:         message,
		result:          category,
		errorFields:     errFields,
		succeededEvents: succeeded,
		failedEvents:    failed,
	}
}
