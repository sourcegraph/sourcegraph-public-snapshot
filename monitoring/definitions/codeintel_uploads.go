package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func CodeIntelUploads() *monitoring.Container {
	return &monitoring.Container{
		Name:        "codeintel-uploads",
		Title:       "Code Intelligence > Uploads",
		Description: "The service at `internal/codeintel/uploads`.",
		Variables:   []monitoring.ContainerVariable{},
		Groups: []monitoring.Group{
			shared.CodeIntelligence.NewUploadsServiceGroup(""),
			shared.CodeIntelligence.NewUploadsStoreGroup(""),
			shared.CodeIntelligence.NewUploadsGraphQLTransportGroup(""),
			shared.CodeIntelligence.NewUploadsHTTPTransportGroup(""),
			shared.CodeIntelligence.NewUploadsCleanupTaskGroup(""),
		},
	}
}
