package auth

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"golang.org/x/oauth2"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

var (
	// ErrNoExternalAuthToken occurs when no external auth token exists
	// for a given user and host.
	ErrNoExternalAuthToken = errors.New("no external auth token found for user and host")
)

func FetchGitHubToken(ctx context.Context, uid string) (*sourcegraph.ExternalToken, error) {
	ts := oauth2.ReuseTokenSource(nil, &auth0TokenSource{
		connection: "github",
		uid:        uid,
	})

	token, err := ts.Token()
	if err != nil {
		return nil, err
	}

	scopes, err := getGitHubScopes(ctx, ts)
	if err != nil {
		return nil, err
	}

	return &sourcegraph.ExternalToken{
		UID:   uid,
		Host:  "github.com",
		Token: token.AccessToken,
		Scope: strings.Join(scopes, ","),
	}, nil
}

func getGitHubScopes(ctx context.Context, ts oauth2.TokenSource) ([]string, error) {
	resp, err := oauth2.NewClient(ctx, ts).Get("https://api.github.com/rate_limit")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return strings.Split(resp.Header.Get("X-OAuth-Scopes"), ", "), nil
}

// auth0TokenSource fetches an external token (e.g. for GitHub) for the given user from Auth0.
type auth0TokenSource struct {
	ctx        context.Context // If nil, context.Background() is used.
	connection string
	uid        string
}

func (ts *auth0TokenSource) Token() (*oauth2.Token, error) {
	ctx := ts.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	resp, err := oauth2.NewClient(ctx, auth0ManagementTokenSource).Get("https://" + Auth0Domain + "/api/v2/users/" + ts.uid)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload struct {
		Identities []struct {
			Connection   string `json:"connection"`
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"identities"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	for _, identity := range payload.Identities {
		if identity.Connection == ts.connection {
			return &oauth2.Token{
				TokenType:    "token",
				AccessToken:  identity.AccessToken,
				RefreshToken: identity.RefreshToken,
			}, nil
		}
	}

	return nil, ErrNoExternalAuthToken
}
