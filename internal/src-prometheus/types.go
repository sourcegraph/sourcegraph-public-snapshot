package srcprometheus

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/conf"
)

type AlertsStatus struct {
	Warning          int `json:"warning"`
	Silenced         int `json:"silenced"`
	Critical         int `json:"critical"`
	ServicesCritical int `json:"services_critical"`
}

// MonitoringAlert implements the GraphQL type MonitoringAlert.
//
// Internal fields named to accomodate GraphQL getters and setters, see grapqhlbackend.MonitoringAlert
type MonitoringAlert struct {
	TimestampValue   time.Time `json:"timestamp"`
	NameValue        string    `json:"name"`
	ServiceNameValue string    `json:"service_name"`
	OwnerValue       string    `json:"owner"`
	// AverageValue indicates average over past 12 hours, see alertsHistoryQuery and GraphQL schema docs for MonitoringAlert
	AverageValue float64 `json:"average"`
}

type MonitoringAlerts []*MonitoringAlert

// Less determined by timestamp -> serviceName -> alert name
func (a MonitoringAlerts) Less(i, j int) bool {
	if a[i].TimestampValue.Equal(a[j].TimestampValue) {
		if a[i].ServiceNameValue == a[j].ServiceNameValue {
			return a[i].NameValue < a[j].NameValue
		}
		return a[i].ServiceNameValue < a[j].ServiceNameValue
	}
	return a[i].TimestampValue.Before(a[j].TimestampValue)
}
func (a MonitoringAlerts) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a MonitoringAlerts) Len() int { return len(a) }

type AlertsHistory struct {
	Alerts MonitoringAlerts `json:"alerts"`
}

type ConfigStatus struct {
	Problems conf.Problems `json:"problems"`
}
