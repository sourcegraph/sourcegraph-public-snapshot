package repos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	gh "github.com/google/go-github/v43/github"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Payload struct {
	Name   string   `json:"name"`
	ID     int      `json:"id,omitempty"`
	Config Config   `json:"config"`
	Events []string `json:"events"`
	Active bool     `json:"active"`
	URL    string   `json:"url"`
}

type Config struct {
	Url          string `json:"url"`
	Content_type string `json:"content_type"`
	Secret       string `json:"secret"`
	Insecure_ssl string `json:"insecure_ssl"`
	Token        string `json:"token"`
	Digest       string `json:"digest,omitempty"`
}

type GithubWebhookAPI struct {
	Client *Client
}

func NewGithubWebhookAPI() (*GithubWebhookAPI, error) {
	cf := httpcli.ExternalClientFactory
	opts := []httpcli.Opt{}

	doer, err := cf.Doer(opts...)
	if err != nil {
		return nil, err
	}

	client, err := NewClient(doer)
	if err != nil {
		return nil, err
	}

	return &GithubWebhookAPI{Client: client}, nil
}

func (g *GithubWebhookAPI) Register(router *webhooks.GitHubWebhook) {
	router.Register(g.handleGitHubWebhook, "push")
}

func (g *GithubWebhookAPI) handleGitHubWebhook(ctx context.Context, extSvc *types.ExternalService, payload any) error {
	event, ok := payload.(*gh.PushEvent)
	if !ok {
		return errors.New("error converting to push event")
	}
	repoName := *event.Repo.URL
	name := api.RepoName(repoName[7:])
	fmt.Println("Repo:", name)
	repoupdater.DefaultClient.EnqueueRepoUpdate(ctx, name)
	return nil
}

func (g GithubWebhookAPI) CreateSyncWebhook(ctx context.Context,
	repoName string, targetUrl string, secret string, token string) (id int, err error) {
	url, err := urlBuilder(repoName)
	if err != nil {
		return -1, err
	}

	payload := Payload{
		Name:   "web",
		Active: true,
		Config: Config{
			Url:          fmt.Sprintf("%s/github-webhooks", targetUrl), // need to see from batch changes
			Content_type: "json",
			Secret:       secret,
			Token:        token,
			Insecure_ssl: "0",
		},
		Events: []string{
			"push",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return -1, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return -1, err
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("Authorization", fmt.Sprintf("token %s", token))

	resp, err := g.Client.do(ctx, req)
	if err != nil {
		return -1, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, err
	}
	// fmt.Println("res:", string(respBody))

	var obj Payload
	if err := json.Unmarshal(respBody, &obj); err != nil {
		return -1, err
	}

	return obj.ID, nil
}

func (g GithubWebhookAPI) ListSyncWebhooks(ctx context.Context, repoName string, token string) ([]Payload, error) {
	url, err := urlBuilder(repoName)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte("")))
	if err != nil {
		fmt.Println("making new request error:", err)
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("Authorization", fmt.Sprintf("token %s", token))

	resp, err := g.Client.do(ctx, req)
	if err != nil {
		return nil, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// fmt.Println(string(respBody))

	var obj []Payload
	if err := json.Unmarshal(respBody, &obj); err != nil {
		return nil, err
	}

	return obj, nil
}

func (g GithubWebhookAPI) FindSyncWebhook(ctx context.Context, repoName string, token string) (int, bool) {
	payloads, err := g.ListSyncWebhooks(ctx, repoName, token)
	if err != nil {
		return -1, false
	}

	for _, payload := range payloads {
		endpoint := payload.Config.Url
		parts := strings.Split(endpoint, "/")
		if parts[len(parts)-1] == "enqueue-repo-update" {
			return payload.ID, true
		}
	}

	return -1, false
}

func (g GithubWebhookAPI) DeleteSyncWebhook(ctx context.Context, repoName string, hookID int, token string) (bool, error) {
	url, err := urlBuilderWithID(repoName, hookID)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer([]byte("")))
	if err != nil {
		return false, err
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("Authorization", fmt.Sprintf("token %s", token))

	_, err = g.Client.do(ctx, req)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (g GithubWebhookAPI) TestPushSyncWebhook(ctx context.Context, repoName string, hookID int, token string) (bool, error) {
	u, err := urlBuilderWithID(repoName, hookID)
	if err != nil {
		return false, err
	}

	url := fmt.Sprintf("%s/tests", u)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte("")))
	if err != nil {
		return false, err
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("Authorization", fmt.Sprintf("token %s", token))

	resp, err := g.Client.do(ctx, req)
	if err != nil {
		return false, err
	}

	if resp.StatusCode != 204 {
		return false, errors.Newf("non-204 status code: %d", resp.StatusCode)
	}

	return true, nil
}

func urlBuilder(repoName string) (newUrl string, err error) {
	repoName = fmt.Sprintf("//%s", repoName)
	u, err := url.Parse(repoName)
	if err != nil {
		return "", errors.Newf("error parsing URL:", err)
	}

	if u.Host == "github.com" {
		newUrl = fmt.Sprintf("https://api.github.com/repos%s/hooks", u.Path)
	} else {
		newUrl = fmt.Sprintf("https://%s/api/v3/repos%s/hooks", u.Host, u.Path)
	}
	return newUrl, nil
}

func urlBuilderWithID(repoName string, hookID int) (newUrl string, err error) {
	repoName = fmt.Sprintf("//%s", repoName)
	u, err := url.Parse(repoName)
	if err != nil {
		return "", errors.Newf("error parsing URL:", err)
	}

	if u.Host == "github.com" {
		newUrl = fmt.Sprintf("https://api.github.com/repos%s/hooks/%d", u.Path, hookID)
	} else {
		newUrl = fmt.Sprintf("https://%s/api/v3/repos%s/hooks/%d", u.Host, u.Path, hookID)
	}
	return newUrl, nil
}
