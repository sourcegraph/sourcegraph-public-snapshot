package authzchecked

import (
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/store"
)

// ExternalAuthTokens wraps base's methods with authorization checks.
func ExternalAuthTokens(base store.ExternalAuthTokens) store.ExternalAuthTokens {
	return &externalAuthTokens{base}
}

// externalAuthTokens adds authorization checks to an underlying
// ExternalAuthTokens.
type externalAuthTokens struct {
	noauthz store.ExternalAuthTokens
}

func (s *externalAuthTokens) GetUserToken(ctx context.Context, user int, host, clientID string) (*auth.ExternalAuthToken, error) {
	if err := checkActorUID(ctx, user); err != nil {
		return nil, err
	}
	return s.noauthz.GetUserToken(ctx, user, host, clientID)
}

func (s *externalAuthTokens) SetUserToken(ctx context.Context, tok *auth.ExternalAuthToken) error {
	if err := checkActorUID(ctx, tok.User); err != nil {
		return err
	}
	return s.noauthz.SetUserToken(ctx, tok)
}
