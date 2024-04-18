package operationdocs

import (
	"github.com/vvakame/gcplogurl"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/operationdocs/internal/markdown"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
)

// entitleEditorLinksByCategory is a mapping of preconfigured Markdown-formatted
// links for requesting the appropriate 'mspServiceReader' role.
var entitleReaderLinksByCategory = map[spec.EnvironmentCategory]string{
	spec.EnvironmentCategoryTest: markdown.Link("Read-only Entitle request for the 'Engineering Projects' folder",
		"https://app.entitle.io/request?data=eyJkdXJhdGlvbiI6IjIxNjAwIiwianVzdGlmaWNhdGlvbiI6IkVOVEVSIEpVU1RJRklDQVRJT04gSEVSRSIsInJvbGVJZHMiOlt7ImlkIjoiZGY3NWJkNWMtYmUxOC00MjhmLWEzNjYtYzlhYTU1MGIwODIzIiwidGhyb3VnaCI6ImRmNzViZDVjLWJlMTgtNDI4Zi1hMzY2LWM5YWE1NTBiMDgyMyIsInR5cGUiOiJyb2xlIn1dfQ%3D%3D"),
	spec.EnvironmentCategoryInternal: markdown.Link("Read-only Entitle request for the 'Internal Services' folder",
		"https://app.entitle.io/request?data=eyJkdXJhdGlvbiI6IjEwODAwIiwianVzdGlmaWNhdGlvbiI6IkVOVEVSIEpVU1RJRklDQVRJT04gSEVSRSIsInJvbGVJZHMiOlt7ImlkIjoiNzg0M2MxYWYtYzU2MS00ZDMyLWE3ZTAtYjZkNjY0NDM4MzAzIiwidGhyb3VnaCI6Ijc4NDNjMWFmLWM1NjEtNGQzMi1hN2UwLWI2ZDY2NDQzODMwMyIsInR5cGUiOiJyb2xlIn1dfQ%3D%3D"),
	spec.EnvironmentCategoryExternal: markdown.Link("Read-only Entitle request for the 'Managed Services ' folder",
		"https://app.entitle.io/request?data=eyJkdXJhdGlvbiI6IjEwODAwIiwianVzdGlmaWNhdGlvbiI6IkVOVEVSIEpVU1RJRklDQVRJT04gSEVSRSIsInJvbGVJZHMiOlt7ImlkIjoiYTQ4OWM2MDktNTBlYy00ODAzLWIzZjItMzYzZGJhMTgwMWJhIiwidGhyb3VnaCI6ImE0ODljNjA5LTUwZWMtNDgwMy1iM2YyLTM2M2RiYTE4MDFiYSIsInR5cGUiOiJyb2xlIn1dfQ%3D%3D"),
}

// entitleEditorLinksByCategory is a mapping of preconfigured Markdown-formatted
// links for requesting the appropriate 'mspServiceEditor' role.
var entitleEditorLinksByCategory = map[spec.EnvironmentCategory]string{
	spec.EnvironmentCategoryTest: markdown.Link("Write access Entitle request for the 'Engineering Projects' folder",
		"https://app.entitle.io/request?data=eyJkdXJhdGlvbiI6IjIxNjAwIiwianVzdGlmaWNhdGlvbiI6IkVOVEVSIEpVU1RJRklDQVRJT04gSEVSRSIsInJvbGVJZHMiOlt7ImlkIjoiYzJkMTUwOGEtMGQ0ZS00MjA1LWFiZWUtOGY1ODg1ZGY3ZDE4IiwidGhyb3VnaCI6ImMyZDE1MDhhLTBkNGUtNDIwNS1hYmVlLThmNTg4NWRmN2QxOCIsInR5cGUiOiJyb2xlIn1dfQ%3D%3D"),
	spec.EnvironmentCategoryInternal: markdown.Link("Write access Entitle request for the 'Internal Services' folder",
		"https://app.entitle.io/request?data=eyJkdXJhdGlvbiI6IjEwODAwIiwianVzdGlmaWNhdGlvbiI6IkVOVEVSIEpVU1RJRklDQVRJT04gSEVSRSIsInJvbGVJZHMiOlt7ImlkIjoiZTEyYTJkZDktYzY1ZC00YzM0LTlmNDgtMzYzNTNkZmY0MDkyIiwidGhyb3VnaCI6ImUxMmEyZGQ5LWM2NWQtNGMzNC05ZjQ4LTM2MzUzZGZmNDA5MiIsInR5cGUiOiJyb2xlIn1dfQ%3D%3D"),
	spec.EnvironmentCategoryExternal: markdown.Link("Write access Entitle request for the 'Managed Services' folder",
		"https://app.entitle.io/request?data=eyJkdXJhdGlvbiI6IjEwODAwIiwianVzdGlmaWNhdGlvbiI6IkVOVEVSIEpVU1RJRklDQVRJT04gSEVSRSIsInJvbGVJZHMiOlt7ImlkIjoiODQzNTYxNzktZjkwMi00MDVlLTlhMTQtNTY3YTY1NmM5MzdmIiwidGhyb3VnaCI6Ijg0MzU2MTc5LWY5MDItNDA1ZS05YTE0LTU2N2E2NTZjOTM3ZiIsInR5cGUiOiJyb2xlIn1dfQ%3D%3D"),
}

func ServiceLogsURL(serviceKind spec.ServiceKind, envProjectID string) string {
	switch serviceKind {
	case spec.ServiceKindJob:
		return (&gcplogurl.Explorer{
			ProjectID: envProjectID,
			Query:     gcplogurl.Query(`resource.type = "cloud_run_job"`),
			SummaryFields: &gcplogurl.SummaryFields{
				Fields: []string{
					// execution identifier
					`labels/"run.googleapis.com/execution_name"`,
					// fields from structured logs by sourcegraph/log
					"jsonPayload/InstrumentationScope",
					"jsonPayload/Body",
					"jsonPayload/Attributes/error",
				},
			},
		}).String()

	default:
		return (&gcplogurl.Explorer{
			ProjectID: envProjectID,
			Query:     gcplogurl.Query(`resource.type = "cloud_run_revision" -logName=~"logs/run.googleapis.com%2Frequests"`),
			SummaryFields: &gcplogurl.SummaryFields{
				Fields: []string{
					// fields from structured logs by sourcegraph/log
					"jsonPayload/InstrumentationScope",
					"jsonPayload/Body",
					"jsonPayload/Attributes/error",
				},
			},
		}).String()
	}
}

// AlertPolicyDashboardURL creates a link to the GCP monitoring dashboards page
// with a filter for `custom` type (user created) dashboards
func AlertPolicyDashboardURL(envProjectID string) string {
	return markdown.Linkf("Dashboard", "https://console.cloud.google.com/monitoring/dashboards?project=%s&pageState=(%%22dashboards%%22:(%%22t%%22:%%22All%%22),%%22dashboardList%%22:(%%22f%%22:%%22%%255B%%257B_22k_22_3A_22Type_22_2C_22t_22_3A10_2C_22v_22_3A_22_5C_22Custom_5C_22_22_2C_22s_22_3Atrue_2C_22i_22_3A_22category_22%%257D%%255D%%22))", envProjectID)
}
