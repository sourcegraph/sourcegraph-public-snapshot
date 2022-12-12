package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

var _ graphqlbackend.ScopedInsightQueryPayloadResolver = &scopedInsightQueryPreviewResolver{}

func (r *Resolver) ValidateScopedInsightQuery(ctx context.Context, args graphqlbackend.ValidateScopedInsightQueryArgs) (graphqlbackend.ScopedInsightQueryPayloadResolver, error) {
	return nil, nil
}

type scopedInsightQueryPreviewResolver struct {
	numberOfRepositories int32
	query                string
}

func (s *scopedInsightQueryPreviewResolver) NumberOfRepositories(ctx context.Context) int32 {
	return s.numberOfRepositories
}

func (s *scopedInsightQueryPreviewResolver) Query(ctx context.Context) string {
	return s.query
}
