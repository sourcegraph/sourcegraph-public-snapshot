package shared

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

// Observation exports available shared observable and group constructors related
// to the metrics emitted by internal/metrics.NewREDMetrics in the Go backend.
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

type SharedObservationGroupOptions struct {
	// Total transforms the default observable used to construct the operation count panel.
	Total ObservableOption

	// Duration transforms the default observable used to construct the duration histogram panel.
	Duration ObservableOption

	// Errors transforms the default observable used to construct the error count panel.
	Errors ObservableOption

	// ErrorRate transforms the default observable used to construct the error rate panel.
	ErrorRate ObservableOption
}

type ObservationGroupOptions struct {
	GroupConstructorOptions
	SharedObservationGroupOptions

	// Aggregate is the option container for the group's aggregate panels.
	// This option should only be supplied if a label is supplied (via the By option) by which to split the data.
	Aggregate *SharedObservationGroupOptions
}

// NewGroup creates a group containing panels displaying the total number of operations, operation
// duration histogram, number of errors, and error rate for the given observable within the given
// container, based on the RED methodology.
//
// Requires a:
//   - counter of the format `src_{options.MetricNameRoot}_total`
//   - histogram of the format `src_{options.MetricNameRoot}_duration_seconds_bucket`
//   - counter of the format `src_{options.MetricNameRoot}_errors_total`
//
// These metrics can be created via internal/metrics.NewREDMetrics in the Go backend.
func (observationConstructor) NewGroup(containerName string, owner monitoring.ObservableOwner, options ObservationGroupOptions) monitoring.Group {
	rows := make([]monitoring.Row, 0, 2)
	if options.JobLabel == "" {
		options.JobLabel = "job"
	}

	if len(options.By) == 0 {
		if options.Aggregate != nil {
			panic("Aggregate must not be supplied when By is not set")
		}
	} else if options.Aggregate != nil {
		aggregateOptions := options.ObservableConstructorOptions
		aggregateOptions.By = nil
		aggregateOptions.MetricDescriptionRoot = "aggregate " + aggregateOptions.MetricDescriptionRoot

		aggregateRow := Observation.newRow(containerName, owner, *options.Aggregate, aggregateOptions)
		if len(aggregateRow) > 0 {
			rows = append(rows, aggregateRow)
		}
	}

	splitRow := Observation.newRow(containerName, owner, options.SharedObservationGroupOptions, options.ObservableConstructorOptions)
	if len(splitRow) > 0 {
		rows = append(rows, splitRow)
	}

	if len(rows) == 0 {
		panic("No rows were constructed. Supply at least one ObservableOption to this group constructor.")
	}

	return monitoring.Group{
		Title:  fmt.Sprintf("%s: %s", titlecase(options.Namespace), options.DescriptionRoot),
		Hidden: options.Hidden,
		Rows:   rows,
	}
}

// newRow constructs a single row of (up to) four panels composing observation metrics.
func (c observationConstructor) newRow(containerName string, owner monitoring.ObservableOwner, groupOptions SharedObservationGroupOptions, observableOptions ObservableConstructorOptions) monitoring.Row {
	row := make(monitoring.Row, 0, 4)
	if groupOptions.Total != nil {
		row = append(row, groupOptions.Total(Observation.Total(observableOptions)(containerName, owner)).Observable())
	}
	if groupOptions.Duration != nil {
		row = append(row, groupOptions.Duration(Observation.Duration(observableOptions)(containerName, owner)).Observable())
	}
	if groupOptions.Errors != nil {
		row = append(row, groupOptions.Errors(Observation.Errors(observableOptions)(containerName, owner)).Observable())
	}
	if groupOptions.ErrorRate != nil {
		row = append(row, groupOptions.ErrorRate(Observation.ErrorRate(observableOptions)(containerName, owner)).Observable())
	}

	return row
}
