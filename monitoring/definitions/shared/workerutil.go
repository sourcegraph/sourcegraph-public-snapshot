package shared

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

var (
	// WorkerutilProcessorTotal creates an observable from the given options backed by
	// the counter specifying the number of handler invocations performed by workerutil.
	//
	// Requires a counter of the format `src_{options.MetricNameRoot}_processor_total`
	WorkerutilProcessorTotal observableConstructor = func(options ObservableConstructorOptions) sharedObservable {
		options.MetricNameRoot += "_processor"
		return ObservationTotal(options)
	}

	// WorkerutilProcessorDuration creates an observable from the given options backed by
	// the histogram specifying the duration of handler invocations performed by workerutil.
	//
	// Requires a histogram of the format `src_{options.MetricNameRoot}_processor_duration_seconds_bucket`
	WorkerutilProcessorDuration observableConstructor = func(options ObservableConstructorOptions) sharedObservable {
		options.MetricNameRoot += "_processor"
		return ObservationDuration(options)
	}

	// WorkerutilProcessorErrors creates an observable from the given options backed by
	// the counter specifying the number of handler invocations that resulted in an error
	// performed by workerutil.
	//
	// Requires a counter of the format `src_{options.MetricNameRoot}_processor_errors_total`
	WorkerutilProcessorErrors observableConstructor = func(options ObservableConstructorOptions) sharedObservable {
		options.MetricNameRoot += "_processor"
		return ObservationDuration(options)
	}

	// WorkerutilProcessorHandlers creates an observable from the given options backed by
	// the gauge specifying the number of handler invocations performed by workerutil.
	//
	// Requires a gauge of the format `src_{options.MetricNameRoot}_processor_handlers`
	WorkerutilProcessorHandlers observableConstructor = func(options ObservableConstructorOptions) sharedObservable {
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
)

type WorkerutilGroupOptions struct {
	GroupConstructorOptions

	// Total transforms the default observable used to construct the processor operation count panel.
	Total ObservableOption

	// Duration transforms the default observable used to construct the processor duration histogram panel.
	Duration ObservableOption

	// Errors transforms the default observable used to construct the processor error rate panel.
	Errors ObservableOption

	// Handlers transforms the default observable used to construct the processor count panel.
	Handlers ObservableOption
}

// NewWorkerutilGroup creates a group containing panels displaying the total number of jobs,
// duration of processing, error rate, and number of workers operating on the queue for the
// given worker observable within the given container.
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
func NewWorkerutilGroup(containerName string, owner monitoring.ObservableOwner, options WorkerutilGroupOptions) monitoring.Group {
	return monitoring.Group{
		Title:  fmt.Sprintf("[%s] Queue handler: %s", options.Namespace, options.DescriptionRoot),
		Hidden: options.Hidden,
		Rows: []monitoring.Row{
			{
				options.Total.safeApply(WorkerutilProcessorTotal(options.ObservableConstructorOptions)(containerName, owner)).Observable(),
				options.Duration.safeApply(WorkerutilProcessorDuration(options.ObservableConstructorOptions)(containerName, owner)).Observable(),
				options.Errors.safeApply(WorkerutilProcessorErrors(options.ObservableConstructorOptions)(containerName, owner)).Observable(),
				options.Handlers.safeApply(WorkerutilProcessorErrors(options.ObservableConstructorOptions)(containerName, owner)).Observable(),
			},
		},
	}
}
