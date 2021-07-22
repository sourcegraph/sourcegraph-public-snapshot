package shared

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

var (
	// ResetterRecordResets creates an observable from the given options backed by
	// the counter specifying the number of records reset to queued state.
	//
	// Requires a counter of the format `src_{options.MetricNameRoot}_record_resets_total`
	ResetterRecordResets observableConstructor = func(options ObservableConstructorOptions) sharedObservable {
		options.MetricNameRoot += "_record_resets"
		return StandardCount("records reset to queued state")(options)
	}

	// ResetterRecordResetFailures creates an observable from the given options backed by
	// the counter specifying the number of records reset to errored state.
	//
	// Requires a counter of the format `src_{options.MetricNameRoot}_record_reset_failures_total`
	ResetterRecordResetFailures observableConstructor = func(options ObservableConstructorOptions) sharedObservable {
		options.MetricNameRoot += "_record_reset_failures"
		return StandardCount("records reset to errored state")(options)
	}
)

type ResetterGroupOptions struct {
	GroupConstructorOptions

	// Total transforms the default observable used to construct the reset count panel.
	RecordResets ObservableOption

	// Duration transforms the default observable used to construct the reset failure count panel.
	RecordResetFailures ObservableOption

	// Errors transforms the default observable used to construct the resetter error rate panel.
	Errors ObservableOption
}

// NewResetterGroup creates a group containing panels displaying the total number of records
// reset, the number of records moved to errored, and the error rate of the resetter operating
// within the given container.
//
// Requires a:
//   - counter of the format `src_{options.MetricNameRoot}_record_resets_total`
//   - counter of the format `src_{options.MetricNameRoot}_record_reset_failures_total`
//   - counter of the format `src_{options.MetricNameRoot}_record_reset_errors_total`
//
// These metrics are currently created by hand and assigned as fields of an instance of an
// internal/workerutil/dbworker/ResetterMetrics struct in the Go backend. Metrics are emitted
// by the resetter processes themselves.
func NewResetterGroup(containerName string, owner monitoring.ObservableOwner, options ResetterGroupOptions) monitoring.Group {
	errorsOptions := options.ObservableConstructorOptions
	errorsOptions.MetricNameRoot += "_record_reset"

	return monitoring.Group{
		Title:  fmt.Sprintf("[%s] Queue resetter: %s", options.Namespace, options.DescriptionRoot),
		Hidden: options.Hidden,
		Rows: []monitoring.Row{
			{
				options.RecordResets.safeApply(ResetterRecordResets(options.ObservableConstructorOptions)(containerName, owner)).Observable(),
				options.RecordResetFailures.safeApply(ResetterRecordResetFailures(options.ObservableConstructorOptions)(containerName, owner)).Observable(),
				options.Errors.safeApply(ObservationErrors(errorsOptions)(containerName, owner)).Observable(),
			},
		},
	}
}
