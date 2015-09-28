package oauth2util

// AuthorizeParams holds OAuth2 authorization params.
type AuthorizeParams struct {
	ClientID    string `schema:"client_id" url:"client_id,omitempty"`
	RedirectURI string `schema:"redirect_uri" url:"redirect_uri,omitempty"`
	Scope       string `schema:"scope" url:"scope,omitempty"`
	State       string `schema:"state" url:"state,omitempty"`

	// JWKS is the JWKS (JSON Web Key Set) for the OAuth2 client. It
	// is a non-standard extension to the OAuth2 authorize params that
	// is used by clients that use their public keys (e.g., the
	// Sourcegraph ID key) to authenticate with the server. Including
	// it in the URL in every OAuth2 authorize flow means that the
	// user can be prompted to register the client if it has not yet
	// been registered, which greatly simplifies the process.
	JWKS string `schema:"jwks" url:"jwks,omitempty"`
}

// TokenParams holds OAuth2 token params.
type TokenParams struct {
	GrantType   string `schema:"grant_type" url:"grant_type,omitempty"`
	Code        string `schema:"code" url:"code,omitempty"`
	Assertion   string `schema:"assertion" url:"assertion,omitempty"`
	RedirectURI string `schema:"redirect_uri" url:"redirect_uri,omitempty"`
	ClientID    string `schema:"client_id" url:"client_id,omitempty"`
}

// ReceiveParams holds OAuth2 receive-token params.
type ReceiveParams struct {
	ClientID string `schema:"client_id" url:"client_id,omitempty"`
	Code     string `schema:"code" url:"code,omitempty"`
	Scope    string `schema:"scope" url:"scope,omitempty"`
	State    string `schema:"state" url:"state,omitempty"`
}
