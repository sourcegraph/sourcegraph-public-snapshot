package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func CodeIntelAutoIndexing() *monitoring.Dashboard {
	return &monitoring.Dashboard{
		Name:        "codeintel-autoindexing",
		Title:       "Code Intelligence > Autoindexing",
		Description: "The service at `internal/codeintel/autoindexing`.",
		Variables:   []monitoring.ContainerVariable{},
		Groups: []monitoring.Group{
			shared.CodeIntelligence.NewAutoindexingSummaryGroup(""),
			shared.CodeIntelligence.NewAutoindexingServiceGroup(""),
			// shared.CodeIntelligence.NewAutoindexingGraphQLTransportGroup(""),
			shared.CodeIntelligence.NewAutoindexingStoreGroup(""),
			shared.CodeIntelligence.NewAutoindexingInferenceServiceGroup(""),
			shared.CodeIntelligence.NewLuasandboxServiceGroup(""),
		},
	}
}
