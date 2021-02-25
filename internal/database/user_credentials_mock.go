package database

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
)

type MockUserCredentials struct {
	Create     func(context.Context, UserCredentialScope, auth.Authenticator) (*UserCredential, error)
	Update     func(ctx context.Context, credential *UserCredential) error
	Delete     func(context.Context, int64) error
	GetByID    func(context.Context, int64) (*UserCredential, error)
	GetByScope func(context.Context, UserCredentialScope) (*UserCredential, error)
	List       func(context.Context, UserCredentialsListOpts) ([]*UserCredential, int, error)
}
