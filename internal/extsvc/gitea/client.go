package gitea

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/peterhellberg/link"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// Client access a Gitea via the REST API.
type Client struct {
	// Config is the code host connection config for this client
	Config *schema.GiteaConnection

	// URL is the base URL of Gitea.
	URL *url.URL

	// HTTP Client used to communicate with the API
	httpClient httpcli.Doer

	// RateLimit is the self-imposed rate limiter
	rateLimit *ratelimit.InstrumentedLimiter
}

// NewClient returns an optionally authenticated Gitea API client with the
// provided configuration. If a nil httpClient is provided,
// httpli.ExternalDoer will be used.
func NewClient(urn string, config *schema.GiteaConnection, httpClient httpcli.Doer) (*Client, error) {
	u, err := url.Parse(config.Url)
	if err != nil {
		return nil, err
	}

	if httpClient == nil {
		httpClient = httpcli.ExternalDoer
	}

	return &Client{
		Config:     config,
		URL:        u,
		httpClient: httpClient,
		rateLimit:  ratelimit.DefaultRegistry.Get(urn),
	}, nil
}

// Repository is a subset of the fields Gitea model Repository.
type Repository struct {
	ID          int64  `json:"id"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	Private     bool   `json:"private"`
	Fork        bool   `json:"fork"`
	CloneURL    string `json:"clone_url"`
	Stars       int    `json:"stars_count"`
	Forks       int    `json:"forks_count"`
	Archived    bool   `json:"archived"`
}

// ReposSearch https://try.gitea.io/api/swagger#/repository/repoSearch
func (c *Client) ReposSearch(ctx context.Context, params url.Values) *RepositoryIterator {
	nextPage := c.URL.ResolveReference(&url.URL{Path: "/api/v1/repos/search", RawQuery: params.Encode()}).String()

	next := func() ([]*Repository, error) {
		if nextPage == "" {
			return nil, nil
		}

		req, err := http.NewRequestWithContext(ctx, "GET", nextPage, nil)
		if err != nil {
			return nil, err
		}

		var searchResult struct {
			Data []*Repository
		}

		respHeaders, err := c.do(ctx, req, &searchResult)
		if err != nil {
			return nil, err
		}

		if l := link.ParseHeader(respHeaders)["next"]; l != nil {
			nextPage = l.String()
		} else {
			nextPage = ""
		}

		return searchResult.Data, nil
	}

	return &RepositoryIterator{next: next}
}

func (c *Client) do(ctx context.Context, req *http.Request, result any) (http.Header, error) {
	if c.Config.Token != "" {
		req.Header.Add("Authorization", "token "+c.Config.Token)
	}
	req.Header.Add("Accept", "application/json")

	if err := c.rateLimit.Wait(ctx); err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, errors.WithStack(&httpError{
			URL:        req.URL,
			StatusCode: resp.StatusCode,
			Body:       bs,
		})
	}

	return resp.Header, json.Unmarshal(bs, result)
}

type httpError struct {
	StatusCode int
	URL        *url.URL
	Body       []byte
}

func (e *httpError) Error() string {
	return fmt.Sprintf("Gitea API HTTP error: code=%d url=%q body=%q", e.StatusCode, e.URL, e.Body)
}

func (e *httpError) Unauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

func (e *httpError) NotFound() bool {
	return e.StatusCode == http.StatusNotFound
}
