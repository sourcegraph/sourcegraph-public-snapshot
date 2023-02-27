package monitoring

import (
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring/internal/promql"
)

// ListMetrics lists the metrics used by each dashboard, deduplicating metrics by
// dashboard.
func ListMetrics(dashboards ...*Dashboard) (map[*Dashboard][]string, error) {
	results := make(map[*Dashboard][]string)
	for _, d := range dashboards {
		// Deduplicate metrics by dashboard
		foundMetrics := make(map[string]struct{})
		addMetrics := func(metrics []string) {
			for _, m := range metrics {
				if _, exists := foundMetrics[m]; !exists {
					foundMetrics[m] = struct{}{}
					results[d] = append(results[d], m)
				}
			}
		}

		// Add metrics used by fixed variables added in generateDashboards(). This is kind
		// of hack, but easiest to do manually.
		addMetrics([]string{"ALERTS", "alert_count", "src_service_metadata"})

		// Add variable queries if any
		for _, v := range d.Variables {
			if v.OptionsLabelValues.Query != "" {
				metrics, err := promql.ListMetrics(v.OptionsLabelValues.Query, nil)
				if err != nil {
					return nil, errors.Wrapf(err, "%s: %s", d.Name, v.Name)
				}
				addMetrics(metrics)
			}
		}
		// Iterate for Observables
		for _, g := range d.Groups {
			for _, r := range g.Rows {
				for _, o := range r {
					metrics, err := promql.ListMetrics(o.Query, newVariableApplier(d.Variables))
					if err != nil {
						return nil, errors.Wrapf(err, "%s: %s", d.Name, o.Name)
					}
					addMetrics(metrics)
				}
			}
		}
	}

	return results, nil
}
