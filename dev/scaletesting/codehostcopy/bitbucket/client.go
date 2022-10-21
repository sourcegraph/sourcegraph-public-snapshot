package bitbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/group"
)

type cache struct {
	Dir string
}

type Client struct {
	username   string
	password   string
	apiURL     *url.URL
	FetchLimit int
}

type getFunc func(context.Context) (*PagedResp, error)

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

func (c *Client) Get(ctx context.Context, url string, start int) (*PagedResp, error) {
	url = fmt.Sprintf("%s?start=%d&limit=%d", url, start, c.FetchLimit)
	log.Printf("GET %s\n", url)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.SetBasicAuth(c.username, c.password)

	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.Newf("Get failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result PagedResp
	json.NewDecoder(resp.Body).Decode(&result)

	return &result, nil
}

func (c *Client) Post(ctx context.Context, url string, data []byte) (*http.Response, error) {
	log.Printf("POST %s\n", url)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.SetBasicAuth(c.username, c.password)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.Newf("Post failed with status %d: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

func WithRetry(ctx context.Context, getFn getFunc, retries int) (*PagedResp, error) {
	for i := 0; i < retries; i++ {
		r, err := getFn(ctx)
		if err != nil {
			log.Println(err)
			log.Println("get request failed - waiting 15 seconds before retry")
			time.Sleep(15 * time.Second)
		} else {
			return r, nil
		}
	}

	return nil, fmt.Errorf("retries exhausted with gets")
}
func GetAll[T any](ctx context.Context, c *Client, url string) []T {
	start := 0
	count := 0
	items := make([]T, 0)
	for {
		ctx := ctx
		r, err := WithRetry(ctx, func(ctx context.Context) (*PagedResp, error) { return c.Get(ctx, url, start) }, 5)
		if err != nil {
			log.Printf("failed to get '%s': %v", url, err)
			log.Printf("continuing...")
			continue
		}

		count += r.Size
		for _, v := range r.Values {
			var tmp T
			err := json.Unmarshal(v, &tmp)
			if err != nil {
				log.Printf("failed to unmarshall item: %v", err)
				log.Println("continuing...")
			}
			items = append(items, tmp)
		}

		if r.IsLastPage {
			break
		}
		start = r.NextPageStart
	}
	return items
}

func (c *Client) GetProjectByKey(ctx context.Context, key string) (*Project, error) {
	u := c.url(fmt.Sprintf("/rest/api/latest/projects/%s", key))
	resp, err := c.Get(ctx, u, 0)
	if err != nil {
		return nil, err
	}

	var p Project
	if len(resp.Values) == 0 {
		return &p, errors.Newf("Project with key %q not found", key)
	} else if len(resp.Values) > 1 {
		log.Printf("WARN: More than one project returned for key %q", key)
	}
	err = json.Unmarshal(resp.Values[0], &p)
	if err != nil {
		return &p, errors.Wrapf(err, "failed to unmarshall project witht key: %v", key)
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

	resp, err := c.Post(ctx, url, rawRepoData)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result Repo
	err = json.Unmarshal(data, &result)
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
		links     map[string]any `json:"links"`
	}{
		Key: p.Key,
	})
	if err != nil {
		return nil, err
	}

	resp, err := c.Post(ctx, url, rawProjectData)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result Project
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	return &result, err
}

func (c *Client) ListProjects(ctx context.Context) []Project {
	cacheFilename := "cache.projects.json"
	projects, err := Load[[]Project](cacheFilename)
	if err != nil {
		url := c.url("/rest/api/latest/projects")
		projects = GetAll[Project](ctx, c, url)
		Save(cacheFilename, projects)
	}
	return projects
}

func (c *Client) Groups(ctx context.Context) []map[string]json.RawMessage {
	url := c.url("/rest/api/latest/admin/groups")
	return GetAll[map[string]json.RawMessage](ctx, c, url)
}
func (c *Client) ListRepos(ctx context.Context) []Repo {
	projects := c.ListProjects(ctx)
	return c.ListReposForProjects(ctx, projects)
}

func (c *Client) ListReposForProjects(ctx context.Context, projects []Project) []Repo {
	g := group.NewWithResults[[]Repo]().WithMaxConcurrency(10)
	repos := make([]Repo, 0)
	for _, p := range projects {
		g.Go(func() []Repo {
			url := c.url(fmt.Sprintf("/rest/api/latest/projects/%s/repos", p.Key))
			return GetAll[Repo](ctx, c, url)
		})
	}
	results := g.Wait()
	for _, r := range results {
		repos = append(repos, r...)

	}
	return repos
}

func Save(dst string, v any) error {
	f, err := os.Create(dst)
	defer f.Close()
	if err != nil {
		return err
	}

	return json.NewEncoder(f).Encode(&v)
}

func Load[T any](src string) (T, error) {
	f, err := os.Open(src)
	defer f.Close()
	var result T
	if err != nil {
		return result, err
	}

	err = json.NewDecoder(f).Decode(&result)
	if err != nil {
		return result, err
	}

	return result, err
}
