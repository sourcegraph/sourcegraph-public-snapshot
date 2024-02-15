package alertpolicy

import (
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
			Aggregations:   config.ThresholdAggregation.buildThresholdAggregations(),
			Comparison:     pointers.Ptr(string(config.ThresholdAggregation.Comparison)),
			Duration:       pointers.Ptr(config.ThresholdAggregation.Duration),
			Filter:         pointers.Ptr(config.ThresholdAggregation.buildFilter()),
			ThresholdValue: pointers.Float64(config.ThresholdAggregation.Threshold),
			Trigger:        config.ThresholdAggregation.buildThresholdTrigger(),
		},
	}, nil
}
