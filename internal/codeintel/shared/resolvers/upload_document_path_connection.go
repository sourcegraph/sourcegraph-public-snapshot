package sharedresolvers

import "context"

type LSIFUploadDocumentPathsConnectionResolver interface {
	Nodes(ctx context.Context) ([]string, error)
	TotalCount(ctx context.Context) (*int32, error)
}

type UploadDocumentPathsConnectionResolver interface {
	Nodes(ctx context.Context) ([]string, error)
	TotalCount(ctx context.Context) (*int32, error)
}

type uploadDocumentPathsConnectionResolver struct {
	totalCount int
	documents  []string
}

func NewUploadDocumentPathsConnectionResolver(totalCount int, documents []string) UploadDocumentPathsConnectionResolver {
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
