package sharedresolvers

import (
	"context"

	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
)

type uploadDocumentPathsConnectionResolver struct {
	totalCount int
	documents  []string
}

func NewUploadDocumentPathsConnectionResolver(totalCount int, documents []string) resolverstubs.UploadDocumentPathsConnectionResolver {
	return &uploadDocumentPathsConnectionResolver{
		totalCount: totalCount,
		documents:  documents,
	}
}

func (r *uploadDocumentPathsConnectionResolver) Nodes(ctx context.Context) ([]string, error) {
	return r.documents, nil
}

func (r *uploadDocumentPathsConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	count := int32(r.totalCount)
	return &count, nil
}
