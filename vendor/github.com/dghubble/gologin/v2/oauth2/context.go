package oauth2

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
)

// unexported key type prevents collisions
type key int

const (
	tokenKey key = iota
	stateKey
)

// WithState returns a copy of ctx that stores the state value.
func WithState(ctx context.Context, state string) context.Context {
	return context.WithValue(ctx, stateKey, state)
}

// StateFromContext returns the state value from the ctx.
func StateFromContext(ctx context.Context) (string, error) {
	state, ok := ctx.Value(stateKey).(string)
	if !ok {
		return "", fmt.Errorf("oauth2: Context missing state value")
	}
	return state, nil
}

// WithToken returns a copy of ctx that stores the Token.
func WithToken(ctx context.Context, token *oauth2.Token) context.Context {
	return context.WithValue(ctx, tokenKey, token)
}

// TokenFromContext returns the Token from the ctx.
func TokenFromContext(ctx context.Context) (*oauth2.Token, error) {
	token, ok := ctx.Value(tokenKey).(*oauth2.Token)
	if !ok {
		return nil, fmt.Errorf("oauth2: Context missing Token")
	}
	return token, nil
}
