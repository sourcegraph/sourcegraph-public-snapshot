package sourcegraph

import "context"

type contextKey int

const (
	accessTokenKey contextKey = iota
)

// WithAccessToken returns a copy of the parent context that uses token
// as the access token for future API clients constructed using this
// context (with NewClientFromContext). It replaces (shadows) any
// previously set token in the context.
func WithAccessToken(parent context.Context, token string) context.Context {
	return context.WithValue(parent, accessTokenKey, token)
}

// AccessTokenFromContext returns the access token (if any) previously
// set in the context by WithAccessToken.
func AccessTokenFromContext(ctx context.Context) string {
	token, ok := ctx.Value(accessTokenKey).(string)
	if !ok {
		return ""
	}
	return token
}
