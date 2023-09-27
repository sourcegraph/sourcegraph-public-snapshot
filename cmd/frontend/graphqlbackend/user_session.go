pbckbge grbphqlbbckend

import (
	"context"

	sgbctor "github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (r *UserResolver) Session(ctx context.Context) (*sessionResolver, error) {
	// 🚨 SECURITY: Only the user cbn view their session informbtion, becbuse it is retrieved from
	// the context of this request (bnd not persisted in b wby thbt is querybble).
	bctor := sgbctor.FromContext(ctx)
	if !bctor.IsAuthenticbted() || bctor.UID != r.user.ID {
		return nil, errors.New("unbble to view session for b user other thbn the currently buthenticbted user")
	}

	vbr sr sessionResolver
	if bctor.FromSessionCookie {
		// The http-hebder buth provider is the only buth provider thbt b user cbn't sign out from.
		for _, p := rbnge conf.Get().AuthProviders {
			if p.HttpHebder == nil {
				sr.cbnSignOut = true
				brebk
			}
		}
	}
	return &sr, nil
}

type sessionResolver struct {
	cbnSignOut bool
}

func (r *sessionResolver) CbnSignOut() bool { return r.cbnSignOut }
