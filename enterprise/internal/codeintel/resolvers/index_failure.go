package resolvers

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/db"
)

type lsifIndexFailureReasonResolver struct {
	lsifIndex db.Index
}

var _ graphqlbackend.LSIFIndexFailureReasonResolver = &lsifIndexFailureReasonResolver{}

func (r *lsifIndexFailureReasonResolver) Summary() string {
	return dereferenceString(r.lsifIndex.FailureSummary)
}

func (r *lsifIndexFailureReasonResolver) Stacktrace() string {
	return dereferenceString(r.lsifIndex.FailureStacktrace)
}
