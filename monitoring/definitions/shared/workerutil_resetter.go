package shared

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

// WorkerutilResetter exports available shared observable and group constructors related to workerutil
// resetter metrics emitted by instances of internal/workerutil/dbworker/ResetterMetrics in the Go backend.
var WorkerutilResetter workerutilResetterConstructor

// workerutilResetterConstructor provides `WorkerutilResetter` implementations.
type workerutilResetterConstructor struct{}

// Resets creates an observable from the given options backed by the counter specifying the
// number of records reset to queued state.
//
// Requires a counter of the format `src_{options.MetricNameRoot}_record_resets_total`
func (workerutilResetterConstructor) Resets(options ObservableConstructorOptions) sharedObservable {
	options.MetricNameRoot += "_record_resets"
	return Standard.Count("records reset to queued state")(options)
}

// ResetFailures creates an observable from the given options backed by the counter specifying
// the number of records reset to errored state.
//
// Requires a counter of the format `src_{options.MetricNameRoot}_record_reset_failures_total`
func (workerutilResetterConstructor) ResetFailures(options ObservableConstructorOptions) sharedObservable {
	options.MetricNameRoot += "_record_reset_failures"
	return Standard.Count("records reset to errored state")(options)
}

type ResetterGroupOptions struct {
	GroupConstructorOptions

	// Total transforms the default observable used to construct the reset count panel.
	RecordResets ObservableOption

	// Duration transforms the default observable used to construct the reset failure count panel.
	RecordResetFailures ObservableOption

	// Errors transforms the default observable used to construct the resetter error rate panel.
	Errors ObservableOption
}

// NewGroup creates a group containing panels displaying the total number of records reset, the number
// of records moved to errored, and the error rate of the resetter operating within the given container.
//
// Requires any of the following:
//   - counter of the format `src_{options.MetricNameRoot}_record_resets_total`
//   - counter of the format `src_{options.MetricNameRoot}_record_reset_failures_total`
//   - counter of the format `src_{options.MetricNameRoot}_record_reset_errors_total`
//
// These metrics are currently created by hand and assigned as fields of an instance of an
// internal/workerutil/dbworker/ResetterMetrics struct in the Go backend. Metrics are emitted
// by the resetter processes themselves.
func (workerutilResetterConstructor) NewGroup(containerName string, owner monitoring.ObservableOwner, options ResetterGroupOptions) monitoring.Group {
	row := make(monitoring.Row, 0, 3)
	if options.RecordResets != nil {
		row = append(row, options.RecordResets(WorkerutilResetter.Resets(options.ObservableConstructorOptions)(containerName, owner)).Observable())
	}
	if options.RecordResetFailures != nil {
		row = append(row, options.RecordResetFailures(WorkerutilResetter.ResetFailures(options.ObservableConstructorOptions)(containerName, owner)).Observable())
	}
	if options.Errors != nil {
		errorsOptions := options.ObservableConstructorOptions
		errorsOptions.MetricNameRoot += "_record_reset"
		row = append(row, options.Errors(Observation.Errors(errorsOptions)(containerName, owner)).Observable())
	}

	if len(row) == 0 {
		panic("No rows were constructed. Supply at least one ObservableOption to this group constructor.")
	}

	return monitoring.Group{
		Title:  fmt.Sprintf("%s: %s", titlecase(options.Namespace), options.DescriptionRoot),
		Hidden: options.Hidden,
		Rows:   []monitoring.Row{row},
	}
}
