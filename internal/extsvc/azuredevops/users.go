package azuredevops

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"golang.org/x/oauth2"
)

const VISUAL_STUDIO_APP_URL = "https://app.vssps.visualstudio.com"

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
