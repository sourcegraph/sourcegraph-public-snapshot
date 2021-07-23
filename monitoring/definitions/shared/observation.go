package shared

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

// Observation exports available shared observable and group constructors related
// to the metrics emitted by internal/metrics.NewOperationMetrics in the Go backend.
var Observation = observationConstructor{
	Total:    Standard.Count("operations"),
	Duration: Standard.Duration("operation"),
	Errors:   Standard.Errors("operation"),
}

// observationConstructor provides `Observation` implementations.
type observationConstructor struct {
	// Total creates an observable from the given options backed by the counter specifying
	// the number of operatons.
	//
	// Requires a counter of the format `src_{options.MetricNameRoot}_total`
	Total observableConstructor

	// Duration creates an observable from the given options backed by the histogram
	// specifying the duration of operatons.
	//
	// Requires a histogram of the format `src_{options.MetricNameRoot}_duration_seconds_bucket`
	Duration observableConstructor

	// Errors creates an observable from the given options backed by the counter specifying
	// the number of operatons that resulted in an error.
	//
	// Requires a counter of the format `src_{options.MetricNameRoot}_errors_total`
	Errors observableConstructor
}

type ObservationGroupOptions struct {
	GroupConstructorOptions

	// Total transforms the default observable used to construct the operation count panel.
	Total ObservableOption

	// Duration transforms the default observable used to construct the duration histogram panel.
	Duration ObservableOption

	// Errors transforms the default observable used to construct the error rate panel.
	Errors ObservableOption
}

// NewGroup creates a group containing panels displaying the total number of operations, operation
// duration histogram, and number of errors for the given observable within the given container.
//
// Requires a:
//   - counter of the format `src_{options.MetricNameRoot}_total`
//   - histogram of the format `src_{options.MetricNameRoot}_duration_seconds_bucket`
//   - counter of the format `src_{options.MetricNameRoot}_errors_total`
//
// These metrics can be created via internal/metrics.NewOperationMetrics in the Go backend.
func (observationConstructor) NewGroup(containerName string, owner monitoring.ObservableOwner, options ObservationGroupOptions) monitoring.Group {
	return monitoring.Group{
		Title:  fmt.Sprintf("[%s] Observable: %s", options.Namespace, options.DescriptionRoot),
		Hidden: options.Hidden,
		Rows: []monitoring.Row{
			{
				options.Total.safeApply(Observation.Total(options.ObservableConstructorOptions)(containerName, owner)).Observable(),
				options.Duration.safeApply(Observation.Duration(options.ObservableConstructorOptions)(containerName, owner)).Observable(),
				options.Errors.safeApply(Observation.Errors(options.ObservableConstructorOptions)(containerName, owner)).Observable(),
			},
		},
	}
}
