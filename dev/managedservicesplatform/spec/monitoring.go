package spec

import (
	"time"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var codeClassPattern = regexp.MustCompile(`\dx+`)

type MonitoringSpec struct {
	// Alerts is a list of alert configurations for the deployment
	Alerts MonitoringAlertsSpec `json:"alerts"`
}

func (s *MonitoringSpec) Validate() []error {
	var errs []error
	errs = append(errs, s.Alerts.Validate()...)
	return errs
}

type MonitoringAlertsSpec struct {
	ResponseCodeRatios []ResponseCodeRatioSpec `json:"responseCodeRatios"`
}

type ResponseCodeRatioSpec struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  *string  `json:"description,omitempty"`
	Code         *int     `json:"code,omitempty"`
	CodeClass    *string  `json:"codeClass,omitempty"`
	ExcludeCodes []string `json:"excludeCodes,omitempty"`
	Duration     *string  `json:"duration,omitempty"`
	Ratio        float64  `json:"ratio"`
}

func (s *MonitoringAlertsSpec) Validate() []error {
	var errs []error
	for _, r := range s.ResponseCodeRatios {
		errs = append(errs, r.Validate()...)
	}
	return errs
}

func (r *ResponseCodeRatioSpec) Validate() []error {
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
			errs = append(errs, errors.New("responseCodeRatios[].codeClass must match the format NXX (e.g. 4xx, 5xx)"))
		}
	}

	if r.Duration != nil {
		duration, err := time.ParseDuration(*r.Duration)
		if err != nil {
			errs = append(errs, errors.New("responseCodeRatios[].duration must be in the format of XXs"))
		} else if duration%time.Minute != 0 {
			errs = append(errs, errors.New("responseCodeRatios[].duration must be a multiple of 60s"))
		}
	}

	return errs
}
