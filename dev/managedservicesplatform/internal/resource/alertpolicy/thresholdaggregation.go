package alertpolicy

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringalertpolicy"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func newThresholdAggregationCondition(config *Config) (*monitoringalertpolicy.MonitoringAlertPolicyConditions, error) {
	if len(config.ThresholdAggregation.Filters) == 0 {
		return nil, errors.New("must specify at least one filter for threshold aggregation")
	}

	if _, ok := config.ThresholdAggregation.Filters["metric.type"]; !ok {
		return nil, errors.New("must specify filter for `metric.type`")
	}

	// Set some defaults for threshold aggregations
	switch config.ThresholdAggregation.ResourceKind {
	case CloudRunService:
		config.ThresholdAggregation.GroupByFields = append(
			[]string{"resource.label.revision_name"},
			config.ThresholdAggregation.GroupByFields...)

	case CloudSQLDatabase:
		config.ThresholdAggregation.GroupByFields = append(
			[]string{"resource.label.database"},
			config.ThresholdAggregation.GroupByFields...)

	case CloudRunJob, CloudRedis, URLUptime, CloudSQL, "":
		// No defaults

	default:
		return nil, errors.Newf("invalid service kind %q", config.ThresholdAggregation.ResourceKind)
	}

	if config.ThresholdAggregation.Comparison == "" {
		config.ThresholdAggregation.Comparison = ComparisonGT
	}

	if config.ThresholdAggregation.Duration == "" {
		config.ThresholdAggregation.Duration = "0s"
	}
	return &monitoringalertpolicy.MonitoringAlertPolicyConditions{
		DisplayName: pointers.Ptr(config.Name),
		ConditionThreshold: &monitoringalertpolicy.MonitoringAlertPolicyConditionsConditionThreshold{
			Aggregations: []monitoringalertpolicy.MonitoringAlertPolicyConditionsConditionThresholdAggregations{
				{
					AlignmentPeriod:    pointers.Ptr(config.ThresholdAggregation.Period),
					PerSeriesAligner:   pointers.NonZeroPtr(string(config.ThresholdAggregation.Aligner)),
					CrossSeriesReducer: pointers.NonZeroPtr(string(config.ThresholdAggregation.Reducer)),
					GroupByFields:      pointers.Ptr(pointers.Slice(config.ThresholdAggregation.GroupByFields)),
				},
			},
			Comparison:     pointers.Ptr(string(config.ThresholdAggregation.Comparison)),
			Duration:       pointers.Ptr(config.ThresholdAggregation.Duration),
			Filter:         pointers.Ptr(buildThresholdAggregationFilter(config)),
			ThresholdValue: pointers.Float64(config.ThresholdAggregation.Threshold),
			Trigger: func() *monitoringalertpolicy.MonitoringAlertPolicyConditionsConditionThresholdTrigger {
				switch config.ThresholdAggregation.Trigger {
				case TriggerKindAllInViolation:
					return &monitoringalertpolicy.MonitoringAlertPolicyConditionsConditionThresholdTrigger{
						Percent: pointers.Float64(100),
					}

				case TriggerKindAnyViolation:
					fallthrough
				default:
					return &monitoringalertpolicy.MonitoringAlertPolicyConditionsConditionThresholdTrigger{
						Count: pointers.Float64(1),
					}
				}
			}(),
		},
	}, nil
}

// buildThresholdAggregationFilter creates the Filter string for a ThresholdAggregation alert condition
func buildThresholdAggregationFilter(config *Config) string {
	filters := make([]string, 0)
	for key, val := range config.ThresholdAggregation.Filters {
		filters = append(filters, fmt.Sprintf(`%s = "%s"`, key, val))
	}

	// Sort to ensure stable output for testing, because
	// config.ThresholdAggregation.Filters is a map.
	sort.Strings(filters)

	switch config.ThresholdAggregation.ResourceKind {
	case CloudRunService:
		filters = append(filters,
			`resource.type = "cloud_run_revision"`,
			fmt.Sprintf(`resource.labels.service_name = starts_with("%s")`, config.ThresholdAggregation.ResourceName),
		)
	case CloudRunJob:
		filters = append(filters,
			`resource.type = "cloud_run_job"`,
			fmt.Sprintf(`resource.labels.job_name = starts_with("%s")`, config.ThresholdAggregation.ResourceName),
		)
	case CloudRedis:
		filters = append(filters,
			`resource.type = "redis_instance"`,
			fmt.Sprintf(`resource.labels.instance_id = "%s"`, config.ThresholdAggregation.ResourceName),
		)
	case CloudSQL:
		filters = append(filters,
			`resource.type = "cloudsql_database"`,
			fmt.Sprintf(`resource.labels.database_id = "%s"`, config.ThresholdAggregation.ResourceName))
	case CloudSQLDatabase:
		filters = append(filters,
			`resource.type = "cloudsql_instance_database"`,
			fmt.Sprintf(`resource.labels.resource_id = "%s"`, config.ThresholdAggregation.ResourceName))
	case URLUptime:
		filters = append(filters,
			`resource.type = "uptime_url"`,
			fmt.Sprintf(`metric.labels.check_id = "%s"`, config.ThresholdAggregation.ResourceName),
		)
	}

	return strings.Join(filters, " AND ")
}
