package shared

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

// Observation exports available shared observable and group constructors related
// to the metrics emitted by internal/metrics.NewOperationMetrics in the Go backend.
var Observation = observationConstructor{
	Total:     Standard.Count("operations"),
	Duration:  Standard.Duration("operation"),
	Errors:    Standard.Errors("operation"),
	ErrorRate: Standard.ErrorRate("operation"),
}

// observationConstructor provides `Observation` implementations.
type observationConstructor struct {
	// Total creates an observable from the given options backed by the counter specifying
	// the number of operations.
	//
	// Requires a counter of the format `src_{options.MetricNameRoot}_total`
	Total observableConstructor

	// Duration creates an observable from the given options backed by the histogram
	// specifying the duration of operations.
	//
	// Requires a histogram of the format `src_{options.MetricNameRoot}_duration_seconds_bucket`
	Duration observableConstructor

	// Errors creates an observable from the given options backed by the counter specifying
	// the number of operations that resulted in an error.
	//
	// Requires a counter of the format `src_{options.MetricNameRoot}_errors_total`
	Errors observableConstructor

	// ErrorRate creates an observable from the given options backed by the counters specifying
	// the number of operations that resulted in success and error, respectively.
	//
	// Requires a:
	//   - counter of the format `src_{options.MetricNameRoot}_total`
	//   - counter of the format `src_{options.MetricNameRoot}_errors_total`
	ErrorRate observableConstructor
}

type ObservationGroupOptions struct {
	GroupConstructorOptions

	// Total transforms the default observable used to construct the operation count panel.
	Total ObservableOption

	// Duration transforms the default observable used to construct the duration histogram panel.
	Duration ObservableOption

	// Errors transforms the default observable used to construct the error count panel.
	Errors ObservableOption

	// ErrorRate transforms the default observable used to construct the error rate panel.
	ErrorRate ObservableOption

	// AggregateTotal transforms the default observable used to construct the aggregate operation count panel.
	// This option should only be supplied if a label is supplied (via the By option) by which to split the data
	// series.
	AggregateTotal ObservableOption

	// AggregateDuration transforms the default observable used to construct the aggregate duration histogram panel.
	// This option should only be supplied if a label is supplied (via the By option) by which to split the data
	// series.
	AggregateDuration ObservableOption

	// AggregateErrors transforms the default observable used to construct the aggregate error rate panel.
	// This option should only be supplied if a label is supplied (via the By option) by which to split the data
	// series.
	AggregateErrors ObservableOption

	// AggregateErrorRate transforms the default observable used to construct the aggregate error rate panel.
	// This option should only be supplied if a label is supplied (via the By option) by which to split the data
	AggregateErrorRate ObservableOption
}

// NewGroup creates a group containing panels displaying the total number of operations, operation
// duration histogram, number of errors, and error rate for the given observable within the given
// icontainer.
//
// Requires a:
//   - counter of the format `src_{options.MetricNameRoot}_total`
//   - histogram of the format `src_{options.MetricNameRoot}_duration_seconds_bucket`
//   - counter of the format `src_{options.MetricNameRoot}_errors_total`
//
// These metrics can be created via internal/metrics.NewOperationMetrics in the Go backend.
func (observationConstructor) NewGroup(containerName string, owner monitoring.ObservableOwner, options ObservationGroupOptions) monitoring.Group {
	if len(options.By) == 0 {
		if options.AggregateTotal != nil || options.AggregateDuration != nil || options.AggregateErrors != nil || options.AggregateErrorRate != nil {
			panic("AggregateTotal, AggregateDuration, AggregateErrors, and AggregateErrorRate must not be supplied when By is not set")
		}
	} else {
		if options.AggregateTotal == nil || options.AggregateDuration == nil || options.AggregateErrors == nil || options.AggregateErrorRate == nil {
			panic("AggregateTotal, AggregateDuration, AggregateErrors, and AggregateErrorRate must be supplied when By is set")
		}
	}

	rows := []monitoring.Row{
		{
			options.Total.safeApply(Observation.Total(options.ObservableConstructorOptions)(containerName, owner)).Observable(),
			options.Duration.safeApply(Observation.Duration(options.ObservableConstructorOptions)(containerName, owner)).Observable(),
			options.Errors.safeApply(Observation.Errors(options.ObservableConstructorOptions)(containerName, owner)).Observable(),
			options.ErrorRate.safeApply(Observation.ErrorRate(options.ObservableConstructorOptions)(containerName, owner)).Observable(),
		},
	}

	if len(options.By) > 0 {
		aggregateOptions := options.ObservableConstructorOptions
		aggregateOptions.By = nil
		aggregateOptions.MetricDescriptionRoot = "aggregate " + aggregateOptions.MetricDescriptionRoot

		aggregateRow := monitoring.Row{
			options.AggregateTotal.safeApply(Observation.Total(aggregateOptions)(containerName, owner)).Observable(),
			options.AggregateDuration.safeApply(Observation.Duration(aggregateOptions)(containerName, owner)).Observable(),
			options.AggregateErrors.safeApply(Observation.Errors(aggregateOptions)(containerName, owner)).Observable(),
			options.AggregateErrorRate.safeApply(Observation.ErrorRate(aggregateOptions)(containerName, owner)).Observable(),
		}

		rows = append([]monitoring.Row{aggregateRow}, rows...)
	}

	return monitoring.Group{
		Title:  fmt.Sprintf("%s: %s", titlecase(options.Namespace), options.DescriptionRoot),
		Hidden: options.Hidden,
		Rows:   rows,
	}
}
