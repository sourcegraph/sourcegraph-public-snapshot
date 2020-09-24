package graphqlbackend

import (
	"context"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
)

type personResolver struct {
	name  string
	email string

	// fetch + serve sourcegraph stored user information
	includeUserInfo bool

	// cache result because it is used by multiple fields
	once sync.Once
	user *types.User
	err  error
}

// resolveUser resolves the person to a user (using the email address). Not all persons can be
// resolved to a user.
func (r *personResolver) resolveUser(ctx context.Context) (*types.User, error) {
	r.once.Do(func() {
		if r.includeUserInfo && r.email != "" {
			r.user, r.err = db.Users.GetByVerifiedEmail(ctx, r.email)
			if errcode.IsNotFound(r.err) {
				r.err = nil
			}
		}
	})
	return r.user, r.err
}

func (r *personResolver) Name(ctx context.Context) (string, error) {
	user, err := r.resolveUser(ctx)
	if err != nil && !errcode.IsNotFound(err) {
		return "", err
	}
	if user != nil && user.Username != "" {
		return user.DisplayName, nil
	}

	return r.name, nil
}

func (r *personResolver) Email() string {
	return r.email
}

func (r *personResolver) DisplayName(ctx context.Context) (string, error) {
	user, err := r.resolveUser(ctx)
	if err != nil && !errcode.IsNotFound(err) {
		return "", err
	}
	if user != nil && user.DisplayName != "" {
		return user.DisplayName, nil
	}

	if name := strings.TrimSpace(r.name); name != "" {
		return name, nil
	}
	if r.email != "" {
		return r.email, nil
	}
	return "unknown", nil
}

func (r *personResolver) AvatarURL(ctx context.Context) (string, error) {
	user, err := r.resolveUser(ctx)
	if err != nil && !errcode.IsNotFound(err) {
		return "", err
	}
	if user != nil && user.AvatarURL != "" {
		return user.AvatarURL, nil
	}
	return "", nil
}

func (r *personResolver) User(ctx context.Context) (*UserResolver, error) {
	user, err := r.resolveUser(ctx)
	if user == nil || err != nil {
		return nil, err
	}
	return &UserResolver{user: user}, nil
}
