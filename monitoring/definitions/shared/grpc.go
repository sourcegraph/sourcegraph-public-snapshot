package shared

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

type GRPCServerMetricsOptions struct {
	// ServiceName is the short, lowercase, human-readable name of the grpc service that we're gathering metrics for.
	//
	// Example: "gitserver"
	ServiceName string

	// MetricNamespace is the (optional) namespace that the service uses to prefix its 'mount_point_info' metric.
	//
	// Example: "gitserver"
	MetricNamespace string

	// InstanceFilterRegex is the PromQL regex that's used to filter the
	// disk metrics to only those emitted by the instance(s) that were interested in.
	//
	// Example: (gitserver-0 | gitserver-1)
	InstanceFilterRegex string
}

func NewGRPCServerMetricsGroup(opts GRPCServerMetricsOptions, owner monitoring.ObservableOwner) monitoring.Group {
	title := "GRPC server metrics"

	grpcMetricsQuery := func(metric string) string {
		if opts.MetricNamespace != "" {
			return opts.MetricNamespace + "_" + metric
		}

		return metric
	}

	return monitoring.Group{
		Title:  title,
		Hidden: false,
		Rows: []monitoring.Row{
			{
				monitoring.Observable{
					Name:        fmt.Sprintf("%s_grpc_request_rate_all_methods_aggregate", opts.ServiceName),
					Description: "request rate across all methods over 1m (aggregate)",
					Query:       fmt.Sprintf(`sum(rate(%s[1m]))`, grpcMetricsQuery("grpc_server_started_total")),
					Panel: monitoring.Panel().
						Unit(monitoring.RequestsPerSecond).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: fmt.Sprintf("The number of gRPC requests per second across all methods, aggregated across all instances."),
				},
				monitoring.Observable{
					Name:        fmt.Sprintf("%s_grpc_request_rate_per_method_aggregate", opts.ServiceName),
					Description: "request rate per-method over 1m (aggregate)",
					Query:       fmt.Sprintf(`sum(rate(%s[1m])) by (grpc_method)`, grpcMetricsQuery("grpc_server_started_total")),
					Panel: monitoring.Panel().LegendFormat("{{grpc_method}} {{instance}}").
						Unit(monitoring.RequestsPerSecond).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: fmt.Sprintf("The number of gRPC requests per second broken out per method, aggreagated across all instances."),
				},
			},

			{
				monitoring.Observable{
					Name:        fmt.Sprintf("%s_grpc_request_rate_overall_per_instance", opts.ServiceName),
					Description: "request rate across all methods over 1m (per instance)",
					Query:       fmt.Sprintf(`sum(rate(%s[1m])) by (instance)`, grpcMetricsQuery("grpc_server_started_total")),
					Panel: monitoring.Panel().LegendFormat("{{instance}}").
						Unit(monitoring.RequestsPerSecond).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: fmt.Sprintf("The number of gRPC requests per second, across all methods broken out per method."),
				},
				monitoring.Observable{
					Name:        fmt.Sprintf("%s_grpc_request_rate_per_method_per_instance", opts.ServiceName),
					Description: "request rate per-method over 1m (per instance)",
					Query:       fmt.Sprintf(`sum(rate(%s[1m])) by (grpc_method, instance)`, grpcMetricsQuery("grpc_server_started_total")),
					Panel: monitoring.Panel().LegendFormat("{{grpc_method}} {{instance}}").
						Unit(monitoring.RequestsPerSecond).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: fmt.Sprintf("The number of gRPC requests per second, broken out per method and instance."),
				},
			},

			{
				monitoring.Observable{
					Name:        fmt.Sprintf("%s_error_rate_overall_per_instance", opts.ServiceName),
					Description: "error rate across all methods over 1m (per instance)",
					Query:       fmt.Sprintf(`sum(rate(%s{grpc_code!="Ok"}[1m]))`, grpcMetricsQuery("grpc_server_started_total")),
					Panel: monitoring.Panel().LegendFormat("{{instance}}").
						Unit(monitoring.RequestsPerSecond).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: fmt.Sprintf("The number of gRPC requests per second, across all methods broken out per method."),
				},
				monitoring.Observable{
					Name:        fmt.Sprintf("%s_grpc_error_rate_per_method_per_instance", opts.ServiceName),
					Description: "request rate per-method over 1m (per instance)",
					Query:       fmt.Sprintf(`sum(rate(%s[1m])) by (grpc_method, instance)`, grpcMetricsQuery("grpc_server_started_total")),
					Panel: monitoring.Panel().LegendFormat("{{grpc_method}} {{instance}}").
						Unit(monitoring.RequestsPerSecond).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: fmt.Sprintf("The number of gRPC requests per second, broken out per method and instance."),
				},
			},
		},
	}
}
