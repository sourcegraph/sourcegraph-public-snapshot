package resolvers

import "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"

type noLSIFDataResolver struct {
	message string
}

var _ graphqlbackend.NoLSIFDataResolver = &noLSIFDataResolver{}

func (r *noLSIFDataResolver) Message() string {
	return r.message
}

func (r *noLSIFDataResolver) ToLocationConnection() (graphqlbackend.LocationConnectionResolver, bool) {
	return nil, false
}

func (r *noLSIFDataResolver) ToMarkdown() (graphqlbackend.MarkdownResolver, bool) {
	return nil, false
}

func (r *noLSIFDataResolver) ToNoLSIFData() (graphqlbackend.NoLSIFDataResolver, bool) {
	return r, true
}
