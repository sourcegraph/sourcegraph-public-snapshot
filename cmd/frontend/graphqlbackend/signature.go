package graphqlbackend

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type signatureResolver struct {
	person *personResolver
	date   time.Time
}

func (r signatureResolver) Person() *personResolver {
	return r.person
}

func (r signatureResolver) Date() string {
	return r.date.Format(time.RFC3339)
}

func toSignatureResolver(sig *git.Signature, includeUserInfo bool) *signatureResolver {
	if sig == nil {
		return nil
	}
	return &signatureResolver{
		person: &personResolver{
			name:            sig.Name,
			email:           sig.Email,
			includeUserInfo: includeUserInfo,
		},
		date: sig.Date,
	}
}
