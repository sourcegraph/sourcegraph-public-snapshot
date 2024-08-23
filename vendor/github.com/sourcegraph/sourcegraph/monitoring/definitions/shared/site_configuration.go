package shared

import (
	"fmt"
	"strings"
	"time"

	"github.com/iancoleman/strcase"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

type SiteConfigurationMetricsOptions struct {
	// HumanServiceName is the short, lowercase, snake_case, human-readable name of the service that we're gathering metrics for.
	//
	// Example: "gitserver"
	HumanServiceName string

	// InstanceFilterRegex is the PromQL regex that's used to filter the
	// site configuration client metrics to only those emitted by the instance(s) that were interested in.
	//
	// Example: (gitserver-0 | gitserver-1)
	InstanceFilterRegex string

	// JobFilterRegex is the PromQL regex that's used to filter the
	// site configuration client metrics to only those emitted by the Prometheus scrape job(s) that were interested in.
	//
	// Example: `.*gitserver`
	JobFilterRegex string
}

// NewSiteConfigurationClientMetricsGroup creates a group containing site configuration fetching latency statistics for the service
// specified in the given options.
func NewSiteConfigurationClientMetricsGroup(opts SiteConfigurationMetricsOptions, owner monitoring.ObservableOwner) monitoring.Group {
	opts.HumanServiceName = strcase.ToSnake(opts.HumanServiceName)

	jobFilter := fmt.Sprintf("job=~`%s`", opts.JobFilterRegex)

	metric := func(base string, labelFilters ...string) string {
		metric := base

		instanceLabelFilter := fmt.Sprintf("instance=~`%s`", opts.InstanceFilterRegex)

		labelFilters = append(labelFilters, instanceLabelFilter)

		if len(labelFilters) > 0 {
			metric = fmt.Sprintf("%s{%s}", metric, strings.Join(labelFilters, ","))
		}

		return metric
	}

	return monitoring.Group{
		Title:  "Site configuration client update latency",
		Hidden: true,
		Rows: []monitoring.Row{
			{
				{
					Name:           fmt.Sprintf("%s_site_configuration_duration_since_last_successful_update_by_instance", opts.HumanServiceName),
					Description:    "duration since last successful site configuration update (by instance)",
					Query:          metric("src_conf_client_time_since_last_successful_update_seconds", jobFilter),
					Panel:          monitoring.Panel().LegendFormat("{{instance}}").Unit(monitoring.Seconds),
					Owner:          owner,
					NoAlert:        true,
					Interpretation: fmt.Sprintf("The duration since the configuration client used by the %q service last successfully updated its site configuration. Long durations could indicate issues updating the site configuration.", opts.HumanServiceName),
				},
				{
					Name:        fmt.Sprintf("%s_site_configuration_duration_since_last_successful_update_by_instance", opts.HumanServiceName),
					Description: fmt.Sprintf("maximum duration since last successful site configuration update (all %q instances)", opts.HumanServiceName),
					Query:       fmt.Sprintf("max(max_over_time(%s[1m]))", metric("src_conf_client_time_since_last_successful_update_seconds", jobFilter)),
					Panel:       monitoring.Panel().Unit(monitoring.Seconds),
					Owner:       owner,
					Critical:    monitoring.Alert().GreaterOrEqual((5 * time.Minute).Seconds()),
					NextSteps: fmt.Sprintf(`
								- This indicates that one or more %q instances have not successfully updated the site configuration in over 5 minutes. This could be due to networking issues between services or problems with the site configuration service itself.
								- Check for relevant errors in the %q logs, as well as frontend's logs.
							`, opts.HumanServiceName, opts.HumanServiceName),
				},
			},
		},
	}
}
