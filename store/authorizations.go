package store

import (
	"errors"
	"time"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

// Authorizations manages OAuth2 authorization code grants, access
// tokens, etc. (maybe refresh tokens and revocations of all tokens in
// the future).
type Authorizations interface {
	// CreateAuthCode creates a new authorization code grant that may
	// later be exchanged for an access token.
	CreateAuthCode(ctx context.Context, req *sourcegraph.AuthorizationCodeRequest, expires time.Duration) (string, error)

	// MarkExchanged exchanges an authorization code for an access
	// token (per the OAuth2 spec). It invalidates the auth code and
	// returns the information associated with it. The caller can then
	// construct an access token using the information from the
	// exchanged grant. A grant may be exchanged at most once.
	//
	// TODO(sqs): if a code is exchanged twice, disable ALL access
	// tokens generated from it, per the OAuth2 spec (for security).
	MarkExchanged(ctx context.Context, code *sourcegraph.AuthorizationCode, clientID string) (*sourcegraph.AuthorizationCodeRequest, error)
}

var (
	ErrAuthCodeNotFound         = errors.New("OAuth2 auth code not found")
	ErrAuthCodeAlreadyExchanged = errors.New("OAuth2 auth code was already exchanged (is an attack in progress?)")
)
