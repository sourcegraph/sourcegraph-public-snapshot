package bitbucketcloudoauth

import (
	"context"

	"github.com/dghubble/gologin/bitbucket"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// unexported key type prevents collisions
type key int

const userKey key = iota

// WithUser returns a copy of ctx that stores the Bitbucket User.
func WithUser(ctx context.Context, user *bitbucket.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// UserFromContext returns the Bitbucket User from the ctx.
func UserFromContext(ctx context.Context) (*bitbucket.User, error) {
	user, ok := ctx.Value(userKey).(*bitbucket.User)
	if !ok {
		return nil, errors.Errorf("bitbucketcloud: Context missing Bitbucket User")
	}
	return user, nil
}
