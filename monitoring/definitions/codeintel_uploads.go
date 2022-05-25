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
		},
	}
}

// TODO: Add these
// src_codeintel_background_upload_records_removed_total
// src_codeintel_background_index_records_removed_total
// src_codeintel_background_uploads_purged_total
// src_codeintel_background_audit_log_records_expired_t
