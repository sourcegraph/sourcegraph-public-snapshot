package releaseregistry

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

const Endpoint = "https://releaseregistry.sourcegraph.com/v1/"

type Version struct {
	ID            int32      `json:"id"`
	Name          string     `json:"name"`
	Public        bool       `json:"public"`
	CreatedAt     time.Time  `json:"created_at"`
	PromotedAt    *time.Time `json:"promoted_at"`
	Version       string     `json:"version"`
	GitSHA        string     `json:"git_sha"`
	IsDevelopment bool       `json:"is_development"`
}

type Client struct {
	endpoint string
	client   http.Client
}

func NewClient(endpoint string) *Client {
	return &Client{
		endpoint: endpoint,
		client:   http.Client{},
	}
}

func (r *Client) newRequest(method, path string) (*http.Request, error) {
	urlPath, err := url.JoinPath(r.endpoint, path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, urlPath, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func (r *Client) ListVersions(ctx context.Context) ([]Version, error) {
	req, err := r.newRequest(http.MethodGet, "releases")
	if err != nil {
		return nil, err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	results := []Version{}

	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		return nil, err
	}

	return results, nil
}
