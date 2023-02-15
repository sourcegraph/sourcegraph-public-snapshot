package azuredevops

import (
	"context"
	"net/http"
	"net/url"
)

// AzureServicesProfile is used to return information about the authorized user, should only be used for Azure Services (https://dev.azure.com)
func (c *Client) AzureServicesProfile(ctx context.Context) (Profile, error) {
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
