package main

import (
	"github.com/prometheus/alertmanager/api/v2/models"
	"github.com/sourcegraph/sourcegraph/schema"
)

func stringP(v string) *string {
	return &v
}

func boolP(v bool) *bool {
	return &v
}

func newMatchersFromSilence(silence schema.ObservabilitySilenceAlerts) models.Matchers {
	return models.Matchers{{
		Name:    stringP("name"),
		Value:   stringP(silence.Name),
		IsRegex: boolP(false),
	}, {
		Name:    stringP("service_name"),
		Value:   stringP(silence.Service),
		IsRegex: boolP(false),
	}, {
		Name:    stringP("level"),
		Value:   stringP(silence.Level),
		IsRegex: boolP(false),
	}}
}

func newSilenceFromMatchers(matchers models.Matchers) schema.ObservabilitySilenceAlerts {
	var silencedAlert schema.ObservabilitySilenceAlerts
	for _, m := range matchers {
		switch *m.Name {
		case "name":
			silencedAlert.Name = *m.Value
		case "service_name":
			silencedAlert.Service = *m.Value
		case "level":
			silencedAlert.Level = *m.Value
		}
	}
	return silencedAlert
}
