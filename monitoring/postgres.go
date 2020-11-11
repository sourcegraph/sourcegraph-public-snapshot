package main

const (
	codeintel = "codeintel-db"
	db        = "pgsql"
)

func Postgres() *Container {
	return &Container{
		Name:        "postgres",
		Title:       "Postgres",
		Description: "Metrics from postgres_exporter.",
		Groups: []Group{
			{
				Title: "Database settings",
				Rows: []Row{
					{
						postgresVersion(db, ObservableOwnerCloud),
						postgresMaxConnections(db, ObservableOwnerCloud),
						postgresSharedBuffers(db, ObservableOwnerCloud),

						//{
						//	Name:              "effective_cache",
						//	Description:       "effective cache",
						//	Query:             "pg_settings_effective_cache_size_bytes{instance=\"$instance\"}",
						//	Owner:             ObservableOwnerCloud,
						//	DataMayNotExist:   false,
						//	DataMayNotBeNaN:   false,
						//	Warning:           Alert().LessOrEqual(4000000),
						//	PossibleSolutions: "none",
						//	PanelOptions:      PanelOptions().LegendFormat("time_series").Unit(Bytes),
						//},
					},
				},
			},
			//{
			//	Title:  "Internal service requests",
			//	Hidden: true,
			//	Rows: []Row{
			//		{
			//			{
			//				Name:            "internal_indexed_search_error_responses",
			//				Description:     "internal indexed search error responses every 5m",
			//				Query:           `sum by(code) (increase(src_zoekt_request_duration_seconds_count{code!~"2.."}[5m])) / ignoring(code) group_left sum(increase(src_zoekt_request_duration_seconds_count[5m])) * 100`,
			//				DataMayNotExist: true,
			//				Warning:         Alert().GreaterOrEqual(5).For(15 * time.Minute),
			//				PanelOptions:    PanelOptions().LegendFormat("{{code}}").Unit(Percentage),
			//				Owner:           ObservableOwnerSearch,
			//				PossibleSolutions: `
			//					- Check the Zoekt Web Server dashboard for indications it might be unhealthy.
			//				`,
			//			},
			//			{
			//				Name:            "internal_unindexed_search_error_responses",
			//				Description:     "internal unindexed search error responses every 5m",
			//				Query:           `sum by(code) (increase(searcher_service_request_total{code!~"2.."}[5m])) / ignoring(code) group_left sum(increase(searcher_service_request_total[5m])) * 100`,
			//				DataMayNotExist: true,
			//				Warning:         Alert().GreaterOrEqual(5).For(15 * time.Minute),
			//				PanelOptions:    PanelOptions().LegendFormat("{{code}}").Unit(Percentage),
			//				Owner:           ObservableOwnerSearch,
			//				PossibleSolutions: `
			//					- Check the Searcher dashboard for indications it might be unhealthy.
			//				`,
			//			},
			//			{
			//				Name:            "internal_api_error_responses",
			//				Description:     "internal API error responses every 5m by route",
			//				Query:           `sum by(category) (increase(src_frontend_internal_request_duration_seconds_count{code!~"2.."}[5m])) / ignoring(code) group_left sum(increase(src_frontend_internal_request_duration_seconds_count[5m])) * 100`,
			//				DataMayNotExist: true,
			//				Warning:         Alert().GreaterOrEqual(5).For(15 * time.Minute),
			//				PanelOptions:    PanelOptions().LegendFormat("{{category}}").Unit(Percentage),
			//				Owner:           ObservableOwnerCloud,
			//				PossibleSolutions: `
			//					- May not be a substantial issue, check the 'frontend' logs for potential causes.
			//				`,
			//			},
			//		},
			//
			//		{
			//			{
			//				Name:              "observability_test_alert_warning",
			//				Description:       "warning test alert metric",
			//				Query:             `max by(owner) (observability_test_metric_warning)`,
			//				DataMayNotExist:   true,
			//				Warning:           Alert().GreaterOrEqual(1),
			//				PanelOptions:      PanelOptions().Max(1),
			//				Owner:             ObservableOwnerDistribution,
			//				PossibleSolutions: "This alert is triggered via the `triggerObservabilityTestAlert` GraphQL endpoint, and will automatically resolve itself.",
			//			},
			//			{
			//				Name:              "observability_test_alert_critical",
			//				Description:       "critical test alert metric",
			//				Query:             `max by(owner) (observability_test_metric_critical)`,
			//				DataMayNotExist:   true,
			//				Critical:          Alert().GreaterOrEqual(1),
			//				PanelOptions:      PanelOptions().Max(1),
			//				Owner:             ObservableOwnerDistribution,
			//				PossibleSolutions: "This alert is triggered via the `triggerObservabilityTestAlert` GraphQL endpoint, and will automatically resolve itself.",
			//			},
			//		},
			//	},
			//},
			//{
			//	Title:  "Container monitoring (not available on server)",
			//	Hidden: true,
			//	Rows: []Row{
			//		{
			//			sharedContainerCPUUsage("frontend", ObservableOwnerCloud),
			//			sharedContainerMemoryUsage("frontend", ObservableOwnerCloud),
			//		},
			//		{
			//			sharedContainerRestarts("frontend", ObservableOwnerCloud),
			//			sharedContainerFsInodes("frontend", ObservableOwnerCloud),
			//		},
			//	},
			//},
			{
				Title:  "Provisioning indicators (not available on server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedProvisioningCPUUsageLongTerm("pgsql", ObservableOwnerCloud),
						sharedProvisioningMemoryUsageLongTerm("pgsql", ObservableOwnerCloud),
					},
					{
						sharedProvisioningCPUUsageShortTerm("pgsql", ObservableOwnerCloud),
						sharedProvisioningMemoryUsageShortTerm("pgsql", ObservableOwnerCloud),
					},
				},
			},
		},
	}
}
