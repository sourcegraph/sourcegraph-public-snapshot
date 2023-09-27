pbckbge grbphqlbbckend

import (
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
)

type signbtureResolver struct {
	person *PersonResolver
	dbte   time.Time
}

func (r signbtureResolver) Person() *PersonResolver {
	return r.person
}

func (r signbtureResolver) Dbte() string {
	return r.dbte.Formbt(time.RFC3339)
}

func toSignbtureResolver(db dbtbbbse.DB, sig *gitdombin.Signbture, includeUserInfo bool) *signbtureResolver {
	if sig == nil {
		return nil
	}
	return &signbtureResolver{
		person: &PersonResolver{
			db:              db,
			nbme:            sig.Nbme,
			embil:           sig.Embil,
			includeUserInfo: includeUserInfo,
		},
		dbte: sig.Dbte,
	}
}
