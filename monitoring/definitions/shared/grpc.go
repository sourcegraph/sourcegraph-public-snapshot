package shared

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type GRPCServerMetricsOptions struct {
	// HumanServiceName is the short, lowercase, snake_case, human-readable name of the grpc service that we're gathering metrics for.
	//
	// Example: "gitserver"
	HumanServiceName string

	// RawGRPCServiceName is the full, dot-separated, code-generated gRPC service name that we're gathering metrics for.
	//
	// Example: "gitserver.v1.GitserverService"
	RawGRPCServiceName string

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

	// MessageSizeNamespace is the Prometheus namespace that total message size metrics will be placed under.
	//
	// Example: "src"
	MessageSizeNamespace string
}

// NewGRPCServerMetricsGroup creates a group containing statistics (request rate, request duration, etc.) for the grpc service
// specified in the given opts.
func NewGRPCServerMetricsGroup(opts GRPCServerMetricsOptions, owner monitoring.ObservableOwner) monitoring.Group {
	opts.HumanServiceName = strcase.ToSnake(opts.HumanServiceName)

	namespaced := func(base, namespace string) string {
		if namespace != "" {
			return namespace + "_" + base
		}

		return base
	}

	metric := func(base string, labelFilters ...string) string {
		metric := base

		serverLabelFilter := fmt.Sprintf("grpc_service=~%q", opts.RawGRPCServiceName)
		labelFilters = append(labelFilters, serverLabelFilter)

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

	titleCaser := cases.Title(language.English)

	return monitoring.Group{
		Title:  fmt.Sprintf("%s GRPC server metrics", titleCaser.String(strings.ReplaceAll(opts.HumanServiceName, "_", " "))),
		Hidden: true,
		Rows: []monitoring.Row{

			// Track QPS
			{
				monitoring.Observable{
					Name:        fmt.Sprintf("%s_grpc_request_rate_all_methods", opts.HumanServiceName),
					Description: "request rate across all methods over 2m",
					Query:       fmt.Sprintf(`sum(rate(%s[2m]))`, metric("grpc_server_started_total", instanceLabelFilter)),
					Panel: monitoring.Panel().
						Unit(monitoring.RequestsPerSecond).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "The number of gRPC requests received per second across all methods, aggregated across all instances.",
				},
				monitoring.Observable{
					Name:        fmt.Sprintf("%s_grpc_request_rate_per_method", opts.HumanServiceName),
					Description: "request rate per-method over 2m",
					Query:       fmt.Sprintf("sum(rate(%s[2m])) by (grpc_method)", metric("grpc_server_started_total", methodLabelFilter, instanceLabelFilter)),
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
					Name:        fmt.Sprintf("%s_error_percentage_all_methods", opts.HumanServiceName),
					Description: "error percentage across all methods over 2m",
					Query: percentageQuery(
						fmt.Sprintf("sum(rate(%s[2m]))", metric("grpc_server_handled_total", failingCodeFilter, instanceLabelFilter)),
						fmt.Sprintf("sum(rate(%s[2m]))", metric("grpc_server_handled_total", instanceLabelFilter)),
					),
					Panel: monitoring.Panel().
						Unit(monitoring.Percentage).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "The percentage of gRPC requests that fail across all methods, aggregated across all instances.",
				},

				monitoring.Observable{
					Name:        fmt.Sprintf("%s_grpc_error_percentage_per_method", opts.HumanServiceName),
					Description: "error percentage per-method over 2m",
					Query: percentageQuery(
						fmt.Sprintf("sum(rate(%s[2m])) by (grpc_method)", metric("grpc_server_handled_total", methodLabelFilter, failingCodeFilter, instanceLabelFilter)),
						fmt.Sprintf("sum(rate(%s[2m])) by (grpc_method)", metric("grpc_server_handled_total", methodLabelFilter, instanceLabelFilter)),
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
					Name:        fmt.Sprintf("%s_p99_response_time_per_method", opts.HumanServiceName),
					Description: "99th percentile response time per method over 2m",
					Query:       fmt.Sprintf("histogram_quantile(0.99, sum by (le, name, grpc_method)(rate(%s[2m])))", metric("grpc_server_handling_seconds_bucket", methodLabelFilter, instanceLabelFilter)),
					Panel: monitoring.Panel().LegendFormat("{{grpc_method}}").
						Unit(monitoring.Seconds).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "The 99th percentile response time per method, aggregated across all instances.",
				},

				monitoring.Observable{
					Name:        fmt.Sprintf("%s_p90_response_time_per_method", opts.HumanServiceName),
					Description: "90th percentile response time per method over 2m",
					Query:       fmt.Sprintf("histogram_quantile(0.90, sum by (le, name, grpc_method)(rate(%s[2m])))", metric("grpc_server_handling_seconds_bucket", methodLabelFilter, instanceLabelFilter)),
					Panel: monitoring.Panel().LegendFormat("{{grpc_method}}").
						Unit(monitoring.Seconds).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "The 90th percentile response time per method, aggregated across all instances.",
				},

				monitoring.Observable{
					Name:        fmt.Sprintf("%s_p75_response_time_per_method", opts.HumanServiceName),
					Description: "75th percentile response time per method over 2m",
					Query:       fmt.Sprintf("histogram_quantile(0.75, sum by (le, name, grpc_method)(rate(%s[2m])))", metric("grpc_server_handling_seconds_bucket", methodLabelFilter, instanceLabelFilter)),
					Panel: monitoring.Panel().LegendFormat("{{grpc_method}}").
						Unit(monitoring.Seconds).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "The 75th percentile response time per method, aggregated across all instances.",
				},
			},

			// Track total response size per method

			{
				monitoring.Observable{
					Name:        fmt.Sprintf("%s_p99_9_response_size_per_method", opts.HumanServiceName),
					Description: "99.9th percentile total response size per method over 2m",
					Query:       fmt.Sprintf("histogram_quantile(0.999, sum by (le, name, grpc_method)(rate(%s[2m])))", metric(namespaced("grpc_server_sent_bytes_per_rpc_bucket", opts.MessageSizeNamespace), methodLabelFilter, instanceLabelFilter)),
					Panel: monitoring.Panel().LegendFormat("{{grpc_method}}").
						Unit(monitoring.Bytes).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "The 99.9th percentile total per-RPC response size per method, aggregated across all instances.",
				},
				monitoring.Observable{
					Name:        fmt.Sprintf("%s_p90_response_size_per_method", opts.HumanServiceName),
					Description: "90th percentile total response size per method over 2m",
					Query:       fmt.Sprintf("histogram_quantile(0.90, sum by (le, name, grpc_method)(rate(%s[2m])))", metric(namespaced("grpc_server_sent_bytes_per_rpc_bucket", opts.MessageSizeNamespace), methodLabelFilter, instanceLabelFilter)),
					Panel: monitoring.Panel().LegendFormat("{{grpc_method}}").
						Unit(monitoring.Bytes).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "The 90th percentile total per-RPC response size per method, aggregated across all instances.",
				},
				monitoring.Observable{
					Name:        fmt.Sprintf("%s_p75_response_size_per_method", opts.HumanServiceName),
					Description: "75th percentile total response size per method over 2m",
					Query:       fmt.Sprintf("histogram_quantile(0.75, sum by (le, name, grpc_method)(rate(%s[2m])))", metric(namespaced("grpc_server_sent_bytes_per_rpc_bucket", opts.MessageSizeNamespace), methodLabelFilter, instanceLabelFilter)),
					Panel: monitoring.Panel().LegendFormat("{{grpc_method}}").
						Unit(monitoring.Bytes).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "The 75th percentile total per-RPC response size per method, aggregated across all instances.",
				},
			},

			// Track individual message size per method

			{
				monitoring.Observable{
					Name:        fmt.Sprintf("%s_p99_9_invididual_sent_message_size_per_method", opts.HumanServiceName),
					Description: "99.9th percentile individual sent message size per method over 2m",
					Query:       fmt.Sprintf("histogram_quantile(0.999, sum by (le, name, grpc_method)(rate(%s[2m])))", metric(namespaced("grpc_server_sent_individual_message_size_bytes_per_rpc_bucket", opts.MessageSizeNamespace), methodLabelFilter, instanceLabelFilter)),
					Panel: monitoring.Panel().LegendFormat("{{grpc_method}}").
						Unit(monitoring.Bytes).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "The 99.9th percentile size of every individual protocol buffer size sent by the service per method, aggregated across all instances.",
				},
				monitoring.Observable{
					Name:        fmt.Sprintf("%s_p90_invididual_sent_message_size_per_method", opts.HumanServiceName),
					Description: "90th percentile individual sent message size per method over 2m",
					Query:       fmt.Sprintf("histogram_quantile(0.90, sum by (le, name, grpc_method)(rate(%s[2m])))", metric(namespaced("grpc_server_sent_individual_message_size_bytes_per_rpc_bucket", opts.MessageSizeNamespace), methodLabelFilter, instanceLabelFilter)),
					Panel: monitoring.Panel().LegendFormat("{{grpc_method}}").
						Unit(monitoring.Bytes).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "The 90th percentile size of every individual protocol buffer size sent by the service per method, aggregated across all instances.",
				},
				monitoring.Observable{
					Name:        fmt.Sprintf("%s_p75_invididual_sent_message_size_per_method", opts.HumanServiceName),
					Description: "75th percentile individual sent message size per method over 2m",
					Query:       fmt.Sprintf("histogram_quantile(0.75, sum by (le, name, grpc_method)(rate(%s[2m])))", metric(namespaced("grpc_server_sent_individual_message_size_bytes_per_rpc_bucket", opts.MessageSizeNamespace), methodLabelFilter, instanceLabelFilter)),
					Panel: monitoring.Panel().LegendFormat("{{grpc_method}}").
						Unit(monitoring.Bytes).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: "The 75th percentile size of every individual protocol buffer size sent by the service per method, aggregated across all instances.",
				},
			},

			// Track average response stream size per-method
			{
				monitoring.Observable{
					Name:        fmt.Sprintf("%s_grpc_response_stream_message_count_per_method", opts.HumanServiceName),
					Description: "average streaming response message count per-method over 2m",
					Query: fmt.Sprintf(`((%s)/(%s))`,
						fmt.Sprintf("sum(rate(%s[2m])) by (grpc_method)", metric("grpc_server_msg_sent_total", grpcStreamTypeFilter, instanceLabelFilter)),
						fmt.Sprintf("sum(rate(%s[2m])) by (grpc_method)", metric("grpc_server_started_total", grpcStreamTypeFilter, instanceLabelFilter)),
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
					Name:        fmt.Sprintf("%s_grpc_all_codes_per_method", opts.HumanServiceName),
					Description: "response codes rate per-method over 2m",
					Query:       fmt.Sprintf(`sum(rate(%s[2m])) by (grpc_method, grpc_code)`, metric("grpc_server_handled_total", methodLabelFilter, instanceLabelFilter)),
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

type GRPCInternalErrorMetricsOptions struct {
	// HumanServiceName is the short, lowercase, snake_case, human-readable name of the grpc service that we're gathering metrics for.
	//
	// Example: "gitserver"
	HumanServiceName string

	// RawGRPCServiceName is the full, dot-separated, code-generated gRPC service name that we're gathering metrics for.
	//
	// Example: "gitserver.v1.GitserverService"
	RawGRPCServiceName string

	// MethodFilterRegex is the PromQL regex that's used to filter the
	// GRPC server metrics to only those emitted by the method(s) that were interested in.
	//
	// Example: (Search | Exec)
	MethodFilterRegex string

	// Namespace is the Prometheus metrics namespace for metrics emitted by this service.
	Namespace string
}

// NewGRPCInternalErrorMetricsGroup creates a Group containing metrics that track "internal" gRPC errors.
func NewGRPCInternalErrorMetricsGroup(opts GRPCInternalErrorMetricsOptions, owner monitoring.ObservableOwner) monitoring.Group {
	opts.HumanServiceName = strcase.ToSnake(opts.HumanServiceName)

	metric := func(base string, labelFilters ...string) string {
		m := base

		if opts.Namespace != "" {
			m = fmt.Sprintf("%s_%s", opts.Namespace, m)
		}

		if len(labelFilters) > 0 {
			m = fmt.Sprintf("%s{%s}", m, strings.Join(labelFilters, ","))
		}

		return m
	}

	sum := func(metric, duration string, groupByLabels ...string) string {
		base := fmt.Sprintf("sum(rate(%s[%s]))", metric, duration)

		if len(groupByLabels) > 0 {
			base = fmt.Sprintf("%s by (%s)", base, strings.Join(groupByLabels, ", "))
		}

		return fmt.Sprintf("(%s)", base)
	}

	methodLabelFilter := fmt.Sprintf(`grpc_method=~"%s"`, opts.MethodFilterRegex)
	serviceLabelFilter := fmt.Sprintf(`grpc_service=~"%s"`, opts.RawGRPCServiceName)
	isInternalErrorFilter := fmt.Sprintf(`is_internal_error="%s"`, "true")
	failingCodeFilter := fmt.Sprintf("grpc_code!=%q", "OK")

	percentageQuery := func(numerator, denominator string) string {
		ratio := fmt.Sprintf("((%s) / (%s))", numerator, denominator)
		return fmt.Sprintf("(100.0 * (%s))", ratio)
	}

	sharedInternalErrorNote := func() string {
		first := strings.Join([]string{
			"**Note**: Internal errors are ones that appear to originate from the https://github.com/grpc/grpc-go library itself, rather than from any user-written application code.",
			fmt.Sprintf("These errors can be caused by a variety of issues, and can originate from either the code-generated %q gRPC client or gRPC server.", opts.HumanServiceName),
			"These errors might be solvable by adjusting the gRPC configuration, or they might indicate a bug from Sourcegraph's use of gRPC.",
		}, " ")

		second := "When debugging, knowing that a particular error comes from the grpc-go library itself (an 'internal error') as opposed to 'normal' application code can be helpful when trying to fix it."

		third := strings.Join([]string{
			"**Note**: Internal errors are detected via a very coarse heuristic (seeing if the error starts with 'grpc:', etc.).",
			"Because of this, it's possible that some gRPC-specific issues might not be categorized as internal errors.",
		}, " ")

		return fmt.Sprintf("%s\n\n%s\n\n%s", first, second, third)
	}()

	titleCaser := cases.Title(language.English)

	return monitoring.Group{
		Title:  fmt.Sprintf("%s GRPC %q metrics", titleCaser.String(strings.ReplaceAll(opts.HumanServiceName, "_", " ")), "internal error"),
		Hidden: true,
		Rows: []monitoring.Row{
			{
				monitoring.Observable{
					Name:        fmt.Sprintf("%s_grpc_clients_error_percentage_all_methods", opts.HumanServiceName),
					Description: "client baseline error percentage across all methods over 2m",
					Query: percentageQuery(
						sum(metric("grpc_method_status", serviceLabelFilter, failingCodeFilter), "2m"),
						sum(metric("grpc_method_status", serviceLabelFilter), "2m"),
					),
					Panel: monitoring.Panel().
						Unit(monitoring.Percentage).
						With(monitoring.PanelOptions.LegendOnRight()).
						With(monitoring.PanelOptions.ZeroIfNoData()),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: fmt.Sprintf("The percentage of gRPC requests that fail across all methods (regardless of whether or not there was an internal error), aggregated across all %q clients.", opts.HumanServiceName),
				},

				monitoring.Observable{
					Name:        fmt.Sprintf("%s_grpc_clients_error_percentage_per_method", opts.HumanServiceName),
					Description: "client baseline error percentage per-method over 2m",
					Query: percentageQuery(
						sum(metric("grpc_method_status", serviceLabelFilter, methodLabelFilter, failingCodeFilter), "2m", "grpc_method"),
						sum(metric("grpc_method_status", serviceLabelFilter, methodLabelFilter), "2m", "grpc_method"),
					),

					Panel: monitoring.Panel().LegendFormat("{{grpc_method}}").
						Unit(monitoring.Percentage).
						With(monitoring.PanelOptions.LegendOnRight()).
						With(monitoring.PanelOptions.ZeroIfNoData("grpc_method")),

					Owner:          owner,
					NoAlert:        true,
					Interpretation: fmt.Sprintf("The percentage of gRPC requests that fail per method (regardless of whether or not there was an internal error), aggregated across all %q clients.", opts.HumanServiceName),
				},

				monitoring.Observable{
					Name:        fmt.Sprintf("%s_grpc_clients_all_codes_per_method", opts.HumanServiceName),
					Description: "client baseline response codes rate per-method over 2m",
					Query:       sum(metric("grpc_method_status", serviceLabelFilter, methodLabelFilter), "2m", "grpc_method", "grpc_code"),
					Panel: monitoring.Panel().LegendFormat("{{grpc_method}}: {{grpc_code}}").
						Unit(monitoring.RequestsPerSecond).
						With(monitoring.PanelOptions.LegendOnRight()).
						With(monitoring.PanelOptions.ZeroIfNoData("grpc_method", "grpc_code")),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: fmt.Sprintf("The rate of all generated gRPC response codes per method (regardless of whether or not there was an internal error), aggregated across all %q clients.", opts.HumanServiceName),
				},
			},
			{
				monitoring.Observable{
					Name:        fmt.Sprintf("%s_grpc_clients_internal_error_percentage_all_methods", opts.HumanServiceName),
					Description: "client-observed gRPC internal error percentage across all methods over 2m",
					Query: percentageQuery(
						sum(metric("grpc_method_status", serviceLabelFilter, failingCodeFilter, isInternalErrorFilter), "2m"),
						sum(metric("grpc_method_status", serviceLabelFilter), "2m"),
					),
					Panel: monitoring.Panel().
						Unit(monitoring.Percentage).
						With(monitoring.PanelOptions.LegendOnRight()).
						With(monitoring.PanelOptions.ZeroIfNoData()),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: fmt.Sprintf("The percentage of gRPC requests that appear to fail due to gRPC internal errors across all methods, aggregated across all %q clients.\n\n%s", opts.HumanServiceName, sharedInternalErrorNote),
				},

				monitoring.Observable{
					Name:        fmt.Sprintf("%s_grpc_clients_internal_error_percentage_per_method", opts.HumanServiceName),
					Description: "client-observed gRPC internal error percentage per-method over 2m",
					Query: percentageQuery(
						sum(metric("grpc_method_status", serviceLabelFilter, methodLabelFilter, failingCodeFilter, isInternalErrorFilter), "2m", "grpc_method"),
						sum(metric("grpc_method_status", serviceLabelFilter, methodLabelFilter), "2m", "grpc_method"),
					),

					Panel: monitoring.Panel().LegendFormat("{{grpc_method}}").
						Unit(monitoring.Percentage).
						With(monitoring.PanelOptions.LegendOnRight()).
						With(monitoring.PanelOptions.ZeroIfNoData("grpc_method")),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: fmt.Sprintf("The percentage of gRPC requests that appear to fail to due to gRPC internal errors per method, aggregated across all %q clients.\n\n%s", opts.HumanServiceName, sharedInternalErrorNote),
				},

				monitoring.Observable{
					Name:        fmt.Sprintf("%s_grpc_clients_internal_error_all_codes_per_method", opts.HumanServiceName),
					Description: "client-observed gRPC internal error response code rate per-method over 2m",
					Query:       sum(metric("grpc_method_status", serviceLabelFilter, isInternalErrorFilter, methodLabelFilter), "2m", "grpc_method", "grpc_code"),
					Panel: monitoring.Panel().LegendFormat("{{grpc_method}}: {{grpc_code}}").
						Unit(monitoring.RequestsPerSecond).
						With(monitoring.PanelOptions.LegendOnRight()).
						With(monitoring.PanelOptions.ZeroIfNoData("grpc_method", "grpc_code")),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: fmt.Sprintf("The rate of gRPC internal-error response codes per method, aggregated across all %q clients.\n\n%s", opts.HumanServiceName, sharedInternalErrorNote),
				},
			},
		},
	}
}

// GRPCMethodVariable creates a container variable that contains all the gRPC methods
// exposed by the given service.
//
// humanServiceName is the short, lowercase, snake_case,
// human-readable name of the grpc service that we're gathering metrics for.
//
// Example: "gitserver"
//
// services is a dot-separated, code-generated gRPC service name that we're gathering metrics for
// (e.g. "gitserver.v1.GitserverService").
func GRPCMethodVariable(humanServiceName string, service string) monitoring.ContainerVariable {
	humanServiceName = strcase.ToSnake(humanServiceName)

	query := fmt.Sprintf("grpc_server_started_total{grpc_service=%q}", service)

	titleCaser := cases.Title(language.English)

	return monitoring.ContainerVariable{
		Label: fmt.Sprintf("%s RPC Method", titleCaser.String(strings.ReplaceAll(humanServiceName, "_", " "))),
		Name:  fmt.Sprintf("%s_method", humanServiceName),
		OptionsLabelValues: monitoring.ContainerVariableOptionsLabelValues{
			Query:         query,
			LabelName:     "grpc_method",
			ExampleOption: "Exec",
		},

		Multi: true,
	}
}
