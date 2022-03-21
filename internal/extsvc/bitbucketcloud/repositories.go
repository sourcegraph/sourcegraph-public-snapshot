package bitbucketcloud

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

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

func (c *Client) NewRepos(accountName string) (*ResultSet[Repo], error) {
	u, err := url.Parse(fmt.Sprintf("/2.0/repositories/%s", accountName))
	if err != nil {
		return nil, err
	}

	return newResultSet[Repo](c, u), nil
}

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
