package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

var Auth0Domain = os.Getenv("AUTH0_DOMAIN")

var Auth0Config = &oauth2.Config{
	ClientID:     os.Getenv("AUTH0_CLIENT_ID"),
	ClientSecret: os.Getenv("AUTH0_CLIENT_SECRET"),
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://" + Auth0Domain + "/authorize",
		TokenURL: "https://" + Auth0Domain + "/oauth/token",
	},
}

var auth0ManagementTokenSource = oauth2.StaticTokenSource(&oauth2.Token{AccessToken: os.Getenv("AUTH0_MANAGEMENT_API_TOKEN")})

func SetAppMetadata(ctx context.Context, uid string, key string, value interface{}) error {
	body, err := json.Marshal(struct {
		AppMetadata map[string]interface{} `json:"app_metadata"`
	}{
		AppMetadata: map[string]interface{}{
			key: value,
		},
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", "https://"+Auth0Domain+"/api/v2/users/"+uid, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := oauth2.NewClient(ctx, auth0ManagementTokenSource).Do(req)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to set app metadata")
	}

	return nil
}
