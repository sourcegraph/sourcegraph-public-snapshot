package golly

type TestingCredentials struct {
	Endpoint string
	// ProductionAccessToken is the access token to use when communicating with the production API.
	// If this is empty, the redacted token will be used instead.
	ProductionAccessToken string
	RedactedToken         string
}

func (t *TestingCredentials) AccessToken() string {
	if t.ProductionAccessToken != "" {
		return t.ProductionAccessToken
	}
	return t.RedactedToken
}
