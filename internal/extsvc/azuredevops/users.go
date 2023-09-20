package azuredevops

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"golang.org/x/oauth2"
)

const VisualStudioAppURL = "https://app.vssps.visualstudio.com/"

var MockVisualStudioAppURL string

// GetAuthorizedProfile is used to return information about the currently authorized user. Should
// only be used for Azure Services (https://dev.azure.com).
// See this link in the docs where the "/me" is documented in the URI parameters:
// https://learn.microsoft.com/en-us/rest/api/azure/devops/profile/profiles/get?source=recommendations&view=azure-devops-rest-7.0&tabs=HTTP#uri-parameters
func (c *client) GetAuthorizedProfile(ctx context.Context) (Profile, error) {
	reqURL := url.URL{Path: "/_apis/profile/profiles/me"}

	apiURL := VisualStudioAppURL
	if MockVisualStudioAppURL != "" {
		apiURL = MockVisualStudioAppURL
	}

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return Profile{}, err
	}

	var p Profile
	if _, err = c.do(ctx, req, apiURL, &p); err != nil {
		return Profile{}, err
	}

	return p, nil
}

func (c *client) ListAuthorizedUserOrganizations(ctx context.Context, profile Profile) ([]Org, error) {
	if MockVisualStudioAppURL == "" && !c.IsAzureDevOpsServices() {
		return nil, errors.New("ListAuthorizedUserOrganizations can only be used with Azure DevOps Services")
	}

	reqURL := url.URL{Path: "_apis/accounts"}

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, err
	}

	queryParams := req.URL.Query()
	queryParams.Set("memberId", profile.PublicAlias)
	req.URL.RawQuery = queryParams.Encode()

	apiURL := VisualStudioAppURL
	if MockVisualStudioAppURL != "" {
		apiURL = MockVisualStudioAppURL
	}

	response := ListAuthorizedUserOrgsResponse{}
	if _, err := c.do(ctx, req, apiURL, &response); err != nil {
		return nil, err
	}

	return response.Value, nil
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
func GetExternalAccountData(ctx context.Context, data *extsvc.AccountData) (profile *Profile, tok *oauth2.Token, err error) {
	if data.Data != nil {
		profile, err = encryption.DecryptJSON[Profile](ctx, data.Data)
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

	return profile, tok, nil
}

func GetPublicExternalAccountData(ctx context.Context, accountData *extsvc.AccountData) (*extsvc.PublicAccountData, error) {
	data, _, err := GetExternalAccountData(ctx, accountData)
	if err != nil {
		return nil, err
	}

	email := strings.ToLower(data.EmailAddress)

	return &extsvc.PublicAccountData{
		DisplayName: data.DisplayName,
		Login:       email,
	}, nil
}
