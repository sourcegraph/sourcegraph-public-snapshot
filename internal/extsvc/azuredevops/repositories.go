package azuredevops

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (c *client) GetRepo(ctx context.Context, args OrgProjectRepoArgs) (Repository, error) {
	reqURL := url.URL{Path: fmt.Sprintf("%s/%s/_apis/git/repositories/%s", args.Org, args.Project, args.RepoNameOrID)}

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return Repository{}, err
	}

	var repo Repository
	if _, err = c.do(ctx, req, "", &repo); err != nil {
		return Repository{}, err
	}

	return repo, nil
}

func (c *client) ListRepositoriesByProjectOrOrg(ctx context.Context, args ListRepositoriesByProjectOrOrgArgs) ([]Repository, error) {
	reqURL := url.URL{Path: fmt.Sprintf("%s/_apis/git/repositories", args.ProjectOrOrgName)}

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, err
	}

	var repos ListRepositoriesResponse
	if _, err = c.do(ctx, req, "", &repos); err != nil {
		return nil, err
	}

	return repos.Value, nil
}

func (c *client) ForkRepository(ctx context.Context, org string, input ForkRepositoryInput) (Repository, error) {
	data, err := json.Marshal(&input)
	if err != nil {
		return Repository{}, errors.Wrap(err, "marshalling request")
	}

	reqURL := url.URL{Path: fmt.Sprintf("%s/_apis/git/repositories", org)}

	req, err := http.NewRequest("POST", reqURL.String(), bytes.NewBuffer(data))
	if err != nil {
		return Repository{}, err
	}

	var repo Repository
	if _, err = c.do(ctx, req, "", &repo); err != nil {
		return Repository{}, err
	}

	return repo, nil
}

func (c *client) GetRepositoryBranch(ctx context.Context, args OrgProjectRepoArgs, branchName string) (Ref, error) {
	var allRefs []Ref
	continuationToken := ""
	queryParams := make(url.Values)
	// The filter here by branch name is only a substring match, so we aren't guaranteed to only get one result.
	queryParams.Set("filter", fmt.Sprintf("heads/%s", branchName))
	reqURL := url.URL{Path: fmt.Sprintf("%s/%s/_apis/git/repositories/%s/refs", args.Org, args.Project, args.RepoNameOrID)}
	for {
		if continuationToken != "" {
			queryParams.Set("continuationToken", continuationToken)
		}
		reqURL.RawQuery = queryParams.Encode()
		req, err := http.NewRequest("GET", reqURL.String(), nil)
		if err != nil {
			return Ref{}, err
		}

		var refs ListRefsResponse
		continuationToken, err = c.do(ctx, req, "", &refs)
		if err != nil {
			return Ref{}, err
		}
		allRefs = append(allRefs, refs.Value...)

		if continuationToken == "" {
			break
		}
	}

	for _, ref := range allRefs {
		if ref.Name == fmt.Sprintf("refs/heads/%s", branchName) {
			return ref, nil
		}
	}

	return Ref{}, errors.Newf("branch %q not found", branchName)
}
