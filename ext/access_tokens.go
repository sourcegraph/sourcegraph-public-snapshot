package ext

import (
	"os"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/platform/storage"
)

const credentialsBucket = "credentials"

type TokenNotFoundError struct {
	msg string
}

func (e TokenNotFoundError) Error() string { return e.msg }

// AccessTokens contains methods for storing global access tokens needed
// to authenticate against external services.
type AuthStore struct{}

type Credentials struct {
	Token string
}

func (s *AuthStore) storage(ctx context.Context) storage.System {
	return storage.Namespace(ctx, "core.external-auth", "")
}

func (s *AuthStore) Set(ctx context.Context, host string, cred Credentials) error {
	fs := s.storage(ctx)
	return storage.PutJSON(fs, credentialsBucket, host, cred)
}

func (s *AuthStore) Get(ctx context.Context, host string) (Credentials, error) {
	cred := Credentials{}
	fs := s.storage(ctx)
	err := storage.GetJSON(fs, credentialsBucket, host, &cred)
	if err != nil {
		if os.IsNotExist(err) {
			return Credentials{}, nil
		}
		return Credentials{}, err
	}

	return cred, nil
}
