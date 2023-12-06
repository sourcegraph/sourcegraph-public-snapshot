package monitoringalertpolicy

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringalertpolicy"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Aligner string

const (
	MonitoringAlignNone          Aligner = "ALIGN_NONE"
	MonitoringAlignDelta         Aligner = "ALIGN_DELTA"
	MonitoringAlignRate          Aligner = "ALIGN_RATE"
	MonitoringAlignInterpolate   Aligner = "ALIGN_INTERPOLATE"
	MonitoringAlignNextOrder     Aligner = "ALIGN_NEXT_ORDER"
	MonitoringAlignMin           Aligner = "ALIGN_MIN"
	MonitoringAlignMax           Aligner = "ALIGN_MAX"
	MonitoringAlignMean          Aligner = "ALIGN_MEAN"
	MonitoringAlignCount         Aligner = "ALIGN_COUNT"
	MonitoringAlignSum           Aligner = "ALIGN_SUM"
	MonitoringAlignStddev        Aligner = "ALIGN_STDDEV"
	MonitoringAlignCountTrue     Aligner = "ALIGN_COUNT_TRUE"
	MonitoringAlignCountFalse    Aligner = "ALIGN_COUNT_FALSE"
	MonitoringAlignFractionTrue  Aligner = "ALIGN_FRACTION_TRUE"
	MonitoringAlignPercentile99  Aligner = "ALIGN_PERCENTILE_99"
	MonitoringAlignPercentile95  Aligner = "ALIGN_PERCENTILE_95"
	MonitoringAlignPercentile50  Aligner = "ALIGN_PERCENTILE_50"
	MonitoringAlignPercentile05  Aligner = "ALIGN_PERCENTILE_05"
	MonitoringAlignPercentChange Aligner = "ALIGN_PERCENT_CHANGE"
)

type Reducer string

const (
	MonitoringReduceNone         Reducer = "REDUCE_NONE"
	MonitoringReduceMean         Reducer = "REDUCE_MEAN"
	MonitoringReduceMin          Reducer = "REDUCE_MIN"
	MonitoringReduceMax          Reducer = "REDUCE_MAX"
	MonitoringReduceSum          Reducer = "REDUCE_SUM"
	MonitoringReduceStddev       Reducer = "REDUCE_STDDEV"
	MonitoringReduceCount        Reducer = "REDUCE_COUNT"
	MonitoringReduceCountTrue    Reducer = "REDUCE_COUNT_TRUE"
	MonitoringReduceCountFalse   Reducer = "REDUCE_COUNT_FALSE"
	MonitoringReduceFractionTrue Reducer = "REDUCE_FRACTION_TRUE"
	MonitoringReducePercentile99 Reducer = "REDUCE_PERCENTILE_99"
	MonitoringReducePercentile95 Reducer = "REDUCE_PERCENTILE_95"
	MonitoringReducePercentile50 Reducer = "REDUCE_PERCENTILE_50"
	MonitoringReducePercentile05 Reducer = "REDUCE_PERCENTILE_05"
)

type Comparison string

const (
	ComparisonGT Comparison = "COMPARISON_GT"
	ComparisonLT Comparison = "COMPARISON_LT"
)

// ThresholdAggregation for alerting when a metric exceeds a defined threshold
//
// Must specify a `metric.type` filter. Additional filters are optional.
//
// All filters are joined with ` AND `
type ThresholdAggregation struct {
	Filters      map[string]string
	GroupByField string
	Comparison   Comparison
	Aligner      Aligner
	Reducer      Reducer
	Period       string
	Threshold    float64
	Duration     string
}

// ResponseCodeMetric for alerting when the numer of a certain response code exceeds a threshold
//
// Must specify either `Code` (e.g. 404) or `CodeClass` (e.g. 4xx)
//
// `ExcludeCodes` allows filtering out specific response codes from the `CodeClass`
type ResponseCodeMetric struct {
	Code         *int
	CodeClass    *string
	ExcludeCodes []string
	Ratio        float64
	Duration     *string
}

// Config for a Monitoring Alert Policy
// Must define either `ThresholdAggregation` or `ResponseCodeMetric`
type Config struct {
	// A unique identifier
	ID          string
	Name        string
	Description *string
	ProjectID   string
	// Name of the service/job to filter the alert on
	ServiceName string
	ServiceKind spec.ServiceKind

	ThresholdAggregation *ThresholdAggregation
	ResponseCodeMetric   *ResponseCodeMetric
}

type Output struct {
}

