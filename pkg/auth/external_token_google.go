package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	googleoauth2 "google.golang.org/api/oauth2/v2"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/google.golang.org/api/source/v1"
)

// FetchGoogleToken returns an access token for user uid.
// It's fetched from Google, after fetching a refresh token from Auth0.
func FetchGoogleToken(ctx context.Context, uid string) (*sourcegraph.ExternalToken, error) {
	refreshToken, err := FetchGoogleRefreshToken(ctx, uid)
	if err != nil {
		return nil, err
	}

	config := oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Endpoint:     google.Endpoint,
	}
	token, err := config.TokenSource(ctx, &oauth2.Token{
		RefreshToken: refreshToken.Token,
	}).Token()
	if err != nil {
		return nil, err
	}

	return &sourcegraph.ExternalToken{
		UID:   uid,
		Host:  "source.developers.google.com",
		Token: token.AccessToken,
		Scope: strings.Join(googleScopes, ","),
	}, nil
}

// FetchGoogleRefreshToken returns a refresh token, not access token.
// It fetches it from Auth0.
func FetchGoogleRefreshToken(ctx context.Context, uid string) (*sourcegraph.ExternalToken, error) {
	token, err := (&auth0TokenSource{
		ctx:        ctx,
		connection: "google-oauth2",
		uid:        uid,
	}).Token()
	if err != nil {
		return nil, err
	}

	return &sourcegraph.ExternalToken{
		UID:   uid,
		Host:  "source.developers.google.com",
		Token: token.RefreshToken,
		Scope: strings.Join(googleScopes, ","),
	}, nil
}

// FetchGoogleUsername fetches the Google username for Sourcegraph user uid.
// It's their Google email.
func FetchGoogleUsername(ctx context.Context, uid string) (string, error) {
	resp, err := oauth2.NewClient(ctx, auth0ManagementTokenSource).Get("https://" + Auth0Domain + "/api/v2/users/" + uid)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var payload struct {
		Identities []struct {
			Connection  string `json:"connection"`
			ProfileData struct {
				Email string `json:"email"`
			} `json:"profileData"`
		} `json:"identities"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}

	for _, identity := range payload.Identities {
		if identity.Connection == "google-oauth2" {
			return identity.ProfileData.Email, nil
		}
	}

	return "", fmt.Errorf("no email available")
}

var googleScopes = []string{
	googleoauth2.UserinfoProfileScope,
	googleoauth2.UserinfoEmailScope,

	source.CloudPlatformScope, // For source.projects.repos.list method.
}
