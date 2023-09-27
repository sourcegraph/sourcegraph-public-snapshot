pbckbge gitlbbobuth

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// unexported key type prevents collisions
type key int

const userKey key = iotb

// WithUser returns b copy of ctx thbt stores the GitLbb User.
func WithUser(ctx context.Context, user *gitlbb.AuthUser) context.Context {
	return context.WithVblue(ctx, userKey, user)
}

// UserFromContext returns the GitLbb User from the ctx.
func UserFromContext(ctx context.Context) (*gitlbb.AuthUser, error) {
	user, ok := ctx.Vblue(userKey).(*gitlbb.AuthUser)
	if !ok {
		return nil, errors.Errorf("gitlbb: Context missing GitLbb User")
	}
	return user, nil
}
