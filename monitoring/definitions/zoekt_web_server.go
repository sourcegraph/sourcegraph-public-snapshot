package definitions

import (
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func ZoektWebServer() *monitoring.Container {
	return &monitoring.Container{
		Name:        "zoekt-webserver",
		Title:       "Zoekt Web Server",
		Description: "Serves indexed search requests using the search index.",
		Groups: []monitoring.Group{
			{
				Title: "General",
				Rows: []monitoring.Row{
					{
						{
							Name:              "indexed_search_request_errors",
							Description:       "indexed search request errors every 5m by code",
							Query:             `sum by (code)(increase(src_zoekt_request_duration_seconds_count{code!~"2.."}[5m])) / ignoring(code) group_left sum(increase(src_zoekt_request_duration_seconds_count[5m])) * 100`,
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(5).For(5 * time.Minute),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("{{code}}").Unit(monitoring.Percentage),
							Owner:             monitoring.ObservableOwnerSearch,
							PossibleSolutions: "none",
						},
					},
				},
			},
			{
				Title:  "Container monitoring (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						sharedContainerCPUUsage("zoekt-webserver", monitoring.ObservableOwnerSearch),
						sharedContainerMemoryUsage("zoekt-webserver", monitoring.ObservableOwnerSearch),
					},
					{
						sharedContainerRestarts("zoekt-webserver", monitoring.ObservableOwnerSearch),
						sharedContainerFsInodes("zoekt-webserver", monitoring.ObservableOwnerSearch),
					},
					{
						{
							Name:              "fs_io_operations",
							Description:       "filesystem reads and writes by instance rate over 1h",
							Query:             fmt.Sprintf(`sum by(name) (rate(container_fs_reads_total{%[1]s}[1h]) + rate(container_fs_writes_total{%[1]s}[1h]))`, promCadvisorContainerMatchers("zoekt-webserver")),
							DataMayNotExist:   true,
							Warning:           monitoring.Alert().GreaterOrEqual(5000),
							PanelOptions:      monitoring.PanelOptions().LegendFormat("{{name}}"),
							Owner:             monitoring.ObservableOwnerSearch,
							PossibleSolutions: "none",
						},
					},
				},
			},
			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						sharedProvisioningCPUUsageLongTerm("zoekt-webserver", monitoring.ObservableOwnerSearch),
						sharedProvisioningMemoryUsageLongTerm("zoekt-webserver", monitoring.ObservableOwnerSearch),
					},
					{
						sharedProvisioningCPUUsageShortTerm("zoekt-webserver", monitoring.ObservableOwnerSearch),
						sharedProvisioningMemoryUsageShortTerm("zoekt-webserver", monitoring.ObservableOwnerSearch),
					},
				},
			},
			// kubernetes monitoring for zoekt-web-server is provided by zoekt-index-server,
			// since both services are deployed together
		},
	}
}
