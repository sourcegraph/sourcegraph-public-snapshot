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
type MonitoringAlert struct {
	TimestampValue   time.Time
	NameValue        string
	ServiceNameValue string
	OwnerValue       string
	AverageValue     float64
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
