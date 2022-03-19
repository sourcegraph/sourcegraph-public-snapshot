package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type StatusApiClient struct {
	baseURL    *url.URL
	httpClient httpcli.Doer
}

type Status struct {
	Description string `json:"description,omitempty"`
	Indicator   string `json:"indicator,omitempty"`
}

var baseURL = url.URL{
	Scheme: "https",
	Host:   "www.githubstatus.com",
	Path:   "/api/v2",
}

func NewStatusApiClient(baseURL *url.URL, httpClient httpcli.Doer) *StatusApiClient {
	return &StatusApiClient{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

func (c *StatusApiClient) request(ctx context.Context) (status *Status, err error) {
	req, err := http.NewRequest("GET", c.baseURL.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	var resp *http.Response
	resp, err = c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		_, err := ioutil.ReadAll(resp.Body)
		return nil, errors.Wrap(err, fmt.Sprintf("unexpected response from Github Status API (%s)", req.URL))
	}

	var result Status
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result, nil

}

func (c *StatusApiClient) isGithubDown(ctx context.Context) bool {
	res, _ := c.request(ctx)

	if strings.EqualFold(res.State, "major") || strings.EqualFold(res.State, "critical") {
		return true
	}

	return false
}
