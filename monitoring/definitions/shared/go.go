package shared

import (
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

// Golang monitoring overviews.
//
// Uses metrics exported by the Prometheus Golang library, so is available on all
// deployment types.
const TitleGolangMonitoring = "Golang runtime monitoring"

var (
	GoGoroutines sharedObservable = func(containerName string, owner monitoring.ObservableOwner) Observable {
		return Observable{
			Name:              "go_goroutines",
			Description:       "maximum active goroutines",
			Query:             fmt.Sprintf(`max by(instance) (go_goroutines{job=~".*%s"})`, containerName),
			Warning:           monitoring.Alert().GreaterOrEqual(10000, nil).For(10 * time.Minute),
			Panel:             monitoring.Panel().LegendFormat("{{name}}"),
			Owner:             owner,
			Interpretation:    "A high value here indicates a possible goroutine leak.",
			PossibleSolutions: "none",
		}
	}

	GoGcDuration sharedObservable = func(containerName string, owner monitoring.ObservableOwner) Observable {
		return Observable{
			Name:              "go_gc_duration_seconds",
			Description:       "maximum go garbage collection duration",
			Query:             fmt.Sprintf(`max by(instance) (go_gc_duration_seconds{job=~".*%s"})`, containerName),
			Warning:           monitoring.Alert().GreaterOrEqual(2, nil),
			Panel:             monitoring.Panel().LegendFormat("{{name}}").Unit(monitoring.Seconds),
			Owner:             owner,
			PossibleSolutions: "none",
		}
	}
)
