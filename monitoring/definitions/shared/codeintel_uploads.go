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
			DescriptionRoot: "Uploads > Store",
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

// src_codeintel_background_upload_records_removed_total
// src_codeintel_background_index_records_removed_total
// src_codeintel_background_uploads_purged_total
// src_codeintel_background_audit_log_records_expired_total
// src_codeintel_uploads_background_cleanup_errors_total
// src_codeintel_autoindexing_background_cleanup_errors_total
func (codeIntelligence) NewUploadsCleanupTaskGroup(containerName string) monitoring.Group {
	return monitoring.Group{
		Title:  "Codeintel: Uploads > Cleanup task",
		Hidden: false,
		Rows: []monitoring.Row{
			{
				Standard.Count("records deleted")(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_background_upload_records_removed",
					MetricDescriptionRoot: "lsif upload",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					Number of LSIF upload records deleted due to expiration or unreachability every 5m
				`).Observable(),

				Standard.Count("records deleted")(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_background_index_records_removed",
					MetricDescriptionRoot: "lsif index",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					Number of LSIF index records deleted due to expiration or unreachability every 5m
				`).Observable(),

				Standard.Count("data bundles deleted")(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_background_uploads_purged",
					MetricDescriptionRoot: "lsif upload",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					Number of LSIF upload data bundles purged from the codeintel-db database every 5m
				`).Observable(),

				Standard.Count("records deleted")(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_background_audit_log_records_expired",
					MetricDescriptionRoot: "lsif upload audit log",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					Number of LSIF upload audit log records deleted due to expiration every 5m
				`).Observable(),
			},
			{
				Observation.Errors(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_uploads_background_cleanup",
					MetricDescriptionRoot: "cleanup task",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					Number of code intelligence uploads cleanup task errors every 5m
				`).Observable(),
				Observation.Errors(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_autoindexing_background_cleanup",
					MetricDescriptionRoot: "cleanup task",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					Number of code intelligence autoindexing cleanup task errors every 5m
				`).Observable(),
			},
		},
	}
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
					MetricNameRoot:        "codeintel_background_repositories_scanned_total",
					MetricDescriptionRoot: "lsif upload repository scan",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					Number of repositories scanned for data retention
				`).Observable(),

				Standard.Count("records scanned")(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_background_upload_records_scanned_total",
					MetricDescriptionRoot: "lsif upload records scan",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					Number of codeintel upload records scanned for data retention
				`).Observable(),

				Standard.Count("commits scanned")(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_background_commits_scanned_total",
					MetricDescriptionRoot: "lsif upload commits scanned",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					Number of commits reachable from a codeintel upload record scanned for data retention
				`).Observable(),

				Standard.Count("uploads scanned")(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_background_upload_records_expired_total",
					MetricDescriptionRoot: "lsif upload records expired",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					Number of codeintel upload records marked as expired
				`).Observable(),
			},
		},
	}
}
