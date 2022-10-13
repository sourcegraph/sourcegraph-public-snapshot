package sharedresolvers

import (
	"strings"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
)

type LSIFUploadsAuditLogsResolver interface {
	LogTimestamp() DateTime
	UploadDeletedAt() *DateTime
	Reason() *string
	ChangedColumns() []AuditLogColumnChange
	UploadID() graphql.ID
	InputCommit() string
	InputRoot() string
	InputIndexer() string
	UploadedAt() DateTime
	Operation() string
	// AssociatedIndex(ctx context.Context) (LSIFIndexResolver, error)
}

type AuditLogColumnChange interface {
	Column() string
	Old() *string
	New() *string
}

type lsifUploadsAuditLogResolver struct {
	log types.UploadLog
}

func NewLSIFUploadsAuditLogsResolver(log types.UploadLog) LSIFUploadsAuditLogsResolver {
	return &lsifUploadsAuditLogResolver{log: log}
}

func (r *lsifUploadsAuditLogResolver) Reason() *string { return r.log.Reason }
func (r *lsifUploadsAuditLogResolver) ChangedColumns() (values []AuditLogColumnChange) {
	for _, transition := range r.log.TransitionColumns {
		values = append(values, &auditLogColumnChangeResolver{transition})
	}
	return values
}

func (r *lsifUploadsAuditLogResolver) LogTimestamp() DateTime {
	return DateTime{Time: r.log.LogTimestamp}
}

func (r *lsifUploadsAuditLogResolver) UploadDeletedAt() *DateTime {
	return DateTimeOrNil(r.log.RecordDeletedAt)
}

func (r *lsifUploadsAuditLogResolver) UploadID() graphql.ID {
	return marshalLSIFUploadGQLID(int64(r.log.UploadID))
}
func (r *lsifUploadsAuditLogResolver) InputCommit() string  { return r.log.Commit }
func (r *lsifUploadsAuditLogResolver) InputRoot() string    { return r.log.Root }
func (r *lsifUploadsAuditLogResolver) InputIndexer() string { return r.log.Indexer }
func (r *lsifUploadsAuditLogResolver) UploadedAt() DateTime {
	return DateTime{Time: r.log.UploadedAt}
}

func (r *lsifUploadsAuditLogResolver) Operation() string {
	return strings.ToUpper(r.log.Operation)
}
