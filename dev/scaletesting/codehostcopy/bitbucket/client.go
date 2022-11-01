package bitbucket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/lib/group"
)

type getResult[T any] struct {
	Result T
	Err    error
}

// APIError represents
type APIError struct {
	StatusCode int
	Message    string
}

var _ error = (*APIError)(nil)

func (apiErr *APIError) Error() string {
	return fmt.Sprintf("Status Code: %d Message: %s", apiErr.StatusCode, apiErr.Message)
}

// Client is a bitbucket Client for the Bitbucket Server REST API v1.0
type Client struct {
	setAuth SetAuthFunc
	apiURL  *url.URL
	http    http.Client
	// FetchLimit The amount of records to request per page
	FetchLimit int
}

type ClientOpt func(client *Client)
type SetAuthFunc func(req *http.Request)

func setBasicAuth(username, password string) SetAuthFunc {
	return func(req *http.Request) {
		req.SetBasicAuth(username, password)
	}
}

func setTokenAuth(token string) SetAuthFunc {
	return func(req *http.Request) {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	}
}

func WithTimeout(n time.Duration) ClientOpt {
	return func(client *Client) {
		client.http.Timeout = n
	}
}

// NewBasicAuthClient creates a Client that uses Basic Authentication. By default the FetchLimit is set to 150.
// To set the Timeout, use WithTimeout and pass it as a ClientOpt to this method.
func NewBasicAuthClient(username, password string, url *url.URL, opts ...ClientOpt) *Client {
	client := &Client{
		apiURL:     url,
		setAuth:    setBasicAuth(username, password),
		FetchLimit: 150,
	}
	for _, opt := range opts {
		opt(client)
	}

	return client
}

// NewTokenClient creates a Client that uses Token based authentication. By default the FetchLimit is set to 150.
// To set the Timout, use WithTimeout and pass it as a ClientOpt to this method.
func NewTokenClient(token string, url *url.URL, opts ...ClientOpt) *Client {
	client := &Client{
		apiURL:     url,
		setAuth:    setTokenAuth(token),
		FetchLimit: 150,
	}
	for _, opt := range opts {
		opt(client)
	}

	return client
}

func (c *Client) url(fragment string) string {
	return fmt.Sprintf("%s%s", c.apiURL.String(), fragment)
}

// getPaged issues a get request against a url that returns a paged response. The response is marshalled into
// a PagedResponse and returned. Otherwise an APIError is returned
func (c *Client) getPaged(ctx context.Context, url string, start int) (*PagedResp, error) {
	url = fmt.Sprintf("%s?start=%d&limit=%d", url, start, c.FetchLimit)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	c.setAuth(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	var pageResp PagedResp
	err = json.NewDecoder(resp.Body).Decode(&pageResp)
	if err != nil {
		return nil, err
	}

	return &pageResp, nil
}

func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	c.setAuth(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	return ioutil.ReadAll(resp.Body)
}

func (c *Client) post(ctx context.Context, url string, data []byte) ([]byte, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(data))
	c.setAuth(req)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	return ioutil.ReadAll(resp.Body)
}

// getAll continuously calls getPaged by adjusting the start query parameter based on the previous
// paged response. A GetResult is returned which contains the results as well as any errors that were
// encountered.
func getAll[T any](ctx context.Context, c *Client, url string) ([]getResult[T], error) {
	start := 0
	count := 0
	items := make([]getResult[T], 0)
	var apiErr *APIError
	for {
		ctx := ctx
		resp, err := c.getPaged(ctx, url, start)
		// If the error is a APIError we store the error and continue, otherwise
		// something severe is wrong and we stop and exit early
		if err != nil && errors.As(err, &apiErr) {
			// record the error and move on
			var value getResult[T]
			value.Err = err
			items = append(items, value)
			continue
		} else if err != nil {
			return nil, err
		}

		count += resp.Size
		for _, v := range resp.Values {
			var value getResult[T]
			value.Err = json.Unmarshal(v, &value.Result)
			items = append(items, value)
		}

		if resp.IsLastPage {
			break
		}
		start = resp.NextPageStart
	}
	return items, nil
}

