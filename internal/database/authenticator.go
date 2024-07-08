package database

import (
	"context"
	"database/sql/driver"
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	ghauth "github.com/sourcegraph/sourcegraph/internal/github_apps/auth"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// AuthenticatorType defines all possible types of authenticators stored in the database.
type AuthenticatorType string

// Define credential type strings that we'll use when encoding credentials.
const (
	AuthenticatorTypeOAuthClient                        AuthenticatorType = "OAuthClient"
	AuthenticatorTypeBasicAuth                          AuthenticatorType = "BasicAuth"
	AuthenticatorTypeBasicAuthWithSSH                   AuthenticatorType = "BasicAuthWithSSH"
	AuthenticatorTypeOAuthBearerToken                   AuthenticatorType = "OAuthBearerToken"
	AuthenticatorTypeOAuthBearerTokenWithSSH            AuthenticatorType = "OAuthBearerTokenWithSSH"
	AuthenticatorTypeBitbucketServerSudoableOAuthClient AuthenticatorType = "BitbucketSudoableOAuthClient"
	AuthenticatorTypeGitLabSudoableToken                AuthenticatorType = "GitLabSudoableToken"
	AuthenticatorTypeGitHubApp                          AuthenticatorType = "GitHubApp"
)

// NullAuthenticator represents an authenticator that may be null. It implements
// the sql.Scanner interface so it can be used as a scan destination, similar to
// sql.NullString. When the scanned value is null, the authenticator will be nil.
// It handles marshalling and unmarshalling the authenticator from and to JSON.
type NullAuthenticator struct{ A *auth.Authenticator }

// Scan implements the Scanner interface.
func (n *NullAuthenticator) Scan(value any) (err error) {
	switch value := value.(type) {
	case string:
		*n.A, err = UnmarshalAuthenticator(value)
		return err
	case nil:
		return nil
	default:
		return errors.Errorf("value is not string: %T", value)
	}
}

// Value implements the driver Valuer interface.
func (n NullAuthenticator) Value() (driver.Value, error) {
	if *n.A == nil {
		return nil, nil
	}
	return MarshalAuthenticator(*n.A)
}

// EncryptAuthenticator encodes an authenticator into a byte slice. If the given
// key is non-nil, it will also be encrypted.
func EncryptAuthenticator(ctx context.Context, key encryption.Key, a auth.Authenticator) ([]byte, string, error) {
	raw, err := MarshalAuthenticator(a)
	if err != nil {
		return nil, "", errors.Wrap(err, "marshalling authenticator")
	}

	data, keyID, err := encryption.MaybeEncrypt(ctx, key, raw)
	return []byte(data), keyID, err
}

// MarshalAuthenticator encodes an Authenticator into a JSON string.
func MarshalAuthenticator(a auth.Authenticator) (string, error) {
	var t AuthenticatorType
	switch a.(type) {
	case *auth.OAuthClient:
		t = AuthenticatorTypeOAuthClient
	case *auth.BasicAuth:
		t = AuthenticatorTypeBasicAuth
	case *auth.BasicAuthWithSSH:
		t = AuthenticatorTypeBasicAuthWithSSH
	case *auth.OAuthBearerToken:
		t = AuthenticatorTypeOAuthBearerToken
	case *auth.OAuthBearerTokenWithSSH:
		t = AuthenticatorTypeOAuthBearerTokenWithSSH
	case *bitbucketserver.SudoableOAuthClient:
		t = AuthenticatorTypeBitbucketServerSudoableOAuthClient
	case *gitlab.SudoableToken:
		t = AuthenticatorTypeGitLabSudoableToken
	case *ghauth.InstallationAuthenticator:
	case *ghauth.GitHubAppAuthenticator:
		t = AuthenticatorTypeGitHubApp
	default:
		return "", errors.Errorf("unknown Authenticator implementation type: %T", a)
	}

	raw, err := json.Marshal(struct {
		Type AuthenticatorType
		Auth auth.Authenticator
	}{
		Type: t,
		Auth: a,
	})
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

// UnmarshalAuthenticator decodes a JSON string into an Authenticator.
func UnmarshalAuthenticator(raw string) (auth.Authenticator, error) {
	// We do two unmarshals: the first just to get the type, and then the second
	// to actually unmarshal the authenticator itself.
	var partial struct {
		Type AuthenticatorType
		Auth json.RawMessage
	}
	if err := json.Unmarshal([]byte(raw), &partial); err != nil {
		return nil, err
	}

	var a any
	switch partial.Type {
	case AuthenticatorTypeOAuthClient:
		a = &auth.OAuthClient{}
	case AuthenticatorTypeBasicAuth:
		a = &auth.BasicAuth{}
	case AuthenticatorTypeBasicAuthWithSSH:
		a = &auth.BasicAuthWithSSH{}
	case AuthenticatorTypeOAuthBearerToken:
		a = &auth.OAuthBearerToken{}
	case AuthenticatorTypeOAuthBearerTokenWithSSH:
		a = &auth.OAuthBearerTokenWithSSH{}
	case AuthenticatorTypeBitbucketServerSudoableOAuthClient:
		a = &bitbucketserver.SudoableOAuthClient{}
	case AuthenticatorTypeGitLabSudoableToken:
		a = &gitlab.SudoableToken{}
	default:
		return nil, errors.Errorf("unknown credential type: %s", partial.Type)
	}

	if err := json.Unmarshal(partial.Auth, &a); err != nil {
		return nil, err
	}

	return a.(auth.Authenticator), nil
}
