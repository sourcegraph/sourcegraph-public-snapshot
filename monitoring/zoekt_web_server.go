package main

import "time"

func ZoektWebServer() *Container {
	return &Container{
		Name:        "zoekt-webserver",
		Title:       "Zoekt Web Server",
		Description: "Serves indexed search requests using the search index.",
		Groups: []Group{
			{
				Title: "General",
				Rows: []Row{
					{
						{
							Name:              "indexed_search_request_errors",
							Description:       "indexed search request errors every 5m by code",
							Query:             `sum by (code)(increase(src_zoekt_request_duration_seconds_count{code!~"2.."}[5m])) / ignoring(code) group_left sum(increase(src_zoekt_request_duration_seconds_count[5m])) * 100`,
							DataMayNotExist:   true,
							DataMayBeNaN:      true, // denominator may be zero
							Warning:           Alert{GreaterOrEqual: 5, For: 5 * time.Minute},
							PanelOptions:      PanelOptions().LegendFormat("{{code}}").Unit(Percentage),
							Owner:             ObservableOwnerSearch,
							PossibleSolutions: "none",
						},
					},
				},
			},
			{
				Title:  "Container monitoring (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerCPUUsage("zoekt-webserver"),
						sharedContainerMemoryUsage("zoekt-webserver"),
					},
					{
						sharedContainerRestarts("zoekt-webserver"),
						sharedContainerFsInodes("zoekt-webserver"),
					},
				},
			},
			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedProvisioningCPUUsageLongTerm("zoekt-webserver"),
						sharedProvisioningMemoryUsageLongTerm("zoekt-webserver"),
					},
					{
						sharedProvisioningCPUUsageShortTerm("zoekt-webserver"),
						sharedProvisioningMemoryUsageShortTerm("zoekt-webserver"),
					},
				},
			},
			// kubernetes monitoring for zoekt-web-server is provided by zoekt-index-server,
			// since both services are deployed together
		},
	}
}
