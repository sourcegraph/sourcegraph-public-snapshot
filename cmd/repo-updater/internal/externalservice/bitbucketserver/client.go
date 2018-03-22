package bitbucketserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"golang.org/x/net/context/ctxhttp"
)

// Client access a Bitbucket Server via the REST API.
type Client struct {
	// URL is the base URL of Bitbucket Server.
	URL *url.URL

	// Token is the personal access token for accessing the
	// server. https://bitbucket.example.com/plugins/servlet/access-tokens/manage
	Token string

	// The username and password credentials for accessing the server. Typically these are only
	// used when the server doesn't support personal access tokens (such as Bitbucket Server
	// version 5.4 and older). If both Token and Username/Password are specifed, Token is used.
	Username, Password string

	// HTTPClient is the client used to access Bitbucket Server. To enabled
	// tracing, ensure the transport includes nethttp.Transport.
	HTTPClient *http.Client
}

func (c *Client) Repo(ctx context.Context, projectKey, repoSlug string) (*Repo, error) {
	u := fmt.Sprintf("rest/api/1.0/projects/%s/repos/%s", projectKey, repoSlug)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	var resp Repo
	err = c.do(ctx, req, &resp)
	return &resp, err
}

func (c *Client) Repos(ctx context.Context, pageToken *PageToken) ([]*Repo, *PageToken, error) {
	u := fmt.Sprintf("rest/api/1.0/repos%s", pageToken.Query())
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}
	var resp struct {
		*PageToken
		Values []*Repo
	}
	err = c.do(ctx, req, &resp)
	if err != nil {
		return nil, nil, err
	}
	return resp.Values, resp.PageToken, nil
}

func (c *Client) RecentRepos(ctx context.Context, pageToken *PageToken) ([]*Repo, *PageToken, error) {
	u := fmt.Sprintf("rest/api/1.0/profile/recent/repos%s", pageToken.Query())
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}
	var resp struct {
		*PageToken
		Values []*Repo
	}
	err = c.do(ctx, req, &resp)
	if err != nil {
		return nil, nil, err
	}
	return resp.Values, resp.PageToken, nil
}

func (c *Client) do(ctx context.Context, req *http.Request, result interface{}) error {
	req.URL = c.URL.ResolveReference(req.URL)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	// Authenticate request, preferring token.
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	} else if c.Username != "" || c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	req, ht := nethttp.TraceRequest(opentracing.GlobalTracer(), req,
		nethttp.OperationName("Bitbucket Server"),
		nethttp.ClientTrace(false))
	defer ht.Finish()

	resp, err := ctxhttp.Do(ctx, c.HTTPClient, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.WithStack(&httpError{URL: req.URL, StatusCode: resp.StatusCode})
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

type PageToken struct {
	Size          int  `json:"size"`
	Limit         int  `json:"limit"`
	IsLastPage    bool `json:"isLastPage"`
	Start         int  `json:"start"`
	NextPageStart int  `json:"nextPageStart"`
}

func (t *PageToken) HasMore() bool {
	if t == nil {
		return true
	}
	return !t.IsLastPage
}

func (t *PageToken) Query() string {
	if t == nil {
		return ""
	}
	v := url.Values{}
	if t.NextPageStart != 0 {
		v.Set("start", strconv.Itoa(t.NextPageStart))
	}
	if t.Limit != 0 {
		v.Set("limit", strconv.Itoa(t.Limit))
	}
	if len(v) == 0 {
		return ""
	}
	return "?" + v.Encode()
}

type Repo struct {
	Slug          string   `json:"slug"`
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	SCMID         string   `json:"scmId"`
	State         string   `json:"state"`
	StatusMessage string   `json:"statusMessage"`
	Forkable      bool     `json:"forkable"`
	Origin        *Repo    `json:"origin"`
	Project       *Project `json:"project"`
	Public        bool     `json:"public"`
	Links         struct {
		Clone []struct {
			Href string `json:"href"`
			Name string `json:"name"`
		} `json:"clone"`
		Self []struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
}

type Project struct {
	Key    string `json:"key"`
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Public bool   `json:"public"`
	Type   string `json:"type"`
	Links  struct {
		Self []struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"links"`
}

type httpError struct {
	StatusCode int
	URL        *url.URL
}

func (e *httpError) Error() string {
	return fmt.Sprintf("unexpected %d response from BitBucket Server REST API at %s", e.StatusCode, e.URL)
}

func (e *httpError) Unauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

func (e *httpError) NotFound() bool {
	return e.StatusCode == http.StatusNotFound
}
