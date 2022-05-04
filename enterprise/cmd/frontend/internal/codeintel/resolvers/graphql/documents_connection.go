package graphql

import "context"

type uploadDocumentsConnectionResolver struct {
	totalCount int
	documents  []string
}

func (r *uploadDocumentsConnectionResolver) Nodes(ctx context.Context) ([]string, error) {
	return r.documents, nil
}

func (r *uploadDocumentsConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	count := int32(r.totalCount)
	return &count, nil
}
