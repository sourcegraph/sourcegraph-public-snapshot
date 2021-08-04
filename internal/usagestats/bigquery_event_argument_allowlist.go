package usagestats

// bigqueryEventsWithArgumentsAllowlist contains the event names of events
// that have arguments (event properties) that contain *only* public data. If you have an event log
// that has attached event properties, and you want the event properties data sent to our BigQuery
// instance for Cloud data, the event name must be added to this list.
//
// 🚨 PRIVACY: Do NOT add an event to this list if the event properties contain any explicitly or
// potentially private data, including but not limited to repository names, file names, search queries,
// and clone URLs.
var bigqueryEventsWithArgumentsAllowlist map[string]string = map[string]string{
	"ExtensionToggled":          "ExtensionToggled",
	"ExternalAuthSignupClicked": "ExternalAuthSignupClicked",
	"DynamicFilterClicked":      "DynamicFilterClicked",
	"InsightHover":              "InsightHover",
	"InsightUICustomization":    "InsightUICustomization",
	"InsightDataPointClick":     "InsightDataPointClick",
	"InsightsGroupedCount":      "InsightsGroupedCount",
	"InsightGroupedStepSizes":   "InsightGroupedStepSizes",
	"InsightRemoval":            "InsightRemoval",
	"InsightAddition":           "InsightAddition",
	"InsightEdit":               "InsightEdit",
	// RepogroupPageRepoLinkClicked passes the repo_name, but these are all public repos defined by Sourcegraph.
	// If we start allowing Cloud users to create private repogroups, we must remove this from the allowlist.
	"RepogroupPageRepoLinkClicked":          "RepogroupPageRepoLinkClicked",
	"CloseOnboardingTourClicked":            "CloseOnboardingTourClicked",
	"SearchNotebookRunBlock":                "SearchNotebookRunBlock",
	"SearchNotebookAddBlock":                "SearchNotebookAddBlock",
	"SearchNotebookMoveBlock":               "SearchNotebookMoveBlock",
	"SearchNotebookDuplicateBlock":          "SearchNotebookDuplicateBlock",
	"RecentFilesPanelLoaded":                "RecentFilesPanelLoaded",
	"RecentSearchesPanelLoaded":             "RecentSearchesPanelLoaded",
	"RepositoriesPanelLoaded":               "RepositoriesPanelLoaded",
	"SavedSearchesPanelLoaded":              "SavedSearchesPanelLoaded",
	"SavedSearchesPanelCreateButtonClicked": "SavedSearchesPanelCreateButtonClicked",
	"SavedQueriesToggleCreating":            "SavedQueriesToggleCreating",
	"search.latencies.frontend.code-load":   "search.latencies.frontend.code-load",
	"SearchResultsFetch":                    "SearchResultsFetch",
	"SearchResultsFetchFailed":              "SearchResultsFetchFailed",
	"SurveyButtonClicked":                   "SurveyButtonClicked",
}
