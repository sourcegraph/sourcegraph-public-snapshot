package auth0

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

var Domain = env.Get("AUTH0_DOMAIN", "", "domain of the Auth0 account")

var Config = &oauth2.Config{
	ClientID:     env.Get("AUTH0_CLIENT_ID", "", "OAuth client ID for Auth0"),
	ClientSecret: env.Get("AUTH0_CLIENT_SECRET", "", "OAuth client secret for Auth0"),
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://" + Domain + "/authorize",
		TokenURL: "https://" + Domain + "/oauth/token",
	},
}

var auth0ManagementTokenSource = (&clientcredentials.Config{
	ClientID:     Config.ClientID,
	ClientSecret: Config.ClientSecret,
	TokenURL:     "https://" + Domain + "/oauth/token",
	EndpointParams: url.Values{
		"audience": []string{"https://" + Domain + "/api/v2/"},
	},
}).TokenSource(context.Background())

// User represents the user information returned from Auth0 profile information
type User struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	FamilyName    string `json:"family_name"`
	Gender        string `json:"gender"`
	GivenName     string `json:"given_name"`
	AppMetadata   struct {
		DidLoginBefore bool `json:"did_login_before"`
	} `json:"app_metadata"`
	Identities []struct {
		Provider   string `json:"provider"`
		UserID     string `json:"user_id"`
		Connection string `json:"connection"`
		IsSocial   bool   `json:"isSocial"`
	} `json:"identities"`
	Locale   string `json:"locale"`
	Name     string `json:"name"`
	Nickname string `json:"nickname"`
	Picture  string `json:"picture"`
	UserID   string `json:"user_id"`
}

func GetAuth0User(ctx context.Context) (*User, error) {
	actor := actor.FromContext(ctx)
	uid := actor.AuthInfo().UID
	resp, err := oauth2.NewClient(ctx, auth0ManagementTokenSource).Get("https://" + Domain + "/api/v2/users/" + uid)
	if err != nil {
		return nil, err
	}
	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func GetEmailVerificationStatus(ctx context.Context) (bool, error) {
	user, err := GetAuth0User(ctx)
	if err != nil {
		return false, err
	}
	return user.EmailVerified, nil
}

type AppMetadata struct {
	AppMetadata map[string]interface{} `json:"app_metadata"`
}

func SetAppMetadata(ctx context.Context, uid string, key string, value interface{}) error {
	body, err := json.Marshal(AppMetadata{
		AppMetadata: map[string]interface{}{
			key: value,
		},
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", "https://"+Domain+"/api/v2/users/"+uid, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := oauth2.NewClient(ctx, auth0ManagementTokenSource).Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to set app metadata")
	}

	return nil
}
