package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GithubStatusClient struct {
	baseURL    *url.URL
	httpClient *http.Client
}

type status struct {
	Description string `json:"description,omitempty"`
	Indicator   string `json:"indicator,omitempty"`
}

var baseURL = url.URL{
	Scheme: "https",
	Host:   "www.githubstatus.com",
	Path:   "/api/v2",
}

func NewGithubStatusClient() *GithubStatusClient {
	return &GithubStatusClient{
		baseURL:    &baseURL,
		httpClient: &http.Client{},
	}
}

func (c *GithubStatusClient) request(ctx context.Context) (*status, error) {
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

	var s status
	err = json.NewDecoder(resp.Body).Decode(&s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (c *GithubStatusClient) IsServiceDown(ctx context.Context) bool {
	res, err := c.request(ctx)
	if err != nil {
		return false
	}

	if res.Indicator == "critical" || res.Indicator == "major" {
		return true
	}
	return false
}
