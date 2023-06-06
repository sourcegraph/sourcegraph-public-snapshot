package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func Embeddings() *monitoring.Dashboard {
	const containerName = "embeddings"

	return &monitoring.Dashboard{
		Name:        "embeddings",
		Title:       "Embeddings",
		Description: "Handles embeddings searches.",
		Groups: []monitoring.Group{
			shared.NewDatabaseConnectionsMonitoringGroup(containerName),
			shared.NewFrontendInternalAPIErrorResponseMonitoringGroup(containerName, monitoring.ObservableOwnerCody, nil),
			shared.NewContainerMonitoringGroup(containerName, monitoring.ObservableOwnerCody, nil),
			shared.NewProvisioningIndicatorsGroup(containerName, monitoring.ObservableOwnerCody, nil),
			shared.NewGolangMonitoringGroup(containerName, monitoring.ObservableOwnerCody, nil),
			shared.NewKubernetesMonitoringGroup(containerName, monitoring.ObservableOwnerCody, nil),
			{
				Title:  "Cache",
				Hidden: true,
				Rows: []monitoring.Row{{
					{
						Name:           "hit_ratio",
						Description:    "hit ratio of the embeddings cache",
						Owner:          monitoring.ObservableOwner{},
						Query:          "rate(src_embeddings_cache_hit_count[30m]) / (rate(src_embeddings_cache_hit_count[30m]) + rate(src_embeddings_cache_miss_count[30m]))",
						NoAlert:        true,
						Interpretation: "A low hit rate indicates your cache is not well utilized. Consider increasing the cache size.",
						Panel:          monitoring.Panel().Unit(monitoring.Number),
					},
					{
						Name:           "evicted_bytes",
						Description:    "bytes evicted from the embeddings cache",
						Owner:          monitoring.ObservableOwner{},
						Query:          "rate(src_embeddings_cache_evicted_bytes[10m])",
						NoAlert:        true,
						Interpretation: "A high eviction rate indicates that large numbers of embeddings are being removed from the cache, and will have to subsequently be re-fetched on a new query. Consider increasing the cache size.",
						Panel:          monitoring.Panel().Unit(monitoring.Bytes),
					},
				}},
			},
		},
	}
}
