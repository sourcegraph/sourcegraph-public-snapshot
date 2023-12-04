package shared

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

// Container monitoring overviews - these provide short-term overviews of container
// behaviour for a service.
//
// These observables should only use cAdvisor metrics, and are thus only available on
// Kubernetes and docker-compose deployments.
//
// cAdvisor metrics reference: https://github.com/google/cadvisor/blob/master/docs/storage/prometheus.md#prometheus-container-metrics
const TitleContainerMonitoring = "Container monitoring (not available on server)"

var (
	ContainerMissing sharedObservable = func(containerName string, owner monitoring.ObservableOwner) Observable {
		return Observable{
			Name:        "container_missing",
			Description: "container missing",
			// inspired by https://awesome-prometheus-alerts.grep.to/rules#docker-containers
			Query:   fmt.Sprintf(`count by(name) ((time() - container_last_seen{%s}) > 60)`, CadvisorContainerNameMatcher(containerName)),
			NoAlert: true,
			Panel:   monitoring.Panel().LegendFormat("{{name}}"),
			Owner:   owner,
			Interpretation: strings.ReplaceAll(`
				This value is the number of times a container has not been seen for more than one minute. If you observe this
				value change independent of deployment events (such as an upgrade), it could indicate pods are being OOM killed or terminated for some other reasons.

				- **Kubernetes:**
					- Determine if the pod was OOM killed using 'kubectl describe pod {{CONTAINER_NAME}}' (look for 'OOMKilled: true') and, if so, consider increasing the memory limit in the relevant 'Deployment.yaml'.
					- Check the logs before the container restarted to see if there are 'panic:' messages or similar using 'kubectl logs -p {{CONTAINER_NAME}}'.
				- **Docker Compose:**
					- Determine if the pod was OOM killed using 'docker inspect -f \'{{json .State}}\' {{CONTAINER_NAME}}' (look for '"OOMKilled":true') and, if so, consider increasing the memory limit of the {{CONTAINER_NAME}} container in 'docker-compose.yml'.
					- Check the logs before the container restarted to see if there are 'panic:' messages or similar using 'docker logs {{CONTAINER_NAME}}' (note this will include logs from the previous and currently running container).
			`, "{{CONTAINER_NAME}}", containerName),
		}
	}

	ContainerMemoryUsage sharedObservable = func(containerName string, owner monitoring.ObservableOwner) Observable {
		return Observable{
			Name:        "container_memory_usage",
			Description: "container memory usage by instance",
			Query:       fmt.Sprintf(`cadvisor_container_memory_usage_percentage_total{%s}`, CadvisorContainerNameMatcher(containerName)),
			Warning:     monitoring.Alert().GreaterOrEqual(99),
			Panel:       monitoring.Panel().LegendFormat("{{name}}").Unit(monitoring.Percentage).Interval(100).Max(100).Min(0),
			Owner:       owner,
			NextSteps: strings.ReplaceAll(`
			- **Kubernetes:** Consider increasing memory limit in relevant 'Deployment.yaml'.
			- **Docker Compose:** Consider increasing 'memory:' of {{CONTAINER_NAME}} container in 'docker-compose.yml'.
		`, "{{CONTAINER_NAME}}", containerName),
		}
	}

	ContainerCPUUsage sharedObservable = func(containerName string, owner monitoring.ObservableOwner) Observable {
		return Observable{
			Name:        "container_cpu_usage",
			Description: "container cpu usage total (1m average) across all cores by instance",
			Query:       fmt.Sprintf(`cadvisor_container_cpu_usage_percentage_total{%s}`, CadvisorContainerNameMatcher(containerName)),
			Warning:     monitoring.Alert().GreaterOrEqual(99),
			Panel:       monitoring.Panel().LegendFormat("{{name}}").Unit(monitoring.Percentage).Interval(100).Max(100).Min(0),
			Owner:       owner,
			NextSteps: strings.ReplaceAll(`
			- **Kubernetes:** Consider increasing CPU limits in the the relevant 'Deployment.yaml'.
			- **Docker Compose:** Consider increasing 'cpus:' of the {{CONTAINER_NAME}} container in 'docker-compose.yml'.
		`, "{{CONTAINER_NAME}}", containerName),
		}
	}

	// ContainerIOUsage monitors filesystem reads and writes, and is useful for services
	// that use disk.
	ContainerIOUsage sharedObservable = func(containerName string, owner monitoring.ObservableOwner) Observable {
		return Observable{
			Name:        "fs_io_operations",
			Description: "filesystem reads and writes rate by instance over 1h",
			Query:       fmt.Sprintf(`sum by(name) (rate(container_fs_reads_total{%[1]s}[1h]) + rate(container_fs_writes_total{%[1]s}[1h]))`, CadvisorContainerNameMatcher(containerName)),
			NoAlert:     true,
			Panel:       monitoring.Panel().LegendFormat("{{name}}"),
			Owner:       owner,
			Interpretation: `
				This value indicates the number of filesystem read and write operations by containers of this service.
				When extremely high, this can indicate a resource usage problem, or can cause problems with the service itself, especially if high values or spikes correlate with {{CONTAINER_NAME}} issues.
			`,
		}
	}
)

type ContainerMonitoringGroupOptions struct {
	// ContainerMissing transforms the default observable used to construct the container missing panel.
	ContainerMissing ObservableOption

	// CPUUsage transforms the default observable used to construct the CPU usage panel.
	CPUUsage ObservableOption

	// MemoryUsage transforms the default observable used to construct the memory usage panel.
	MemoryUsage ObservableOption

	// IOUsage transforms the default observable used to construct the IO usage panel.
	IOUsage ObservableOption

	// CustomTitle, if provided, provides a custom title for this monitoring group that will be displayed in Grafana.
	CustomTitle string
}

// NewContainerMonitoringGroup creates a group containing panels displaying
// container monitoring metrics - cpu, memory, io resource usage as well as
// a container missing alert - for the given container.
func NewContainerMonitoringGroup(containerName string, owner monitoring.ObservableOwner, options *ContainerMonitoringGroupOptions) monitoring.Group {
	if options == nil {
		options = &ContainerMonitoringGroupOptions{}
	}

	title := TitleContainerMonitoring
	if options.CustomTitle != "" {
		title = options.CustomTitle
	}

	return monitoring.Group{
		Title:  title,
		Hidden: true,
		Rows: []monitoring.Row{
			{
				options.ContainerMissing.safeApply(ContainerMissing(containerName, owner)).Observable(),
				options.CPUUsage.safeApply(ContainerCPUUsage(containerName, owner)).Observable(),
				options.MemoryUsage.safeApply(ContainerMemoryUsage(containerName, owner)).Observable(),
				options.IOUsage.safeApply(ContainerIOUsage(containerName, owner)).Observable(),
			},
		},
	}
}
