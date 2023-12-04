package grafana

import "github.com/grafana-tools/sdk"

func NewContainerAlertsDefinedTable(target sdk.Target) *sdk.Panel {
	alertsDefined := sdk.NewCustom("Alerts defined")
	alertsDefined.Type = "table"

	var panelTemplateLink = "/-/debug/grafana/d/${__data.fields.service_name}/${__data.fields.service_name}?viewPanel=${__data.fields.grafana_panel_id}"
	alertsDefined.CustomPanel = &sdk.CustomPanel{
		"fieldConfig": map[string]any{
			"overrides": []*Override{
				{
					Matcher: matcherByName("level"),
					Properties: []OverrideProperty{
						propertyWidth(80),
					},
				},
				{
					Matcher: matcherByName("description"),
					Properties: []OverrideProperty{
						{ID: "custom.filterable", Value: true},
						propertyLinks([]*sdk.Link{{
							Title: "Graph panel",
							URL:   &panelTemplateLink,
						}}),
					},
				},
				alertsFiringOverride(),
				{
					Matcher: matcherByName("grafana_panel_id"),
					Properties: []OverrideProperty{
						propertyWidth(0.1),
					},
				},
				{
					Matcher: matcherByName("service_name"),
					Properties: []OverrideProperty{
						propertyWidth(0.1),
					},
				},
			},
		},
		"options": map[string]any{
			"showHeader": true,
			"sortBy": []map[string]any{{
				"desc":        true,
				"displayName": "firing?",
			}},
		},
		"transformations": []map[string]any{{
			"id": "organize",
			"options": map[string]map[string]any{
				"excludeByName": {
					"Time": true,
				},
				"indexByName": {
					"Time":        0,
					"level":       1,
					"description": 2,
					"Value":       3,
				},
			},
		}},
		"targets": []*sdk.Target{&target},
	}
	return alertsDefined
}

func alertsFiringOverride() *Override {
	return &Override{
		Matcher: matcherByName("Value"),
		Properties: []OverrideProperty{
			{ID: "displayName", Value: "firing?"},
			{ID: "custom.displayMode", Value: "color-background"},
			{ID: "custom.align", Value: "center"},
			propertyWidth(80),
			{ID: "unit", Value: "short"},
			{
				ID: "thresholds",
				Value: map[string]any{
					"mode": "absolute",
					"steps": []map[string]any{{
						"color": "rgba(50, 172, 45, 0.97)",
						"value": nil,
					}, {
						"color": "rgba(245, 54, 54, 0.9)",
						"value": 1,
					}},
				},
			},
			{
				ID: "mappings",
				Value: []map[string]any{{
					"from":  "",
					"id":    1,
					"text":  "false",
					"to":    "",
					"type":  1,
					"value": "0",
				}, {
					"from":  "",
					"id":    2,
					"text":  "true",
					"to":    "",
					"type":  1,
					"value": "1",
				}},
			},
		},
	}
}
