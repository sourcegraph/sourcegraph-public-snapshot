package bitbucket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
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
		req.Header.Set("Authorization", "Bearer "+token)
	}
}

func WithTimeout(n time.Duration) ClientOpt {
	return func(client *Client) {
		client.http.Timeout = n
	}
}

// NewBasicAuthClient creates a Client that uses Basic Authentication. By default the FetchLimit is set to 150.
// To set the Timeout, use WithTimeout and pass it as a ClientOpt to this method. This is the preferred client
// interacting with the REST API, since it is able to perform some calls the Token based client is not allowed
// to do by the Bitbucket API ie. CreateRepo
func NewBasicAuthClient(username, password string, url *url.URL, opts ...ClientOpt) *Client {
	client := &Client{
		apiURL:  url,
		setAuth: setBasicAuth(username, password),
	}
	for _, opt := range opts {
		opt(client)
	}

	return client
}

// NewTokenClient creates a Client that uses Token based authentication. By default the FetchLimit is set to 150.
// To set the Timout, use WithTimeout and pass it as a ClientOpt to this method. This client is more restrictive
// than the BasicAuth client. The restriction is not imposed by the client itself, but by the nature of the Bitbucket
// REST API. For more power like create projects and repos, use the Basic auth client.
func NewTokenClient(token string, url *url.URL, opts ...ClientOpt) *Client {
	client := &Client{
		apiURL:  url,
		setAuth: setTokenAuth(token),
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
func (c *Client) getPaged(ctx context.Context, url string, start int, perPage int) (*PagedResp, error) {
	url = fmt.Sprintf("%s?start=%d&limit=%d", url, start, perPage)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	c.setAuth(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
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
		body, _ := io.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	return io.ReadAll(resp.Body)
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
		body, _ := io.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	return io.ReadAll(resp.Body)
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
		resp, err := c.getPaged(ctx, url, start, 30)
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

func (c *Client) GetRepo(ctx context.Context, key string, name string) (*Repo, error) {
	key = strings.ToUpper(key)
	u := c.url(fmt.Sprintf("/rest/api/latest/projects/%s/repos/%s", key, name))
	respData, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}

	var repo Repo
	err = json.Unmarshal(respData, &repo)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshall repo: %s", name)
	}

	return &repo, nil
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

// CreateRepo creates a repo within the given project with the given name.
func (c *Client) CreateRepo(ctx context.Context, p *Project, repoName string) (*Repo, error) {
	endpointUrl := c.url(fmt.Sprintf("/rest/api/latest/projects/%s/repos", p.Key))

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

	respData, err := c.post(ctx, endpointUrl, rawRepoData)
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

// CreateProject creates a Project. The PROJECT_CREATE permission is required for this call, which one cannot
// assign to a Token in the Bitbucket admin interface. The only available Token permissions are PROJECT_ADMIN,
// PROJECT_READ, PROJECT_WRITE.
//
// PROJECT_ADMIN does not imply PROJECT_CREATE which is only associated with a
// authenticated user. Therefore, it is strongly recommended, that if you want to create a project to use the
// BasicAuth client.
func (c *Client) CreateProject(ctx context.Context, p *Project) (*Project, error) {
	endpointUrl := c.url("/rest/api/latest/projects")

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

	respData, err := c.post(ctx, endpointUrl, rawProjectData)
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
	endpointUrl := c.url("/rest/api/latest/projects")
	all, err := getAll[*Project](ctx, c, endpointUrl)
	if err != nil {
		return nil, err
	}

	results, err := extractResults(all)
	return results, err
}

func (c *Client) ListRepos(ctx context.Context, project *Project, page int, perPage int) ([]*Repo, int, error) {
	return c.ListReposForProject(ctx, project, page, perPage)
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

func (c *Client) ListReposForProject(ctx context.Context, project *Project, page int, perPage int) ([]*Repo, int, error) {
	repos := make([]*Repo, 0)
	endpointUrl := c.url(fmt.Sprintf("/rest/api/latest/projects/%s/repos", project.Key))
	resp, err := c.getPaged(ctx, endpointUrl, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	for _, v := range resp.Values {
		var repo Repo
		err := json.Unmarshal(v, &repo)
		if err != nil {
			return nil, 0, err
		}
		repos = append(repos, &repo)
	}
	return repos, resp.NextPageStart, nil
}
