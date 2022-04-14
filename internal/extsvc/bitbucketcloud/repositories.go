package bitbucketcloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Repo returns a single repository, based on its namespace and slug.
func (c *Client) Repo(ctx context.Context, namespace, slug string) (*Repo, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("/2.0/repositories/%s/%s", namespace, slug), nil)
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}

	var repo Repo
	if err := c.do(ctx, req, &repo); err != nil {
		return nil, errors.Wrap(err, "sending request")
	}

	return &repo, nil
}

// Repos returns a list of repositories that are fetched and populated based on given account
// name and pagination criteria. If the account requested is a team, results will be filtered
// down to the ones that the app password's user has access to.
// If the argument pageToken.Next is not empty, it will be used directly as the URL to make
// the request. The PageToken it returns may also contain the URL to the next page for
// succeeding requests if any.
func (c *Client) Repos(ctx context.Context, pageToken *PageToken, accountName string) ([]*Repo, *PageToken, error) {
	var repos []*Repo
	var next *PageToken
	var err error
	if pageToken.HasMore() {
		next, err = c.reqPage(ctx, pageToken.Next, &repos)
	} else {
		next, err = c.page(ctx, fmt.Sprintf("/2.0/repositories/%s", accountName), nil, pageToken, &repos)
	}
	return repos, next, err
}

type ForkInputProject struct {
	Key string `json:"key"`
}

type ForkInputWorkspace string

func (fiw ForkInputWorkspace) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Slug string `json:"slug"`
	}{
		Slug: string(fiw),
	})
}

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

func (c *Client) ForkRepository(ctx context.Context, upstream *Repo, input ForkInput) (*Repo, error) {
	data, err := json.Marshal(&input)
	if err != nil {
		return nil, errors.Wrap(err, "marshalling request")
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("/2.0/repositories/%s/forks", upstream.FullName), bytes.NewBuffer(data))
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}

	var fork Repo
	if err := c.do(ctx, req, &fork); err != nil {
		return nil, errors.Wrap(err, "sending request")
	}

	return &fork, nil
}

// Repo represents the Repository type returned by Bitbucket Cloud.
//
// When used as an input into functions, only the FullName field is actually
// read.
type Repo struct {
	Slug        string     `json:"slug"`
	Name        string     `json:"name"`
	FullName    string     `json:"full_name"`
	UUID        string     `json:"uuid"`
	SCM         string     `json:"scm"`
	Description string     `json:"description"`
	Parent      *Repo      `json:"parent"`
	IsPrivate   bool       `json:"is_private"`
	Links       RepoLinks  `json:"links"`
	ForkPolicy  ForkPolicy `json:"fork_policy"`
}

func (r *Repo) Namespace() (string, error) {
	// Bitbucket Cloud will return cut down versions of the repository in some
	// cases (for example, embedded in pull requests), but we always have the
	// full name, so let's parse the namespace out of that.

	// TODO: replace with strings.Cut() once we upgrade to Go 1.18.
	parts := strings.SplitN(r.FullName, "/", 2)
	if len(parts) < 2 {
		return "", errors.New("cannot split namespace from repo name")
	}

	return parts[0], nil
}

type ForkPolicy string

const (
	ForkPolicyAllow    ForkPolicy = "allow_forks"
	ForkPolicyNoPublic ForkPolicy = "no_public_forks"
	ForkPolicyNone     ForkPolicy = "no_forks"
)

type RepoLinks struct {
	Clone CloneLinks `json:"clone"`
	HTML  Link       `json:"html"`
}

type CloneLinks []Link

// HTTPS returns clone link named "https", it returns an error if not found.
func (cl CloneLinks) HTTPS() (string, error) {
	for _, l := range cl {
		if l.Name == "https" {
			return l.Href, nil
		}
	}
	return "", errors.New("HTTPS clone link not found")
}
