package azuredevops

import (
	"context"
	"net/http"
	"net/url"
)

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
	if _, err = c.do(ctx, req, "https://app.vssps.visualstudio.com", &p); err != nil {
		return Profile{}, err
	}

	return p, nil
}
