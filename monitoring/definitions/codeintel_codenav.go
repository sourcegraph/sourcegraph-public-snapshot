package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func CodeIntelCodeNav() *monitoring.Dashboard {
	return &monitoring.Dashboard{
		Name:        "codeintel-codenav",
		Title:       "Code Intelligence > Code Nav",
		Description: "The service at `internal/codeintel/codenav`.",
		Variables:   []monitoring.ContainerVariable{},
		Groups: []monitoring.Group{
			shared.CodeIntelligence.NewCodeNavServiceGroup(""),
			shared.CodeIntelligence.NewCodeNavLsifStoreGroup(""),
			shared.CodeIntelligence.NewCodeNavGraphQLTransportGroup(""),
			shared.CodeIntelligence.NewCodeNavStoreGroup(""),
		},
	}
}
