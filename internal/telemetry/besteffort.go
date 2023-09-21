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
	return &BestEffortEventRecorder{logger: logger, recorder: recorder}
}

func (r *BestEffortEventRecorder) Record(ctx context.Context, feature eventFeature, action eventAction, parameters EventParameters) {
	if err := r.recorder.Record(ctx, feature, action, parameters); err != nil {
		trace.Logger(ctx, r.logger).Error("failed to record telemetry event",
			log.String("feature", string(feature)),
			log.String("action", string(action)),
			log.Int("parameters.version", parameters.Version),
			log.Error(err))
	}
}
