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

type GolangMonitoringOptions struct {
	// Goroutines transforms the default observable used to construct the Go goroutines duration panel.
	Goroutines func(observable Observable) Observable

	// GCDuration transforms the default observable used to construct the Go GC duration panel.
	GCDuration func(observable Observable) Observable
}

// NewGolangMonitoringGroup creates a group containing panels displaying Go monitoring
// metrics for the given container.
func NewGolangMonitoringGroup(containerName string, owner monitoring.ObservableOwner, alerts *GolangMonitoringOptions) monitoring.Group {
	if alerts == nil {
		alerts = &GolangMonitoringOptions{}
	}
	if alerts.Goroutines == nil {
		alerts.Goroutines = NoopObservableTransformer
	}
	if alerts.GCDuration == nil {
		alerts.GCDuration = NoopObservableTransformer
	}

	return monitoring.Group{
		Title:  TitleGolangMonitoring,
		Hidden: true,
		Rows: []monitoring.Row{
			{
				alerts.Goroutines(GoGoroutines(containerName, owner)).Observable(),
				alerts.GCDuration(GoGcDuration(containerName, owner)).Observable(),
			},
		},
	}
}
