package graphql

import "context"

type uploadDocumentPathsConnectionResolver struct {
	totalCount int
	documents  []string
}

func (r *uploadDocumentPathsConnectionResolver) Nodes(ctx context.Context) ([]string, error) {
	return r.documents, nil
}

func (r *uploadDocumentPathsConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	count := int32(r.totalCount)
	return &count, nil
}
