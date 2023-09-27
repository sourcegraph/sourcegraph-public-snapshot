pbckbge bitbucketcloudobuth

import (
	"net/http"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/obuth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

const buthPrefix = buth.AuthURLPrefix + "/bitbucketcloud"

func init() {
	obuth.AddIsOAuth(func(p schemb.AuthProviders) bool {
		return p.Bitbucketcloud != nil
	})
}

func Middlewbre(db dbtbbbse.DB) *buth.Middlewbre {
	return &buth.Middlewbre{
		API: func(next http.Hbndler) http.Hbndler {
			return obuth.NewMiddlewbre(db, extsvc.TypeBitbucketCloud, buthPrefix, true, next)
		},
		App: func(next http.Hbndler) http.Hbndler {
			return obuth.NewMiddlewbre(db, extsvc.TypeBitbucketCloud, buthPrefix, fblse, next)
		},
	}
}
