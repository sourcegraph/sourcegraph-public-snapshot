package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func CodeIntelUploads() *monitoring.Dashboard {
	return &monitoring.Dashboard{
		Name:        "codeintel-uploads",
		Title:       "Code Intelligence > Uploads",
		Description: "The service at `internal/codeintel/uploads`.",
		Variables:   []monitoring.ContainerVariable{},
		Groups: []monitoring.Group{
			shared.CodeIntelligence.NewUploadsServiceGroup(""),
			shared.CodeIntelligence.NewUploadsStoreGroup(""),
			shared.CodeIntelligence.NewUploadsBackgroundGroup(""),
			shared.CodeIntelligence.NewUploadsGraphQLTransportGroup(""),
			shared.CodeIntelligence.NewUploadsHTTPTransportGroup(""),
			shared.CodeIntelligence.NewUploadsCleanupTaskGroup(""),
			shared.CodeIntelligence.NewCommitGraphQueueGroup(""),
			shared.CodeIntelligence.NewUploadsExpirationTaskGroup(""),
		},
	}
}
