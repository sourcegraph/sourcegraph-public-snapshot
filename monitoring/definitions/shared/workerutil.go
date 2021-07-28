package shared

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

// Workerutil exports available shared observable and group constructors related to workerutil
// metrics emitted by internal/workerutil.NewMetrics in the Go backend.
var Workerutil workerutilConstructor

// workerutilConstructor provides `Workerutil` implementations.
type workerutilConstructor struct{}

// Total creates an observable from the given options backed by the counter specifying the
// number of handler invocations performed by workerutil.
//
// Requires a counter of the format `src_{options.MetricNameRoot}_processor_total`
func (workerutilConstructor) Total(options ObservableConstructorOptions) sharedObservable {
	options.MetricNameRoot += "_processor"
	return Observation.Total(options)
}

// Duration creates an observable from the given options backed by the histogram specifying
// the duration of handler invocations performed by workerutil.
//
// Requires a histogram of the format `src_{options.MetricNameRoot}_processor_duration_seconds_bucket`
func (workerutilConstructor) Duration(options ObservableConstructorOptions) sharedObservable {
	options.MetricNameRoot += "_processor"
	return Observation.Duration(options)
}

// Errors creates an observable from the given options backed by the counter specifying the number
// of handler invocations that resulted in an error performed by workerutil.
//
// Requires a counter of the format `src_{options.MetricNameRoot}_processor_errors_total`
func (workerutilConstructor) Errors(options ObservableConstructorOptions) sharedObservable {
	options.MetricNameRoot += "_processor"
	return Observation.Errors(options)
}

// ErrorRate creates an observable from the given options backed by the counters specifying
// the number of operations that resulted in success and error, respectively.
//
// Requires a:
//   - counter of the format `src_{options.MetricNameRoot}_total`
//   - counter of the format `src_{options.MetricNameRoot}_errors_total`
func (workerutilConstructor) ErrorRate(options ObservableConstructorOptions) sharedObservable {
	options.MetricNameRoot += "_processor"
	return Observation.ErrorRate(options)
}

// Handlers creates an observable from the given options backed by the gauge specifying the number
// of handler invocations performed by workerutil.
//
// Requires a gauge of the format `src_{options.MetricNameRoot}_processor_handlers`
func (workerutilConstructor) Handlers(options ObservableConstructorOptions) sharedObservable {
	return func(containerName string, owner monitoring.ObservableOwner) Observable {
		filters := makeFilters(containerName, options.Filters...)
		by, legendPrefix := makeBy(options.By...)

		return Observable{
			Name:        fmt.Sprintf("%s_handlers", options.MetricNameRoot),
			Description: fmt.Sprintf("%s active handlers", options.MetricDescriptionRoot),
			Query:       fmt.Sprintf(`sum%s(src_%s_processor_handlers{%s})`, by, options.MetricNameRoot, filters),
			Panel:       monitoring.Panel().LegendFormat(fmt.Sprintf("%shandlers", legendPrefix)),
			Owner:       owner,
		}
	}
}

type WorkerutilGroupOptions struct {
	GroupConstructorOptions

	// Total transforms the default observable used to construct the processor operation count panel.
	Total ObservableOption

	// Duration transforms the default observable used to construct the processor duration histogram panel.
	Duration ObservableOption

	// Errors transforms the default observable used to construct the processor error count panel.
	Errors ObservableOption

	// ErrorRate transforms the default observable used to construct the processor error rate panel.
	ErrorRate ObservableOption

	// Handlers transforms the default observable used to construct the processor count panel.
	Handlers ObservableOption
}

// NewGroup creates a group containing panels displaying the total number of jobs, duration of
// processing, error count, error rate, and number of workers operating on the queue for the given
// worker observable within the given container.
//
// Requires a:
//   - counter of the format `src_{options.MetricNameRoot}_processor_total`
//   - histogram of the format `src_{options.MetricNameRoot}_processor_duration_seconds_bucket`
//   - counter of the format `src_{options.MetricNameRoot}_processor_errors_total`
//   - gauge of the format `src_{options.MetricNameRoot}_processor_handlers`
//
// These metrics can be created via internal/workerutil.NewMetrics("..._processor", ...) in the Go
// backend. Note that we supply the `_processor` suffix here explicitly so that we can differentiate
// metrics for the worker and the queue that backs the worker while still using the same metric name
// root.
func (workerutilConstructor) NewGroup(containerName string, owner monitoring.ObservableOwner, options WorkerutilGroupOptions) monitoring.Group {
	return monitoring.Group{
		Title:  fmt.Sprintf("%s: %s", titlecase(options.Namespace), options.DescriptionRoot),
		Hidden: options.Hidden,
		Rows: []monitoring.Row{
			{
				options.Handlers.safeApply(Workerutil.Handlers(options.ObservableConstructorOptions)(containerName, owner)).Observable(),
			}, {
				options.Total.safeApply(Workerutil.Total(options.ObservableConstructorOptions)(containerName, owner)).Observable(),
				options.Duration.safeApply(Workerutil.Duration(options.ObservableConstructorOptions)(containerName, owner)).Observable(),
				options.Errors.safeApply(Workerutil.Errors(options.ObservableConstructorOptions)(containerName, owner)).Observable(),
				options.ErrorRate.safeApply(Workerutil.ErrorRate(options.ObservableConstructorOptions)(containerName, owner)).Observable(),
			},
		},
	}
}
