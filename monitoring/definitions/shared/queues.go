package shared

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

var (
	// WorkerutilQueueSize creates an observable from the given options backed by
	// the gauge specifying the number of pending records in a given queue.
	//
	// Requires a gauge of the format `src_{options.MetricNameRoot}_total`
	WorkerutilQueueSize observableConstructor = func(options ObservableConstructorOptions) sharedObservable {
		return func(containerName string, owner monitoring.ObservableOwner) Observable {
			filters := makeFilters(containerName, options.Filters...)
			by, legendPrefix := makeBy(options.By...)

			return Observable{
				Name:        fmt.Sprintf("%s_queue_size", options.MetricNameRoot),
				Description: fmt.Sprintf("%s queue size", options.MetricDescriptionRoot),
				Query:       fmt.Sprintf(`max%s(src_%s_total{%s})`, by, options.MetricDescriptionRoot, filters),
				Panel:       monitoring.Panel().LegendFormat(fmt.Sprintf("%s records", legendPrefix)),
				Owner:       owner,
			}
		}
	}

	// WorkerutilQueueGrowthRate creates an observable from the given options backed by
	// the rate of increase of enqueues compared to the processing rate.
	//
	// Requires a:
	//   - gauge of the format `src_{options.MetricNameRoot}_total`
	//   - counter of the format `src_{options.MetricNameRoot}_processor_total`
	WorkerutilQueueGrowthRate observableConstructor = func(options ObservableConstructorOptions) sharedObservable {
		return func(containerName string, owner monitoring.ObservableOwner) Observable {
			filters := makeFilters(containerName, options.Filters...)
			by, legendPrefix := makeBy(options.By...)

			return Observable{
				Name:        fmt.Sprintf("%s_queue_growth_rate", options.MetricNameRoot),
				Description: fmt.Sprintf("%s queue growth rate over 30m", options.MetricDescriptionRoot),
				Query:       fmt.Sprintf(`sum%[1]s(increase(src_%[2]s_total{%[3]s}[30m])) / sum%[1]s(increase(src_%[2]s_processor_total{%[3]s}[30m]))`, by, options.MetricNameRoot, filters),
				Panel:       monitoring.Panel().LegendFormat(fmt.Sprintf("%s queue growth rate", legendPrefix)),
				Owner:       owner,
			}
		}
	}
)

type QueueSizeGroupOptions struct {
	GroupConstructorOptions

	// QueueSize transforms the default observable used to construct the queue sizes panel.
	QueueSize ObservableOption

	// QueueGrowthRate transforms the default observable used to construct the queue growth rate panel.
	QueueGrowthRate ObservableOption
}

// NewQueueSizeGroup creates a group containing panels displaying metrics to monitor the
// size and growth rate of a queue of work within the given container.
//
// Requires a:
//   - gauge of the format `src_{options.MetricNameRoot}_total`
//   - counter of the format `src_{options.MetricNameRoot}_processor_total`
//
// The queue size metric should be created via a Prometheus gauge function in the Go backend. For
// instructions on how to create the processor metrics, see the `NewWorkerutilGroup` function in
// this package.
func NewQueueSizeGroup(containerName string, owner monitoring.ObservableOwner, options QueueSizeGroupOptions) monitoring.Group {
	return monitoring.Group{
		Title:  fmt.Sprintf("[%s] Queue: %s", options.Namespace, options.DescriptionRoot),
		Hidden: options.Hidden,
		Rows: []monitoring.Row{
			{
				options.QueueSize.safeApply(WorkerutilQueueSize(options.ObservableConstructorOptions)(containerName, owner)).Observable(),
				options.QueueGrowthRate.safeApply(WorkerutilQueueGrowthRate(options.ObservableConstructorOptions)(containerName, owner)).Observable(),
			},
		},
	}
}
