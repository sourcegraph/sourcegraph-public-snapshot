package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"

	"golang.org/x/oauth2"
)

var Auth0Domain = env.Get("AUTH0_DOMAIN", "", "domain of the Auth0 account")

var Auth0Config = &oauth2.Config{
	ClientID:     env.Get("AUTH0_CLIENT_ID", "", "OAuth client ID for Auth0"),
	ClientSecret: env.Get("AUTH0_CLIENT_SECRET", "", "OAuth client secret for Auth0"),
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://" + Auth0Domain + "/authorize",
		TokenURL: "https://" + Auth0Domain + "/oauth/token",
	},
}

var auth0ManagementTokenSource = oauth2.StaticTokenSource(&oauth2.Token{
	AccessToken: env.Get("AUTH0_MANAGEMENT_API_TOKEN", "", "management token for accessing the Auth0 user database"),
})

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
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to set app metadata")
	}

	return nil
}

// ListUsersByGitHubID lists registered Sourcegraph users by their GitHub ID.
func ListUsersByGitHubID(ctx context.Context, ghIDs []string) (map[string]User, error) {
	if len(ghIDs) == 0 {
		return nil, errors.New("Array of GitHub IDs is required")
	}

	resp, err := oauth2.NewClient(ctx, auth0ManagementTokenSource).Get("https://" + Auth0Domain + "/api/v2/users?q=identities.user_id%3A(" + url.QueryEscape(strings.Join(ghIDs, " ")) + ")")
	if err != nil {
		return nil, err
	}

	var users []User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, err
	}

	rUsers := make(map[string]User)
	for _, user := range users {
		for _, identity := range user.Identities {
			if identity.Provider == "github" {
				rUsers[identity.UserID] = user
			}
		}
	}
	for _, id := range ghIDs {
		if _, ok := rUsers[id]; !ok {
			delete(rUsers, id)
		}
	}

	return rUsers, nil
}

// User represents the user information returned from Auth0 profile information
type User struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	FamilyName    string `json:"family_name"`
	Gender        string `json:"gender"`
	GivenName     string `json:"given_name"`
	Identities    []struct {
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

// LinkAccount links account with uid with linkWithProvider provider and linkWithUID uid.
func LinkAccount(ctx context.Context, uid string, linkWithProvider, linkWithUID string) error {
	body, err := json.Marshal(struct {
		Provider string `json:"provider"`
		UserID   string `json:"user_id"`
	}{
		Provider: linkWithProvider,
		UserID:   linkWithUID,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://"+Auth0Domain+"/api/v2/users/"+uid+"/identities", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := oauth2.NewClient(ctx, auth0ManagementTokenSource).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("did not get acceptable status code: %v body: %q", resp.Status, body)
	}
	io.Copy(ioutil.Discard, resp.Body)

	return nil
}
