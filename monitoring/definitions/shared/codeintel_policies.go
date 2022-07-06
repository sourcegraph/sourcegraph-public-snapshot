package shared

import "github.com/sourcegraph/sourcegraph/monitoring/monitoring"

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
				})("job", containerName, monitoring.ObservableOwnerCodeIntel).WithNoAlerts(`
					Number of configuration policies whose repository membership list was updated
				`).Observable(),
			},
		},
	}
}
