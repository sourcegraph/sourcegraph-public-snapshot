package graphql

import resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"

type gitObjectFilterPreviewResolver struct {
	name string
	rev  string
}

func NewGitObjectFilterPreviewResolver(name, rev string) resolverstubs.GitObjectFilterPreviewResolver {
	return &gitObjectFilterPreviewResolver{name: name, rev: rev}
}

func (r *gitObjectFilterPreviewResolver) Name() string {
	return r.name
}

func (r *gitObjectFilterPreviewResolver) Rev() string {
	return r.rev
}
