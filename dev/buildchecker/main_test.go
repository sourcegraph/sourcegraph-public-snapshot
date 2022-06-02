package main

import (
	"testing"

	"github.com/hexops/autogold"
)

func TestGenerateSummaryMessage(t *testing.T) {
	fromString := "2006-01-02"
	toString := "2006-01-03"
	got := generateSummaryMessage(fromString, toString, 5, 1, 20, 150)
	want := autogold.Want("name", `:bar_chart: Welcome to your weekly CI report for period *2006-01-02* to *2006-01-03*!
• Total builds: *5*
• Total flakes: *1*
• Average % of build flakes: *20%*
• Total incident duration: *150ns*

For more information, view the dashboards at <https://app.okayhq.com/dashboards/3856903d-33ea-4d60-9719-68fec0eb4313/build-stats-kpis|OkayHQ>.
`)
	want.Equal(t, got)
}
