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
		Variables: []monitoring.ContainerVariable{
			{
				Label: "Source",
				Name:  "source",
				OptionsLabelValues: monitoring.ContainerVariableOptionsLabelValues{
					Query:         "src_codeintel_uploads_total{}",
					LabelName:     "app",
					ExampleOption: "frontend",
				},
				WildcardAllValue: true,
				Multi:            false,
			},
		},
		Groups: []monitoring.Group{
			shared.CodeIntelligence.NewUploadsServiceGroup("$source"),
			shared.CodeIntelligence.NewUploadsStoreGroup("$source"),
			shared.CodeIntelligence.NewUploadsBackgroundGroup("$source"),
			shared.CodeIntelligence.NewUploadsGraphQLTransportGroup("$source"),
			shared.CodeIntelligence.NewUploadsHTTPTransportGroup("$source"),
			shared.CodeIntelligence.NewUploadsCleanupTaskGroup("$source"),
			shared.CodeIntelligence.NewCommitGraphQueueGroup("$source"),
			shared.CodeIntelligence.NewUploadsExpirationTaskGroup("$source"),
		},
	}
}
