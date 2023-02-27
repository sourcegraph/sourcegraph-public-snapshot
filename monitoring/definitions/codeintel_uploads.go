package definitions

import (
	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func CodeIntelUploads() *monitoring.Dashboard {
	return &monitoring.Dashboard{
		Name:        "codeintel-uploads",
		Title:       "Code Intelligence > Uploads",
		Description: "The service at `enterprise/internal/codeintel/uploads`.",
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
			shared.CodeIntelligence.NewUploadsServiceGroup("${source:regex}"),
			shared.CodeIntelligence.NewUploadsStoreGroup("${source:regex}"),
			shared.CodeIntelligence.NewUploadsBackgroundGroup("${source:regex}"),
			shared.CodeIntelligence.NewUploadsGraphQLTransportGroup("${source:regex}"),
			shared.CodeIntelligence.NewUploadsHTTPTransportGroup("${source:regex}"),
			shared.CodeIntelligence.NewUploadsCleanupTaskGroup("${source:regex}"),
			shared.CodeIntelligence.NewCommitGraphQueueGroup("${source:regex}"),
			shared.CodeIntelligence.NewUploadsExpirationTaskGroup("${source:regex}"),
		},
	}
}
