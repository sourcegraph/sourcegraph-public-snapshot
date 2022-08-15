package resolvers

import (
	"context"

	store "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
)

func (r *resolver) AuditLogsForUpload(ctx context.Context, id int) ([]store.UploadLog, error) {
	return r.dbStore.GetAuditLogsForUpload(ctx, id)
}
