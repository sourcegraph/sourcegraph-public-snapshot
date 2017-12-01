package graphqlbackend

type commitInfoResolver struct {
	repository *repositoryResolver

	oid       gitObjectID
	author    signatureResolver
	committer *signatureResolver
	message   string
}

func (r *commitInfoResolver) Repository() *repositoryResolver { return r.repository }

func (r *commitInfoResolver) OID() gitObjectID { return r.oid }

func (r *commitInfoResolver) AbbreviatedOID() string { return string(r.oid)[:6] }

func (r *commitInfoResolver) Rev() string {
	return string(r.oid)
}

func (r *commitInfoResolver) Author() *signatureResolver {
	return &r.author
}

func (r *commitInfoResolver) Committer() *signatureResolver {
	return r.committer
}

func (r *commitInfoResolver) Message() string {
	return r.message
}
