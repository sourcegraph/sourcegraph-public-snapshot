package sharedresolvers

import (
	"strings"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

type lsifUploadsAuditLogResolver struct {
	log types.UploadLog
}

func NewLSIFUploadsAuditLogsResolver(log types.UploadLog) resolverstubs.LSIFUploadsAuditLogsResolver {
	return &lsifUploadsAuditLogResolver{log: log}
}

func (r *lsifUploadsAuditLogResolver) Reason() *string { return r.log.Reason }
func (r *lsifUploadsAuditLogResolver) ChangedColumns() (values []resolverstubs.AuditLogColumnChange) {
	for _, transition := range r.log.TransitionColumns {
		values = append(values, &auditLogColumnChangeResolver{transition})
	}
	return values
}

func (r *lsifUploadsAuditLogResolver) LogTimestamp() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.log.LogTimestamp}
}

func (r *lsifUploadsAuditLogResolver) UploadDeletedAt() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(r.log.RecordDeletedAt)
}

func (r *lsifUploadsAuditLogResolver) UploadID() graphql.ID {
	return marshalLSIFUploadGQLID(int64(r.log.UploadID))
}
func (r *lsifUploadsAuditLogResolver) InputCommit() string  { return r.log.Commit }
func (r *lsifUploadsAuditLogResolver) InputRoot() string    { return r.log.Root }
func (r *lsifUploadsAuditLogResolver) InputIndexer() string { return r.log.Indexer }
func (r *lsifUploadsAuditLogResolver) UploadedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.log.UploadedAt}
}

func (r *lsifUploadsAuditLogResolver) Operation() string {
	return strings.ToUpper(r.log.Operation)
}
