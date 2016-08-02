package oauth2client

import (
	"bytes"
	"errors"
)

// oauthAuthorizeClientState holds the state that the OAuth2 client
// passes to the provider and expects to receive back, during the
// OAuth2 authorization flow.
//
// No authentication of these values is performed; callers should
// check that, e.g., the State field matches the cookie nonce.
type oauthAuthorizeClientState struct {
	// Nonce is the state nonce that the OAuth2 client expects to
	// receive from the provider. It is generated and stored in a
	// cookie by writeNonceCookie.
	Nonce string

	// ReturnTo is the request URI on the client app that the resource owner
	// should be redirected to, after successful completion of OAuth2
	// authorization.
	ReturnTo string
}

func (s oauthAuthorizeClientState) MarshalText() ([]byte, error) {
	return []byte(s.Nonce + ":" + s.ReturnTo), nil
}

func (s *oauthAuthorizeClientState) UnmarshalText(text []byte) error {
	parts := bytes.SplitN(text, []byte(":"), 2)
	if len(parts) != 2 {
		return errors.New("invalid OAuth2 authorize client state: no ':' delimiter")
	}
	*s = oauthAuthorizeClientState{Nonce: string(parts[0]), ReturnTo: string(parts[1])}
	return nil
}
