package shared

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

// Queue exports available shared observable and group constructors related to queue sizes
// and process rates.
var Queue queueConstructor

// queueConstructor provides `Queue` implementations.
type queueConstructor struct{}

// Size creates an observable from the given options backed by the gauge specifying the number
// of pending records in a given queue.
//
// Requires a gauge of the format `src_{options.MetricNameRoot}_total`
func (queueConstructor) Size(options ObservableConstructorOptions) sharedObservable {
	return func(containerName string, owner monitoring.ObservableOwner) Observable {
		filters := makeFilters(options.JobLabel, containerName, options.Filters...)
		by, legendPrefix := makeBy(options.By...)

		return Observable{
			Name:        fmt.Sprintf("%s_queue_size", options.MetricNameRoot),
			Description: fmt.Sprintf("%s queue size", options.MetricDescriptionRoot),
			Query:       fmt.Sprintf(`max%s(src_%s_total{%s})`, by, options.MetricNameRoot, filters),
			Panel:       monitoring.Panel().LegendFormat(fmt.Sprintf("%s records", legendPrefix)),
			Owner:       owner,
		}
	}
}

// GrowthRate creates an observable from the given options backed by the rate of increase of
// enqueues compared to the processing rate.
//
// Requires a:
//   - gauge of the format `src_{options.MetricNameRoot}_total`
//   - counter of the format `src_{options.MetricNameRoot}_processor_total`
func (queueConstructor) GrowthRate(options ObservableConstructorOptions) sharedObservable {
	return func(containerName string, owner monitoring.ObservableOwner) Observable {
		filters := makeFilters(options.JobLabel, containerName, options.Filters...)
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

// MaxAge creates an observable from the given options backed by the max of the counters
// specifying the age of the oldest unprocessed record in the queue.
//
// Requires a:
//   - counter of the format `src_{options.MetricNameRoot}_queued_duration_seconds_total`
func (queueConstructor) MaxAge(options ObservableConstructorOptions) sharedObservable {
	return func(containerName string, owner monitoring.ObservableOwner) Observable {
		filters := makeFilters(options.JobLabel, containerName, options.Filters...)
		by, legendPrefix := makeBy(options.By...)

		return Observable{
			Name:        fmt.Sprintf("%s_queued_max_age", options.MetricNameRoot),
			Description: fmt.Sprintf("%s queue longest time in queue", options.MetricDescriptionRoot),
			Query:       fmt.Sprintf(`max%[1]s(src_%[2]s_queued_duration_seconds_total{%[3]s})`, by, options.MetricNameRoot, filters),
			Panel:       monitoring.Panel().LegendFormat(fmt.Sprintf("%s max queued age", legendPrefix)).Unit(monitoring.Seconds),
			Owner:       owner,
		}
	}
}

func (queueConstructor) DequeueCacheSize(options ObservableConstructorOptions) sharedObservable {
	return func(containerName string, owner monitoring.ObservableOwner) Observable {
		filters := makeFilters(options.JobLabel, containerName, options.Filters...)
		_, legendPrefix := makeBy(options.By...)

		return Observable{
			Name:        fmt.Sprintf("multiqueue_%s_dequeue_cache_size", options.MetricNameRoot),
			Description: fmt.Sprintf("%s dequeue cache size for multiqueue executors", options.MetricDescriptionRoot),
			Query:       fmt.Sprintf(`multiqueue_%[1]s_dequeue_cache_size{%[2]s}`, options.MetricNameRoot, filters),
			Panel:       monitoring.Panel().LegendFormat(fmt.Sprintf("%s dequeue cache size", legendPrefix)),
			Owner:       owner,
		}
	}
}

//      "expr": "max by (queue) (src_executor_total{job=~\"^(executor|sourcegraph-code-intel-indexers|executor-batches|frontend|sourcegraph-frontend|worker|sourcegraph-executors).*\",queue=~\"$queue\"})",
//      "expr": "multiqueue_executor_dequeue_cache_size{job=~\"^(executor|sourcegraph-code-intel-indexers|executor-batches|frontend|sourcegraph-frontend|worker|sourcegraph-executors).*\",queue=~\"$queue\"}",

type QueueSizeGroupOptions struct {
	GroupConstructorOptions

	// QueueSize transforms the default observable used to construct the queue sizes panel.
	QueueSize ObservableOption

	// QueueGrowthRate transforms the default observable used to construct the queue growth rate panel.
	QueueGrowthRate ObservableOption

	// QueueMaxAge transforms the default observable used to construct the queue's oldest record age panel.
	QueueMaxAge ObservableOption
}

type MultiqueueGroupOptions struct {
	GroupConstructorOptions

	QueueDequeueCacheSize ObservableOption
}

// NewGroup creates a group containing panels displaying metrics to monitor the size and growth rate
// of a queue of work within the given container, as well as the age of the oldest unprocessed entry
// in the queue.
//
// Requires any of the following:
//   - gauge of the format `src_{options.MetricNameRoot}_total`
//   - counter of the format `src_{options.MetricNameRoot}_processor_total`
//   - counter of the format `src_{options.MetricNameRoot}_queued_duration_seconds_total`
//
// The queue size metric should be created via a Prometheus gauge function in the Go backend. For
// instructions on how to create the processor metrics, see the `NewWorkerutilGroup` function in
// this package.
func (queueConstructor) NewGroup(containerName string, owner monitoring.ObservableOwner, options QueueSizeGroupOptions) monitoring.Group {
	row := make(monitoring.Row, 0, 3)
	if options.QueueSize != nil {
		row = append(row, options.QueueSize(Queue.Size(options.ObservableConstructorOptions)(containerName, owner)).Observable())
	}
	if options.QueueGrowthRate != nil {
		row = append(row, options.QueueGrowthRate(Queue.GrowthRate(options.ObservableConstructorOptions)(containerName, owner)).Observable())
	}
	if options.QueueMaxAge != nil {
		row = append(row, options.QueueMaxAge(Queue.MaxAge(options.ObservableConstructorOptions)(containerName, owner)).Observable())
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

func (queueConstructor) NewMultiqueueGroup(containerName string, owner monitoring.ObservableOwner, options MultiqueueGroupOptions) monitoring.Group {
	row := make(monitoring.Row, 0, 1)
	if options.QueueDequeueCacheSize != nil {
		row = append(row, options.QueueDequeueCacheSize(Queue.DequeueCacheSize(options.ObservableConstructorOptions)(containerName, owner)).Observable())
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
