package graphqlbackend

import (
	"context"
)

// A resolver for the GraphQL type GenericSearchMatch
type GenericSearchMatchResolver struct {
	url        string
	body       string
	highlights []*highlightedRange
}

func (m *GenericSearchMatchResolver) URL() string {
	return m.url
}

func (m *GenericSearchMatchResolver) Body(ctx context.Context) string {
	return m.body
}

func (m *GenericSearchMatchResolver) Highlights(ctx context.Context) []*highlightedRange {
	return m.highlights
}
