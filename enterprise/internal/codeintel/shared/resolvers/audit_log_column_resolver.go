package sharedresolvers

import resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"

type auditLogColumnChangeResolver struct {
	columnTransition map[string]*string
}

func NewAuditLogColumnChangeResolver(columnTransition map[string]*string) resolverstubs.AuditLogColumnChangeResolver {
	return &auditLogColumnChangeResolver{columnTransition}
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
