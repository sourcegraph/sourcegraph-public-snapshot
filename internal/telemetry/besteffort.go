package telemetry

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// BestEffortEventRecorder is a version of EventRecorder that logs errors instead
// of returning them. This is useful for non-critical telemetry collection.
type BestEffortEventRecorder struct {
	logger   log.Logger
	recorder *EventRecorder
}

// NewEventRecorder creates a custom event recorder backed by a store
// implementation. In general, prefer to use the telemetryrecorder.NewBestEffort()
// constructor instead.
func NewBestEffortEventRecorder(logger log.Logger, recorder *EventRecorder) *BestEffortEventRecorder {
	return &BestEffortEventRecorder{
		logger:   logger.AddCallerSkip(1), // report from Recorder callsite
		recorder: recorder,
	}
}

// Record records a single telemetry event with the context's Sourcegraph
// actor, logging any recording errors it encounters. Parameters are optional.
func (r *BestEffortEventRecorder) Record(ctx context.Context, feature eventFeature, action eventAction, parameters *EventParameters) {
	if err := r.recorder.Record(ctx, feature, action, parameters); err != nil {
		fields := []log.Field{
			log.String("feature", string(feature)),
			log.String("action", string(action)),
			log.Error(err),
		}
		if parameters != nil {
			fields = append(fields, log.Int("parameters.version", parameters.Version))
		}
		trace.Logger(ctx, r.logger).Error("failed to record telemetry event", fields...)
	}
}
