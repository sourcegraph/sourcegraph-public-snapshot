package shared

import "github.com/sourcegraph/sourcegraph/monitoring/monitoring"

// src_codeintel_ranking_total
// src_codeintel_ranking_duration_seconds_bucket
// src_codeintel_ranking_errors_total
func (codeIntelligence) NewRankingServiceGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "Ranking > Service",
			Hidden:          false,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_ranking",
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

// src_codeintel_ranking_store_total
// src_codeintel_ranking_store_duration_seconds_bucket
// src_codeintel_ranking_store_errors_total
func (codeIntelligence) NewRankingStoreGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "Ranking > Store",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_ranking_store",
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

// src_codeintel_uploads_ranking_uploads_read_total
// src_codeintel_uploads_ranking_bytes_uploaded_total
// src_codeintel_uploads_ranking_stale_uploads_removed_total
// src_codeintel_uploads_ranking_bytes_deleted_total
// src_codeintel_ranking_csv_files_processed_total
// src_codeintel_ranking_csv_files_bytes_read_total
// src_codeintel_ranking_repositories_updated_total
// src_codeintel_ranking_input_rows_processed_total
func (codeIntelligence) NewRankingPageRankGroup(containerName string) monitoring.Group {
	return monitoring.Group{
		Title:  "Codeintel: Ranking > PageRank",
		Hidden: false,
		Rows: []monitoring.Row{
			{
				Standard.Count("repository path ranks updated")(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_ranking_repositories_updated",
					MetricDescriptionRoot: "repository path ranks updated",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					The number of updates to document scores of any repository.
				`).Observable(),

				Standard.Count("files read from GCS")(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_ranking_csv_files_processed",
					MetricDescriptionRoot: "csv files read and processed",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					The number of input CSV records read from GCS.
				`).Observable(),

				Standard.Count("csv result rows processed")(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_ranking_input_rows_processed",
					MetricDescriptionRoot: "csv result rows processed",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					The number of input row records merged into document scores for
				`).Observable(),
			},

			{
				Standard.Count("uploads read for export")(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_uploads_ranking_uploads_read",
					MetricDescriptionRoot: "uploads read",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					The number of upload records read.
				`).Observable(),

				Standard.Count("stale upload records removed")(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_uploads_ranking_stale_uploads_removed",
					MetricDescriptionRoot: "uploads removed",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					The number of stale upload records removed from GCS.
				`).Observable(),
			},

			{
				withBytes(Standard.Count("bytes read from GCS")(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_ranking_csv_files_bytes_read",
					MetricDescriptionRoot: "bytes read",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					The number of bytes read from GCS.
				`).Observable()),

				withBytes(Standard.Count("bytes uploaded to GCS")(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_uploads_ranking_bytes_uploaded",
					MetricDescriptionRoot: "bytes uploaded",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					The number of bytes uploaded to GCS.
				`).Observable()),

				withBytes(Standard.Count("bytes deleted from GCS")(ObservableConstructorOptions{
					MetricNameRoot:        "codeintel_uploads_ranking_bytes_deleted",
					MetricDescriptionRoot: "bytes deleted",
				})(containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					The number of bytes deleted from GCS.
				`).Observable()),
			},
		},
	}
}

func withBytes(observable monitoring.Observable) monitoring.Observable {
	observable.Panel = observable.Panel.Unit(monitoring.Bytes)
	return observable
}
