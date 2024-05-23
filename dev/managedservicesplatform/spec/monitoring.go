package spec

import (
	"strings"
	"time"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var codeClassPattern = regexp.MustCompile(`\dx+`)
var customAlertNamePattern = regexp.MustCompile(`^[-A-Za-z0-9 ]+$`)

type SeverityLevel string

const (
	SeverityLevelWarning  = "WARNING"
	SeverityLevelCritical = "CRITICAL"
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
	Duration     *string  `yaml:"duration,omitempty"`
	Ratio        float64  `yaml:"ratio"`
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
		// custom alert ids are generated from the name during unmarshaling
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

	if r.Duration != nil {
		duration, err := time.ParseDuration(*r.Duration)
		if err != nil {
			errs = append(errs, errors.Wrap(err, "responseCodeRatios[].duration must be in the format of XXs"))
		} else if duration%time.Minute != 0 {
			errs = append(errs, errors.New("responseCodeRatios[].duration must be a multiple of 60s"))
		}
	}

	return errs
}

// CustomAlert
type CustomAlert struct {
	ID            string        `yaml:"-"`
	Name          string        `yaml:"name"`
	Description   string        `yaml:"description,omitempty"`
	SeverityLevel SeverityLevel `yaml:"severityLevel"`
	Duration      *string       `yaml:"duration,omitempty"`
	MQLQuery      *string       `yaml:"mqlQuery,omitempty"`
	PromQLQuery   *string       `yaml:"promQLQuery,omitempty"`
}

func (c *CustomAlert) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Use an alias to prevent cyclical unmarshaling
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
	case SeverityLevelWarning, SeverityLevelCritical:
		break
	default:
		errs = append(errs, errors.New("customAlerts[].severityLevel must be either `WARNING` or `CRITICAL`"))
	}

	if c.Name == "" {
		errs = append(errs, errors.New("customAlerts[].name is required"))
	}

	if c.MQLQuery != nil && c.PromQLQuery != nil {
		errs = append(errs, errors.New("only one of customAlerts[].mqlQuery or customAlerts[].promQLQuery should be specified"))
	}

	if c.MQLQuery == nil && c.PromQLQuery == nil {
		errs = append(errs, errors.New("one of customAlerts[].mqlQuery or customAlerts[].promQLQuery should be specified"))
	}

	if c.Duration != nil {
		duration, err := time.ParseDuration(*c.Duration)
		if err != nil {
			errs = append(errs, errors.Wrap(err, "customAlerts[].duration must be in the format of XXs"))
		} else if duration%time.Minute != 0 {
			errs = append(errs, errors.New("customAlerts[].duration must be a multiple of 60s"))
		}
	}

	return errs
}
