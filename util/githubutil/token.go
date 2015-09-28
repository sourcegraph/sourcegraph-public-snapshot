package githubutil

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// A Token is a user's GitHub OAuth2 token and associated info, taken
// from the GitHub API.
type Token struct {
	ID     int
	URL    string
	Scopes []string
	Token  string
	App    struct {
		URL      string
		Name     string
		ClientID string `json:"client_id"`
	}
}

// GetToken gets info about a user's GitHub OAuth2 token.
func GetToken(clientID, clientSecret, token string) (*Token, error) {
	var ghtok *Token
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/applications/%s/tokens/%s", clientID, token), nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(clientID, clientSecret)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&ghtok); err != nil {
		return nil, err
	}
	return ghtok, nil
}
