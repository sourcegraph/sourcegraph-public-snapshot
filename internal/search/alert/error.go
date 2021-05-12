package alert

import (
	"errors"
	"strings"
)

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

func FromError(err error) *Alert {
	var aErr *AlertableError

	if errors.As(err, &aErr) {
		return aErr.Alert
	} else if strings.Contains(err.Error(), "Worker_oomed") || strings.Contains(err.Error(), "Worker_exited_abnormally") {
		return &Alert{
			PrometheusType: "structural_search_needs_more_memory",
			Title:          "Structural search needs more memory",
			Description:    "Running your structural search may require more memory. If you are running the query on many repositories, try reducing the number of repositories with the `repo:` filter.",
			Priority:       5,
		}
	} else if strings.Contains(err.Error(), "Out of memory") {
		return &Alert{
			PrometheusType: "structural_search_needs_more_memory__give_searcher_more_memory",
			Title:          "Structural search needs more memory",
			Description:    `Running your structural search requires more memory. You could try reducing the number of repositories with the "repo:" filter. If you are an administrator, try double the memory allocated for the "searcher" service. If you're unsure, reach out to us at support@sourcegraph.com.`,
			Priority:       4,
		}
	}
	return nil
}
