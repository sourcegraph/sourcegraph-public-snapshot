package notionapi

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
)

type AuthenticationService interface {
	CreateToken(ctx context.Context, request *TokenCreateRequest) (*TokenCreateResponse, error)
}

type AuthenticationClient struct {
	apiClient *Client
}

// Creates an access token that a third-party service can use to authenticate
// with Notion.
//
// See https://developers.notion.com/reference/create-a-token
func (cc *AuthenticationClient) CreateToken(ctx context.Context, request *TokenCreateRequest) (*TokenCreateResponse, error) {
	res, err := cc.apiClient.requestImpl(ctx, http.MethodPost, "oauth/token", nil, request, true, decodeTokenCreateError)
	if err != nil {
		return nil, err
	}

	defer func() {
		if errClose := res.Body.Close(); errClose != nil {
			log.Println("failed to close body, should never happen")
		}
	}()

	var response TokenCreateResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func decodeTokenCreateError(data []byte) error {
	var apiErr TokenCreateError
	err := json.Unmarshal(data, &apiErr)
	if err != nil {
		return err
	}
	return &apiErr
}

// TokenCreateRequest represents the request body for AuthenticationClient.CreateToken.
type TokenCreateRequest struct {
	// A unique random code that Notion generates to authenticate with your service,
	// generated when a user initiates the OAuth flow.
	Code string `json:"code"`
	// A constant string: "authorization_code".
	GrantType string `json:"grant_type"`
	// The "redirect_uri" that was provided in the OAuth Domain & URI section of
	// the integration's Authorization settings. Do not include this field if a
	// "redirect_uri" query param was not included in the Authorization URL
	// provided to users. In most cases, this field is required.
	RedirectUri string `json:"redirect_uri,omitempty"`
	// Required if and only when building Link Preview integrations (otherwise
	// ignored). An object with key and name properties. key should be a unique
	// identifier for the account. Notion uses the key to determine whether or not
	// the user is re-connecting the same account. name should be some way for the
	// user to know which account they used to authenticate with your service. If
	// a user has authenticated Notion with your integration before and key is the
	// same but name is different, then Notion updates the name associated with
	// your integration.
	ExternalAccount ExternalAccount `json:"external_account,omitempty"`
}

type ExternalAccount struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type TokenCreateResponse struct {
	AccessToken          string `json:"access_token"`
	BotId                string `json:"bot_id"`
	DuplicatedTemplateId string `json:"duplicated_template_id,omitempty"`

	// Owner can be { "workspace": true } OR a User object.
	// Ref: https://developers.notion.com/docs/authorization#step-4-notion-responds-with-an-access_token-and-some-additional-information
	Owner         interface{} `json:"owner,omitempty"`
	WorkspaceIcon string      `json:"workspace_icon"`
	WorkspaceId   string      `json:"workspace_id"`
	WorkspaceName string      `json:"workspace_name"`
}
