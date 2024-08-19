package bitbucket

import (
	"context"
	"fmt"
)

// unexported key type prevents collisions
type key int

const (
	userKey key = iota
)

// WithUser returns a copy of ctx that stores the Bitbucket User.
func WithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// UserFromContext returns the Bitbucket User from the ctx.
func UserFromContext(ctx context.Context) (*User, error) {
	user, ok := ctx.Value(userKey).(*User)
	if !ok {
		return nil, fmt.Errorf("bitbucket: Context missing Bitbucket User")
	}
	return user, nil
}
