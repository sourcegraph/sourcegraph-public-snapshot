package graphqlbackend

import "sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"

type signatureResolver struct {
	person *personResolver
	date   string
}

func (r *signatureResolver) Person() *personResolver {
	return r.person
}

func (r *signatureResolver) Date() string {
	return r.date
}

func toSignatureResolver(sig *vcs.Signature) *signatureResolver {
	if sig == nil {
		return nil
	}
	return &signatureResolver{
		person: &personResolver{
			name:  sig.Name,
			email: sig.Email,
		},
		date: sig.Date.String(),
	}
}
