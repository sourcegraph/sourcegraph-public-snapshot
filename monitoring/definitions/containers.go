package definitions

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/images"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func Containers() *monitoring.Container {
	var (
		serviceNames       = append(images.DeploySourcegraphDockerImages, "pgsql", "codeinsights", "codeintel")
		containerNameQuery = fmt.Sprintf(`name=~"(%s).*"`, strings.Join(serviceNames, "|"))
	)

	return &monitoring.Container{
		Name:                     "containers",
		Title:                    "Global Containers Resource Usage",
		Description:              "Container usage and provisioning indicators of all services.",
		NoSourcegraphDebugServer: true,
		Groups: []monitoring.Group{
			{
				Title: "Containers (not available on server)",
				// This chart is extremely noisy on k8s, so we hide it by default.
				Hidden: true,
				Rows: []monitoring.Row{
					{
						monitoring.Observable{
							Name:        "container_memory_usage",
							Description: "container memory usage of all services",
							Query:       fmt.Sprintf(`cadvisor_container_memory_usage_percentage_total{%s}`, containerNameQuery),
							NoAlert:     true,
							Panel:       monitoring.Panel().With(monitoring.PanelOptions.LegendOnRight()).LegendFormat("{{name}}").Unit(monitoring.Percentage).Interval(100).Max(100).Min(0),
							Owner:       monitoring.ObservableOwnerDevOps,
							Interpretation: `
								This value indicates the memory usage of all containers.
							`,
						},
					},
					{
						monitoring.Observable{
							Name:        "container_cpu_usage",
							Description: "container cpu usage total (1m average) across all cores by instance",
							Query:       fmt.Sprintf(`cadvisor_container_cpu_usage_percentage_total{%s}`, containerNameQuery),
							NoAlert:     true,
							Panel:       monitoring.Panel().With(monitoring.PanelOptions.LegendOnRight()).LegendFormat("{{name}}").Unit(monitoring.Percentage).Interval(100).Max(100).Min(0),
							Owner:       monitoring.ObservableOwnerDevOps,
							Interpretation: `
								This value indicates the CPU usage of all containers.
							`,
						},
					},
				},
			},
			{
				Title:  "Containers: Provisioning Indicators (not available on server)",
				Hidden: false,
				Rows: []monitoring.Row{
					{
						monitoring.Observable{
							Name:        "container_memory_usage_provisioning",
							Description: "container memory usage (5m maximum) of services that exceed 80% memory limit",
							Query:       fmt.Sprintf(`max_over_time(cadvisor_container_memory_usage_percentage_total{%s}[5m]) >= 80`, containerNameQuery),
							NoAlert:     true,
							Panel:       monitoring.Panel().With(monitoring.PanelOptions.LegendOnRight()).LegendFormat("{{name}}").Unit(monitoring.Percentage).Interval(100).Max(100).Min(0),
							Owner:       monitoring.ObservableOwnerDevOps,
							Interpretation: `
								Containers that exceed 80% memory limit. The value indicates potential underprovisioned resources.
							`,
						},
					},
					{
						monitoring.Observable{
							Name:        "container_cpu_usage_provisioning",
							Description: "container cpu usage total (5m maximum) across all cores of services that exceed 80% cpu limit",
							Query:       fmt.Sprintf(`max_over_time(cadvisor_container_cpu_usage_percentage_total{%s}[5m]) >= 80`, containerNameQuery),
							NoAlert:     true,
							Panel:       monitoring.Panel().With(monitoring.PanelOptions.LegendOnRight()).LegendFormat("{{name}}").Unit(monitoring.Percentage).Interval(100).Max(100).Min(0),
							Owner:       monitoring.ObservableOwnerDevOps,
							Interpretation: `
								Containers that exceed 80% CPU limit. The value indicates potential underprovisioned resources.
							`,
						},
					},
					{
						monitoring.Observable{
							Name:        "container_oomkill_events_total",
							Description: "container OOMKILL events total",
							Query:       fmt.Sprintf(`max by (name) (container_oom_events_total{%s}) >= 1`, containerNameQuery),
							NoAlert:     true,
							Panel:       monitoring.Panel().With(monitoring.PanelOptions.LegendOnRight()).LegendFormat("{{name}}"),
							Owner:       monitoring.ObservableOwnerDevOps,
							Interpretation: `
								This value indicates the total number of times the container main process or child processes were terminated by OOM killer.
								When it occurs frequently, it is an indicator of underprovisioning.
							`,
						},
					},
					{
						monitoring.Observable{
							Name:        "container_missing",
							Description: "container missing",
							// inspired by https://awesome-prometheus-alerts.grep.to/rules#docker-containers
							Query:   fmt.Sprintf(`count by(name) ((time() - container_last_seen{%s}) > 60)`, containerNameQuery),
							NoAlert: true,
							Panel:   monitoring.Panel().With(monitoring.PanelOptions.LegendOnRight()).LegendFormat("{{name}}"),
							Owner:   monitoring.ObservableOwnerDevOps,
							Interpretation: `
								This value is the number of times a container has not been seen for more than one minute. If you observe this
								value change independent of deployment events (such as an upgrade), it could indicate pods are being OOM killed or terminated for some other reasons.
							`,
						},
					},
				},
			},
		},
	}
}
