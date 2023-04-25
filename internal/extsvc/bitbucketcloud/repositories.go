package bitbucketcloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Repo returns a single repository, based on its namespace and slug.
func (c *client) Repo(ctx context.Context, namespace, slug string) (*Repo, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("/2.0/repositories/%s/%s", namespace, slug), nil)
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}

	var repo Repo
	if _, err := c.do(ctx, req, &repo); err != nil {
		return nil, errors.Wrap(err, "sending request")
	}

	return &repo, nil
}

type ReposOptions struct {
	RequestOptions
	Role string `url:"role,omitempty"`
}

// Repos returns a list of repositories that are fetched and populated based on given account
// name and pagination criteria. If the account requested is a team, results will be filtered
// down to the ones that the app password's user has access to.
// If the argument pageToken.Next is not empty, it will be used directly as the URL to make
// the request. The PageToken it returns may also contain the URL to the next page for
// succeeding requests if any.
// If the argument accountName is empty, it will return all repositories for
// the authenticated user.
func (c *client) Repos(ctx context.Context, pageToken *PageToken, accountName string, opts *ReposOptions) (repos []*Repo, next *PageToken, err error) {
	if pageToken.HasMore() {
		next, err = c.reqPage(ctx, pageToken.Next, &repos)
		return
	}

	var reposURL string
	if accountName == "" {
		reposURL = "/2.0/repositories"
	} else {
		reposURL = fmt.Sprintf("/2.0/repositories/%s", url.PathEscape(accountName))
	}

	var urlValues url.Values
	if opts != nil && opts.Role != "" {
		urlValues = make(url.Values)
		urlValues.Set("role", opts.Role)
	}

	next, err = c.page(ctx, reposURL, urlValues, pageToken, &repos)

	if opts != nil && opts.FetchAll {
		repos, err = fetchAll(ctx, c, repos, next, err)
	}

	return repos, next, err
}

type ExplicitUserPermsResponse struct {
	User       *Account `json:"user"`
	Permission string   `json:"permission"`
}

func (c *client) ListExplicitUserPermsForRepo(ctx context.Context, pageToken *PageToken, namespace, slug string, opts *RequestOptions) (users []*Account, next *PageToken, err error) {
	var resp []ExplicitUserPermsResponse
	if pageToken.HasMore() {
		next, err = c.reqPage(ctx, pageToken.Next, &resp)
	} else {
		userPermsURL := fmt.Sprintf("/2.0/repositories/%s/%s/permissions-config/users", url.PathEscape(namespace), url.PathEscape(slug))
		next, err = c.page(ctx, userPermsURL, nil, pageToken, &resp)
	}

	if opts != nil && opts.FetchAll {
		resp, err = fetchAll(ctx, c, resp, next, err)
	}

	if err != nil {
		return
	}

	users = make([]*Account, len(resp))
	for i, r := range resp {
		users[i] = r.User
	}

	return
}

type ForkInputProject struct {
	Key string `json:"key"`
}

type ForkInputWorkspace string

// ForkInput defines the options used when forking a repository.
//
// All fields are optional except for the workspace, which must be defined.
type ForkInput struct {
	Name        *string            `json:"name,omitempty"`
	Workspace   ForkInputWorkspace `json:"workspace"`
	Description *string            `json:"description,omitempty"`
	ForkPolicy  *ForkPolicy        `json:"fork_policy,omitempty"`
	Language    *string            `json:"language,omitempty"`
	MainBranch  *string            `json:"mainbranch,omitempty"`
	IsPrivate   *bool              `json:"is_private,omitempty"`
	HasIssues   *bool              `json:"has_issues,omitempty"`
	HasWiki     *bool              `json:"has_wiki,omitempty"`
	Project     *ForkInputProject  `json:"project,omitempty"`
}

// ForkRepository forks the given upstream repository.
func (c *client) ForkRepository(ctx context.Context, upstream *Repo, input ForkInput) (*Repo, error) {
	data, err := json.Marshal(&input)
	if err != nil {
		return nil, errors.Wrap(err, "marshalling request")
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("/2.0/repositories/%s/forks", upstream.FullName), bytes.NewBuffer(data))
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}

	var fork Repo
	if _, err := c.do(ctx, req, &fork); err != nil {
		return nil, errors.Wrap(err, "sending request")
	}

	return &fork, nil
}

var _ json.Marshaler = ForkInputWorkspace("")

func (fiw ForkInputWorkspace) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Slug string `json:"slug"`
	}{
		Slug: string(fiw),
	})
}
