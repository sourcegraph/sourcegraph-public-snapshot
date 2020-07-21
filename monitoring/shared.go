package main

import (
	"fmt"
	"strings"
)

// This file contains shared declarations between dashboards. In general, you should NOT be making
// changes to this file: we use a declarative style for monitoring intentionally, so you should err
// on the side of repeating yourself and NOT writing shared or programatically generated monitoring.
//
// When editing this file or introducing any shared declarations, you should abide strictly by the
// following rules:
//
// 1. Do NOT declare a shared definition unless 5+ dashboards will use it. Sharing dashboard
//    declarations means the codebase becomes more complex and non-declarative which we want to avoid
//    so repeat yourself instead if it applies to less than 5 dashboards.
//
// 2. ONLY declare shared Observables. Introducing shared Rows or Groups prevents individual dashboard
//    maintainers from holistically considering both the layout of dashboards as well as the
//    metrics and alerts defined within them -- which we do not want.
//
// 3. Use the sharedObservable type and do NOT parameterize more than just the container name. It may
//    be tempting to pass an alerting threshold as an argument, or parameterize whether a critical
//    alert is defined -- but this makes reasoning about alerts at a high level much more difficult.
//    If you have a need for this, it is a strong signal you should NOT be using the shared definition
//    anymore and should instead copy it and apply your modifications.
//

type sharedObservable func(containerName string) Observable

var sharedFrontendInternalAPIErrorResponses sharedObservable = func(containerName string) Observable {
	return Observable{
		Name:            "frontend_internal_api_error_responses",
		Description:     "frontend-internal API error responses every 5m by route",
		Query:           fmt.Sprintf(`sum by (category)(increase(src_frontend_internal_request_duration_seconds_count{job="%s",code!~"2.."}[5m]))`, containerName),
		DataMayNotExist: true,
		Warning:         Alert{GreaterOrEqual: 5},
		PanelOptions:    PanelOptions().LegendFormat("{{category}}"),
		PossibleSolutions: strings.Replace(`
			- **Single-container deployments:** Check 'docker logs $CONTAINER_ID' for logs starting with 'repo-updater' that indicate requests to the frontend service are failing.
			- **Kubernetes:**
				- Confirm that 'kubectl get pods' shows the 'frontend' pods are healthy.
				- Check 'kubectl logs {{CONTAINER_NAME}}' for logs indicate request failures to 'frontend' or 'frontend-internal'.
			- **Docker Compose:**
				- Confirm that 'docker ps' shows the 'frontend-internal' container is healthy.
				- Check 'docker logs {{CONTAINER_NAME}}' for logs indicating request failures to 'frontend' or 'frontend-internal'.
		`, "{{CONTAINER_NAME}}", containerName, -1),
	}
}

// promCadvisorContainerMatchers generates Prometheus matchers that capture metrics that match the given container name
// while excluding some irrelevant metrics (namely pods and jaeger sidecars)
func promCadvisorContainerMatchers(containerName string) string {
	return fmt.Sprintf(`name=~".*%s.*",name!~".*(_POD_|_jaeger-agent_).*"`, containerName)
}

// Container monitoring overviews - alert on all container failures, but only alert on extreme resource usage.
// More granular resource usage warnings are provided by the provisioning observables.

