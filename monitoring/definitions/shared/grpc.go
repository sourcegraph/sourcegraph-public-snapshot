package shared

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

type GRPCServerMetricsOptions struct {
	// ServiceName is the short, lowercase, human-readable name of the grpc service that we're gathering metrics for.
	//
	// Example: "gitserver"
	ServiceName string

	// MetricNamespace is the (optional) namespace that the service uses to prefix its grpc server metrics.
	//
	// Example: "gitserver"
	MetricNamespace string

	// MethodFilterRegex is the PromQL regex that's used to filter the
	// GRPC server metrics to only those emitted by the method(s) that were interested in.
	//
	// Example: (Search | Exec)
	MethodFilterRegex string

	// InstanceFilterRegex is the PromQL regex that's used to filter the
	// GRPC server metrics to only those emitted by the instance(s) that were interested in.
	//
	// Example: (gitserver-0 | gitserver-1)
	InstanceFilterRegex string
}

// NewGRPCServerMetricsGroup creates a group containing statistics (request rate, request duration, etc.) for the grpc service
// specified in the given opts.
func NewGRPCServerMetricsGroup(opts GRPCServerMetricsOptions, owner monitoring.ObservableOwner) monitoring.Group {
	metric := func(base string, labelFilters ...string) string {
		metric := base

		if opts.MetricNamespace != "" {
			metric = opts.MetricNamespace + "_" + base
		}

		if len(labelFilters) > 0 {
			metric = fmt.Sprintf("%s{%s}", metric, strings.Join(labelFilters, ","))
		}

		return metric
	}

	methodLabelFilter := fmt.Sprintf("grpc_method=~`%s`", opts.MethodFilterRegex)
	instanceLabelFilter := fmt.Sprintf("instance=~`%s`", opts.InstanceFilterRegex)
	failingCodeFilter := fmt.Sprintf("grpc_code!=%q", "OK")
	grpcStreamTypeFilter := fmt.Sprintf("grpc_type=%q", "server_stream")

	percentageQuery := func(numerator, denominator string) string {
		return fmt.Sprintf("(100.0 * ( (%s) / (%s) ))", numerator, denominator)
	}

	return monitoring.Group{
		Title:  "GRPC server metrics",
		Hidden: true,
		Rows: []monitoring.Row{

			// Track QPS
			{
				monitoring.Observable{
					Name:        fmt.Sprintf("%s_grpc_request_rate_all_methods", opts.ServiceName),
					Description: "request rate across all methods over 1m",
					Query:       fmt.Sprintf(`sum(rate(%s[1m]))`, metric("grpc_server_started_total", instanceLabelFilter)),
					Panel: monitoring.Panel().
						Unit(monitoring.RequestsPerSecond).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "The number of gRPC requests received per second across all methods, aggregated across all instances.",
				},
				monitoring.Observable{
					Name:        fmt.Sprintf("%s_grpc_request_rate_per_method", opts.ServiceName),
					Description: "request rate per-method over 1m",
					Query:       fmt.Sprintf("sum(rate(%s[1m])) by (grpc_method)", metric("grpc_server_started_total", methodLabelFilter, instanceLabelFilter)),
					Panel: monitoring.Panel().LegendFormat("{{grpc_method}}").
						Unit(monitoring.RequestsPerSecond).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "The number of gRPC requests received per second broken out per method, aggregated across all instances.",
				},
			},

			// Track error percentage
			{
				monitoring.Observable{
					Name:        fmt.Sprintf("%s_error_percentage_all_methods", opts.ServiceName),
					Description: "error percentage across all methods over 1m",
					Query: percentageQuery(
						fmt.Sprintf("sum(rate(%s[1m]))", metric("grpc_server_handled_total", failingCodeFilter, instanceLabelFilter)),
						fmt.Sprintf("sum(rate(%s[1m]))", metric("grpc_server_handled_total", instanceLabelFilter)),
					),
					Panel: monitoring.Panel().
						Unit(monitoring.Percentage).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "The percentage of gRPC requests that fail across all methods, aggregated across all instances.",
				},

				monitoring.Observable{
					Name:        fmt.Sprintf("%s_grpc_error_percentage_per_method", opts.ServiceName),
					Description: "error percentage per-method over 1m",
					Query: percentageQuery(
						fmt.Sprintf("sum(rate(%s[1m])) by (grpc_method)", metric("grpc_server_handled_total", methodLabelFilter, failingCodeFilter, instanceLabelFilter)),
						fmt.Sprintf("sum(rate(%s[1m])) by (grpc_method)", metric("grpc_server_handled_total", methodLabelFilter, instanceLabelFilter)),
					),

					Panel: monitoring.Panel().LegendFormat("{{grpc_method}}").
						Unit(monitoring.Percentage).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "The percentage of gRPC requests that fail per method, aggregated across all instances.",
				},
			},

			// Track response time per method
			{

				monitoring.Observable{
					Name:        fmt.Sprintf("%s_p99_response_time_per_method", opts.ServiceName),
					Description: "99th percentile response time per method over 1m",
					Query:       fmt.Sprintf("histogram_quantile(0.99, sum by (le, name, grpc_method)(rate(%s[1m])))", metric("grpc_server_handling_seconds_bucket", methodLabelFilter, instanceLabelFilter)),
					Panel: monitoring.Panel().LegendFormat("{{grpc_method}}").
						Unit(monitoring.Seconds).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "The 99th percentile response time per method, aggregated across all instances.",
				},

				monitoring.Observable{
					Name:        fmt.Sprintf("%s_p90_response_time_per_method", opts.ServiceName),
					Description: "90th percentile response time per method over 1m",
					Query:       fmt.Sprintf("histogram_quantile(0.90, sum by (le, name, grpc_method)(rate(%s[1m])))", metric("grpc_server_handling_seconds_bucket", methodLabelFilter, instanceLabelFilter)),
					Panel: monitoring.Panel().LegendFormat("{{grpc_method}}").
						Unit(monitoring.Seconds).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "The 90th percentile response time per method, aggregated across all instances.",
				},

				monitoring.Observable{
					Name:        fmt.Sprintf("%s_p75_response_time_per_method", opts.ServiceName),
					Description: "75th percentile response time per method over 1m",
					Query:       fmt.Sprintf("histogram_quantile(0.75, sum by (le, name, grpc_method)(rate(%s[1m])))", metric("grpc_server_handling_seconds_bucket", methodLabelFilter, instanceLabelFilter)),
					Panel: monitoring.Panel().LegendFormat("{{grpc_method}}").
						Unit(monitoring.Seconds).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "The 75th percentile response time per method, aggregated across all instances.",
				},
			},

			// Track average response stream size per-method
			{
				monitoring.Observable{
					Name:        fmt.Sprintf("%s_grpc_response_stream_message_count_per_method", opts.ServiceName),
					Description: "average streaming response message count per-method over 1m",
					Query: fmt.Sprintf(`((%s)/(%s))`,
						fmt.Sprintf("sum(rate(%s[1m])) by (grpc_method)", metric("grpc_server_msg_sent_total", grpcStreamTypeFilter, instanceLabelFilter)),
						fmt.Sprintf("sum(rate(%s[1m])) by (grpc_method)", metric("grpc_server_started_total", grpcStreamTypeFilter, instanceLabelFilter)),
					),
					Panel: monitoring.Panel().LegendFormat("{{grpc_method}}").
						Unit(monitoring.Number).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "The average number of response messages sent during a streaming RPC method, broken out per method, aggregated across all instances.",
				},
			},

			// Track rate across all gRPC response codes
			{
				monitoring.Observable{
					Name:        fmt.Sprintf("%s_grpc_all_codes_per_method", opts.ServiceName),
					Description: "response codes rate per-method over 1m",
					Query:       fmt.Sprintf(`sum(rate(%s[1m])) by (grpc_method, grpc_code)`, metric("grpc_server_handled_total", methodLabelFilter, instanceLabelFilter)),
					Panel: monitoring.Panel().LegendFormat("{{grpc_method}}: {{grpc_code}}").
						Unit(monitoring.RequestsPerSecond).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "The rate of all generated gRPC response codes per method, aggregated across all instances.",
				},
			},
		},
	}

}

// GRPCMethodVariable creates a container variable that contains all the gRPC methods
// exposed by the given service.
func GRPCMethodVariable(serviceNamespace string) monitoring.ContainerVariable {
	query := "grpc_server_started_total"
	if serviceNamespace != "" {
		query = fmt.Sprintf("%s_%s", serviceNamespace, query)
	}

	return monitoring.ContainerVariable{
		Label: "RPC Method",
		Name:  "method",
		OptionsLabelValues: monitoring.ContainerVariableOptionsLabelValues{
			Query:         query,
			LabelName:     "grpc_method",
			ExampleOption: "Exec",
		},
		Multi: true,
	}
}
