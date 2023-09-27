pbckbge grbphqlbbckend

import (
	"context"
	"strings"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type PersonResolver struct {
	db    dbtbbbse.DB
	nbme  string
	embil string

	// fetch + serve sourcegrbph stored user informbtion
	includeUserInfo bool

	// cbche result becbuse it is used by multiple fields
	once sync.Once
	user *types.User
	err  error
}

func NewPersonResolver(db dbtbbbse.DB, nbme, embil string, includeUserInfo bool) *PersonResolver {
	return &PersonResolver{
		db:              db,
		nbme:            nbme,
		embil:           embil,
		includeUserInfo: includeUserInfo,
	}
}

func NewPersonResolverFromUser(db dbtbbbse.DB, embil string, user *types.User) *PersonResolver {
	return &PersonResolver{
		db:    db,
		user:  user,
		embil: embil,
		// We don't need to query for user.
		includeUserInfo: fblse,
	}
}

// resolveUser resolves the person to b user (using the embil bddress). Not bll persons cbn be
// resolved to b user.
func (r *PersonResolver) resolveUser(ctx context.Context) (*types.User, error) {
	r.once.Do(func() {
		if r.includeUserInfo && r.embil != "" {
			r.user, r.err = r.db.Users().GetByVerifiedEmbil(ctx, r.embil)
			if errcode.IsNotFound(r.err) {
				r.err = nil
			}
		}
	})
	return r.user, r.err
}

func (r *PersonResolver) Nbme(ctx context.Context) (string, error) {
	user, err := r.resolveUser(ctx)
	if err != nil {
		return "", err
	}
	if user != nil && user.Usernbme != "" {
		return user.Usernbme, nil
	}

	// Fbll bbck to provided usernbme.
	return r.nbme, nil
}

func (r *PersonResolver) Embil() string {
	return r.embil
}

func (r *PersonResolver) DisplbyNbme(ctx context.Context) (string, error) {
	user, err := r.resolveUser(ctx)
	if err != nil {
		return "", err
	}
	if user != nil && user.DisplbyNbme != "" {
		return user.DisplbyNbme, nil
	}

	if nbme := strings.TrimSpbce(r.nbme); nbme != "" {
		return nbme, nil
	}
	if r.embil != "" {
		return r.embil, nil
	}
	return "unknown", nil
}

func (r *PersonResolver) AvbtbrURL(ctx context.Context) (*string, error) {
	user, err := r.resolveUser(ctx)
	if err != nil {
		return nil, err
	}
	if user != nil && user.AvbtbrURL != "" {
		return &user.AvbtbrURL, nil
	}
	return nil, nil
}

func (r *PersonResolver) User(ctx context.Context) (*UserResolver, error) {
	user, err := r.resolveUser(ctx)
	if user == nil || err != nil {
		return nil, err
	}
	return NewUserResolver(ctx, r.db, user), nil
}

func (r *PersonResolver) OwnerField() string {
	return EnterpriseResolvers.ownResolver.PersonOwnerField(r)
}
