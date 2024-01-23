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

type Client struct {
	baseUrl     string
	secretToken string
}

// FetchSubscriptionBySAMSAccountID returns the user's Cody subscription for the sams_account_id.
// It returns nil, nil if the user does not have a Cody Pro subscription.
func (c *Client) FetchSubscriptionBySAMSAccountID(samsAccountID string) (*Subscription, error) {
	if c.baseUrl == "" {
		return nil, errors.New("SSC base url is not set")
	}

	if c.secretToken == "" {
		return nil, errors.New("SSC secret token is not set")
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/rest/svc/subscription/%s", c.baseUrl, samsAccountID), nil)
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
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var subscription Subscription
		err = json.Unmarshal(body, &subscription)
		if err != nil {
			return nil, err
		}

		return &subscription, nil
	}

	// 204 response indicates that the user does not have a Cody Pro subscription
	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	return nil, errors.Errorf("unexpected status code %d while fetching user subscription from SSC", resp.StatusCode)
}

func NewClient() *Client {
	sgconf := conf.Get().SiteConfig()

	return &Client{
		baseUrl:     sgconf.SscApiBaseUrl,
		secretToken: sgconf.SscApiSecret,
	}
}
