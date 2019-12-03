package resolvers

import "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"

//
// Markdown Resolver

type markdownWithConfidenceResolver struct {
	*graphqlbackend.MarkdownResolver
}

var _ graphqlbackend.MarkdownWithConfidenceResolver = &markdownWithConfidenceResolver{}

func (r *markdownWithConfidenceResolver) Placeholder() *string { return nil }
