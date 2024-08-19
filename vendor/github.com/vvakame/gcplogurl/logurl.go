package gcplogurl

import (
	"fmt"
	"net/url"
)

const logBaseURL = "https://console.cloud.google.com/logs/query"

// Explorer is a map of GCP Cloud Logging Log Explorer.
type Explorer struct {
	BaseURL   *url.URL
	ProjectID string
	// Query expression for logs.
	// https://cloud.google.com/logging/docs/view/logging-query-language
	Query Query
	// StorageScope for refine scope.
	StorageScope StorageScope
	// TimeRange for filter logs.
	TimeRange TimeRange
	// SummaryFields for manage summary fields.
	SummaryFields *SummaryFields
}

// String returns represent of Explorer URL.
func (ex *Explorer) String() string {
	var u *url.URL
	if ex.BaseURL != nil {
		tmp := *ex.BaseURL
		u = &tmp
	} else {
		var err error
		u, err = url.Parse(logBaseURL)
		if err != nil {
			panic(err)
		}
	}

	vs := values{}
	if v := ex.Query; v != "" {
		v.marshalURL(vs)
	}
	if v := ex.StorageScope; v != nil {
		v.marshalURL(vs)
	}
	if v := ex.TimeRange; v != nil {
		v.marshalURL(vs)
	}
	if v := ex.SummaryFields; v != nil {
		v.marshalURL(vs)
	}
	if u.RawPath == "" {
		u.RawPath = u.Path
	}
	u.Path = fmt.Sprintf("%s%s%s", u.Path, string(parameterSeparator), vs.RawEncode())
	u.RawPath = fmt.Sprintf("%s%s%s", u.RawPath, string(parameterSeparator), vs.Encode())

	if v := ex.ProjectID; v != "" {
		vs := u.Query()
		vs.Set("project", v)
		u.RawQuery = vs.Encode()
	}

	return u.String()
}
