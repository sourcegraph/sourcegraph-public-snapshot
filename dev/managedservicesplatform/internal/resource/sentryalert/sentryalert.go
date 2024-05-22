package sentryalert

import (
	"encoding/json"

	"github.com/aws/constructs-go/constructs/v10"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	sentryproject "github.com/sourcegraph/managed-services-platform-cdktf/gen/sentry/project"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/sentry/issuealert"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Output struct{}

// Config for a Sentry issue alert.
//
// [Sentry] and [Terraform] document each field
//
// [Sentry]: https://docs.sentry.io/api/alerts/create-an-issue-alert-rule-for-a-project/
// [Terraform]: https://registry.terraform.io/providers/jianyuan/sentry/latest/docs/resources/issue_alert
type Config struct {
	// ID of the issue alert. Must be unique
	ID string
	// SentryProject is the project to set the alert on
	SentryProject sentryproject.Project
	// AlertConfig is the configuration for the Sentry issue alert rule
	AlertConfig AlertConfig
}

// AlertConfig is the configuration for the Sentry issue alert rule
type AlertConfig struct {
	// Name the name of the alert
	Name string
	// Frequency determines how often to perform the actions for an issue, in minutes. (valid range 5-43200)
	Frequency float64
	// ActionMatch trigger actions when an event is captured by Sentry and `ActionMatch` of the
	// specified `Conditions` happen
	ActionMatch ActionMatch
	// Conditions which must be satisfied for actions to trigger.
	// `ActionMatch` determines whether all or any must be true.
	Conditions []Condition
	// Actions the list of actions to run when the conditions match
	Actions []Action

	// FilterMatch (optional) determines which filters need to be true before actions are executed.
	// Required when a value is provided for `Filters`
	FilterMatch FilterMatch
	// Filters determines if a rule fires after the necessary conditions have been met
	Filters []Filter
}

func (a AlertConfig) Validate() error {
	if a.Name == "" {
		return errors.New("Name is required")
	}

	if a.ActionMatch == "" {
		return errors.New("ActionMatch is required")
	}

	if len(a.Actions) == 0 {
		return errors.New("Actions is required with at least one action specified")
	}

	if a.Frequency == 0 {
		return errors.New("Frequency is required and must be a value between 5-43200")
	}

	if a.Frequency < 5 || a.Frequency > 43200 {
		return errors.New("Frequency must be between 5-43200")
	}

	if a.Filters != nil && a.FilterMatch == "" {
		return errors.New("FilterMatch is required when Filters are provided")
	}

	return nil
}

type ActionMatch string

const (
	ActionMatchAll ActionMatch = "all"
	ActionMatchAny ActionMatch = "any"
)

type ConditionID string

const (
	// FirstSeenEventCondition A new issue is created
	FirstSeenEventCondition ConditionID = "sentry.rules.conditions.first_seen_event.FirstSeenEventCondition"
	// RegressionEventCondition The issue changes state from resolved to unresolved
	RegressionEventCondition ConditionID = "sentry.rules.conditions.regression_event.RegressionEventCondition"
	// EventFrequencyCondition The issue is seen more than `value` times in `interval` (valid values are 1m, 5m, 15m, 1h, 1d, 1w and 30d)
	EventFrequencyCondition ConditionID = "sentry.rules.conditions.event_frequency.EventFrequencyCondition"
	// EventUniqueUserFrequencyCondition The issue is seen by more than `value` users in `interval` (valid values are 1m, 5m, 15m, 1h, 1d, 1w and 30d)
	EventUniqueUserFrequencyCondition ConditionID = "sentry.rules.conditions.event_frequency.EventUniqueUserFrequencyCondition"
	// EventFrequencyPercentCondition The issue affects more than `value` percent (integer 0 to 100) of sessions in `interval` (valid values are 5m, 10m, 30m, and 1h)
	EventFrequencyPercentCondition ConditionID = "sentry.rules.conditions.event_frequency.EventFrequencyPercentCondition"
)

// Condition checked `WHEN` an event is captured by Sentry.
//
// Multiple Conditions can be composed together in an alert requiring `all` or `any` to pass
type Condition struct {
	// ID represents the type of condition
	ID ConditionID `json:"id"`
	// Value is an integer threshold used by certain conditions
	Value *int `json:"value,omitempty"`
	// Interval is a threshold used by certain conditions.
	Interval *string `json:"interval,omitempty"`
}

type ActionID string

const (
	// Send a Slack notification
	SlackNotifyServiceAction ActionID = "sentry.integrations.slack.notify_action.SlackNotifyServiceAction"
	// Send an Opsgenie notification
	OpsgenieNotifyTeamAction ActionID = "sentry.integrations.opsgenie.notify_action.OpsgenieNotifyTeamAction"
)

type Action struct {
	// ID represents the type of action
	ID ActionID
	// ActionParameters define parameters unique to specific actions documented here [body parameters > actions]
	//
	// [body parameters > actions]: https://docs.sentry.io/api/alerts/create-an-issue-alert-rule-for-a-project/
	ActionParameters map[string]any
}

// Custom marshalling to flatten Action struct
func (a Action) MarshalJSON() ([]byte, error) {
	// Create a new map to hold the flattened JSON representation
	flattened := make(map[string]interface{})

	// Copy the fields from the Action struct to the flattened map
	flattened["id"] = a.ID
	for key, value := range a.ActionParameters {
		flattened[key] = value
	}

	// Marshal the flattened map to JSON
	return json.Marshal(flattened)
}

type FilterMatch string

const (
	FilterMatchAll  FilterMatch = "all"
	FilterMatchAny  FilterMatch = "any"
	FilterMatchNone FilterMatch = "none"
)

type FilterID string

const (
	// AgeComparisonFilter the issue `comparison_type` (older, newer) than `value` of `time`
	AgeComparisonFilter FilterID = "sentry.rules.filters.age_comparison.AgeComparisonFilter"
)

type Filter struct {
	// ID represents the type of filter
	ID FilterID

	// FilterParameters define parameters unique to specific filters documented here [body parameters > filters]
	//
	// [body parameters > filters]: https://docs.sentry.io/api/alerts/create-an-issue-alert-rule-for-a-project/
	FilterParameters map[string]any
}

// Custom marshalling to flatten Filter struct
func (f Filter) MarshalJSON() ([]byte, error) {
	// Create a new map to hold the flattened JSON representation
	flattened := make(map[string]interface{})

	// Copy the fields from the Filter struct to the flattened map
	flattened["id"] = f.ID
	for key, value := range f.FilterParameters {
		flattened[key] = value
	}

	// Marshal the flattened map to JSON
	return json.Marshal(flattened)
}

func New(scope constructs.Construct, id resourceid.ID, config Config) (*Output, error) {
	if err := config.AlertConfig.Validate(); err != nil {
		return nil, errors.Wrap(err, "sentry alert validation failed")
	}

	conditions, err := json.Marshal(config.AlertConfig.Conditions)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal AlertConfig.Conditions")
	}

	actions, err := json.Marshal(config.AlertConfig.Actions)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal AlertConfig.Actions")
	}

	// Filters is optional
	filters := []byte{}
	if len(config.AlertConfig.Filters) > 0 {
		filters, err = json.Marshal(config.AlertConfig.Filters)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal AlertConfig.Filters")
		}
	}

	_ = issuealert.NewIssueAlert(scope, id.TerraformID("alert"), &issuealert.IssueAlertConfig{
		Organization: config.SentryProject.Organization(),
		Project:      config.SentryProject.Slug(),
		Name:         pointers.Ptr(config.AlertConfig.Name),
		ActionMatch:  pointers.Ptr(string(config.AlertConfig.ActionMatch)),
		Conditions:   pointers.Ptr(string(conditions)),
		Actions:      pointers.Ptr(string(actions)),
		Frequency:    pointers.Ptr(config.AlertConfig.Frequency),
		FilterMatch: func() *string {
			// Sentry will default to All when none is set which causes issues with the Terraform provider
			if config.AlertConfig.FilterMatch == "" {
				return pointers.Ptr(string(FilterMatchAll))
			}
			return pointers.Ptr(string(config.AlertConfig.FilterMatch))
		}(),
		Filters: pointers.NonZeroPtr(string(filters)),
	})

	return &Output{}, nil
}
