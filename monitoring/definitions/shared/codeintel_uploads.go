package shared

import "github.com/sourcegraph/sourcegraph/monitoring/monitoring"

// src_codeintel_uploads_total
// src_codeintel_uploads_duration_seconds_bucket
// src_codeintel_uploads_errors_total
func (codeIntelligence) NewUploadsServiceGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "Uploads > Service",
			Hidden:          false,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_uploads",
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

// src_codeintel_uploads_store_total
// src_codeintel_uploads_store_duration_seconds_bucket
// src_codeintel_uploads_store_errors_total
func (codeIntelligence) NewUploadsStoreGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "Uploads > Store (internal)",
			Hidden:          false,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_uploads_store",
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

// src_codeintel_uploads_transport_graphql_total
// src_codeintel_uploads_transport_graphql_duration_seconds_bucket
// src_codeintel_uploads_transport_graphql_errors_total
func (codeIntelligence) NewUploadsGraphQLTransportGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "Uploads > GQL Transport",
			Hidden:          false,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_uploads_transport_graphql",
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

// src_codeintel_uploads_transport_http_total
// src_codeintel_uploads_transport_http_duration_seconds_bucket
// src_codeintel_uploads_transport_http_errors_total
func (codeIntelligence) NewUploadsHTTPTransportGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "Uploads > HTTP Transport",
			Hidden:          false,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_uploads_transport_http",
				MetricDescriptionRoot: "http handler",
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

// src_codeintel_background_repositories_scanned_total
// src_codeintel_background_upload_records_scanned_total
// src_codeintel_background_commits_scanned_total
// src_codeintel_background_upload_records_expired_total
func (codeIntelligence) NewUploadsExpirationTaskGroup(containerName string) monitoring.Group {
	return monitoring.Group{
		Title:  "Codeintel: Uploads > Expiration task",
		Hidden: false,
		Rows: []monitoring.Row{
			{
				Standard.Count("repositories scanned")(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_background_repositories_scanned",
					MetricDescriptionRoot: "lsif upload repository scan",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					Number of repositories scanned for data retention
				`).Observable(),

				Standard.Count("records scanned")(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_background_upload_records_scanned",
					MetricDescriptionRoot: "lsif upload records scan",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					Number of codeintel upload records scanned for data retention
				`).Observable(),

				Standard.Count("commits scanned")(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_background_commits_scanned",
					MetricDescriptionRoot: "lsif upload commits scanned",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					Number of commits reachable from a codeintel upload record scanned for data retention
				`).Observable(),

				Standard.Count("uploads scanned")(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_background_upload_records_expired",
					MetricDescriptionRoot: "lsif upload records expired",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					Number of codeintel upload records marked as expired
				`).Observable(),
			},
		},
	}
}

// Tasks:
//   - codeintel_uploads_janitor_unknown_repository
//   - codeintel_uploads_janitor_unknown_commit
//   - codeintel_uploads_janitor_abandoned
//   - codeintel_uploads_expirer_unreferenced
//   - codeintel_uploads_expirer_unreferenced_graph
//   - codeintel_uploads_hard_deleter
//   - codeintel_uploads_janitor_audit_logs
//   - codeintel_uploads_janitor_scip_documents
//
// Suffixes:
//   - _total
//   - _duration_seconds_bucket
//   - _errors_total
//   - _records_scanned_total
//   - _records_altered_total
func (codeIntelligence) NewJanitorTaskGroups(containerName string) []monitoring.Group {
	return CodeIntelligence.newJanitorGroups(
		"Uploads > Janitor task",
		containerName,
		[]string{
			"codeintel_uploads_janitor_unknown_repository",
			"codeintel_uploads_janitor_unknown_commit",
			"codeintel_uploads_janitor_abandoned",
			"codeintel_uploads_expirer_unreferenced",
			"codeintel_uploads_expirer_unreferenced_graph",
			"codeintel_uploads_hard_deleter",
			"codeintel_uploads_janitor_audit_logs",
			"codeintel_uploads_janitor_scip_documents",
		},
	)
}

// Tasks:
//   - codeintel_uploads_reconciler_scip_metadata
//   - codeintel_uploads_reconciler_scip_data
//
// Suffixes:
//   - _total
//   - _duration_seconds_bucket
//   - _errors_total
//   - _records_scanned_total
//   - _records_altered_total
func (codeIntelligence) NewReconcilerTaskGroups(containerName string) []monitoring.Group {
	return CodeIntelligence.newJanitorGroups(
		"Uploads > Reconciler task",
		containerName,
		[]string{
			"codeintel_uploads_reconciler_scip_metadata",
			"codeintel_uploads_reconciler_scip_data",
		},
	)
}