func New(scope constructs.Construct, id resourceid.ID, config *Config) (*Output, error) {
	if config.ThresholdAggregation == nil && config.ResponseCodeMetric == nil {
		return nil, errors.New("Must provide either SingleMetric or ResponseCodeMetric config")
	}

	if config.ThresholdAggregation != nil && config.ResponseCodeMetric != nil {
		return nil, errors.New("Must provide either SingleMetric or ResponseCodeMetric config, not both")
	}

	if config.ThresholdAggregation != nil {
		if config.ServiceKind == spec.ServiceKindService && config.ThresholdAggregation.GroupByField != "" {
			return nil, errors.New("Specifying GroupByField is invalid for Cloud Run Services as the default `resource.label.revision_name` is enforced")
		}

		if len(config.ThresholdAggregation.Filters) == 0 {
			return nil, errors.New("must specify at least one filter for threshold aggregation")
		}

		if _, ok := config.ThresholdAggregation.Filters["metric.type"]; !ok {
			return nil, errors.New("must specify filter for `metric.type`")
		}
		return thresholdAggregation(scope, id, config)
	}
	return responseCodeMetric(scope, id, config)
}

// threshholdAggregation defines a monitoring alert policy based on a single metric threshold
func thresholdAggregation(scope constructs.Construct, id resourceid.ID, config *Config) (*Output, error) {
	// Set some defaults
	if config.ServiceKind == spec.ServiceKindService {
		// For Cloud Run Services we need to group by the revision_name
		config.ThresholdAggregation.GroupByField = "resource.label.revision_name"
	} else if config.ServiceKind == spec.ServiceKindJob {
		// No default for this
	} else {
		return nil, errors.Newf("invalid service kind %q", config.ServiceKind)
	}

	if config.ThresholdAggregation.Comparison == "" {
		config.ThresholdAggregation.Comparison = ComparisonGT
	}

	if config.ThresholdAggregation.Duration == "" {
		config.ThresholdAggregation.Duration = "0s"
	}

	// Optional on Jobs
	var group_by *[]*string
	if config.ThresholdAggregation.GroupByField != "" {
		group_by = pointers.Ptr([]*string{pointers.Ptr(config.ThresholdAggregation.GroupByField)})
	}

	_ = monitoringalertpolicy.NewMonitoringAlertPolicy(scope,
		id.TerraformID(config.ID), &monitoringalertpolicy.MonitoringAlertPolicyConfig{
			Project:     pointers.Ptr(config.ProjectID),
			DisplayName: pointers.Ptr(config.Name),
			Documentation: &monitoringalertpolicy.MonitoringAlertPolicyDocumentation{
				Content:  config.Description,
				MimeType: pointers.Ptr("text/markdown"),
			},
			Combiner: pointers.Ptr("OR"),
			Conditions: []monitoringalertpolicy.MonitoringAlertPolicyConditions{
				{
					DisplayName: pointers.Ptr(config.Name),
					ConditionThreshold: &monitoringalertpolicy.MonitoringAlertPolicyConditionsConditionThreshold{
						Aggregations: []monitoringalertpolicy.MonitoringAlertPolicyConditionsConditionThresholdAggregations{
							{
								AlignmentPeriod:    pointers.Ptr(config.ThresholdAggregation.Period),
								PerSeriesAligner:   pointers.Ptr(string(config.ThresholdAggregation.Aligner)),
								CrossSeriesReducer: pointers.Ptr(string(config.ThresholdAggregation.Reducer)),
								GroupByFields:      group_by,
							},
						},
						Comparison:     pointers.Ptr(string(config.ThresholdAggregation.Comparison)),
						Duration:       pointers.Ptr(config.ThresholdAggregation.Duration),
						Filter:         pointers.Ptr(filter(config)),
						ThresholdValue: pointers.Float64(config.ThresholdAggregation.Threshold),
						Trigger: &monitoringalertpolicy.MonitoringAlertPolicyConditionsConditionThresholdTrigger{
							Count: pointers.Float64(1),
						},
					},
				},
			},
			AlertStrategy: &monitoringalertpolicy.MonitoringAlertPolicyAlertStrategy{
				AutoClose: pointers.Ptr("604800s"),
			},
		})
	return &Output{}, nil
}

