package main

import "fmt"

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
	}
}

var sharedContainerRestarts sharedObservable = func(containerName string) Observable {
	return Observable{
		Name:            "container_restarts",
		Description:     "container restarts every 5m by instance (not available on k8s or server)",
		Query:           fmt.Sprintf(`increase(cadvisor_container_restart_count{name=~".*%s.*"}[5m])`, containerName),
		DataMayNotExist: true,
		Warning:         Alert{GreaterOrEqual: 1},
		PanelOptions:    PanelOptions().LegendFormat("{{name}}"),
	}
}

var sharedContainerMemoryUsage sharedObservable = func(containerName string) Observable {
	return Observable{
		Name:            "container_memory_usage",
		Description:     "container memory usage by instance (not available on k8s or server)",
		Query:           fmt.Sprintf(`cadvisor_container_memory_usage_percentage_total{name=~".*%s.*"}`, containerName),
		DataMayNotExist: true,
		Warning:         Alert{GreaterOrEqual: 90},
		PanelOptions:    PanelOptions().LegendFormat("{{name}}").Unit(Percentage),
	}
}

var sharedContainerCPUUsage sharedObservable = func(containerName string) Observable {
	return Observable{
		Name:            "container_cpu_usage",
		Description:     "container cpu usage total (5m average) across all cores by instance (not available on k8s or server)",
		Query:           fmt.Sprintf(`cadvisor_container_cpu_usage_percentage_total{name=~".*%s.*"}`, containerName),
		DataMayNotExist: true,
		Warning:         Alert{GreaterOrEqual: 90},
		PanelOptions:    PanelOptions().LegendFormat("{{name}}").Unit(Percentage),
	}
}
