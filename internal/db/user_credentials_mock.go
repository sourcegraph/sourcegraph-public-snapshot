package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
)

type MockUserCredentials struct {
	Delete     func(context.Context, int64) error
	GetByID    func(context.Context, int64) (*UserCredential, error)
	GetByScope func(context.Context, UserCredentialScope) (*UserCredential, error)
	List       func(context.Context, UserCredentialsListOpts) ([]*UserCredential, int, error)
	Upsert     func(context.Context, UserCredentialScope, auth.Authenticator) (*UserCredential, error)
}
