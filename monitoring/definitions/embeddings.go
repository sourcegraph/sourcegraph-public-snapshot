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
		Variables: []monitoring.ContainerVariable{
			{
				Label: "instance",
				Name:  "instance",
				OptionsLabelValues: monitoring.ContainerVariableOptionsLabelValues{
					Query:         "src_embeddings_cache_hit_count",
					LabelName:     "instance",
					ExampleOption: "embeddings:6099",
				},
				Multi: true,
			},
		},
		Groups: []monitoring.Group{
			shared.NewSiteConfigurationClientMetricsGroup(shared.SiteConfigurationMetricsOptions{
				HumanServiceName:    "embeddings",
				InstanceFilterRegex: `${instance:regex}`,
			}, monitoring.ObservableOwnerDevOps),
			shared.NewDatabaseConnectionsMonitoringGroup(containerName, monitoring.ObservableOwnerCody),
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
						Name:           "missed_bytes",
						Description:    "bytes fetched due to a cache miss",
						Owner:          monitoring.ObservableOwner{},
						Query:          "rate(src_embeddings_cache_miss_bytes[10m])",
						NoAlert:        true,
						Interpretation: "A high volume of misses indicates that the many searches are not hitting the cache. Consider increasing the cache size.",
						Panel:          monitoring.Panel().Unit(monitoring.Bytes),
					},
				}},
			},
		},
	}
}
