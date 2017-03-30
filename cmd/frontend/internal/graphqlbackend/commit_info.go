package graphqlbackend

type commitInfoResolver struct {
	rev       string
	author    *signatureResolver
	committer *signatureResolver
	message   string
}

func (r *commitInfoResolver) Rev() string {
	return r.rev
}

func (r *commitInfoResolver) Author() *signatureResolver {
	return r.author
}

func (r *commitInfoResolver) Committer() *signatureResolver {
	return r.committer
}

func (r *commitInfoResolver) Message() string {
	return r.message
}
