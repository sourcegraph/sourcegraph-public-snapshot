package graphql

import gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"

type gitObjectFilterPreviewResolver struct {
	name string
	rev  string
}

var _ gql.GitObjectFilterPreviewResolver = &gitObjectFilterPreviewResolver{}

func (r *gitObjectFilterPreviewResolver) Name() string {
	return r.name
}

func (r *gitObjectFilterPreviewResolver) Rev() string {
	return r.rev
}
