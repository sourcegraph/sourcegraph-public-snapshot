package bitbucket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/lib/group"
)

type cache struct {
	Dir string
}

type GetResult[T any] struct {
	Result T
	Err    error
}

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
	username string
	password string
	apiURL   *url.URL
	// FetchLimit The amount of records to request per page
	FetchLimit int
}

type getFunc func(context.Context) (*PagedResp, error)

// NewClient creates a Client with the username, password and url. The url is the base url which should have the following form
// http://host:port. The client will append /rest/api/latest to the base url. By default the FetchLimit is set to 150
func NewClient(username, password string, url *url.URL) *Client {
	return &Client{
		username:   username,
		password:   password,
		apiURL:     url,
		FetchLimit: 150,
	}
}

func (c *Client) Domain() string {
	return c.apiURL.Hostname()
}

func (c *Client) url(fragment string) string {
	return fmt.Sprintf("%s%s", c.apiURL.String(), fragment)
}

func (c *Client) getPaged(ctx context.Context, url string, start int) (*PagedResp, error) {
	url = fmt.Sprintf("%s?start=%d&limit=%d", url, start, c.FetchLimit)
	log.Printf("GET %s\n", url)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.SetBasicAuth(c.username, c.password)

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

func (c *Client) Get(ctx context.Context, url string) ([]byte, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.SetBasicAuth(c.username, c.password)

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

func (c *Client) Post(ctx context.Context, url string, data []byte) ([]byte, error) {
	log.Printf("POST %s\n", url)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(data))
	req.SetBasicAuth(c.username, c.password)
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

func getAll[T any](ctx context.Context, c *Client, url string) []GetResult[T] {
	start := 0
	count := 0
	items := make([]GetResult[T], 0)
	for {
		ctx := ctx
		resp, err := c.getPaged(ctx, url, start)
		if err != nil {
			// record the error and move on
			var value GetResult[T]
			value.Err = err
			items = append(items, value)
			continue
		}

		count += resp.Size
		for _, v := range resp.Values {
			var value GetResult[T]
			value.Err = json.Unmarshal(v, &value.Result)
			items = append(items, value)
		}

		if resp.IsLastPage {
			break
		}
		start = resp.NextPageStart
	}
	return items
}

func (c *Client) GetProjectByKey(ctx context.Context, key string) (*Project, error) {
	key = strings.ToUpper(key)
	u := c.url(fmt.Sprintf("/rest/api/latest/projects/%s", key))
	respData, err := c.Get(ctx, u)
	if err != nil {
		return nil, err
	}

	var p Project
	err = json.Unmarshal(respData, &p)
	if err != nil {
		return &p, errors.Wrapf(err, "failed to unmarshall project witht key: %s", key)
	}

	return &p, nil
}

func (c *Client) CreateRepo(ctx context.Context, p *Project, repoName string) (*Repo, error) {
	url := c.url(fmt.Sprintf("/rest/api/latest/projects/%s/repos", p.Key))

	rawRepoData, err := json.Marshal(struct {
		Name  string         `json:"name"`
		ScmId string         `json:"scmId"`
		Slug  string         `json:"slug"`
		links map[string]any `json:"links"`
	}{
		Name:  repoName,
		ScmId: "git",
		Slug:  repoName,
	})
	if err != nil {
		return nil, err
	}

	respData, err := c.Post(ctx, url, rawRepoData)
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

	respData, err := c.Post(ctx, url, rawProjectData)
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
	all := getAll[*Project](ctx, c, url)

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

func extractResults[T any](items []GetResult[T]) ([]T, error) {
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
	g := group.NewWithResults[GetResult[[]*Repo]]().WithMaxConcurrency(10)
	repos := make([]*Repo, 0)
	for _, p := range projects {

		g.Go(func() GetResult[[]*Repo] {
			url := c.url(fmt.Sprintf("/rest/api/latest/projects/%s/repos", p.Key))
			all := getAll[*Repo](ctx, c, url)
			results, error := extractResults(all)
			return GetResult[[]*Repo]{
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
		} else {
			repos = append(repos, r.Result...)
		}

	}

	return repos, err
}
