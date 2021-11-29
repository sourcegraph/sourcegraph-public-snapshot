package graphqlbackend

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

type signatureResolver struct {
	person *PersonResolver
	date   time.Time
}

func (r signatureResolver) Person() *PersonResolver {
	return r.person
}

func (r signatureResolver) Date() string {
	return r.date.Format(time.RFC3339)
}

func toSignatureResolver(db database.DB, sig *gitdomain.Signature, includeUserInfo bool) *signatureResolver {
	if sig == nil {
		return nil
	}
	return &signatureResolver{
		person: &PersonResolver{
			db:              db,
			name:            sig.Name,
			email:           sig.Email,
			includeUserInfo: includeUserInfo,
		},
		date: sig.Date,
	}
}
