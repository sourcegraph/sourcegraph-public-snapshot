package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	amclient "github.com/prometheus/alertmanager/api/v2/client"
	"github.com/prometheus/alertmanager/api/v2/client/alert"
	"github.com/sourcegraph/log"

	srcprometheus "github.com/sourcegraph/sourcegraph/internal/src-prometheus"
)

// AlertsStatusReporter summarizes alert activity from Alertmanager
type AlertsStatusReporter struct {
	log          log.Logger
	alertmanager *amclient.Alertmanager
}

func NewAlertsStatusReporter(logger log.Logger, alertmanager *amclient.Alertmanager) *AlertsStatusReporter {
	return &AlertsStatusReporter{
		log:          logger.Scoped("alerts-status"),
		alertmanager: alertmanager,
	}
}

func (s *AlertsStatusReporter) Handler() http.Handler {
	handler := mux.NewRouter()
	handler.StrictSlash(true)
	// see EndpointAlertsStatus usages
	handler.HandleFunc(srcprometheus.EndpointAlertsStatus, func(w http.ResponseWriter, req *http.Request) {
		if noAlertmanager == "true" {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("alertmanager is disabled"))
			return
		}
		t := true
		f := false
		results, err := s.alertmanager.Alert.GetAlerts(&alert.GetAlertsParams{
			Active:    &t,
			Inhibited: &f,
			Context:   req.Context(),
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		var criticalAlerts, warningAlerts, silencedAlerts int
		servicesWithCriticalAlerts := map[string]struct{}{}
		for _, a := range results.GetPayload() {
			if len(a.Status.SilencedBy) > 0 {
				silencedAlerts++
				continue
			}
			level := a.Labels["level"]
			switch level {
			case "warning":
				warningAlerts++
			case "critical":
				criticalAlerts++
				svc := a.Labels["service_name"]
				servicesWithCriticalAlerts[svc] = struct{}{}
			}
		}
		// summarize alerts status
		b, err := json.Marshal(&srcprometheus.AlertsStatus{
			Silenced:         silencedAlerts,
			Warning:          warningAlerts,
			Critical:         criticalAlerts,
			ServicesCritical: len(servicesWithCriticalAlerts),
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		_, _ = w.Write(b)
	})
	return handler
}