// filter creates the Filter string for a ThresholdAggregation monitoring alert policy
func filter(config *Config) string {
	filters := make([]string, 0)
	for key, val := range config.ThresholdAggregation.Filters {
		filters = append(filters, fmt.Sprintf(`%s = "%s"`, key, val))
	}

	// Sort to ensure stable output for testing
	// This code runs so infreuqently that sort performance is not a concern and there are only a couple elements
	sort.Strings(filters)

	if config.ServiceKind == spec.ServiceKindService {
		filters = append(filters, `resource.type = "cloud_run_revision"`)
		filters = append(filters, fmt.Sprintf(`resource.labels.service_name = "%s"`, config.ServiceName))
	} else if config.ServiceKind == spec.ServiceKindJob {
		filters = append(filters, `resource.type = "cloud_run_job"`)
		filters = append(filters, fmt.Sprintf(`resource.labels.job_name = "%s"`, config.ServiceName))
	}

	return strings.Join(filters, " AND ")
}

// responseCodeMetric defines the MonitoringAlertPolicy for response code metrics
// Supports a single Code e.g. 404 or an entire Code Class e.g. 4xx
// Optionally when using a Code Class, codes to exclude can be defined
func responseCodeMetric(scope constructs.Construct, id resourceid.ID, config *Config) (*Output, error) {
	query := responseCodeBuilder(config)

	if config.ResponseCodeMetric.Duration == nil {
		config.ResponseCodeMetric.Duration = pointers.Ptr("60s")
	}

	_ = monitoringalertpolicy.NewMonitoringAlertPolicy(scope,
		id.TerraformID(config.ID), &monitoringalertpolicy.MonitoringAlertPolicyConfig{
			Project:     pointers.Ptr(config.ProjectID),
			DisplayName: pointers.Ptr(fmt.Sprintf("High Ratio of %s Responses", config.Name)),
			Documentation: &monitoringalertpolicy.MonitoringAlertPolicyDocumentation{
				Content:  config.Description,
				MimeType: pointers.Ptr("text/markdown"),
			},
			Combiner: pointers.Ptr("OR"),
			Conditions: []monitoringalertpolicy.MonitoringAlertPolicyConditions{
				{
					DisplayName: pointers.Ptr(fmt.Sprintf("High Ratio of %s Responses", config.Name)),
					ConditionMonitoringQueryLanguage: &monitoringalertpolicy.MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguage{
						Query:    pointers.Ptr(query),
						Duration: config.ResponseCodeMetric.Duration,
						Trigger: &monitoringalertpolicy.MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTrigger{
							Count: pointers.Float64(1),
						},
					},
				},
			},
			AlertStrategy: &monitoringalertpolicy.MonitoringAlertPolicyAlertStrategy{
				AutoClose: pointers.Ptr("604800s"),
			},
		})
	return &Output{}, nil
}

// responseCodeBuilder builds the MQL for a response code metric alert
func responseCodeBuilder(config *Config) string {
	var builder strings.Builder

	builder.WriteString("fetch cloud_run_revision\n")
	builder.WriteString("| metric 'run.googleapis.com/request_count'\n")
	builder.WriteString("| group_by 15s, [value_request_count_aggregate: aggregate(value.request_count)]\n")
	builder.WriteString("| every 15s\n")
	builder.WriteString("| {\n")
	if config.ResponseCodeMetric.CodeClass != nil {
		builder.WriteString("  group_by [metric.response_code, metric.response_code_class],\n")
	} else {
		builder.WriteString("  group_by [metric.response_code],\n")
	}
	builder.WriteString("  [response_code_count_aggregate: aggregate(value_request_count_aggregate)]\n")
	if config.ResponseCodeMetric.Code != nil {
		builder.WriteString(fmt.Sprintf("  | filter (metric.response_code = '%d')\n", *config.ResponseCodeMetric.Code))
	} else {
		builder.WriteString(fmt.Sprintf("  | filter (metric.response_code_class = '%s')\n", *config.ResponseCodeMetric.CodeClass))
	}
	if config.ResponseCodeMetric.ExcludeCodes != nil && len(config.ResponseCodeMetric.ExcludeCodes) > 0 {
		for _, code := range config.ResponseCodeMetric.ExcludeCodes {
			builder.WriteString(fmt.Sprintf("  | filter (metric.response_code != '%s')\n", code))
		}
	}
	builder.WriteString("; group_by [],\n")
	builder.WriteString("  [value_request_count_aggregate_aggregate: aggregate(value_request_count_aggregate)]\n")
	builder.WriteString("}\n")
	builder.WriteString("| join\n")
	builder.WriteString("| value [response_code_ratio: val(0) / val(1)]\n")
	builder.WriteString(fmt.Sprintf("| condition gt(val(), %s)\n", strconv.FormatFloat(config.ResponseCodeMetric.Ratio, 'f', -1, 64)))
	return builder.String()
}
