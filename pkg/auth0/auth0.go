// Package auth0 can be removed when we are fully migrated off of Auth0.
package auth0

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

var domain = env.Get("AUTH0_DOMAIN", "", "domain of the Auth0 account")

var config = &oauth2.Config{
	ClientID:     env.Get("AUTH0_CLIENT_ID", "", "OAuth client ID for Auth0"),
	ClientSecret: env.Get("AUTH0_CLIENT_SECRET", "", "OAuth client secret for Auth0"),
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://" + domain + "/authorize",
		TokenURL: "https://" + domain + "/oauth/token",
	},
}

var auth0ManagementTokenSource = (&clientcredentials.Config{
	ClientID:     config.ClientID,
	ClientSecret: config.ClientSecret,
	TokenURL:     "https://" + domain + "/oauth/token",
	EndpointParams: url.Values{
		"audience": []string{"https://" + domain + "/api/v2/"},
	},
}).TokenSource(context.Background())

// CheckPassword asks Auth0 if the given username and password are valid.
//
// TODO(sqs): migrate the Auth0 password hashes to our own DB and remove this func
func CheckPassword(ctx context.Context, username, password string) (bool, error) {
	// Resource Owner Password (OAuth 2.0 grant)
	body, err := json.Marshal(map[string]string{
		"grant_type":    "password",
		"username":      username,
		"password":      password,
		"scope":         "",
		"client_id":     config.ClientID,
		"client_secret": config.ClientSecret,
	})
	if err != nil {
		return false, err
	}

	resp, err := http.Post(config.Endpoint.TokenURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}
