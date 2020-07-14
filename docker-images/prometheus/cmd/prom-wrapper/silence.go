package main

import (
	"github.com/prometheus/alertmanager/api/v2/models"
)

func stringP(v string) *string {
	return &v
}

func boolP(v bool) *bool {
	return &v
}

func newMatchersFromSilence(silence string) models.Matchers {
	return models.Matchers{{
		Name:    stringP("alertname"),
		Value:   stringP(silence),
		IsRegex: boolP(false),
	}}
}

func newSilenceFromMatchers(matchers models.Matchers) string {
	for _, m := range matchers {
		if *m.Name == "alertname" {
			return *m.Value
		}
	}
	return ""
}
