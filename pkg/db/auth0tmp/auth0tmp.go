// Package auth0tmp contains helpers that we can remove when we are fully migrated off
// of Auth0.
package auth0tmp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth0"
)

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
		"client_id":     auth0.Config.ClientID,
		"client_secret": auth0.Config.ClientSecret,
	})
	if err != nil {
		return false, err
	}

	resp, err := http.Post(auth0.Config.Endpoint.TokenURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}
