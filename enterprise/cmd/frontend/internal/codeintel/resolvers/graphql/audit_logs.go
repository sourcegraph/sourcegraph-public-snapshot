package graphql

import (
	"strings"

	"github.com/graph-gophers/graphql-go"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	uploadsShared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
)

type auditLogColumnChangeResolver struct {
	columnTransition map[string]*string
}

func (r *auditLogColumnChangeResolver) Column() string {
	return *r.columnTransition["column"]
}

func (r *auditLogColumnChangeResolver) Old() *string {
	return r.columnTransition["old"]
}

func (r *auditLogColumnChangeResolver) New() *string {
	return r.columnTransition["new"]
}

type lsifUploadsAuditLogResolver struct {
	log uploadsShared.UploadLog
}

func (r *lsifUploadsAuditLogResolver) Reason() *string { return r.log.Reason }
func (r *lsifUploadsAuditLogResolver) ChangedColumns() (values []gql.AuditLogColumnChange) {
	for _, transition := range r.log.TransitionColumns {
		values = append(values, &auditLogColumnChangeResolver{transition})
	}
	return
}

func (r *lsifUploadsAuditLogResolver) LogTimestamp() gql.DateTime {
	return gql.DateTime{Time: r.log.LogTimestamp}
}

func (r *lsifUploadsAuditLogResolver) UploadDeletedAt() *gql.DateTime {
	return gql.DateTimeOrNil(r.log.RecordDeletedAt)
}

func (r *lsifUploadsAuditLogResolver) UploadID() graphql.ID {
	return marshalLSIFUploadGQLID(int64(r.log.UploadID))
}
func (r *lsifUploadsAuditLogResolver) InputCommit() string  { return r.log.Commit }
func (r *lsifUploadsAuditLogResolver) InputRoot() string    { return r.log.Root }
func (r *lsifUploadsAuditLogResolver) InputIndexer() string { return r.log.Indexer }
func (r *lsifUploadsAuditLogResolver) UploadedAt() gql.DateTime {
	return gql.DateTime{Time: r.log.UploadedAt}
}

func (r *lsifUploadsAuditLogResolver) Operation() string {
	return strings.ToUpper(r.log.Operation)
}
