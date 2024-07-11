package spec

import (
	"strings"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var codeClassPattern = regexp.MustCompile(`\dx+`)
var customAlertNamePattern = regexp.MustCompile(`^[-A-Za-z0-9 ]+$`)

// AlertSeverityLevel is used for configuration notification urgency for monitoring alerts.
type AlertSeverityLevel string

const (
	AlertSeverityLevelWarning  AlertSeverityLevel = "WARNING"
	AlertSeverityLevelCritical AlertSeverityLevel = "CRITICAL"
)

type MonitoringSpec struct {
	// Alerts is a list of alert configurations for the deployment
	Alerts MonitoringAlertsSpec `yaml:"alerts"`
	// Nobl9 determines whether to provision a Nobl9 project.
	// Currently for trial purposes only
	Nobl9 *bool `yaml:"nobl9,omitempty"`
}

func (s *MonitoringSpec) Validate() []error {
	if s == nil {
		return nil
	}
	var errs []error
	errs = append(errs, s.Alerts.Validate()...)
	return errs
}

type MonitoringAlertsSpec struct {
	ResponseCodeRatios []ResponseCodeRatioAlertSpec `yaml:"responseCodeRatios,omitempty"`
	CustomAlerts       []CustomAlert                `yaml:"customAlerts,omitempty"`
}

type ResponseCodeRatioAlertSpec struct {
	ID           string   `yaml:"id"`
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description,omitempty"`
	Code         *int     `yaml:"code,omitempty"`
	CodeClass    *string  `yaml:"codeClass,omitempty"`
	ExcludeCodes []string `yaml:"excludeCodes,omitempty"`
	// DurationMinutes is the time in minutes the query must violate the threshold
	// to trigger the alert. Defaults to one minute.
	DurationMinutes *uint   `yaml:"durationMinutes,omitempty"`
	Ratio           float64 `yaml:"ratio"`
}

func (s *MonitoringAlertsSpec) Validate() []error {
	var errs []error
	// Use map to contain seen IDs to ensure uniqueness
	responceCodeRatioIDs := make(map[string]struct{})
	for _, r := range s.ResponseCodeRatios {
		if r.ID == "" {
			errs = append(errs, errors.New("responseCodeRatios[].id is required and cannot be empty"))
		}
		if _, ok := responceCodeRatioIDs[r.ID]; ok {
			errs = append(errs, errors.Newf("response code alert IDs must be unique, found duplicate ID: %s", r.ID))
		}
		responceCodeRatioIDs[r.ID] = struct{}{}
		errs = append(errs, r.Validate()...)
	}

	customAlertIDs := make(map[string]struct{})
	for _, c := range s.CustomAlerts {
		// Custom alert IDs are generated from the name during unmarshaling.
		if _, ok := customAlertIDs[c.ID]; ok {
			errs = append(errs, errors.Newf("custom alert names must be unique, found duplicate Name: `%s`", c.Name))
		}

		customAlertIDs[c.ID] = struct{}{}
		errs = append(errs, c.Validate()...)
	}
	return errs
}

func (r *ResponseCodeRatioAlertSpec) Validate() []error {
	var errs []error

	if r.ID == "" {
		errs = append(errs, errors.New("responseCodeRatios[].id is required"))
	}

	if r.Name == "" {
		errs = append(errs, errors.New("responseCodeRatios[].name is required"))
	}

	if r.Ratio < 0 || r.Ratio > 1 {
		errs = append(errs, errors.New("responseCodeRatios[].ratio must be between 0 and 1"))
	}

	if r.CodeClass != nil && r.Code != nil {
		errs = append(errs, errors.New("only one of responseCodeRatios[].code or responseCodeRatios[].codeClass should be specified"))
	}

	if r.Code != nil && *r.Code <= 0 {
		errs = append(errs, errors.New("responseCodeRatios[].code must be positive"))
	}

	if r.CodeClass != nil {
		if !codeClassPattern.MatchString(*r.CodeClass) {
			errs = append(errs, errors.New("responseCodeRatios[].codeClass must match the format Nxx (e.g. 4xx, 5xx)"))
		}
	}

	if r.DurationMinutes != nil {
		if *r.DurationMinutes > 1440 { // 24 hours
			errs = append(errs, errors.New("responseCodeRatios[].duration must be less than 1440 minutes"))
		}
	}

	return errs
}

type CustomAlertQueryType string

const (
	CustomAlertQueryTypeMQL    CustomAlertQueryType = "mql"
	CustomAlertQueryTypePromQL CustomAlertQueryType = "promql"
)

// CustomAlert defines a custom alert on a mql or promql query.
type CustomAlert struct {
	ID string `yaml:"-"` // set by custom unmarshaller
	// Human readable name of the alert.
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	// SeverityLevel is the severity level of the alert.
	// Valid values are "WARNING" and "CRITICAL".
	// Alerts with severity level WARNING are not sent to Opsgenie
	SeverityLevel AlertSeverityLevel   `yaml:"severityLevel"`
	Condition     CustomAlertCondition `yaml:"condition"`
}

type CustomAlertCondition struct {
	// Type describes the format of the configured query, and is one of `mql` or
	// `promql`:
	//
	// - MQL: GCP monitoring's custom 'Monitoring Query Language'. Good for more
	//   complicated alerts.
	// - PromQL: Prometheus's query language. Good for simple alerts that you
	//   already know how to express in PromQL.
	Type CustomAlertQueryType `yaml:"type"`
	// Query is the MQL or PromQL query to execute, based on your configured 'type'.
	// Refer to the following guides for more details on how to write queries
	// for each type:
	//
	// - PromQL: https://cloud.google.com/monitoring/promql
	// - MQL: https://cloud.google.com/monitoring/mql
	Query string `yaml:"query"`
	// DurationMinutes is the time in minutes the query must violate the threshold.
	// to trigger the alert. Defaults to one minute.
	DurationMinutes *uint `yaml:"durationMinutes,omitempty"`
}

func (c *CustomAlert) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Use an alias to prevent cyclical unmarshalling
	type CustomAlertAlias CustomAlert
	if err := unmarshal((*CustomAlertAlias)(c)); err != nil {
		return err
	}

	// Set ID to lower case name with spaces replaced with dashes
	c.ID = strings.ToLower(strings.Replace(c.Name, " ", "-", -1))
	return nil
}

func (c *CustomAlert) Validate() []error {
	var errs []error
	if !customAlertNamePattern.MatchString(c.ID) {
		errs = append(errs, errors.Newf("custom alert name must match the format %s, got: `%s`", customAlertNamePattern.String(), c.Name))
	}

	switch c.SeverityLevel {
	case AlertSeverityLevelWarning, AlertSeverityLevelCritical:
		break
	default:
		errs = append(errs, errors.New("customAlerts[].severityLevel must be either `WARNING` or `CRITICAL`"))
	}

	switch c.Condition.Type {
	case CustomAlertQueryTypeMQL, CustomAlertQueryTypePromQL:
		break
	default:
		errs = append(errs, errors.New("customAlerts[].condition.type must be either `mql` or `promql`"))
	}

	if c.Name == "" {
		errs = append(errs, errors.New("customAlerts[].name is required"))
	}

	if c.Condition.Query == "" {
		errs = append(errs, errors.New("customAlerts[].condition.query cannot be empty"))
	}

	if c.Condition.DurationMinutes != nil {
		if *c.Condition.DurationMinutes > 1440 { // 24 hours
			errs = append(errs, errors.New("customAlerts[].condition.durationMinutes must be less than 1440 minutes"))
		}
	}

	return errs
}
