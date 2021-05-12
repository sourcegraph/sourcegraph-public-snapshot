package alert

import (
	"github.com/sourcegraph/sourcegraph/internal/search/query"
)

type Alert struct {
	PrometheusType  string
	Title           string
	Description     string
	ProposedQueries []ProposedQuery
	// The higher the priority the more important is the alert.
	Priority int
}

type ProposedQuery struct {
	Description string
	Query       string
	PatternType query.SearchType
}

// AlertableError is an error that contains an alert.
//
// The intent of the type is to allow creation of alerts at the site of the
// error rather than inspecting the propagated error and creating an alert
// based on the error text. This is fragile since the error text may change
// over time. Additionally, when creating alerts away from the source, it's
// difficult to tell where the inital error source was.
type AlertableError struct {
	Alert *Alert
	Err   error
}

func (a AlertableError) Error() string {
	return a.Err.Error()
}

func (a AlertableError) Unwrap() error {
	return a.Err
}

// Wrap wraps an error into an AlertableError given
// the associated alert.
func Wrap(err error, a *Alert) error {
	return &AlertableError{
		Alert: a,
		Err:   err,
	}
}
