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
	SharedObservationGroupOptions

	// Handlers transforms the default observable used to construct the processor count panel.
	Handlers ObservableOption
}

// NewGroup creates a group containing panels displaying the total number of jobs, duration of
// processing, error count, error rate, and number of workers operating on the queue for the given
// worker observable within the given container.
//
// Requires any of the following:
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
	row := make(monitoring.Row, 0, 5)
	if options.Handlers != nil {
		row = append(row, options.Handlers(Workerutil.Handlers(options.ObservableConstructorOptions)(containerName, owner)).Observable())
	}
	if options.Total != nil {
		row = append(row, options.Total(Workerutil.Total(options.ObservableConstructorOptions)(containerName, owner)).Observable())
	}
	if options.Duration != nil {
		row = append(row, options.Duration(Workerutil.Duration(options.ObservableConstructorOptions)(containerName, owner)).Observable())
	}
	if options.Errors != nil {
		row = append(row, options.Errors(Workerutil.Errors(options.ObservableConstructorOptions)(containerName, owner)).Observable())
	}
	if options.ErrorRate != nil {
		row = append(row, options.ErrorRate(Workerutil.ErrorRate(options.ObservableConstructorOptions)(containerName, owner)).Observable())
	}

	if len(row) == 0 {
		panic("No rows were constructed. Supply at least one ObservableOption to this group constructor.")
	}

	rows := []monitoring.Row{row}
	if len(row) == 5 {
		// If we have all 5 metrics, put handlers on a row by itself first,
		// followed by the standard observation group panels.
		rows = []monitoring.Row{row[:1], row[1:]}
	}

	return monitoring.Group{
		Title:  fmt.Sprintf("%s: %s", titlecase(options.Namespace), options.DescriptionRoot),
		Hidden: options.Hidden,
		Rows:   rows,
	}
}
