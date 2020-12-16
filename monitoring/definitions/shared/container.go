package shared

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

// Container monitoring overviews - alert on all container failures, but only alert on extreme resource usage.
// More granular resource usage warnings are provided by the provisioning observables.

var (
	ContainerRestarts sharedObservable = func(containerName string, owner monitoring.ObservableOwner) monitoring.Observable {
		return monitoring.Observable{
			Name:            "container_restarts",
			Description:     "container restarts every 5m by instance",
			Query:           fmt.Sprintf(`increase(cadvisor_container_restart_count{%s}[5m])`, CadvisorNameMatcher(containerName)),
			DataMayNotExist: true,
			Warning:         monitoring.Alert().GreaterOrEqual(1),
			PanelOptions:    monitoring.PanelOptions().LegendFormat("{{name}}"),
			Owner:           owner,
			PossibleSolutions: strings.Replace(`
			- **Kubernetes:**
				- Determine if the pod was OOM killed using 'kubectl describe pod {{CONTAINER_NAME}}' (look for 'OOMKilled: true') and, if so, consider increasing the memory limit in the relevant 'Deployment.yaml'.
				- Check the logs before the container restarted to see if there are 'panic:' messages or similar using 'kubectl logs -p {{CONTAINER_NAME}}'.
			- **Docker Compose:**
				- Determine if the pod was OOM killed using 'docker inspect -f \'{{json .State}}\' {{CONTAINER_NAME}}' (look for '"OOMKilled":true') and, if so, consider increasing the memory limit of the {{CONTAINER_NAME}} container in 'docker-compose.yml'.
				- Check the logs before the container restarted to see if there are 'panic:' messages or similar using 'docker logs {{CONTAINER_NAME}}' (note this will include logs from the previous and currently running container).
		`, "{{CONTAINER_NAME}}", containerName, -1),
		}
	}

	ContainerMemoryUsage sharedObservable = func(containerName string, owner monitoring.ObservableOwner) monitoring.Observable {
		return monitoring.Observable{
			Name:            "container_memory_usage",
			Description:     "container memory usage by instance",
			Query:           fmt.Sprintf(`cadvisor_container_memory_usage_percentage_total{%s}`, CadvisorNameMatcher(containerName)),
			DataMayNotExist: true,
			Warning:         monitoring.Alert().GreaterOrEqual(99),
			PanelOptions:    monitoring.PanelOptions().LegendFormat("{{name}}").Unit(monitoring.Percentage).Interval(100).Max(100).Min(0),
			Owner:           owner,
			PossibleSolutions: strings.Replace(`
			- **Kubernetes:** Consider increasing memory limit in relevant 'Deployment.yaml'.
			- **Docker Compose:** Consider increasing 'memory:' of {{CONTAINER_NAME}} container in 'docker-compose.yml'.
		`, "{{CONTAINER_NAME}}", containerName, -1),
		}
	}

	ContainerCPUUsage sharedObservable = func(containerName string, owner monitoring.ObservableOwner) monitoring.Observable {
		return monitoring.Observable{
			Name:            "container_cpu_usage",
			Description:     "container cpu usage total (1m average) across all cores by instance",
			Query:           fmt.Sprintf(`cadvisor_container_cpu_usage_percentage_total{%s}`, CadvisorNameMatcher(containerName)),
			DataMayNotExist: true,
			Warning:         monitoring.Alert().GreaterOrEqual(99),
			PanelOptions:    monitoring.PanelOptions().LegendFormat("{{name}}").Unit(monitoring.Percentage).Interval(100).Max(100).Min(0),
			Owner:           owner,
			PossibleSolutions: strings.Replace(`
			- **Kubernetes:** Consider increasing CPU limits in the the relevant 'Deployment.yaml'.
			- **Docker Compose:** Consider increasing 'cpus:' of the {{CONTAINER_NAME}} container in 'docker-compose.yml'.
		`, "{{CONTAINER_NAME}}", containerName, -1),
		}
	}

	ContainerFsInodes sharedObservable = func(containerName string, owner monitoring.ObservableOwner) monitoring.Observable {
		return monitoring.Observable{
			Name:            "fs_inodes_used",
			Description:     "fs inodes in use by instance",
			Query:           fmt.Sprintf(`sum by (name)(container_fs_inodes_total{%s})`, CadvisorNameMatcher(containerName)),
			DataMayNotExist: true,
			Warning:         monitoring.Alert().GreaterOrEqual(3e+06),
			PanelOptions:    monitoring.PanelOptions().LegendFormat("{{name}}"),
			Owner:           owner,
			PossibleSolutions: `
			- Refer to your OS or cloud provider's documentation for how to increase inodes.
			- **Kubernetes:** consider provisioning more machines with less resources.`,
		}
	}
)
