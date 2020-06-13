package main

import (
	"fmt"

	"github.com/grafana-tools/sdk"
)

func newDefaultAlertsPanelAlert(level string) *sdk.Alert {
	return &sdk.Alert{
		Name:      fmt.Sprintf("Alert: %s alerts are firing", level),
		Frequency: "1m",
		For:       "1m",
		Conditions: []sdk.AlertCondition{{
			Type: "query",
			Reducer: sdk.AlertReducer{
				Type: "max",
			},
			Query: sdk.AlertQuery{
				Params: []string{
					"A",
					"1m",
					"now",
				},
			},
			Evaluator: sdk.AlertEvaluator{
				Params: []float64{0},
				Type:   "gt",
			},
		}},
		ExecutionErrorState: "alerting",
		NoDataState:         "no_data",
		Notifications:       []sdk.AlertNotification{},
	}
}
