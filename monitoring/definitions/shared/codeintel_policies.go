package shared

import "github.com/sourcegraph/sourcegraph/monitoring/monitoring"

// src_codeintel_policies_total
// src_codeintel_policies_duration_seconds_bucket
// src_codeintel_policies_errors_total
func (codeIntelligence) NewPoliciesServiceGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "Policies > Service",
			Hidden:          false,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_policies",
				MetricDescriptionRoot: "service",
				By:                    []string{"op"},
			},
		},

		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
		Aggregate: &SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
	})
}

// src_codeintel_policies_store_total
// src_codeintel_policies_store_duration_seconds_bucket
// src_codeintel_policies_store_errors_total
func (codeIntelligence) NewPoliciesStoreGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "Policies > Store",
			Hidden:          false,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_policies_store",
				MetricDescriptionRoot: "store",
				By:                    []string{"op"},
			},
		},

		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
		Aggregate: &SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
	})
}

// src_codeintel_policies_transport_graphql_total
// src_codeintel_policies_transport_graphql_duration_seconds_bucket
// src_codeintel_policies_transport_graphql_errors_total
func (codeIntelligence) NewPoliciesGraphQLTransportGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "Policies > GQL Transport",
			Hidden:          false,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_policies_transport_graphql",
				MetricDescriptionRoot: "resolver",
				By:                    []string{"op"},
			},
		},

		SharedObservationGroupOptions: SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
		Aggregate: &SharedObservationGroupOptions{
			Total:     NoAlertsOption("none"),
			Duration:  NoAlertsOption("none"),
			Errors:    NoAlertsOption("none"),
			ErrorRate: NoAlertsOption("none"),
		},
	})
}

// src_codeintel_background_policies_updated_total
func (codeIntelligence) NewRepoMatcherTaskGroup(containerName string) monitoring.Group {
	return monitoring.Group{
		Title:  "Codeintel: Policies > Repository Pattern Matcher task",
		Hidden: false,
		Rows: []monitoring.Row{
			{
				Standard.Count("repositories pattern matcher")(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_background_policies_updated_total",
					MetricDescriptionRoot: "lsif repository pattern matcher",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					Number of configuration policies whose repository membership list was updated
				`).Observable(),
			},
		},
	}
}