var sharedContainerRestarts sharedObservable = func(containerName string) Observable {
	return Observable{
		Name:            "container_restarts",
		Description:     "container restarts every 5m by instance",
		Query:           fmt.Sprintf(`increase(cadvisor_container_restart_count{%s}[5m])`, promCadvisorContainerMatchers(containerName)),
		DataMayNotExist: true,
		Warning:         Alert{GreaterOrEqual: 1},
		PanelOptions:    PanelOptions().LegendFormat("{{name}}"),
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

var sharedContainerMemoryUsage sharedObservable = func(containerName string) Observable {
	return Observable{
		Name:            "container_memory_usage",
		Description:     "container memory usage by instance",
		Query:           fmt.Sprintf(`cadvisor_container_memory_usage_percentage_total{%s}`, promCadvisorContainerMatchers(containerName)),
		DataMayNotExist: true,
		Warning:         Alert{GreaterOrEqual: 99},
		PanelOptions:    PanelOptions().LegendFormat("{{name}}").Unit(Percentage).Interval(100),
		PossibleSolutions: strings.Replace(`
			- **Kubernetes:** Consider increasing memory limit in relevant 'Deployment.yaml'.
			- **Docker Compose:** Consider increasing 'memory:' of {{CONTAINER_NAME}} container in 'docker-compose.yml'.
		`, "{{CONTAINER_NAME}}", containerName, -1),
	}
}

var sharedContainerCPUUsage sharedObservable = func(containerName string) Observable {
	return Observable{
		Name:            "container_cpu_usage",
		Description:     "container cpu usage total (1m average) across all cores by instance",
		Query:           fmt.Sprintf(`cadvisor_container_cpu_usage_percentage_total{%s}`, promCadvisorContainerMatchers(containerName)),
		DataMayNotExist: true,
		Warning:         Alert{GreaterOrEqual: 99},
		PanelOptions:    PanelOptions().LegendFormat("{{name}}").Unit(Percentage).Interval(100),
		PossibleSolutions: strings.Replace(`
			- **Kubernetes:** Consider increasing CPU limits in the the relevant 'Deployment.yaml'.
			- **Docker Compose:** Consider increasing 'cpus:' of the {{CONTAINER_NAME}} container in 'docker-compose.yml'.
		`, "{{CONTAINER_NAME}}", containerName, -1),
	}
}

// Warn that instances might need more resources if short-term usage is high.

var sharedProvisioningCPUUsage5m sharedObservable = func(containerName string) Observable {
	return Observable{
		Name:            "provisioning_container_cpu_usage_5m",
		Description:     "container cpu usage total (5m maximum) across all cores by instance",
		Query:           fmt.Sprintf(`max_over_time(cadvisor_container_cpu_usage_percentage_total{%s}[5m])`, promCadvisorContainerMatchers(containerName)),
		DataMayNotExist: true,
		Warning:         Alert{GreaterOrEqual: 90},
		PanelOptions:    PanelOptions().LegendFormat("{{name}}").Unit(Percentage).Interval(100),
		PossibleSolutions: strings.Replace(`
			- **Kubernetes:** Consider increasing CPU limits in the the relevant 'Deployment.yaml'.
			- **Docker Compose:** Consider increasing 'cpus:' of the {{CONTAINER_NAME}} container in 'docker-compose.yml'.
		`, "{{CONTAINER_NAME}}", containerName, -1),
	}
}

var sharedProvisioningMemoryUsage5m sharedObservable = func(containerName string) Observable {
	return Observable{
		Name:            "provisioning_container_memory_usage_5m",
		Description:     "container memory usage (5m maximum) by instance",
		Query:           fmt.Sprintf(`max_over_time(cadvisor_container_memory_usage_percentage_total{%s}[5m])`, promCadvisorContainerMatchers(containerName)),
		DataMayNotExist: true,
		Warning:         Alert{GreaterOrEqual: 90},
		PanelOptions:    PanelOptions().LegendFormat("{{name}}").Unit(Percentage).Interval(100),
		PossibleSolutions: strings.Replace(`
			- **Kubernetes:** Consider increasing memory limit in relevant 'Deployment.yaml'.
			- **Docker Compose:** Consider increasing 'memory:' of {{CONTAINER_NAME}} container in 'docker-compose.yml'.
		`, "{{CONTAINER_NAME}}", containerName, -1),
	}
}

// Warn that instances might need more/less resources if long-term usage is high or low.

var sharedProvisioningCPUUsage7d sharedObservable = func(containerName string) Observable {
	return Observable{
		Name:            "provisioning_container_cpu_usage_7d",
		Description:     "container cpu usage total (7d maximum) across all cores by instance",
		Query:           fmt.Sprintf(`max_over_time(cadvisor_container_cpu_usage_percentage_total{%s}[7d])`, promCadvisorContainerMatchers(containerName)),
		DataMayNotExist: true,
		Warning:         Alert{LessOrEqual: 30, GreaterOrEqual: 80},
		PanelOptions:    PanelOptions().LegendFormat("{{name}}").Unit(Percentage),
		PossibleSolutions: strings.Replace(`
			- If usage is high:
				- **Kubernetes:** Consider decreasing CPU limits in the the relevant 'Deployment.yaml'.
				- **Docker Compose:** Consider descreasing 'cpus:' of the {{CONTAINER_NAME}} container in 'docker-compose.yml'.
			- If usage is low, consider decreasing the above values.
		`, "{{CONTAINER_NAME}}", containerName, -1),
	}
}

var sharedProvisioningMemoryUsage7d sharedObservable = func(containerName string) Observable {
	return Observable{
		Name:            "provisioning_container_memory_usage_7d",
		Description:     "container memory usage (7d maximum) by instance",
		Query:           fmt.Sprintf(`max_over_time(cadvisor_container_memory_usage_percentage_total{%s}[7d])`, promCadvisorContainerMatchers(containerName)),
		DataMayNotExist: true,
		Warning:         Alert{LessOrEqual: 30, GreaterOrEqual: 80},
		PanelOptions:    PanelOptions().LegendFormat("{{name}}").Unit(Percentage),
		PossibleSolutions: strings.Replace(`
			- If usage is high:
				- **Kubernetes:** Consider decreasing memory limit in relevant 'Deployment.yaml'.
				- **Docker Compose:** Consider decreasing 'memory:' of {{CONTAINER_NAME}} container in 'docker-compose.yml'.
			- If usage is low, consider decreasing the above values.
		`, "{{CONTAINER_NAME}}", containerName, -1),
	}
}