func (c *Client) GetProjectByKey(ctx context.Context, key string) (*Project, error) {
	key = strings.ToUpper(key)
	u := c.url(fmt.Sprintf("/rest/api/latest/projects/%s", key))
	respData, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}

	var p Project
	err = json.Unmarshal(respData, &p)
	if err != nil {
		return &p, errors.Wrapf(err, "failed to unmarshall project with key: %s", key)
	}

	return &p, nil
}

func (c *Client) CreateRepo(ctx context.Context, p *Project, repoName string) (*Repo, error) {
	url := c.url(fmt.Sprintf("/rest/api/latest/projects/%s/repos", p.Key))

	rawRepoData, err := json.Marshal(struct {
		Name  string         `json:"name"`
		ScmId string         `json:"scmId"`
		Slug  string         `json:"slug"`
		Links map[string]any `json:"links"`
	}{
		Name:  repoName,
		ScmId: "git",
		Slug:  repoName,
	})
	if err != nil {
		return nil, err
	}

	respData, err := c.post(ctx, url, rawRepoData)
	if err != nil {
		return nil, err
	}

	var result Repo
	err = json.Unmarshal(respData, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) CreateProject(ctx context.Context, p *Project) (*Project, error) {
	url := c.url("/rest/api/latest/projects")

	rawProjectData, err := json.Marshal(struct {
		Key       string         `json:"key"`
		Avatar    string         `json:"avatar"`
		AvatarUrl string         `json:"avatarUrl"`
		Links     map[string]any `json:"links"`
	}{
		Key:   p.Key,
		Links: make(map[string]any),
	})
	if err != nil {
		return nil, err
	}

	respData, err := c.post(ctx, url, rawProjectData)
	if err != nil {
		return nil, err
	}

	var result Project
	err = json.Unmarshal(respData, &result)
	if err != nil {
		return nil, err
	}

	return &result, err
}

func (c *Client) ListProjects(ctx context.Context) ([]*Project, error) {
	var err error
	url := c.url("/rest/api/latest/projects")
	all, err := getAll[*Project](ctx, c, url)
	if err != nil {
		return nil, err
	}

	results, err := extractResults(all)
	return results, err
}

func (c *Client) ListRepos(ctx context.Context) ([]*Repo, error) {
	projects, err := c.ListProjects(ctx)
	if err != nil {
		return nil, err
	}
	return c.ListReposForProjects(ctx, projects)
}

func extractResults[T any](items []getResult[T]) ([]T, error) {
	var err error
	results := make([]T, 0)
	for _, r := range items {
		if r.Err != nil {
			err = errors.Append(err, r.Err)
		} else {
			results = append(results, r.Result)
		}
	}

	return results, err
}

func (c *Client) ListReposForProjects(ctx context.Context, projects []*Project) ([]*Repo, error) {
	g := group.NewWithResults[getResult[[]*Repo]]().WithMaxConcurrency(10)
	repos := make([]*Repo, 0)
	for _, p := range projects {

		g.Go(func() getResult[[]*Repo] {
			url := c.url(fmt.Sprintf("/rest/api/latest/projects/%s/repos", p.Key))
			all, err := getAll[*Repo](ctx, c, url)
			if err != nil {
				return getResult[[]*Repo]{
					Err: err,
				}
			}
			results, error := extractResults(all)
			return getResult[[]*Repo]{
				Result: results,
				Err:    error,
			}

		})
	}
	results := g.Wait()
	var err error
	for _, r := range results {
		if r.Err != nil {
			err = errors.Append(err, r.Err)
		} else if r.Result != nil {
			repos = append(repos, r.Result...)
		}

	}

	return repos, err
}
