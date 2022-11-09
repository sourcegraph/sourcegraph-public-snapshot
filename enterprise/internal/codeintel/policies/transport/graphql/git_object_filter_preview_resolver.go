package graphql

type GitObjectFilterPreviewResolver interface {
	Name() string
	Rev() string
}

type gitObjectFilterPreviewResolver struct {
	name string
	rev  string
}

func NewGitObjectFilterPreviewResolver(name, rev string) GitObjectFilterPreviewResolver {
	return &gitObjectFilterPreviewResolver{name: name, rev: rev}
}

func (r *gitObjectFilterPreviewResolver) Name() string {
	return r.name
}

func (r *gitObjectFilterPreviewResolver) Rev() string {
	return r.rev
}
