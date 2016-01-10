package client

import "golang.org/x/oauth2"

// UpdateGlobalTokenSource updates Credentials.AccessToken with the
// newest access token each time it is refreshed.
type UpdateGlobalTokenSource struct{ oauth2.TokenSource }

func (ts UpdateGlobalTokenSource) Token() (*oauth2.Token, error) {
	tok, err := ts.TokenSource.Token()

	// Compare the token returned by TokenSource with the currently
	// set token. If the token had expired, the new token must be
	// set in the global Credentials for future use.
	//
	// Potentially, another goroutine might execute this same code
	// concurrently, in which case it will generate a different valid
	// token and the goroutine that calls SetAccessToken() last will
	// set its token for future use. This is fine as every goroutine
	// would still use a valid token. The alternative is to make this
	// an atomic "compare and swap" operation, but that is expensive
	// as every goroutine must acquire a read+write lock in every call
	// to this function.
	currTok := Credentials.GetAccessToken()
	if tok != nil && tok.AccessToken != currTok {
		Credentials.SetAccessToken(tok.AccessToken)
	}
	return tok, err
}
