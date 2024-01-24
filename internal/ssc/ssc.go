package ssc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Client is the interface for making requests to the Self-Service Cody backend.
// This uses a REST API exposed from the service, and not the GraphQL API (which is
// instead used for user-facing frontend queries.)
type Client interface {
	FetchSubscriptionBySAMSAccountID(samsAccountID string) (*Subscription, error)
}

type SSCClient struct {
	baseURL     string
	secretToken string
}

func (c *SSCClient) sendRequest(method string, url string, body *Subscription) (code *int, err error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.secretToken))

	resp, err := httpcli.UncachedExternalDoer.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return &resp.StatusCode, err
		}

		err = json.Unmarshal(bytes, body)
		if err != nil {
			return nil, err
		}
	}

	return &resp.StatusCode, nil

}

// FetchSubscriptionBySAMSAccountID returns the user's Cody subscription for the sams_account_id.
// It returns nil, nil if the user does not have a Cody Pro subscription.
func (c *SSCClient) FetchSubscriptionBySAMSAccountID(samsAccountID string) (*Subscription, error) {
	if c.baseURL == "" {
		return nil, errors.New("SSC base url is not set")
	}

	if c.secretToken == "" {
		return nil, errors.New("SSC secret token is not set")
	}

	var subscription Subscription

	code, err := c.sendRequest(http.MethodGet, fmt.Sprintf("%s/rest/svc/subscription/%s", c.baseURL, samsAccountID), &subscription)
	if err != nil {
		return nil, err
	}

	if *code == http.StatusOK {
		return &subscription, nil
	}

	// 204 response indicates that the user does not have a Cody Pro subscription
	if *code == http.StatusNoContent {
		return nil, nil
	}

	return nil, errors.Errorf("unexpected status code %d while fetching user subscription from SSC", *code)
}

func NewClient() *SSCClient {
	sgconf := conf.Get().SiteConfig()

	return &SSCClient{
		baseURL:     sgconf.SscApiBaseUrl,
		secretToken: sgconf.SscApiSecret,
	}
}
