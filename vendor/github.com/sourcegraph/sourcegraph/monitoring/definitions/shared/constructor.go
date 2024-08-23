package shared

import (
	"fmt"
	"strings"

	"github.com/prometheus/common/model"
)

type ObservableConstructorOptions struct {
	// MetricNameRoot is the root of the Prometheus metric name used to construct the query
	// for the target panel. For example:
	//
	// `src_search_query_errors_total`
	//      ^^^^^^^^^^^^ root
	//
	// See the documentation of the observable or group constructor to determine the exact
	// metrics that are expected to be emitted from the backend based on the supplied metric
	// name.
	MetricNameRoot string

	// MetricDescriptionRoot is a human-readable name for the object represented by each
	// metric. This is used to disambiguate more generic terms such as "requests" or "records".
	// The value in the panel description or legend will be generated but made more specific
	// by this value. For example:
	//
	//               code intel resolver operations
	//   metric desc ^^^^^^^^^^^^^^^^^^^ ^^^^^^^^^^ generic term (chosen by constructor)
	//
	// This value should start with a lower-case letter. Note that setting the `By` field
	// will add a prefix to the constructed legend.
	MetricDescriptionRoot string

	// JobLabel is the name of the label used to denote the job name. If unset, "job" is used.
	JobLabel string

	// Filters are additional prometheus filter expressions used to select or hide values
	// for a given label pattern.
	Filters []string

	// By are label names that should not be aggregated together. Supplying labels here
	// will increase the number of series on the target panel. The legends for each series
	// will be updated to include the value of each label group supplied here. For example,
	// assuming options.By = []string{"queue", "shard"}:
	//
	//                             batches-01 store operations
	// queue + shared label values ^^^^^^^^^^ ^^^^^^^^^^^^^^^^ metric desc + generic term (chosen by constructor)
	By []string

	// RangeWindow allows setting a custom window for functions like `rate` and `increase`. By default it is
	// set to 5m.
	RangeWindow model.Duration
}

// observableConstructor is a type of constructor function used in this package that creates
// a shared observable given a set of common observable options.
type observableConstructor func(options ObservableConstructorOptions) sharedObservable

type GroupConstructorOptions struct {
	// ObservableConstructorOptions are shared between child observables of the group.
	ObservableConstructorOptions

	// Namespace specifies the component or team owning the enclosed set of metrics. This
	// value is displayed in the title of the group containing the observable. For example:
	//
	// [codeintel] Queue handler: LSIF uploads
	//  ^^^^^^^^^ namespace
	Namespace string

	// DescriptionRoot is a human-readable value that disambiguates the source of data from
	// similar groups. This value is displayed in the legend of the panel as well as in the
	// title of the group containing the observable (if constructed by this package). For
	// example:
	//
	// [codeintel] Queue handler: LSIF uploads
	//                            ^^^^^^^^^^^^ name root
	DescriptionRoot string

	// Hidden sets the Hidden field of the group containing the observable.
	Hidden bool
}

// makeFilters creates metric filters based on the given container name that matches
// against the container name as well as any additionally supplied label filter
// expressions. The given container name may be string or pattern, which will be matched
// against the prefix of the value of the job label. Note that this excludes replicas like
// -0 and -1 in docker-compose.
func makeFilters(containerLabel, containerName string, filters ...string) string {
	if containerLabel == "" {
		containerLabel = "job"
	}

	filters = append(filters, fmt.Sprintf(`%s=~"^%s.*"`, containerLabel, containerName))
	return strings.Join(filters, ",")
}

// makeBy returns the suffix if the aggregator expression.
//
//	e.g. max by (queue)
//	         ^^^^^^^^^^
//
// legendPrefix is a prefix to be used as part of the legend consisting of
// placeholder values that will render to the value of the label/variable in
// the Grafana UI.
func makeBy(labels ...string) (aggregateExprSuffix string, legendPrefix string) {
	if len(labels) == 0 {
		return "", ""
	}

	placeholders := make([]string, 0, len(labels))
	for _, label := range labels {
		placeholders = append(placeholders, fmt.Sprintf("%[1]s={{%[1]s}}", label))
	}

	aggregateExprSuffix = fmt.Sprintf(" by (%s)", strings.Join(labels, ","))
	legendPrefix = fmt.Sprintf("%s ", strings.Join(placeholders, ","))

	return aggregateExprSuffix, legendPrefix
}
