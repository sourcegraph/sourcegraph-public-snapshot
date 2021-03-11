package graphqlbackend

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
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

func toSignatureResolver(db dbutil.DB, sig *git.Signature, includeUserInfo bool) *signatureResolver {
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
