package resolvers

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
)

type lsifUploadFailureReasonResolver struct {
	lsifUpload db.Upload
}

var _ graphqlbackend.LSIFUploadFailureReasonResolver = &lsifUploadFailureReasonResolver{}

func (r *lsifUploadFailureReasonResolver) Summary() string {
	return dereferenceString(r.lsifUpload.FailureSummary)
}

func (r *lsifUploadFailureReasonResolver) Stacktrace() string {
	return dereferenceString(r.lsifUpload.FailureStacktrace)
}

func dereferenceString(s *string) string {
	if s != nil {
		return *s
	}

	return ""
}
