
package bitbucketoauth

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// unexported key type prevents collisions
type key int

const userKey key = iota

// WithUser returns a copy of ctx that stores the GitLab User.
func WithUser(ctx context.Context, user *bitbucketcloud.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// UserFromContext returns the GitLab User from the ctx.
func UserFromContext(ctx context.Context) (*bitbucketcloud.User, error) {
	user, ok := ctx.Value(userKey).(*bitbucketcloud.User)
	if !ok {
		return nil, errors.Errorf("gitlab: Context missing Bitbucket User")
	}
	return user, nil
}
