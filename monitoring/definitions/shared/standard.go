package shared

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

// Standard exports available standard observable constructors.
var Standard standardConstructor

// standardConstructor provides `Standard` implementations.
type standardConstructor struct{}

// Count creates an observable from the given options backed by the counter specifying
// the number of operations. The legend name supplied to the outermost function will be
// used as the panel's dataset legend. Note that the legend is also supplemented by label
// values if By is also assigned.
//
// Requires a counter of the format `src_{options.MetricNameRoot}_total`
func (standardConstructor) Count(legend string) observableConstructor {
	if legend != "" {
		legend = " " + legend
	}

	return func(options ObservableConstructorOptions) sharedObservable {
		return func(containerName string, owner monitoring.ObservableOwner) Observable {
			filters := makeFilters(containerName, options.Filters...)
			by, legendPrefix := makeBy(options.By...)

			return Observable{
				Name:        fmt.Sprintf("%s_total", options.MetricNameRoot),
				Description: fmt.Sprintf("%s%s every 5m", options.MetricDescriptionRoot, legend),
				Query:       fmt.Sprintf(`sum%s(increase(src_%s_total{%s}[5m]))`, by, options.MetricNameRoot, filters),
				Panel:       monitoring.Panel().LegendFormat(fmt.Sprintf("%s%s", legendPrefix, legend)),
				Owner:       owner,
			}
		}
	}
}

// Duration creates an observable from the given options backed by the histogram specifying
// the duration of operations. The legend name supplied to the outermost function will be
// used as the panel's dataset legend. Note that the legend is also supplemented by label
// values if By is also assigned.
//
// Requires a histogram of the format `src_{options.MetricNameRoot}_duration_seconds_bucket`
func (standardConstructor) Duration(legend string) observableConstructor {
	if legend != "" {
		legend = " " + legend
	}

	return func(options ObservableConstructorOptions) sharedObservable {
		return func(containerName string, owner monitoring.ObservableOwner) Observable {
			filters := makeFilters(containerName, options.Filters...)
			by, _ := makeBy(append([]string{"le"}, options.By...)...)
			_, legendPrefix := makeBy(options.By...)

			return Observable{
				Name:        fmt.Sprintf("%s_99th_percentile_duration", options.MetricNameRoot),
				Description: fmt.Sprintf("99th percentile successful %s%s duration over 5m", options.MetricDescriptionRoot, legend),
				Query:       fmt.Sprintf(`histogram_quantile(0.99, sum %s(rate(src_%s_duration_seconds_bucket{%s}[5m])))`, by, options.MetricNameRoot, filters),
				Panel:       monitoring.Panel().LegendFormat(fmt.Sprintf("%s%s", legendPrefix, legend)).Unit(monitoring.Seconds),
				Owner:       owner,
			}
		}
	}
}

// Errors creates an observable from the given options backed by the counter specifying
// the number of operations that resulted in an error. The legend name supplied to the
// outermost function will be used as the panel's dataset legend. Note that the legend
// is also supplemented by label values if By is also assigned.
//
// Requires a counter of the format `src_{options.MetricNameRoot}_errors_total`
func (standardConstructor) Errors(legend string) observableConstructor {
	if legend != "" {
		legend = " " + legend
	}

	return func(options ObservableConstructorOptions) sharedObservable {
		return func(containerName string, owner monitoring.ObservableOwner) Observable {
			filters := makeFilters(containerName, options.Filters...)
			by, legendPrefix := makeBy(options.By...)

			return Observable{
				Name:        fmt.Sprintf("%s_errors_total", options.MetricNameRoot),
				Description: fmt.Sprintf("%s%s errors every 5m", options.MetricDescriptionRoot, legend),
				Query:       fmt.Sprintf(`sum%s(increase(src_%s_errors_total{%s}[5m]))`, by, options.MetricNameRoot, filters),
				Panel:       monitoring.Panel().LegendFormat(fmt.Sprintf("%s%s errors", legendPrefix, legend)).With(monitoring.PanelOptions.ZeroIfNoData()),
				Owner:       owner,
			}
		}
	}
}

// ErrorRate creates an observable from the given options backed by the counters specifying
// the number of operations that resulted in success and error, respectively. The legend name
// supplied to the outermost function will be used as the panel's dataset legend. Note that
// the legend is also supplemented by label values if By is also assigned.
//
// Requires a:
//   - counter of the format `src_{options.MetricNameRoot}_total`
//   - counter of the format `src_{options.MetricNameRoot}_errors_total`
func (standardConstructor) ErrorRate(legend string) observableConstructor {
	if legend != "" {
		legend = " " + legend
	}

	return func(options ObservableConstructorOptions) sharedObservable {
		return func(containerName string, owner monitoring.ObservableOwner) Observable {
			filters := makeFilters(containerName, options.Filters...)
			by, legendPrefix := makeBy(options.By...)

			return Observable{
				Name:        fmt.Sprintf("%s_error_rate", options.MetricNameRoot),
				Description: fmt.Sprintf("%s%s error rate over 5m", options.MetricDescriptionRoot, legend),
				Query:       fmt.Sprintf(`sum%[1]s(increase(src_%[2]s_errors_total{%[3]s}[5m])) / (sum%[1]s(increase(src_%[2]s_total{%[3]s}[5m])) + sum%[1]s(increase(src_%[2]s_errors_total{%[3]s}[5m]))) * 100`, by, options.MetricNameRoot, filters),
				Panel:       monitoring.Panel().LegendFormat(fmt.Sprintf("%s%s error rate", legendPrefix, legend)).With(monitoring.PanelOptions.ZeroIfNoData()).Unit(monitoring.Percentage),
				Owner:       owner,
			}
		}
	}
}
