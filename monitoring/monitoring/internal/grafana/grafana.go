// Package grafana is home to additional internal data types for Grafana to extend the Grafana SDK library
package grafana

import "github.com/grafana-tools/sdk"

type OverrideMatcher struct {
	ID      string `json:"id"`
	Options string `json:"options" `
}

func matcherByName(name string) OverrideMatcher {
	return OverrideMatcher{ID: "byName", Options: name}
}

type OverrideProperty struct {
	ID    string `json:"id"`
	Value any    `json:"value"`
}

func propertyWidth(width float32) OverrideProperty {
	return OverrideProperty{ID: "custom.width", Value: width}
}

func propertyLinks(links []*sdk.Link) OverrideProperty {
	return OverrideProperty{ID: "links", Value: links}
}

type Override struct {
	Matcher    OverrideMatcher    `json:"matcher"`
	Properties []OverrideProperty `json:"properties"`
}
