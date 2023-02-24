package azuredevops

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"golang.org/x/oauth2"
)

const VISUAL_STUDIO_APP_URL = "https://app.vssps.visualstudio.com/"

// GetAuthorizedProfile is used to return information about the currently authorized user. Should
// only be used for Azure Services (https://dev.azure.com).
func (c *Client) GetAuthorizedProfile(ctx context.Context) (Profile, error) {
	// See this link in the docs where the "/me" is documented in the URI parameters:
	// https://learn.microsoft.com/en-us/rest/api/azure/devops/profile/profiles/get?source=recommendations&view=azure-devops-rest-7.0&tabs=HTTP#uri-parameters
	reqURL := url.URL{Path: "/_apis/profile/profiles/me"}

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return Profile{}, err
	}

	var p Profile
	if _, err = c.do(ctx, req, VISUAL_STUDIO_APP_URL, &p); err != nil {
		return Profile{}, err
	}

	return p, nil
}

// SetExternalAccountData sets the user and token into the external account data blob.
func SetExternalAccountData(data *extsvc.AccountData, user *Profile, token *oauth2.Token) error {
	serializedUser, err := json.Marshal(user)
	if err != nil {
		return err
	}
	serializedToken, err := json.Marshal(token)
	if err != nil {
		return err
	}

	data.Data = extsvc.NewUnencryptedData(serializedUser)
	data.AuthData = extsvc.NewUnencryptedData(serializedToken)
	return nil
}

// GetExternalAccountData returns the deserialized user and token from the external account data
// JSON blob in a typesafe way.
func GetExternalAccountData(ctx context.Context, data *extsvc.AccountData) (usr *Profile, tok *oauth2.Token, err error) {
	if data.Data != nil {
		usr, err = encryption.DecryptJSON[Profile](ctx, data.Data)
		if err != nil {
			return nil, nil, err
		}
	}

	if data.AuthData != nil {
		tok, err = encryption.DecryptJSON[oauth2.Token](ctx, data.AuthData)
		if err != nil {
			return nil, nil, err
		}
	}

	return usr, tok, nil
}

func GetPublicExternalAccountData(ctx context.Context, accountData *extsvc.AccountData) (*extsvc.PublicAccountData, error) {
	data, _, err := GetExternalAccountData(ctx, accountData)
	if err != nil {
		return nil, err
	}

	email := strings.ToLower(data.EmailAddress)

	return &extsvc.PublicAccountData{
		DisplayName: &data.DisplayName,
		Login:       &email,
	}, nil
}
