package alertpolicy

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringalertpolicy"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringnotificationchannel"

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
// All filters are joined with ` AND `
//
// GroupByFields is an optional field specifying time series labels to aggregate:
//   - For services it defaults to `["resource.label.revision_name"]`; additional fields are appended
//   - For jobs there is no default
type ThresholdAggregation struct {
	ConditionBuilder

	// Threshold
	Threshold  float64
	Duration   string
	Comparison Comparison
}

// MetricAbsence for alerting when a metric is missing for a defined amount
// of time.
type MetricAbsence struct {
	ConditionBuilder

	// Duration (must be in seconds)
	Duration string
}

// ResponseCodeMetric for alerting when the number of a certain response code exceeds a threshold
//
// Must specify either `Code` (e.g. 404) or `CodeClass` (e.g. 4xx)
//
// `ExcludeCodes` allows filtering out specific response codes from the `CodeClass`
type ResponseCodeMetric struct {
	Code            *int
	CodeClass       *string
	ExcludeCodes    []string
	Ratio           float64
	DurationMinutes *uint
}

// DescriptionSuffix points to the service page and environment anchor expected to be
// generated at https://handbook.sourcegraph.com/departments/engineering/teams/core-services/managed-services/platform/,
// and should be added as a suffix to all alert descriptions.
func DescriptionSuffix(s spec.ServiceSpec, environmentID string) string {
	return fmt.Sprintf(`See %s -> **%s** for service and infrastructure access details for this environment.
If you need additional assistance, reach out to #discuss-core-services.`,
		s.GetHandbookPageURL(), environmentID)
}

type NotificationChannels map[spec.AlertSeverityLevel][]monitoringnotificationchannel.MonitoringNotificationChannel

// Config for a Monitoring Alert Policy
type Config struct {
	Service       spec.ServiceSpec
	EnvironmentID string
	ProjectID     string

	// ID is unique identifier of the alert policy
	ID string
	// Name is a human-readable name for the alert policy
	Name string
	// Description is a Markdown-format description for the alert policy. Some
	// unified context will be included as well, including links to the service
	// handbook page and so on.
	Description string
	// Severity is the desired level for this alert. It is used to choose from
	// the provided set of NotificationChannels.
	//
	// If not provided, SeverityLevelWarning is used.
	Severity spec.AlertSeverityLevel

	// NotificationChannels to choose from for subscribing on this alert
	NotificationChannels NotificationChannels

	// Only one of the following can be set.
	ThresholdAggregation *ThresholdAggregation
	MetricAbsence        *MetricAbsence
	ResponseCodeMetric   *ResponseCodeMetric
	CustomAlert          *spec.CustomAlertCondition
}

// makeDocsSubject prefixes the name with the service and environment for ease
// of reading in various feeds.
func (c Config) makeDocsSubject() string {
	return fmt.Sprintf("%s (%s): %s",
		c.Service.GetName(), c.EnvironmentID, c.Name)
}

type Output struct {
	AlertPolicy monitoringalertpolicy.MonitoringAlertPolicy
}

func New(scope constructs.Construct, id resourceid.ID, config *Config) (*Output, error) {
	if err := onlyOneNonNil([]any{config.ThresholdAggregation, config.ResponseCodeMetric, config.MetricAbsence, config.CustomAlert}); err != nil {
		return nil, err
	}

	// Universal alert description addendum
	if config.Service.ID == "" {
		return nil, errors.New("Service is required")
	}
	if config.Description == "" {
		return nil, errors.New("Description is required")
	} else {
		config.Description = fmt.Sprintf("%s\n\n%s",
			config.Description,
			DescriptionSuffix(config.Service, config.EnvironmentID))
	}

	// Set default
	if config.Severity == "" {
		config.Severity = spec.AlertSeverityLevelWarning
	}

	// Labels for the alert
	labels := map[string]*string{
		"source": pointers.Ptr("managed-services-platform"),

		"msp_alert_id":       pointers.Ptr(config.ID),
		"msp_service_id":     pointers.Ptr(config.Service.ID),
		"msp_environment_id": pointers.Ptr(config.EnvironmentID),
	}

	// Build the condition, each of which may have special handling and additional
	// defaults based on recommendations and other best practices
	var condition *monitoringalertpolicy.MonitoringAlertPolicyConditions
	switch {
	case config.ThresholdAggregation != nil:
		var err error
		condition, err = newThresholdAggregationCondition(config)
		if err != nil {
			return nil, errors.Wrap(err, "newThresholdAggregationCondition")
		}
		if config.ThresholdAggregation.ResourceKind != "" {
			labels["resource_kind"] = pointers.Ptr(string(config.ThresholdAggregation.ResourceKind))
		}

	case config.MetricAbsence != nil:
		condition = newMetricAbsenceCondition(config)
		if config.MetricAbsence.ResourceKind != "" {
			labels["resource_kind"] = pointers.Ptr(string(config.MetricAbsence.ResourceKind))
		}

	case config.ResponseCodeMetric != nil:
		condition = newResponseCodeMetricCondition(config)

	case config.CustomAlert != nil:
		condition = newCustomAlertCondition(config)
	default:
		return nil, errors.New("no condition configuration provided")
	}

	// Build the final alert policy
	alert := monitoringalertpolicy.NewMonitoringAlertPolicy(scope, id.TerraformID(config.ID),
		&monitoringalertpolicy.MonitoringAlertPolicyConfig{
			Project:     pointers.Ptr(config.ProjectID),
			DisplayName: pointers.Ptr(config.Name),
			Documentation: &monitoringalertpolicy.MonitoringAlertPolicyDocumentation{
				Subject:  pointers.Ptr(config.makeDocsSubject()),
				Content:  pointers.Ptr(config.Description),
				MimeType: pointers.Ptr("text/markdown"),
			},
			UserLabels: &labels,

			// Notification strategy
			AlertStrategy: &monitoringalertpolicy.MonitoringAlertPolicyAlertStrategy{
				AutoClose: pointers.Ptr("86400s"), // 24 hours
			},
			NotificationChannels: notificationChannelIDs(config.NotificationChannels[config.Severity]),
			// Possible values: ["CRITICAL", "ERROR", "WARNING"]
			Severity: pointers.Ptr(string(config.Severity)),

			// Conditions
			Combiner:   pointers.Ptr("OR"),
			Conditions: []*monitoringalertpolicy.MonitoringAlertPolicyConditions{condition},
		})

	return &Output{
		AlertPolicy: alert,
	}, nil
}

func onlyOneNonNil(options []any) error {
	var types []string
	var nonNil []string
	for _, o := range options {
		t := fmt.Sprintf("%T", o)
		types = append(types, t)
		// we must reflect as []any boxes the pointer
		if !reflect.ValueOf(o).IsNil() {
			nonNil = append(nonNil, t)
		}
	}
	if len(nonNil) != 1 {
		return errors.Newf("exactly one of [ %s ] must be specified, found: [ %s ]",
			strings.Join(types, ", "), strings.Join(nonNil, ", "))
	}
	return nil
}
