package definitions

import (
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func ZoektWebServer() *monitoring.Container {
	const containerName = "zoekt-webserver"

	return &monitoring.Container{
		Name:                     "zoekt-webserver",
		Title:                    "Zoekt Web Server",
		Description:              "Serves indexed search requests using the search index.",
		NoSourcegraphDebugServer: true,
		Groups: []monitoring.Group{
			{
				Title: "General",
				Rows: []monitoring.Row{
					{
						{
							Name:              "indexed_search_request_errors",
							Description:       "indexed search request errors every 5m by code",
							Query:             `sum by (code)(increase(src_zoekt_request_duration_seconds_count{code!~"2.."}[5m])) / ignoring(code) group_left sum(increase(src_zoekt_request_duration_seconds_count[5m])) * 100`,
							Warning:           monitoring.Alert().GreaterOrEqual(5, nil).For(5 * time.Minute),
							Panel:             monitoring.Panel().LegendFormat("{{code}}").Unit(monitoring.Percentage),
							Owner:             monitoring.ObservableOwnerSearchCore,
							PossibleSolutions: "none",
						},
					},
				},
			},

			// Note 1:
			// indexed-search does not have zero-downtime deploy, so deploys can cause extended container restarts.
			// We set the default warning alert for extended periods of container restarts as it may still indicate
			// a real problem.
			//
			// Note 2:
			// Kubernetes monitoring for zoekt-webserver is provided by zoekt-indexserver as they are bundled together.

			shared.NewContainerMonitoringGroup(containerName, monitoring.ObservableOwnerSearchCore, nil),
			shared.NewProvisioningIndicatorsGroup(containerName, monitoring.ObservableOwnerSearchCore, nil),
		},
	}
}
