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

// src_codeintel_ranking_lsifstore_total
// src_codeintel_ranking_lsifstore_duration_seconds_bucket
// src_codeintel_ranking_lsifstore_errors_total
func (codeIntelligence) NewRankingLSIFStoreGroup(containerName string) monitoring.Group {
	return Observation.NewGroup(containerName, monitoring.ObservableOwnerCodeIntel, ObservationGroupOptions{
		GroupConstructorOptions: GroupConstructorOptions{
			Namespace:       "codeintel",
			DescriptionRoot: "Ranking > LSIFStore",
			Hidden:          true,

			ObservableConstructorOptions: ObservableConstructorOptions{
				MetricNameRoot:        "codeintel_ranking_lsifstore",
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

// Tasks:
//   - codeintel_ranking_symbol_exporter
//   - codeintel_ranking_file_reference_count_mapper
//   - codeintel_ranking_file_reference_count_reducer
//
// Suffixes:
//   - _total
//   - _duration_seconds_bucket
//   - _errors_total
//   - _records_processed_total
//   - _records_altered_total
func (codeIntelligence) NewRankingPipelineTaskGroups(containerName string) []monitoring.Group {
	return CodeIntelligence.newPipelineGroups(
		"Uploads > Pipeline task",
		containerName,
		[]string{
			"codeintel_ranking_symbol_exporter",
			"codeintel_ranking_file_reference_count_seed_mapper",
			"codeintel_ranking_file_reference_count_mapper",
			"codeintel_ranking_file_reference_count_reducer",
		},
	)
}

// Tasks:
//   - codeintel_ranking_exported_uploads_janitor
//   - codeintel_ranking_deleted_exported_uploads_janitor
//   - codeintel_ranking_abandoned_exported_uploads_janitor
//   - codeintel_ranking_rank_counts_janitor
//   - codeintel_ranking_rank_janitor
//
// Suffixes:
//   - _total
//   - _duration_seconds_bucket
//   - _errors_total
//   - _records_scanned_total
//   - _records_altered_total
func (codeIntelligence) NewRankingJanitorTaskGroups(containerName string) []monitoring.Group {
	return CodeIntelligence.newJanitorGroups(
		"Uploads > Janitor task",
		containerName,
		[]string{
			"codeintel_ranking_processed_references_janitor",
			"codeintel_ranking_processed_paths_janitor",
			"codeintel_ranking_exported_uploads_janitor",
			"codeintel_ranking_deleted_exported_uploads_janitor",
			"codeintel_ranking_abandoned_exported_uploads_janitor",
			"codeintel_ranking_rank_counts_janitor",
			"codeintel_ranking_rank_janitor",
		},
	)
}
