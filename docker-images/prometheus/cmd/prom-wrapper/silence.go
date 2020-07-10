package main

import (
	"github.com/prometheus/alertmanager/api/v2/models"
	"github.com/sourcegraph/sourcegraph/schema"
)

func stringP(v string) *string {
	return &v
}

func newMatchersFromSilence(silence schema.ObservabilitySilenceAlerts) models.Matchers {
	return models.Matchers{{
		Name:  stringP("alert"),
		Value: stringP(silence.Alert),
	}, {
		Name:  stringP("service"),
		Value: stringP(silence.Service),
	}, {
		Name:  stringP("level"),
		Value: stringP(silence.Level),
	}}
}

func newSilenceFromMatchers(matchers models.Matchers) schema.ObservabilitySilenceAlerts {
	var silencedAlert schema.ObservabilitySilenceAlerts
	for _, m := range matchers {
		switch *m.Name {
		case "alert":
			silencedAlert.Alert = *m.Value
		case "service":
			silencedAlert.Service = *m.Value
		case "level":
			silencedAlert.Level = *m.Value
		}
	}
	return silencedAlert
}
